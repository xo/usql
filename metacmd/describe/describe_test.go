package describe

import (
	"slices"
	"testing"

	"github.com/xo/dbmeta"
)

// TestFunctionBody checks the line where the body of a function starts, which
// \sf+ numbers as line 1 and the LINE of \ef counts from, as psql does.
func TestFunctionBody(t *testing.T) {
	t.Parallel()
	tests := []struct {
		def string
		exp int
	}{
		{"CREATE OR REPLACE FUNCTION f()\n RETURNS integer\n LANGUAGE sql\nAS $function$ select 1 $function$", 3},
		{"CREATE OR REPLACE FUNCTION f()\n RETURNS integer\n LANGUAGE sql\nBEGIN ATOMIC\n SELECT 1;\nEND", 3},
		{"CREATE OR REPLACE FUNCTION f()\n RETURNS integer\n LANGUAGE sql\nRETURN 1", 3},
		{"create function f return number as\nbegin\n  return 1;\nend;", 0},
		{"", 0},
	}
	for _, test := range tests {
		if got := FunctionBody(test.def); got != test.exp {
			t.Errorf("FunctionBody(%q): expected %d, got %d", test.def, test.exp, got)
		}
	}
}

// TestTypePatterns checks that psql's type patterns reach dbmeta as LIKE
// patterns, folded as PostgreSQL folds a name, with - kept for no argument.
func TestTypePatterns(t *testing.T) {
	m, err := dbmeta.New(dbmeta.Dialect("postgres"), dbmeta.VersionSet{})
	if err != nil {
		t.Skipf("no PostgreSQL model: %v", err)
	}
	f := NewFormatter(m, nil, nil, "postgres")
	got := f.typePatterns([]string{"INT*", "-", "pg_catalog.int?", `"MyType"`})
	want := []string{"int%", "-", "pg_catalog.int_", "MyType"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := f.typePatterns(nil); got != nil {
		t.Errorf("no patterns gave %q, want nil", got)
	}
}
