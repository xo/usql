package drivers

import (
	"testing"

	"github.com/xo/dburl"
)

// TestProcessOracle checks that Process removes the semicolon of a statement
// on Oracle, and keeps the one after the END of a PL/SQL block, with or
// without the name of the unit after it.
func TestProcessOracle(t *testing.T) {
	t.Parallel()
	u, err := dburl.Parse("oracle://user@localhost/db")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		src, exp string
	}{
		{"select 1 from dual;", "select 1 from dual"},
		{"select 1 from dual", "select 1 from dual"},
		{"begin\n  x := 1;\nend;", "begin\n  x := 1;\nend;"},
		{"create package p as\n  procedure q;\nend p;", "create package p as\n  procedure q;\nend p;"},
		{"create package body p as\n  procedure q as begin null; end;\nEND \"P\" ;", "create package body p as\n  procedure q as begin null; end;\nEND \"P\" ;"},
		{"select legend;", "select legend"},
		{"end;", "end;"},
	}
	for _, test := range tests {
		if _, got, _ := Process(u, "", test.src); got != test.exp {
			t.Errorf("Process(%q): expected %q, got %q", test.src, test.exp, got)
		}
	}
}
