package drivers

import (
	"strings"
	"testing"

	"github.com/xo/dburl"
	"github.com/xo/usql/feed"
)

// TestLanguageBeginEndBlocks checks that the BEGIN ... END body of a stored
// program stays in one statement for each dialect where dbmeta's Syntax has
// BeginEndBlocks, and that the input splits at each ; for the others. The
// statement that follows the program must still be a statement of its own.
func TestLanguageBeginEndBlocks(t *testing.T) {
	t.Parallel()
	const (
		trigger   = "create trigger tr after insert on t begin\n  insert into u values (1);\n  insert into u values (2);\nend;\nselect 3;\n"
		procedure = "create procedure p() begin\n  select 1;\n  select 2;\nend;\nselect 3;\n"
		asBegin   = "create procedure p as begin\n  select 1;\n  select 2;\nend;\nselect 3;\n"
		atomic    = "create function f() returns int language sql begin atomic\n  select 1;\n  select 2;\nend;\nselect 3;\n"
		begin     = "begin;\nselect 1;\nend;\n"
		// an anonymous block of PL/SQL, which only Oracle has
		block   = "begin\n  null;\n  null;\nend;\nselect 3 from dual;\n"
		declare = "declare\n  x number;\nbegin\n  x := 1;\nend;\nselect 3 from dual;\n"
		execute = "execute block as\ndeclare x int;\nbegin\n  x = 1;\n  x = 2;\nend;\nselect 3 from rdb$database;\n"
		do      = "do begin\n  declare x int;\n  x = 1;\nend;\nselect 3 from dummy;\n"
		with    = "with function f(x int) returns int begin declare y int; set y = x; return y; end select f(1);\nselect 2;\n"
	)
	tests := []struct {
		dialect string
		src     string
		want    int
	}{
		{"sqlite3", trigger, 2},
		{"libsql", trigger, 2},
		{"rqlite", trigger, 2},
		{"mysql", procedure, 2},
		{"memsql", asBegin, 2},
		{"vitess", procedure, 2},
		{"sqlserver", asBegin, 2},
		{"oracle", asBegin, 2},
		{"postgres", atomic, 2},
		{"cockroachdb", atomic, 2},
		{"oracle", block, 2},
		{"oracle", declare, 2},
		{"firebirdsql", execute, 2},
		{"hdb", do, 2},
		{"trino", with, 2},
		// a transaction splits at each ;
		{"postgres", begin, 3},
		{"sqlite3", begin, 3},
		{"sqlserver", "declare @x int;\nselect @x;\n", 2},
		// a dialect with no such bodies splits at each ;
		{"tidb", procedure, 4},
		{"duckdb", procedure, 4},
		{"clickhouse", procedure, 4},
	}
	for _, test := range tests {
		t.Run(test.dialect, func(t *testing.T) {
			p, err := feed.NewParser(Language(&dburl.URL{Dialect: test.dialect}))
			if err != nil {
				t.Fatal(err)
			}
			items, err := p.Parse(test.src)
			if err != nil {
				t.Fatal(err)
			}
			var texts []string
			for _, it := range items {
				texts = append(texts, it.Text)
			}
			if len(items) != test.want {
				t.Fatalf("expected %d statements, got %d: %q", test.want, len(items), texts)
			}
			// a program that stays whole is followed by the select alone
			if test.want == 2 && !strings.HasPrefix(texts[1], "select") {
				t.Errorf("expected the second statement to be the select, got: %q", texts)
			}
		})
	}
}

// TestLanguageClickHouse checks that a ; inside a backticked name, a comment or
// a heredoc string of ClickHouse does not end the statement. dbmeta measured
// each form on a live server (its D250 and D251).
func TestLanguageClickHouse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  string
	}{
		{"backticks", "select 1 as `a;b`;\nselect 2;\n"},
		{"block comment", "select 1 /* a; b */;\nselect 2;\n"},
		{"slash comment", "select 1 // a; b\n;\nselect 2;\n"},
		{"hash comment", "select 1 # a; b\n;\nselect 2;\n"},
		{"hash bang comment", "select 1 #! a; b\n;\nselect 2;\n"},
		{"dollar quotes", "select $$a;b$$;\nselect 2;\n"},
		{"tagged dollar quotes", "select $x$a;b$x$;\nselect 2;\n"},
	}
	p, err := feed.NewParser(Language(&dburl.URL{Dialect: "clickhouse"}))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		items, err := p.Parse(test.src)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 2 {
			var texts []string
			for _, it := range items {
				texts = append(texts, it.Text)
			}
			t.Errorf("%s: expected 2 statements, got %d: %q", test.name, len(items), texts)
		}
	}
}
