package highlight

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/xo/transit/grammars/usql"
	"github.com/xo/transit/styles"
	"github.com/xo/usql/env"
)

// newTest returns a highlighter with no line editor, which is all that capture
// needs.
func newTest(t *testing.T, dialect string) *Highlighter {
	t.Helper()
	h := &Highlighter{
		defined: make(map[string]string),
	}
	h.parser, h.cursor = newParser()
	if err := h.SetDialect(dialect, usql.Language()); err != nil {
		t.Fatalf("SetDialect(%q): %v", dialect, err)
	}
	return h
}

// names returns the text of src that each capture name covers, joined by the
// name.
func names(t *testing.T, h *Highlighter, src string) map[string]string {
	t.Helper()
	got, err := h.capture(context.Background(), []byte(src))
	if err != nil {
		t.Fatalf("capture(%q): %v", src, err)
	}
	out := make(map[string]string)
	for i := 0; i < len(src); {
		j := i + 1
		for j < len(src) && got[j] == got[i] {
			j++
		}
		if got[i] != "" {
			out[got[i]] += src[i:j] + "|"
		}
		i = j
	}
	return out
}

func TestCapture(t *testing.T) {
	const sql = "select 1 as n, 'two' as s from t where a = :v; \\echo hi\n"
	tests := []struct {
		dialect string
		src     string
		keyword string
	}{
		{"", sql, "select"},
		{"postgres", sql, "select"},
		{"mysql", sql, "select"},
		{"sqlserver", sql, "select"},
		{"oracle", sql, "select"},
		{"cassandra", sql, "select"},
		{"sqlite3", sql, "select"},
		{"duckdb", sql, "select"},
		{"clickhouse", sql, "select"},
		{"trino", sql, "select"},
		{"surrealdb", sql, "select"},
		{"ydb", sql, "select"},
		// Neo4j speaks Cypher
		{"neo4j", "match (n) where n.a = :v return n; \\echo hi\n", "match"},
	}
	for _, test := range tests {
		t.Run("dialect "+test.dialect, func(t *testing.T) {
			h := newTest(t, test.dialect)
			got := names(t, h, test.src)
			var all strings.Builder
			for name, text := range got {
				all.WriteString(name + "=" + text + " ")
			}
			if len(got) == 0 {
				t.Fatalf("nothing is highlighted")
			}
			t.Log(all.String())
			var found bool
			for name, text := range got {
				if strings.HasPrefix(name, "keyword") && strings.Contains(strings.ToLower(text), test.keyword) {
					found = true
				}
			}
			if !found {
				t.Errorf("%s is not a keyword in %v", test.keyword, got)
			}
		})
	}
}

// TestEdit types a statement a key at a time, with the old tree, and checks
// that the result is the same as one parse of the whole text.
func TestEdit(t *testing.T) {
	const src = "select a, 'b' from t;\n\\set x 1\nselect :x;"
	typed := newTest(t, "postgres")
	var last []string
	for i := 1; i <= len(src); i++ {
		var err error
		if last, err = typed.capture(context.Background(), []byte(src[:i])); err != nil {
			t.Fatalf("key %d: %v", i, err)
		}
	}
	whole, err := newTest(t, "postgres").capture(context.Background(), []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(last, ",") != strings.Join(whole, ",") {
		t.Errorf("typing and one parse disagree\ntyped: %q\nwhole: %q", last, whole)
	}
	// a change in the middle and a deletion also parse as a whole does
	edited := "select a, 'bb' from t;\n\\set x 1\nselect :x;"
	got, err := typed.capture(context.Background(), []byte(edited))
	if err != nil {
		t.Fatal(err)
	}
	want, err := newTest(t, "postgres").capture(context.Background(), []byte(edited))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("an edit in the middle disagrees with one parse\ngot:  %q\nwant: %q", got, want)
	}
}

