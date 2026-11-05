// Package feed reads usql's input: SQL statements, meta commands such as \dt
// and \g, and variables such as :name. It parses the input with the usql
// grammar of github.com/xo/transit, and feeds each statement and meta command
// to the handler, which runs them.
//
// The package holds no SQL splitter of its own. The grammar decides where a
// statement ends, which text is a string or a comment, and what the arguments
// of a meta command are.
package feed

import (
	"context"
	"strings"

	"github.com/xo/dbmeta"
	"github.com/xo/transit"
	"github.com/xo/transit/grammars/usql"
)

// Language returns the usql grammar for a dialect with the lexical forms s.
// The grammar is one language, and its scanner reads the forms of the dialect
// as its options (transit's D108). With batches, a batch from BEGIN BATCH to
// APPLY BATCH stays one statement, as CQL has it. When s has BeginEndBlocks, the
// BEGIN ... END body of a stored program stays in its statement (transit's
// D112), and with AnonymousBlocks, so does an anonymous block of PL/SQL
// (transit's D128). Two dashes start a comment only with DashComments, which
// Cypher and AQL do not have (transit's D139). With Brackets, a name between
// [ and ] is a quoted identifier, as in SQL Server and SQLite (transit's D140).
// A line with only / is statement text, as in psql (transit's D138). A U+FEFF,
// a byte order mark, is white space outside strings, quoted names and comments
// (transit's D141).
func Language(s dbmeta.Syntax, batches bool) *transit.Language {
	return usql.LanguageFor(usql.Options{
		DollarQuotes:    s.DollarQuotes,
		BlockComments:   s.BlockComments,
		SlashComments:   s.SlashComments,
		HashComments:    s.HashComments,
		Backticks:       s.Backticks,
		DashComments:    s.DashComments,
		BeginEndBlocks:  s.BeginEndBlocks,
		AnonymousBlocks: s.AnonymousBlocks,
		Batches:         batches,
		Brackets:        s.Brackets,
	})
}

// Kind is the kind of an Item.
type Kind int

// Kinds of item.
const (
	// Statement is SQL, up to a ';', a meta command or the end of the input.
	Statement Kind = iota
	// Command is a meta command, such as \dt or \g.
	Command
)

// PartKind is the kind of a Part.
type PartKind int

// Kinds of part.
const (
	// Text is SQL text or a word.
	Text PartKind = iota
	// String is a string between single quotes.
	String
	// QuotedIdentifier is a name between double quotes, or backticks where
	// the dialect quotes names with them.
	QuotedIdentifier
	// DollarString is a string between dollar quotes, such as $$...$$.
	DollarString
	// Comment is a comment.
	Comment
	// Variable is a variable, such as :name or :'name'.
	Variable
	// Backtick is a shell command between backticks.
	Backtick
	// Pipe is a | and the shell command after it.
	Pipe
	// Shell is the shell command of \!.
	Shell
	// Options is the (key=value ...) list of \g and its forms.
	Options
)

// Part is one node of an item: a string, a variable, a word and so on.
type Part struct {
	Kind PartKind
	// Start and End are the byte offsets of the part in the item's Text.
	Start, End int
	// Name is the name of a variable.
	Name string
	// Quote is the quote of a variable: 0 for :name, ' for :'name', " for
	// :"name" and ? for :{?name}.
	Quote byte
	// Value is the shell command of a Backtick, a Pipe or a Shell part.
	Value string
	// Options are the keys and values of an Options part, in order.
	Options []Option
}

// Option is one key=value of an option list. Value is empty for a key alone.
type Option struct {
	Key   string
	Value string
}

// Arg is one argument of a meta command: the parts that touch, with no white
// space between them, such as foo:bar or 'a b'c.
type Arg struct {
	Parts []Part
}

// Item is one statement or one meta command.
type Item struct {
	Kind Kind
	// Text is the text of the item as written.
	Text string
	// Terminated reports whether a statement ends with a ';'.
	Terminated bool
	// Name is the name of a meta command, without its backslash, with its
	// modifiers, such as dtS+.
	Name string
	// Parts are the named parts of a statement, in order.
	Parts []Part
	// Args are the arguments of a meta command, in order.
	Args []Arg
	// Start and End are the byte offsets of the item in the parsed text.
	Start, End int
}

