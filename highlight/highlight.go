// Package highlight draws the input of usql in color as it is typed.
//
// The line editor hands the text to a [Highlighter]. The highlighter parses the
// text with the usql grammar of transit, parses each statement with the SQL
// grammar of the dialect, and runs the highlight query of each grammar. The
// capture names that the queries give find their colors in a style of transit's
// styles module.
//
// Four settings of usql control it: syntax_hl turns it on and off,
// syntax_hl_style names the style, syntax_hl_format names the color depth, and
// syntax_hl_override_bg says whether the style keeps the background of the
// terminal.
package highlight

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/xo/rline"
	"github.com/xo/transit"
	"github.com/xo/transit/grammars/usql"
	"github.com/xo/transit/inject"
	"github.com/xo/transit/styles"
	"github.com/xo/usql/env"
)

// Highlighter marks up the text that the line editor holds. It keeps the parse
// of the last text, so that the next key parses with the old tree.
type Highlighter struct {
	prompt *rline.Prompt

	mu sync.Mutex
	// the dialect that the queries and the configurations are for
	dialect   string
	usql      *inject.Config
	sql       *inject.Config
	usqlQ     *transit.Query
	sqlQ      *transit.Query
	parser    *transit.Parser
	cursor    *transit.QueryCursor
	src       []byte
	tree      *transit.Tree
	style     *styles.Style
	styleName string
	defined   map[string]string
}

// New returns a highlighter that defines its styles on the prompt. Its
// language is the one of PostgreSQL until SetDialect changes it.
func New(prompt *rline.Prompt) (*Highlighter, error) {
	h := &Highlighter{
		prompt:  prompt,
		defined: make(map[string]string),
	}
	h.parser, h.cursor = newParser()
	if err := h.SetDialect("", usql.Language()); err != nil {
		return nil, err
	}
	return h, nil
}

// newParser returns the parser and the query cursor that a highlighter keeps.
func newParser() (*transit.Parser, *transit.QueryCursor) {
	return transit.NewParser(), transit.NewQueryCursor()
}

