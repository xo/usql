package metadata

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xo/tblfmt"
)

// TestBoolFormatsAsItself checks that a Bool is written as its plain value.
//
// Bool is a named string type, and a Go type switch does not match a named
// type to the type it is defined from. tblfmt's formatter therefore does not
// reach its string case, and anything it does not recognise is printed as
// JSON. That made \d report a nullable column as "YES", with the quotes, on
// every driver that uses the metadata writer.
func TestBoolFormatsAsItself(t *testing.T) {
	for _, b := range []Bool{YES, NO, UNKNOWN} {
		got, err := format(t, b)
		if err != nil {
			t.Fatalf("formatting %q: %v", string(b), err)
		}
		if want := string(b); got != want {
			t.Errorf("a Bool of %q was written as %q, want %q", string(b), got, want)
		}
	}
}

// format writes v through tblfmt the way the metadata writer does, and returns
// the single cell it produced.
func format(t *testing.T, v interface{}) (string, error) {
	t.Helper()
	rs := NewColumnSet([]Column{{}})
	rs.SetColumns([]string{"c"})
	rs.SetScanValues(func(Result) []interface{} { return []interface{}{v} })
	buf := new(bytes.Buffer)
	if err := tblfmt.EncodeAll(buf, rs, map[string]string{
		"format":      "unaligned",
		"tuples_only": "on",
	}); err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\r\n"), nil
}