// Parser parses usql's input.
type Parser struct {
	p    *transit.Parser
	lang *transit.Language

	// tree is the last parse of src by first, which the next parse reuses
	// when the text has only grown, so that it reads only what was added.
	// point is the row and the column of the end of src.
	tree  *transit.Tree
	src   []byte
	point transit.Point
}

// NewParser returns a Parser for the language.
func NewParser(lang *transit.Language) (*Parser, error) {
	p := transit.NewParser()
	if err := p.SetLanguage(lang); err != nil {
		return nil, err
	}
	return &Parser{p: p, lang: lang}, nil
}

// SetLanguage changes the language, as when a connection changes the dialect.
func (p *Parser) SetLanguage(lang *transit.Language) error {
	if lang == p.lang {
		return nil
	}
	if err := p.p.SetLanguage(lang); err != nil {
		return err
	}
	p.lang = lang
	p.reset()
	return nil
}

// reset drops the tree that first keeps.
func (p *Parser) reset() {
	p.tree.Close()
	p.tree, p.src, p.point = nil, p.src[:0], transit.Point{}
}

// advance returns the point after text, which starts at pt.
func advance(pt transit.Point, text string) transit.Point {
	n := strings.Count(text, "\n")
	if n == 0 {
		pt.Column += len(text)
		return pt
	}
	return transit.Point{Row: pt.Row + n, Column: len(text) - strings.LastIndexByte(text, '\n') - 1}
}

// Parse returns the statements and meta commands of src, in order. Comments
// between them are left out.
func (p *Parser) Parse(src string) ([]Item, error) {
	b := []byte(src)
	tree, err := p.p.Parse(context.Background(), b, nil)
	if err != nil {
		return nil, err
	}
	defer tree.Close()
	root := tree.RootNode()
	var items []Item
	for i := range root.ChildCount() {
		n, _ := root.Child(i)
		switch n.Kind() {
		case "statement":
			items = append(items, statement(n, b))
		case "meta_command":
			items = append(items, command(n, b))
		}
	}
	return items, nil
}

// readyItems parses src, and returns its first items, as many as ready says
// can be fed. ready gets how many items src holds, and the last of them, with
// only its kind, its end and whether it is terminated. Source asks this for
// each line that can end an item, so no item is built until some can be fed,
// and when src only grew since the last call, it parses only what was added.
func (p *Parser) readyItems(src string, grown bool, ready func(n int, last Item) int) ([]Item, error) {
	var old *transit.Tree
	if grown && p.tree != nil && len(src) >= len(p.src) {
		add := src[len(p.src):]
		start, at := len(p.src), p.point
		p.src = append(p.src, add...)
		p.point = advance(at, add)
		p.tree.Edit(transit.InputEdit{
			StartByte: start, OldEndByte: start, NewEndByte: len(p.src),
			StartPoint: at, OldEndPoint: at, NewEndPoint: p.point,
		})
		old = p.tree
	} else {
		p.reset()
		p.src = append(p.src, src...)
		p.point = advance(transit.Point{}, src)
	}
	tree, err := p.p.Parse(context.Background(), p.src, old)
	if err != nil {
		p.reset()
		return nil, err
	}
	old.Close()
	p.tree = tree
	root := tree.RootNode()
	var nodes []transit.Node
	var last Item
	for i := range root.ChildCount() {
		c, _ := root.Child(i)
		switch c.Kind() {
		case "statement":
			last = Item{Kind: Statement, Start: c.StartByte(), End: c.EndByte(), Terminated: terminated(c)}
		case "meta_command":
			last = Item{Kind: Command, Start: c.StartByte(), End: c.EndByte()}
		default:
			continue
		}
		nodes = append(nodes, c)
	}
	items := make([]Item, ready(len(nodes), last))
	for i := range items {
		if nodes[i].Kind() == "statement" {
			items[i] = statement(nodes[i], p.src)
		} else {
			items[i] = command(nodes[i], p.src)
		}
	}
	return items, nil
}

