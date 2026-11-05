package feed

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/xo/dbmeta"
)

// postgres is the lexical forms of PostgreSQL.
var postgres = dbmeta.Syntax{DollarQuotes: true, BlockComments: true, DashComments: true}

// parse parses src in the postgres family.
func parse(t *testing.T, src string) []Item {
	t.Helper()
	p, err := NewParser(Language(postgres, false))
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
	s, err := NewSource(NewLines(strings.NewReader(src)), Language(postgres, false), nil)
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
	s, err := NewSource(e, Language(postgres, false), nil)
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
	s, err := NewSource(e, Language(postgres, false), nil)
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
	mysql := dbmeta.Syntax{BlockComments: true, DashComments: true, HashComments: true, Backticks: true}
	if Language(postgres, false) != Language(postgres, false) {
		t.Error("expected the same language for the same forms")
	}
	if Language(postgres, false) == Language(mysql, false) {
		t.Error("expected a different language for different forms")
	}
	// the forms decide what is a comment: # is one for MySQL and not for
	// PostgreSQL
	p, err := NewParser(Language(mysql, false))
	if err != nil {
		t.Fatal(err)
	}
	items, err := p.Parse("select 1 # a; comment\n;")
	if err != nil {
		t.Fatal(err)
	}
	if got, exp := summary(items), "S:select 1 # a; comment\n;"; got != exp {
		t.Errorf("mysql:\n got: %s\nwant: %s", got, exp)
	}
	if got := summary(parse(t, "select 1 # a; b;")); !strings.HasPrefix(got, "S:select 1 # a; |") {
		t.Errorf("postgres: expected # not to be a comment, got %s", got)
	}
}

// TestBacktickNameWithQuote checks that a quote inside a name between
// backticks, as MySQL and SQLite write names, opens no string, so that the
// string after it keeps both of its backslashes (GitHub xo/usql#587, from
// PR xo/usql#593).
func TestBacktickNameWithQuote(t *testing.T) {
	t.Parallel()
	const q = "select 1 as `we'ird`, hex('a\\\\b');"
	mysql := dbmeta.Syntax{BlockComments: true, DashComments: true, HashComments: true, Backticks: true}
	p, err := NewParser(Language(mysql, false))
	if err != nil {
		t.Fatal(err)
	}
	items, err := p.Parse(q)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one statement, got: %s", summary(items))
	}
	if got := items[0].Expand(func(string) (string, bool) { return "", false }); !strings.Contains(got, `a\\b`) {
		t.Errorf("expected a\\\\b in %q", got)
	}
	var name bool
	for _, part := range items[0].Parts {
		if part.Kind == QuotedIdentifier {
			name = true
			if text := items[0].Text[part.Start:part.End]; text != "`we'ird`" {
				t.Errorf("expected the name `we'ird`, got %q", text)
			}
		}
	}
	if !name {
		t.Errorf("expected a quoted name, got parts %+v", items[0].Parts)
	}
}

// TestBrackets checks that a name between brackets, as SQL Server and SQLite
// write names, holds a quote, a ';' or the start of a comment as text, and
// that the statement after it is a statement of its own (GitHub discussion
// xo/usql#600).
func TestBrackets(t *testing.T) {
	t.Parallel()
	sqlserver := dbmeta.Syntax{BlockComments: true, DashComments: true, Brackets: true, BeginEndBlocks: true}
	p, err := NewParser(Language(sqlserver, false))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		src string
		exp string
	}{
		{"create table t1 ([a'b] int);\nselect 2;", "S:create table t1 ([a'b] int); | S:select 2;"},
		{"select [a;b], [x]]y] from t;\nselect 2;", "S:select [a;b], [x]]y] from t; | S:select 2;"},
		{"select [--a], [/*b], [\\] from t;\nselect 2;", "S:select [--a], [/*b], [\\] from t; | S:select 2;"},
		{"select [a\nb;\nselect 2;", "S:select [a\nb;\nselect 2;(open)"},
	}
	for _, test := range tests {
		items, err := p.Parse(test.src)
		if err != nil {
			t.Fatal(err)
		}
		if got := summary(items); got != test.exp {
			t.Errorf("%q:\n got: %s\nwant: %s", test.src, got, test.exp)
		}
	}
	// PostgreSQL has no names between brackets
	if got := summary(parse(t, "select [a'b];")); !strings.HasSuffix(got, "(open)") {
		t.Errorf("postgres: expected the quote to open a string, got %s", got)
	}
}