func TestSpec(t *testing.T) {
	s, ok := styles.Get("monokai")
	if !ok {
		t.Fatal("no style monokai")
	}
	tests := []struct {
		name    string
		capture string
		d       depth
		keepBG  bool
		want    string
	}{
		{"keyword rgb", "keyword", depthRGB, true, "color=#66d9ef"},
		{"keyword 256", "keyword", depth256, true, "ansi-color=81"},
		{"plain text keeps the terminal", "", depthRGB, true, ""},
		{"plain text takes the background", "", depthRGB, false, "bgcolor=#272822"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := spec(s, tt.capture, tt.d, tt.keepBG); got != tt.want {
				t.Errorf("spec(%q) = %q, want %q", tt.capture, got, tt.want)
			}
		})
	}
}

func TestDepthOf(t *testing.T) {
	for format, want := range map[string]depth{"terminal": depth16, "terminal256": depth256, "terminal16m": depthRGB} {
		if got, ok := depthOf(format); !ok || got != want {
			t.Errorf("depthOf(%q) = %v, %v, want %v, true", format, got, ok, want)
		}
	}
	if _, ok := depthOf("noop"); ok {
		t.Errorf("noop draws color")
	}
}

func TestSGR(t *testing.T) {
	s, ok := styles.Get("monokai")
	if !ok {
		t.Fatal("no style monokai")
	}
	tests := []struct {
		name    string
		capture string
		d       depth
		keepBG  bool
		want    string
	}{
		{"keyword rgb", "keyword", depthRGB, true, "\x1b[38;2;102;217;239m"},
		{"keyword 256", "keyword", depth256, true, "\x1b[38;5;81m"},
		{"plain text keeps the terminal", "", depthRGB, true, ""},
		{"plain text takes the background", "", depthRGB, false, "\x1b[48;2;39;40;34m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sgr(s, tt.capture, tt.d, tt.keepBG); got != tt.want {
				t.Errorf("sgr(%q) = %q, want %q", tt.capture, got, tt.want)
			}
		})
	}
}

func TestAppendSGR(t *testing.T) {
	ansi := func(n uint32) styles.Color { return styles.Color{Kind: styles.ColorANSI, Value: n} }
	tests := []struct {
		name string
		base int
		c    styles.Color
		want string
	}{
		{"red", 38, ansi(1), "31"},
		{"bright red", 38, ansi(9), "91"},
		{"red background", 48, ansi(1), "41"},
		{"bright red background", 48, ansi(9), "101"},
		{"one of 256", 38, styles.Color{Kind: styles.Color256, Value: 208}, "38;5;208"},
		{"not set", 38, styles.Color{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strings.Join(appendSGR(nil, tt.base, tt.c, depthRGB), ";"); got != tt.want {
				t.Errorf("appendSGR = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestText checks that Text colors a statement for the terminal, that no color
// runs past the end of a line, and that the text stays the same.
func TestText(t *testing.T) {
	const src = "select 1 as n, 'two' as s\nfrom t where a = :v;\n\\echo hi"
	for name, value := range map[string]string{
		"syntax_hl":        "true",
		"syntax_hl_format": "terminal16m",
		"syntax_hl_style":  "monokai",
	} {
		old := env.Setting(name)
		if err := env.Vars().SetSetting(name, value); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = env.Vars().SetSetting(name, old) })
	}
	h := newTest(t, "postgres")
	got := h.Text(src)
	if !strings.Contains(got, "\x1b[38;2;102;217;239mselect\x1b[0m") {
		t.Errorf("select is not drawn as a keyword: %q", got)
	}
	if plain := sgrRE.ReplaceAllString(got, ""); plain != src {
		t.Errorf("the text changed:\n got: %q\nwant: %q", plain, src)
	}
	for i, line := range strings.Split(got, "\n") {
		if on, off := len(sgrRE.FindAllString(line, -1)), strings.Count(line, "\x1b[0m"); on != 2*off {
			t.Errorf("line %d leaves a color on: %q", i+1, line)
		}
	}
	// the tree of the line editor stays as it was
	if h.tree != nil {
		t.Errorf("Text kept its tree")
	}
	if err := env.Vars().SetSetting("syntax_hl", "false"); err != nil {
		t.Fatal(err)
	}
	if got := h.Text(src); got != src {
		t.Errorf("highlighting is off, and the text changed: %q", got)
	}
}

// sgrRE matches an escape sequence that sets the style of the text.
var sgrRE = regexp.MustCompile("\x1b\\[[0-9;]*m")
