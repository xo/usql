package dump

import (
	"errors"
	"testing"

	"github.com/xo/dbmeta"
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
