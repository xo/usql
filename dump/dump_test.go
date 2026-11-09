package dump

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xo/dbmeta"
	dbdump "github.com/xo/dbmeta/dump"
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