// terminated reports whether a statement node ends with a ';'.
func terminated(n transit.Node) bool {
	count := n.ChildCount()
	if count == 0 {
		return false
	}
	c, _ := n.Child(count - 1)
	return c.Kind() == ";" && !c.IsNamed()
}

// statement returns the item of a statement node.
func statement(n transit.Node, src []byte) Item {
	start := n.StartByte()
	it := Item{
		Kind:  Statement,
		Text:  n.Text(src),
		Start: start,
		End:   n.EndByte(),
	}
	for i := range n.ChildCount() {
		c, _ := n.Child(i)
		if c.Kind() == ";" && !c.IsNamed() {
			it.Terminated = true
			continue
		}
		if p, ok := part(c, src, start); ok {
			it.Parts = append(it.Parts, p)
		}
	}
	return it
}

// command returns the item of a meta_command node.
func command(n transit.Node, src []byte) Item {
	start := n.StartByte()
	it := Item{
		Kind:  Command,
		Text:  n.Text(src),
		Start: start,
		End:   n.EndByte(),
	}
	var name strings.Builder
	last := -1
	for i := range n.ChildCount() {
		c, _ := n.Child(i)
		switch {
		case n.FieldNameForChild(i) == "name":
			name.WriteString(strings.TrimPrefix(c.Text(src), `\`))
			continue
		case c.Kind() == "modifier":
			name.WriteString(c.Text(src))
			continue
		}
		p, ok := part(c, src, start)
		if !ok {
			continue
		}
		// parts that touch make one argument
		if last != -1 && c.StartByte() == last && len(it.Args) != 0 {
			a := &it.Args[len(it.Args)-1]
			a.Parts = append(a.Parts, p)
		} else {
			it.Args = append(it.Args, Arg{Parts: []Part{p}})
		}
		last = c.EndByte()
	}
	it.Name = name.String()
	return it
}

// part returns the part of a child node, with its offsets made relative to
// start. It returns false for a node that is not a part.
func part(n transit.Node, src []byte, start int) (Part, bool) {
	p := Part{Start: n.StartByte() - start, End: n.EndByte() - start}
	switch n.Kind() {
	case "word":
		p.Kind = Text
	case "string":
		p.Kind = String
	case "quoted_identifier":
		p.Kind = QuotedIdentifier
	case "dollar_string":
		p.Kind = DollarString
	case "comment":
		p.Kind = Comment
	case "variable":
		p.Kind = Variable
		if v, ok := n.ChildByFieldName("name"); ok {
			p.Name = v.Text(src)
		}
		switch sigil, _ := n.Child(0); sigil.Text(src) {
		case `:'`:
			p.Quote = '\''
		case `:"`:
			p.Quote = '"'
		case `:{?`:
			p.Quote = '?'
		}
	case "backtick_command":
		p.Kind = Backtick
		p.Value = shellCommand(n, src)
	case "pipe":
		p.Kind = Pipe
		p.Value = shellCommand(n, src)
	case "shell_command":
		p.Kind = Shell
		p.Value = n.Text(src)
	case "option_list":
		p.Kind = Options
		for i := range n.NamedChildCount() {
			o, _ := n.NamedChild(i)
			var opt Option
			if k, ok := o.ChildByFieldName("key"); ok {
				opt.Key = k.Text(src)
			}
			if v, ok := o.ChildByFieldName("value"); ok {
				opt.Value = v.Text(src)
			}
			p.Options = append(p.Options, opt)
		}
	default:
		return Part{}, false
	}
	return p, true
}

// shellCommand returns the text of the shell_command child of n, or an empty
// string.
func shellCommand(n transit.Node, src []byte) string {
	for i := range n.NamedChildCount() {
		if c, _ := n.NamedChild(i); c.Kind() == "shell_command" {
			return c.Text(src)
		}
	}
	return ""
}

// Finished reports whether an entry typed in the line editor is ready to run:
// its last statement ends with a ';', or it ends with a meta command. An entry
// that holds no statement or meta command, such as an empty one or a lone
// comment, is finished too.
func Finished(items []Item) bool {
	if len(items) == 0 {
		return true
	}
	last := items[len(items)-1]
	return last.Kind == Command || last.Terminated
}
