package highlight

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/xo/transit"
	"github.com/xo/transit/styles"
)

// depth is the number of colors that the setting syntax_hl_format allows.
type depth int

const (
	depth16 depth = iota
	depth256
	depthRGB
)

// depthOf returns the color depth that a value of syntax_hl_format stands for.
// The names are the names of the formatters of Chroma, which usql used before.
// It returns false for noop, which draws no color.
func depthOf(format string) (depth, bool) {
	switch format {
	case "noop":
		return 0, false
	case "terminal16m":
		return depthRGB, true
	case "terminal256":
		return depth256, true
	}
	return depth16, true
}

// spec returns the style spec of the line editor for the capture name, or the
// empty string when the name needs no style.
func spec(s *styles.Style, name string, d depth, keepBG bool) string {
	e, ok := entry(s, name, keepBG)
	if !ok {
		return ""
	}
	var words []string
	if e.Bold {
		words = append(words, "bold")
	}
	if e.Italic {
		words = append(words, "italic")
	}
	if e.Underline {
		words = append(words, "underline")
	}
	words = appendColor(words, "color", e.Fg, d)
	words = appendColor(words, "bgcolor", e.Bg, d)
	return strings.Join(words, " ")
}

// sgr returns the escape sequence of a terminal that turns on the style of the
// capture name, or the empty string when the name needs no style.
func sgr(s *styles.Style, name string, d depth, keepBG bool) string {
	e, ok := entry(s, name, keepBG)
	if !ok {
		return ""
	}
	var params []string
	if e.Bold {
		params = append(params, "1")
	}
	if e.Italic {
		params = append(params, "3")
	}
	if e.Underline {
		params = append(params, "4")
	}
	params = appendSGR(params, 38, e.Fg, d)
	params = appendSGR(params, 48, e.Bg, d)
	if len(params) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(params, ";") + "m"
}

// entry returns the entry of the style for the capture name. A name that the
// style does not give has the entry of its longest prefix, and a text that no
// capture names has no color, except that a background that the style asks
// for is drawn. It returns false when the name needs no style.
func entry(s *styles.Style, name string, keepBG bool) (styles.Entry, bool) {
	var e styles.Entry
	if name != "" {
		e = s.Lookup(name)
	}
	if !keepBG && !e.Bg.IsSet() {
		e.Bg = s.Background()
	}
	if keepBG && name == "" {
		return e, false
	}
	return e, true
}

// appendSGR appends the parameters of an escape sequence that set a color,
// reduced to the depth. base is 38 for the foreground and 48 for the
// background.
func appendSGR(params []string, base int, c styles.Color, d depth) []string {
	if !c.IsSet() {
		return params
	}
	switch d {
	case depth16:
		c = c.To16()
	case depth256:
		c = c.To256()
	}
	switch c.Kind {
	case styles.ColorANSI:
		// 30 to 37 and 90 to 97 for the foreground, 40 to 47 and 100 to 107
		// for the background
		n := base - 8 + int(c.Value)
		if c.Value >= 8 {
			n += 60 - 8
		}
		return append(params, strconv.Itoa(n))
	case styles.Color256:
		return append(params, fmt.Sprintf("%d;5;%d", base, c.Value))
	}
	r, g, b := c.RGB()
	return append(params, fmt.Sprintf("%d;2;%d;%d;%d", base, r, g, b))
}

// appendColor appends the property of a color, reduced to the depth.
func appendColor(words []string, property string, c styles.Color, d depth) []string {
	if !c.IsSet() {
		return words
	}
	switch d {
	case depth16:
		c = c.To16()
	case depth256:
		c = c.To256()
	}
	switch c.Kind {
	case styles.ColorANSI, styles.Color256:
		return append(words, fmt.Sprintf("ansi-%s=%d", property, c.Value))
	}
	return append(words, fmt.Sprintf("%s=#%06x", property, c.Value))
}

// editOf describes the edit that turns old into next, as one change between the
// text that they share at the start and at the end. The line editor hands over
// the whole text and not the edit.
func editOf(old, next []byte) transit.InputEdit {
	prefix := 0
	for prefix < len(old) && prefix < len(next) && old[prefix] == next[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(old)-prefix && suffix < len(next)-prefix &&
		old[len(old)-1-suffix] == next[len(next)-1-suffix] {
		suffix++
	}
	oldEnd, newEnd := len(old)-suffix, len(next)-suffix
	return transit.InputEdit{
		StartByte:   prefix,
		OldEndByte:  oldEnd,
		NewEndByte:  newEnd,
		StartPoint:  pointAt(old, prefix),
		OldEndPoint: pointAt(old, oldEnd),
		NewEndPoint: pointAt(next, newEnd),
	}
}

// pointAt returns the row and the column, in bytes, of offset in src.
func pointAt(src []byte, offset int) transit.Point {
	row := bytes.Count(src[:offset], []byte{'\n'})
	return transit.Point{Row: row, Column: offset - (bytes.LastIndexByte(src[:offset], '\n') + 1)}
}

// placeholder replaces a variable of usql with a text of the same length that a
// SQL grammar accepts: a name for :name, a string for :'name', a quoted name
// for :"name", and TRUE for :{?name}.
func placeholder(_ string, n transit.Node, src []byte) ([]byte, bool) {
	if n.Kind() != "variable" {
		return nil, false
	}
	name, ok := n.ChildByFieldName("name")
	if !ok {
		return nil, false
	}
	text := src[n.StartByte():n.EndByte()]
	id := "_" + string(src[name.StartByte():name.EndByte()])
	switch {
	case bytes.HasPrefix(text, []byte(":{?")):
		return append([]byte("TRUE"), bytes.Repeat([]byte(" "), len(text)-4)...), true
	case bytes.HasPrefix(text, []byte(":'")):
		return []byte("'" + id + "'"), true
	case bytes.HasPrefix(text, []byte(`:"`)):
		return []byte(`"` + id + `"`), true
	}
	return []byte(id), true
}
