package dump

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/xo/dbmeta"
	dbdump "github.com/xo/dbmeta/dump"
	"github.com/xo/usql/feed"
)

func TestDialect(t *testing.T) {
	for _, test := range []struct {
		name string
		want dbmeta.Dialect
	}{
		{"postgres", "postgres"},
		{"pg", "postgres"},
		{"pgx", "postgres"},
		{"sqlite", "sqlite3"},
		{"sqlite3", "sqlite3"},
		{"sq", "sqlite3"},
		{"duckdb", "duckdb"},
		{"dk:", "duckdb"},
		{"mysql", "mysql"},
		{"MariaDB", "mysql"},
	} {
		got, err := Dialect(test.name)
		if err != nil {
			t.Errorf("Dialect(%q) returned an error: %v", test.name, err)
			continue
		}
		if got != test.want {
			t.Errorf("Dialect(%q) is %q, want %q", test.name, got, test.want)
		}
	}
	for _, name := range []string{"", "nosuchdb", "file"} {
		if _, err := Dialect(name); !errors.Is(err, ErrUnknownTarget) {
			t.Errorf("Dialect(%q) returned %v, want %v", name, err, ErrUnknownTarget)
		}
	}
}

func TestTerminate(t *testing.T) {
	for _, test := range []struct{ in, want string }{
		{"SELECT 1", "SELECT 1;"},
		{"SELECT 1;", "SELECT 1;"},
		{"SELECT 1\n\n", "SELECT 1;"},
		{"CREATE TRIGGER t BEGIN SELECT 1; END", "CREATE TRIGGER t BEGIN SELECT 1; END;"},
		{"CREATE VIEW v AS SELECT 1\n  --ORDER BY 1;", "CREATE VIEW v AS SELECT 1\n  --ORDER BY 1;\n;"},
		{"SELECT 1 -- one", "SELECT 1 -- one\n;"},
		{"SELECT '--';", "SELECT '--';\n;"},
	} {
		if got := terminate(test.in); got != test.want {
			t.Errorf("terminate(%q) is %q, want %q", test.in, got, test.want)
		}
	}
}

func TestRandomKey(t *testing.T) {
	a, b := randomKey(), randomKey()
	if len(a) != 32 || a == b {
		t.Errorf("randomKey returned %q and %q, want two keys of 32 characters", a, b)
	}
	for _, c := range a {
		if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9') {
			t.Errorf("randomKey returned %q, which has %q", a, c)
		}
	}
}

// TestNotSupported checks that a dump of a dialect that dbmeta cannot dump
// names the dialects that it can, as a dump of Oracle once said only "no type
// functions". The dialect is the first whose model is built and that cannot
// be dumped, so that the test follows what dbmeta supports.
func TestNotSupported(t *testing.T) {
	var m *dbmeta.Meta
	for _, d := range dbmeta.Dialects() {
		if dbdump.Support(d).Source {
			continue
		}
		if dm, err := dbmeta.New(d, dbmeta.VersionSet{}); err == nil {
			m = dm
			break
		}
	}
	if m == nil {
		t.Skip("every dialect with a model can be dumped")
	}
	_, err := New(context.Background(), m, nil, Options{})
	if err == nil {
		t.Fatalf("New returned no error for %s", m.Dialect())
	}
	msg, _, _ := strings.Cut(err.Error(), " can: ")
	for _, want := range []string{string(m.Dialect()) + " cannot be dumped, and ", "duckdb, ", "oracle, ", "postgres, ", "sqlserver"} {
		if !strings.Contains(msg, want) {
			t.Errorf("New returned %q, want %q in the dialects that can be dumped", err, want)
		}
	}
}

// TestStatementEscapes checks that a procedure of SQL Server with no BEGIN and
// END is written with each of its inner ';' escaped, because usql ends a
// statement at the first ';' that is not escaped. Without the escape, the load
// makes a procedure that holds only its first statement, and runs the others.
func TestStatementEscapes(t *testing.T) {
	lang, ok := feed.DialectLanguage("sqlserver")
	if !ok {
		t.Fatal("no language for sqlserver")
	}
	p, err := feed.NewParser(lang)
	if err != nil {
		t.Fatal(err)
	}
	d := &Dumper{parser: p}
	var b strings.Builder
	w := &errWriter{w: &b}
	d.statement(w, dbdump.Statement{
		Name: "core.p",
		Kind: "routine",
		SQL:  "create procedure core.p as\n  select 1;\n  select 2",
	})
	if w.err != nil {
		t.Fatal(w.err)
	}
	if want := "create procedure core.p as\n  select 1\\;\n  select 2;\n"; !strings.HasSuffix(b.String(), want) {
		t.Errorf("wrote\n%q\nwant it to end with\n%q", b.String(), want)
	}
}

// TestWarningsNotDumped checks that an object that the script does not write
// gets a warning, which usql prints on standard error, and not only a comment
// in the script, which a user who does not read the script never sees.
func TestWarningsNotDumped(t *testing.T) {
	d := &Dumper{
		snap: &dbdump.Snapshot{
			Dialect: "postgres",
			Schemas: []dbdump.Schema{{
				Name:     "app",
				Routines: []dbdump.Routine{{Name: "f", Kind: "function"}},
			}},
		},
		target: "postgres",
	}
	want := "routine app.f is not dumped: the source keeps no statement that creates this function, so it is not written"
	if lines := d.Warnings(); !slices.Contains(lines, want) {
		t.Errorf("Warnings returned %q, want a line %q", lines, want)
	}
	d.opts.DataOnly = true
	if lines := d.Warnings(); slices.Contains(lines, want) {
		t.Errorf("Warnings of a dump of the rows only returned %q, which names a routine", lines)
	}
}
