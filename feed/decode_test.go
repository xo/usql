package feed

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// TestDecode checks that a byte order mark decides the encoding, and that the
// encoding setting decides it otherwise.
func TestDecode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		enc  string
		exp  string
	}{
		{"utf-8", "select 'é';", "", "select 'é';"},
		{"utf-8 mark", "\xef\xbb\xbfselect 1;", "", "select 1;"},
		{"utf-16le mark", "\xff\xfes\x00e\x00l\x00\xe9\x00", "", "selé"},
		{"utf-16be mark", "\xfe\xff\x00s\x00e\x00l\x00\xe9", "", "selé"},
		{"utf-32le mark", "\xff\xfe\x00\x00s\x00\x00\x00\xe9\x00\x00\x00", "", "sé"},
		{"utf-32be mark", "\x00\x00\xfe\xff\x00\x00\x00s\x00\x00\x00\xe9", "", "sé"},
		{"mark over setting", "\xef\xbb\xbf\xc3\xa9", "windows-1252", "é"},
		{"windows-1252", "select '\xe9';", "windows-1252", "select 'é';"},
		{"latin1", "\xe9", "latin1", "é"},
		{"shift_jis", "select '\x95\x5c';", "shift_jis", "select '表';"},
		{"utf-8 by name", "é", "UTF-8", "é"},
	}
	for _, test := range tests {
		r, err := Decode(strings.NewReader(test.in), test.enc)
		if err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}
		b, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}
		if got := string(b); got != test.exp {
			t.Errorf("%s: got %q, want %q", test.name, got, test.exp)
		}
	}
	if _, err := Decode(strings.NewReader("a"), "klingon"); err == nil {
		t.Error("expected an error for an unknown encoding")
	}
}

// TestEncoding checks that a name of an encoding becomes the name of the
// WHATWG Encoding Standard.
func TestEncoding(t *testing.T) {
	t.Parallel()
	for in, exp := range map[string]string{
		"":             "utf-8",
		"UTF8":         "utf-8",
		"latin1":       "windows-1252",
		"iso-8859-1":   "windows-1252",
		"sjis":         "shift_jis",
		"Windows-1251": "windows-1251",
		"gb2312":       "gbk",
	} {
		got, err := Encoding(in)
		if err != nil {
			t.Errorf("%q: %v", in, err)
			continue
		}
		if got != exp {
			t.Errorf("%q: got %q, want %q", in, got, exp)
		}
	}
	if _, err := Encoding("klingon"); err == nil {
		t.Error("expected an error for an unknown encoding")
	}
}

// TestSourceInvalidUTF8 checks that a line that is not valid UTF-8 stops the
// input with an error that names the line.
func TestSourceInvalidUTF8(t *testing.T) {
	t.Parallel()
	s, err := NewSource(NewLines(strings.NewReader("select 1;\nselect '\xe9';\nselect 3;\n")), Language(postgres, false), nil)
	if err != nil {
		t.Fatal(err)
	}
	if it, err := s.Next(); err != nil || it.Text != "select 1;" {
		t.Fatalf("expected select 1;, got %q, %v", it.Text, err)
	}
	_, err = s.Next()
	if !errors.Is(err, ErrInvalidUTF8) || !strings.HasPrefix(err.Error(), "line 2: ") {
		t.Errorf("expected an error for line 2, got %v", err)
	}
}
