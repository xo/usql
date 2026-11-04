package env

import "testing"

// TestUnquote checks that a single-quoted argument decodes as psql 18 decodes
// the argument of a meta command. Each expected value was read from psql.
func TestUnquote(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s, exp string
	}{
		{`'a\tb'`, "a\tb"},
		{`'a''b'`, "a'b"},
		{`'\q'`, "q"},
		{`'\61'`, "1"},
		{`'\101'`, "A"},
		{`'\x41'`, "A"},
		{`'\x4'`, "\x04"},
		{`'\xff'`, "\xff"},
		{"'\\" + "u0041'", "u0041"},
		{`'\\'`, `\`},
		{`'\''`, "'"},
		{`''`, ""},
		{`"a b"`, "a b"},
	}
	for _, test := range tests {
		got, err := Unquote(test.s)
		if err != nil {
			t.Errorf("Unquote(%s): %v", test.s, err)
			continue
		}
		if got != test.exp {
			t.Errorf("Unquote(%s): expected %q, got %q", test.s, test.exp, got)
		}
	}
	for _, s := range []string{`'a`, `'a'b'`} {
		if _, err := Unquote(s); err == nil {
			t.Errorf("Unquote(%s): expected an error", s)
		}
	}
}
