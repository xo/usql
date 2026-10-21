package pgx

import (
	"testing"

	"github.com/xo/tblfmt"
)

func TestBytea(t *testing.T) {
	tests := []struct {
		name string
		src  any
		exp  string
		json string
	}{
		{"bytes that are not utf-8", []byte{0xde, 0xad, 0xbe, 0xef}, `\xdeadbeef`, `\\xdeadbeef`},
		{"bytes that are text", []byte("hello"), `\x68656c6c6f`, `\\x68656c6c6f`},
		{"a nul byte", []byte{0x00, 0xff}, `\x00ff`, `\\x00ff`},
		{"no bytes", []byte{}, `\x`, `\\x`},
		{"null", nil, "", "null"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := new(Bytea)
			if err := b.Scan(test.src); err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if s := format(t, b, false); s != test.exp {
				t.Errorf("expected %q, got: %q", test.exp, s)
			}
			if s := format(t, b, true); s != test.json {
				t.Errorf("expected json %q, got: %q", test.json, s)
			}
		})
	}
}

func TestByteaScanCopies(t *testing.T) {
	buf := []byte{0x01, 0x02}
	b := new(Bytea)
	if err := b.Scan(buf); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	buf[0] = 0xff
	if s := format(t, b, false); s != `\x0102` {
		t.Errorf("expected the scanned value to keep its bytes, got: %q", s)
	}
}

func TestByteaScanRejectsOtherTypes(t *testing.T) {
	if err := new(Bytea).Scan("text"); err == nil {
		t.Error("expected an error for a string")
	}
}

// format formats v with the formatter that usql uses for its tables, or for
// its json output when isJSON is true. json escapes the backslash, and the
// encoder adds the quotes later. A null value is "" in a table and "null" in
// json.
func format(t *testing.T, v any, isJSON bool) string {
	t.Helper()
	var opts []tblfmt.EscapeFormatterOption
	if isJSON {
		opts = append(opts, tblfmt.WithIsJSON(true))
	}
	vals, err := tblfmt.NewEscapeFormatter(opts...).Format([]any{v})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	switch {
	case vals[0] == nil && isJSON:
		return "null"
	case vals[0] == nil:
		return ""
	}
	return string(vals[0].Buf)
}