// TestByteOrderMark checks that a U+FEFF between items, as a script saved by
// an editor on Windows starts with, is white space, and that it stays text
// inside a string (transit's D141).
func TestByteOrderMark(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src string
		exp string
	}{
		{"\uFEFFselect 1;\n\uFEFFselect 2;\uFEFF", "S:select 1; | S:select 2;"},
		{"select\uFEFF1;", "S:select\uFEFF1;"},
		{"select '\uFEFF';", "S:select '\uFEFF';"},
		{"\\echo a\uFEFF\n\uFEFF\\echo b", "C:echo[a] | C:echo[b]"},
	}
	for _, test := range tests {
		if got := summary(parse(t, test.src)); got != test.exp {
			t.Errorf("%q:\n got: %q\nwant: %q", test.src, got, test.exp)
		}
	}
}

func TestSourceReadLine(t *testing.T) {
	t.Parallel()
	e := &entries{"\\prompt 'Name: ' who\nBob\nselect :who;", "select 2;"}
	s, err := NewSource(e, Language(postgres, false), nil)
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

// TestBatches checks that a batch of CQL stays one statement, which is open
// until APPLY BATCH ends it, and that it is only a statement of its own where
// the dialect has batches.
func TestBatches(t *testing.T) {
	t.Parallel()
	const src = "begin batch\ninsert into t (a) values ('x;y');\ninsert into t (a) values (2);\napply batch;\nselect 1;\n"
	for _, tt := range []struct {
		name    string
		batches bool
		want    string
	}{
		{"with", true, "S:begin batch\ninsert into t (a) values ('x;y');\ninsert into t (a) values (2);\napply batch; | S:select 1;"},
		{"without", false, "S:begin batch\ninsert into t (a) values ('x;y'); | S:insert into t (a) values (2); | S:apply batch; | S:select 1;"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser(Language(postgres, tt.batches))
			if err != nil {
				t.Fatal(err)
			}
			items, err := p.Parse(src)
			if err != nil {
				t.Fatal(err)
			}
			if got := summary(items); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
	// an entry typed in the editor is not finished while the batch is open
	p, err := NewParser(Language(postgres, true))
	if err != nil {
		t.Fatal(err)
	}
	items, err := p.Parse("begin batch\ninsert into t (a) values (1);\n")
	if err != nil {
		t.Fatal(err)
	}
	if Finished(items) {
		t.Errorf("an open batch is finished: %q", summary(items))
	}
}

// TestSourceMatchesParse checks that a Source, which reads a line at a time,
// parses only some of them, and parses again only what a line added, feeds
// the items that one parse of the whole text gives.
func TestSourceMatchesParse(t *testing.T) {
	t.Parallel()
	mysql := dbmeta.Syntax{BlockComments: true, DashComments: true, HashComments: true, Backticks: true, BeginEndBlocks: true}
	oracle := dbmeta.Syntax{BlockComments: true, DashComments: true, BeginEndBlocks: true, AnonymousBlocks: true}
	tests := []struct {
		name string
		s    dbmeta.Syntax
		src  string
	}{
		{"statements", postgres, "select\n  1\n;\nselect 'a\nb;';\nselect 2 \\g\nselect 3 -- c;\nfrom t;\n"},
		{"commands", postgres, "\\set n 42\nselect :n\n  ;\n\\echo a;b\nselect 4"},
		{"dollar quotes", postgres, "create function f() returns int as $$\nbegin\n  return 1;\nend\n$$ language plpgsql;\nselect f();\n"},
		{"a body", mysql, "create procedure p() begin\n  select 1;\n  select 2;\nend;\nselect 3;\n# a; comment\nselect 4;\n"},
		{"dates", postgres, "insert into t values\n('7/4/1996',\n '1/2/3');\nselect 1;\n"},
		{"no end", postgres, "select 1;\nselect 2\n  from t"},
		{"oracle slash", oracle, "declare\n  x number;\nbegin\n  x := 1;\nend;\n/\nselect 1 from dual\n/\nselect (1\n/\n2) from dual;\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, err := NewParser(Language(test.s, false))
			if err != nil {
				t.Fatal(err)
			}
			items, err := p.Parse(test.src)
			if err != nil {
				t.Fatal(err)
			}
			s, err := NewSource(NewLines(strings.NewReader(test.src)), Language(test.s, false), nil)
			if err != nil {
				t.Fatal(err)
			}
			if got, exp := feedAll(t, s), summary(items); got != exp {
				t.Errorf("\n got: %s\nwant: %s", got, exp)
			}
		})
	}
}

// TestOracleEscapedUnits checks that a PL/SQL unit whose inner semicolons are
// written as \; is one statement, and that the server gets each as ;. The
// usql grammar keeps a block that starts at BEGIN or DECLARE whole, but not
// the declarations of a unit before its BEGIN, nor a package, which has none.
// The writer escapes those, as in psql. A variable in a body, such as the :new
// of a trigger, does not end it (transit v0.7.2).
func TestOracleEscapedUnits(t *testing.T) {
	t.Parallel()
	oracle := dbmeta.Syntax{BlockComments: true, DashComments: true, BeginEndBlocks: true, AnonymousBlocks: true}
	tests := []struct {
		name, src, exp string
	}{
		{
			"declarations",
			"create function f return number as\n  n number\\;\nbegin\n  n := 1\\;\n  return n\\;\nend;\n",
			"create function f return number as\n  n number;\nbegin\n  n := 1;\n  return n;\nend;",
		},
		{
			"declarations, body unescaped",
			"create function f return number as\n  n number\\;\nbegin\n  n := 1;\n  return n;\nend;\n",
			"create function f return number as\n  n number;\nbegin\n  n := 1;\n  return n;\nend;",
		},
		{
			"trigger",
			"create trigger t before insert on x for each row\ndeclare\n  k number\\;\nbegin\n  k := 1\\;\nend;\n",
			"create trigger t before insert on x for each row\ndeclare\n  k number;\nbegin\n  k := 1;\nend;",
		},
		{
			"trigger with :new",
			"create or replace trigger t before update on d\n  for each row\nbegin\n  :new.updated_at := systimestamp;\nend;\n",
			"create or replace trigger t before update on d\n  for each row\nbegin\n  :new.updated_at := systimestamp;\nend;",
		},
		{
			"package",
			"create package p as\n  function total return number\\;\n  procedure raise_all\\;\nend p;\n",
			"create package p as\n  function total return number;\n  procedure raise_all;\nend p;",
		},
		{
			"package body",
			"create package body p as\n  function total return number as\n    t number\\;\n  begin\n    return t\\;\n  end\\;\nend p;\n",
			"create package body p as\n  function total return number as\n    t number;\n  begin\n    return t;\n  end;\nend p;",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, err := NewParser(Language(oracle, false))
			if err != nil {
				t.Fatal(err)
			}
			items, err := p.Parse(test.src + "select 1 from dual;\n")
			if err != nil {
				t.Fatal(err)
			}
			if len(items) != 2 {
				t.Fatalf("expected the unit and the select, got: %s", summary(items))
			}
			// no variable is set, so :new stays as it is
			none := func(string) (string, bool) { return "", false }
			if got := items[0].Expand(none); got != test.exp {
				t.Errorf("\n got: %q\nwant: %q", got, test.exp)
			}
		})
	}
}

// TestSourceLongLine checks that a line of many statements costs about what
// its size does. A Source once parsed the rest of the line again for each
// statement it fed, so that 4,000 statements on one line took 14 seconds and
// 5 GiB. The fuzz test FuzzSplit found it.
func TestSourceLongLine(t *testing.T) {
	src := strings.Repeat("select 1;", 4000) + "\n"
	s, err := NewSource(NewLines(strings.NewReader(src)), Language(postgres, false), nil)
	if err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if got := strings.Count(feedAll(t, s), "S:select 1;"); got != 4000 {
		t.Fatalf("expected 4000 statements, got %d", got)
	}
	runtime.ReadMemStats(&after)
	// one parse of the 36 KB line takes a few MiB
	if mib := (after.TotalAlloc - before.TotalAlloc) >> 20; mib > 64 {
		t.Errorf("reading %d statements on one line allocated %d MiB", 4000, mib)
	}
}

// TestSourceLargeStatement checks that a statement of many lines costs about
// what its size does. A Source once parsed the whole text again for each line
// it read, so that an INSERT of 830 rows read from a file took gigabytes.
func TestSourceLargeStatement(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("insert into t values\n")
	for i := range 5000 {
		fmt.Fprintf(&sb, "  (%d, 'a row on 7/4/1996', null),\n", i)
	}
	sb.WriteString("  (0, 'the last row', null);\nselect 1;\n")
	s, err := NewSource(NewLines(strings.NewReader(sb.String())), Language(postgres, false), nil)
	if err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if got := feedAll(t, s); !strings.HasSuffix(got, "S:select 1;") {
		t.Fatalf("expected the insert and the select, got: %.80s", got)
	}
	runtime.ReadMemStats(&after)
	// one parse of the 200 KB text takes a few MiB
	if mib := (after.TotalAlloc - before.TotalAlloc) >> 20; mib > 64 {
		t.Errorf("reading a statement of %d bytes allocated %d MiB", sb.Len(), mib)
	}
}

// TestBackslashEscapes checks that a backslash escapes in a string and in a
// name between double quotes only where the dialect says so, as MySQL does,
// and that it is text elsewhere, as psql reads it (transit's D142).
func TestBackslashEscapes(t *testing.T) {
	t.Parallel()
	mysql := dbmeta.Syntax{BlockComments: true, DashComments: true, HashComments: true, Backticks: true, BackslashEscapes: true}
	tests := []struct {
		syntax dbmeta.Syntax
		src    string
		exp    string
	}{
		{postgres, "select 'C:\\data\\';\nselect 2;", "S:select 'C:\\data\\'; | S:select 2;"},
		{postgres, "select E'a\\'b';\nselect 2;", "S:select E'a\\'b'; | S:select 2;"},
		{postgres, "select 1 as \"a\\\";\nselect 2;", "S:select 1 as \"a\\\"; | S:select 2;"},
		{mysql, "select 'a\\'b';\nselect 2;", "S:select 'a\\'b'; | S:select 2;"},
		{mysql, "select 'C:\\data\\';\nselect 2;", "S:select 'C:\\data\\';\nselect 2;(open)"},
		{mysql, "select 1 as \"a\\\"b\";\nselect 2;", "S:select 1 as \"a\\\"b\"; | S:select 2;"},
	}
	for _, test := range tests {
		p, err := NewParser(Language(test.syntax, false))
		if err != nil {
			t.Fatal(err)
		}
		items, err := p.Parse(test.src)
		if err != nil {
			t.Fatal(err)
		}
		if got := summary(items); got != test.exp {
			t.Errorf("%q:\n got: %q\nwant: %q", test.src, got, test.exp)
		}
	}
}
