package feed

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/xo/dbmeta"
)

// postgres is the lexical forms of PostgreSQL.
var postgres = dbmeta.Syntax{DollarQuotes: true, BlockComments: true}

// parse parses src in the postgres family.
func parse(t *testing.T, src string) []Item {
	t.Helper()
	p, err := NewParser(Language(postgres))
	if err != nil {
		t.Fatal(err)
	}
	items, err := p.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	return items
}

// summary writes each item as S:text; or C:name[args], so a test can compare
// a whole parse in one string.
func summary(items []Item) string {
	var v []string
	for _, it := range items {
		switch it.Kind {
		case Statement:
			s := "S:" + strings.TrimSpace(it.Text)
			if !it.Terminated {
				s += "(open)"
			}
			v = append(v, s)
		case Command:
			var args []string
			for _, a := range it.Args {
				var parts []string
				for _, p := range a.Parts {
					parts = append(parts, it.Text[p.Start:p.End])
				}
				args = append(args, strings.Join(parts, "+"))
			}
			v = append(v, "C:"+it.Name+"["+strings.Join(args, " ")+"]")
		}
	}
	return strings.Join(v, " | ")
}

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src, exp string
	}{
		{``, ``},
		{`select 1;`, `S:select 1;`},
		{`select 1; select 2;`, `S:select 1; | S:select 2;`},
		{`select 1`, `S:select 1(open)`},
		{"select 1\n\\g", `S:select 1(open) | C:g[]`},
		{`select 1 \g out.txt`, `S:select 1(open) | C:g[out.txt]`},
		{`select 1\g`, `S:select 1(open) | C:g[]`},
		{`\dtS+ public.*`, `C:dtS+[public.*]`},
		{`\echo foo:v 'a b'x "q r"`, `C:echo[foo+:v 'a b'+x "q r"]`},
		{"\\set n 42\nselect :n;", `C:set[n 42] | S:select :n;`},
		{`select 'a;b', "c;d", $$e;f$$;`, `S:select 'a;b', "c;d", $$e;f$$;`},
		{"select 1 -- a; comment\n;", "S:select 1 -- a; comment\n;"},
		{`select (1; 2);`, `S:select (1; 2);`},
		{`select 'open`, `S:select 'open(open)`},
		{"select 1 \\g \\x", `S:select 1(open) | C:g[] | C:x[]`},
		{"\\! ls -l", `C:![ls -l]`},
		{"\\\\ select 1;", `C:\[] | S:select 1;`},
	}
	for _, test := range tests {
		if got := summary(parse(t, test.src)); got != test.exp {
			t.Errorf("%q:\n got: %s\nwant: %s", test.src, got, test.exp)
		}
	}
}

func TestFinished(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src string
		exp bool
	}{
		{``, true},
		{`-- just a comment`, true},
		{`select 1;`, true},
		{`select 1`, false},
		{`select 1 \g`, true},
		{`\dt`, true},
		{`select 'open;`, false},
		{`select (1;`, false},
		{`select /* open;`, false},
		{`select $$open;`, false},
		{"select 1;\nselect 2", false},
	}
	for _, test := range tests {
		if got := Finished(parse(t, test.src)); got != test.exp {
			t.Errorf("%q: expected %t, got %t", test.src, test.exp, got)
		}
	}
}

func TestExpand(t *testing.T) {
	t.Parallel()
	vars := map[string]string{
		"x":    "hello",
		"q":    "it's",
		"id":   `a"b`,
		"int":  "WRONG",
		"semi": "1; drop table t",
	}
	lookup := func(name string) (string, bool) {
		v, ok := vars[name]
		return v, ok
	}
	tests := []struct {
		src, exp string
	}{
		{`select :x;`, `select hello;`},
		{`select :'x';`, `select 'hello';`},
		{`select :'q';`, `select 'it''s';`},
		{`select :"id";`, `select "a""b";`},
		{`select :{?x}, :{?nope};`, `select TRUE, FALSE;`},
		{`select :nope;`, `select :nope;`},
		// a cast is not a variable, even when a variable has its name
		{`select y::int;`, `select y::int;`},
		// a value is not read again, so its ; ends nothing
		{`select :semi;`, `select 1; drop table t;`},
		// variables inside strings and names stay as they are
		{`select ':x', ":x";`, `select ':x', ":x";`},
		{`select 1\; select 2;`, `select 1; select 2;`},
	}
	for _, test := range tests {
		items := parse(t, test.src)
		if len(items) != 1 {
			t.Fatalf("%q: expected one item, got %d", test.src, len(items))
		}
		if got := items[0].Expand(lookup); got != test.exp {
			t.Errorf("%q: expected %q, got %q", test.src, test.exp, got)
		}
	}
}

func TestPrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src, exp string
	}{
		{`select * from t;`, `SELECT`},
		{`create table t (a int);`, `CREATE TABLE T`},
		{`/* x */ insert into t values (1);`, `INSERT INTO T VALUES`},
		{"-- c\nselect 1;", `SELECT`},
		{`select/* x */into t;`, `SELECT INTO T`},
		{`begin batch insert into t values (1);`, `BEGIN BATCH INSERT INTO T VALUES`},
		{`  `, ``},
	}
	for _, test := range tests {
		items := parse(t, test.src)
		var got string
		if len(items) != 0 {
			got = items[0].Prefix()
		}
		if got != test.exp {
			t.Errorf("%q: expected %q, got %q", test.src, test.exp, got)
		}
	}
}

// entries is a LineReader that returns whole entries, as the line editor does
// at a terminal.
type entries []string

func (e *entries) ReadLine(string) (string, error) {
	if len(*e) == 0 {
		return "", io.EOF
	}
	s := (*e)[0]
	*e = (*e)[1:]
	return s, nil
}

// feedAll returns the summary of every item that s feeds.
func feedAll(t *testing.T, s *Source) string {
	t.Helper()
	var items []Item
	for {
		it, err := s.Next()
		if errors.Is(err, io.EOF) {
			return summary(items)
		}
		if err != nil {
			t.Fatal(err)
		}
		items = append(items, it)
	}
}

func TestSourceLines(t *testing.T) {
	t.Parallel()
	src := "select\n  1\n;\n\\set n 42\nselect 'a\nb';\nselect 2 \\g\n-- end\nselect 3"
	s, err := NewSource(NewLines(strings.NewReader(src)), Language(postgres), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := feedAll(t, s)
	exp := "S:select\n  1\n; | C:set[n 42] | S:select 'a\nb'; | S:select 2(open) | C:g[] | S:select 3(open)"
	if got != exp {
		t.Errorf("\n got: %q\nwant: %q", got, exp)
	}
}

func TestSourceEntries(t *testing.T) {
	t.Parallel()
	e := &entries{"select 1; select 2;", "\\dt", "select 3 \\gx", "\\set x 1\n\\echo :x"}
	s, err := NewSource(e, Language(postgres), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := feedAll(t, s)
	exp := "S:select 1; | S:select 2; | C:dt[] | S:select 3(open) | C:gx[] | C:set[x 1] | C:echo[:x]"
	if got != exp {
		t.Errorf("\n got: %q\nwant: %q", got, exp)
	}
}

func TestSourcePush(t *testing.T) {
	t.Parallel()
	e := &entries{"select 2;"}
	s, err := NewSource(e, Language(postgres), nil)
	if err != nil {
		t.Fatal(err)
	}
	s.Push("select 1;")
	if got, exp := feedAll(t, s), "S:select 1; | S:select 2;"; got != exp {
		t.Errorf("\n got: %q\nwant: %q", got, exp)
	}
}

func TestLanguage(t *testing.T) {
	t.Parallel()
	names := func(s dbmeta.Syntax) string {
		return Language(s).Name()
	}
	tests := []struct {
		s   dbmeta.Syntax
		exp string
	}{
		{dbmeta.Syntax{DollarQuotes: true, BlockComments: true}, "usql_postgres"},
		{dbmeta.Syntax{DollarQuotes: true, BlockComments: true, SlashComments: true}, "usql_cql"},
		{dbmeta.Syntax{BlockComments: true, HashComments: true, Backticks: true}, "usql_mysql"},
		{dbmeta.Syntax{BlockComments: true, Backticks: true}, "usql_sqlite"},
		{dbmeta.Syntax{BlockComments: true}, "usql_standard"},
		{dbmeta.Syntax{}, "usql_plain"},
	}
	for _, test := range tests {
		if got := names(test.s); got != test.exp {
			t.Errorf("%+v: expected %s, got %s", test.s, test.exp, got)
		}
	}
}

func TestSourceReadLine(t *testing.T) {
	t.Parallel()
	e := &entries{"\\prompt 'Name: ' who\nBob\nselect :who;", "select 2;"}
	s, err := NewSource(e, Language(postgres), nil)
	if err != nil {
		t.Fatal(err)
	}
	it, err := s.Next()
	if err != nil || it.Name != "prompt" {
		t.Fatalf("expected \\prompt, got %q, %v", it.Name, err)
	}
	// \prompt takes the next line of the entry as its answer
	line, err := s.ReadLine("")
	if err != nil || line != "Bob" {
		t.Fatalf("expected Bob, got %q, %v", line, err)
	}
	if got, exp := feedAll(t, s), "S:select :who; | S:select 2;"; got != exp {
		t.Errorf("\n got: %q\nwant: %q", got, exp)
	}
}