// SetDialect changes the dialect of the highlighter. lang is the usql grammar
// for the dialect, with the options that the input reads with, so that the
// highlighter splits statements as the input does.
func (h *Highlighter) SetDialect(dialect string, lang *transit.Language) error {
	g := grammarFor(dialect)
	u, err := inject.NewConfig(lang, "usql", query(usql.Queries, "injections.scm"))
	if err != nil {
		return fmt.Errorf("compiling the injection query of usql: %w", err)
	}
	s, err := inject.NewConfig(g.language(), dialect, "")
	if err != nil {
		return fmt.Errorf("making the configuration of %s: %w", dialect, err)
	}
	uq, err := transit.NewQuery(lang, query(usql.Queries, "highlights.scm"))
	if err != nil {
		return fmt.Errorf("compiling the highlight query of usql: %w", err)
	}
	// a grammar with no highlight query keeps the colors of the usql layer
	var sq *transit.Query
	if text := query(g.queries, "highlights.scm"); text != "" {
		if sq, err = transit.NewQuery(g.language(), text); err != nil {
			return fmt.Errorf("compiling the highlight query of %s: %w", dialect, err)
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dialect, h.usql, h.sql, h.usqlQ, h.sqlQ = dialect, u, s, uq, sq
	// the old tree is of another language
	h.src, h.tree = nil, nil
	return nil
}

// Highlight marks up the text of the line editor. It satisfies
// [rline.Highlighter]. A setting that turns highlighting off, or a text that
// does not parse, leaves the text plain.
func (h *Highlighter) Highlight(l *rline.LineStyle) {
	h.mu.Lock()
	defer h.mu.Unlock()
	style, depth, keepBG, ok := h.settings()
	if !ok {
		return
	}
	src := []byte(l.Text())
	names, err := h.capture(context.Background(), src)
	if err != nil {
		return
	}
	for i := 0; i < len(src); {
		j := next(names, i)
		if spec := spec(style, names[i], depth, keepBG); spec != "" {
			l.Style(i, j-i, h.define(spec))
		}
		i = j
	}
}

// Text returns src in the colors of the style, as the escape sequences of a
// terminal, for text that usql prints, such as the query buffer of \p and the
// definition of \sv. No color runs past the end of a line, so a caller can
// split the result into lines. A setting that turns highlighting off, or a
// text that does not parse, returns src as it is.
func (h *Highlighter) Text(src string) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	style, depth, keepBG, ok := h.settings()
	if !ok || src == "" {
		return src
	}
	ctx, b := context.Background(), []byte(src)
	// a tree of its own, so that the tree of the line editor stays
	tree, err := h.parse(ctx, b, nil)
	if err != nil {
		return src
	}
	defer tree.Close()
	names, err := h.names(ctx, tree, b)
	if err != nil {
		return src
	}
	var sb strings.Builder
	for i := 0; i < len(src); {
		j := next(names, i)
		on := sgr(style, names[i], depth, keepBG)
		for k, line := range strings.Split(src[i:j], "\n") {
			if k != 0 {
				sb.WriteByte('\n')
			}
			if on != "" && line != "" {
				sb.WriteString(on + line + "\x1b[0m")
			} else {
				sb.WriteString(line)
			}
		}
		i = j
	}
	return sb.String()
}

// settings returns the style, the color depth and whether the terminal keeps
// its background, from the settings of usql. It returns false when the
// settings turn highlighting off. h.mu must be held.
func (h *Highlighter) settings() (*styles.Style, depth, bool, bool) {
	if env.Setting("syntax_hl") != "true" {
		return nil, 0, false, false
	}
	d, ok := depthOf(env.Setting("syntax_hl_format"))
	if !ok {
		return nil, 0, false, false
	}
	style := h.styleFor(env.Setting("syntax_hl_style"))
	if style == nil {
		return nil, 0, false, false
	}
	return style, d, env.Setting("syntax_hl_override_bg") != "false", true
}

// next returns the end of the run of bytes from i that have the capture name
// of byte i.
func next(names []string, i int) int {
	j := i + 1
	for j < len(names) && names[j] == names[i] {
		j++
	}
	return j
}

// styleFor returns the style of a name, and keeps it, because a style is read
// from a file each time it is asked for.
func (h *Highlighter) styleFor(name string) *styles.Style {
	if h.style != nil && h.styleName == name {
		return h.style
	}
	s, ok := styles.Get(name)
	if !ok {
		return nil
	}
	h.style, h.styleName = s, name
	return s
}

// define gives the line editor a name for a style spec, once, and returns the
// name.
func (h *Highlighter) define(spec string) string {
	if name, ok := h.defined[spec]; ok {
		return name
	}
	name := fmt.Sprintf("sqlhl%d", len(h.defined))
	h.prompt.DefineStyle(name, spec)
	h.defined[spec] = name
	return name
}

// capture parses src with the tree of the last text, and returns the capture
// name of each byte of src. A byte that no capture names has the empty name.
// It keeps the tree for the next text.
func (h *Highlighter) capture(ctx context.Context, src []byte) ([]string, error) {
	if h.tree != nil {
		h.tree.Edit(editOf(h.src, src))
	}
	tree, err := h.parse(ctx, src, h.tree)
	if err != nil {
		return nil, err
	}
	h.src, h.tree = append(h.src[:0], src...), tree
	return h.names(ctx, tree, src)
}

// parse parses src with the usql grammar, from the old tree when it is not
// nil.
func (h *Highlighter) parse(ctx context.Context, src []byte, old *transit.Tree) (*transit.Tree, error) {
	if err := h.parser.SetLanguage(h.usql.Language()); err != nil {
		return nil, fmt.Errorf("setting the language of usql: %w", err)
	}
	tree, err := h.parser.Parse(ctx, src, old)
	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}
	return tree, nil
}

// names returns the capture name of each byte of src, whose parse is tree.
func (h *Highlighter) names(ctx context.Context, tree *transit.Tree, src []byte) ([]string, error) {
	layers, err := h.usql.Layers(ctx, h.parser, src, func(name string) (*inject.Config, bool) {
		return h.sql, name == "sql"
	}, inject.WithReplacer(placeholder))
	if err != nil {
		return nil, fmt.Errorf("finding the layers: %w", err)
	}
	names := make([]string, len(src))
	h.mark(ctx, h.usqlQ, tree.RootNode(), src, names)
	// the root layer comes first, so the layers of the statements paint over it
	if h.sqlQ != nil {
		for _, l := range layers[1:] {
			if l.Name == "sql" {
				h.mark(ctx, h.sqlQ, l.Tree.RootNode(), l.Text, names)
			}
		}
	}
	return names, nil
}

// mark runs the highlight query q on the tree of root, and writes the capture
// name of each byte that a capture covers to names. The innermost capture
// wins, and of two patterns that capture one node the last wins. A name that
// starts with an underscore belongs to a predicate and is not a highlight.
func (h *Highlighter) mark(ctx context.Context, q *transit.Query, root transit.Node, src []byte, names []string) {
	captureNames := q.CaptureNames()
	h.cursor.SetByteRange(root.StartByte(), root.EndByte())
	for m, i := range h.cursor.Captures(ctx, q, root, src) {
		c := m.Captures[i]
		name := captureNames[c.Index]
		if strings.HasPrefix(name, "_") {
			continue
		}
		for b := c.Node.StartByte(); b < c.Node.EndByte() && b < len(names); b++ {
			names[b] = name
		}
	}
}
