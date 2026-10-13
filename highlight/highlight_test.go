package highlight

import (
	"context"
	"strings"
	"testing"

	"github.com/xo/transit/grammars/usql"
	"github.com/xo/transit/styles"
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
	for _, dialect := range []string{"", "postgres", "mysql", "sqlserver", "oracle", "cassandra", "sqlite3"} {
		t.Run("dialect "+dialect, func(t *testing.T) {
			h := newTest(t, dialect)
			got := names(t, h, "select 1 as n, 'two' as s from t where a = :v; \\echo hi\n")
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
				if strings.HasPrefix(name, "keyword") && strings.Contains(strings.ToLower(text), "select") {
					found = true
				}
			}
			if !found {
				t.Errorf("select is not a keyword in %v", got)
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
