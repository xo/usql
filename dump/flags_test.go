package dump

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestFlags sets each option of a dump.
func TestFlags(t *testing.T) {
	var o Options
	for _, s := range [][2]string{
		{"target", "duckdb"},
		{"schema-only", ""},
		{"data-only", "true"},
		{"schema", "a"},
		{"schema", "b"},
		{"exclude-schema", "c"},
		{"table", "t"},
		{"exclude-table", "x"},
		{"allow-lossy", ""},
		{"allow-text", "false"},
		{"restrict-key", "k"},
	} {
		f, ok := LookupFlag(s[0])
		if !ok {
			t.Fatalf("no option %s", s[0])
		}
		if err := f.Set(&o, s[1]); err != nil {
			t.Fatalf("%s: %v", s[0], err)
		}
	}
	if o.Target != "duckdb" || !o.SchemaOnly || !o.DataOnly || !slices.Equal(o.Schemas, []string{"a", "b"}) ||
		!slices.Equal(o.ExcludeSchemas, []string{"c"}) || !slices.Equal(o.Tables, []string{"t"}) ||
		!slices.Equal(o.ExcludeTables, []string{"x"}) || !o.Lossy || o.Text || o.RestrictKey != "k" {
		t.Errorf("options = %+v", o)
	}
	f, _ := LookupFlag("schema-only")
	if err := f.Set(&o, "maybe"); !errors.Is(err, ErrNotBoolean) {
		t.Errorf("schema-only=maybe: %v, want %v", err, ErrNotBoolean)
	}
	if _, ok := LookupFlag("nope"); ok {
		t.Errorf("LookupFlag(nope) found an option")
	}
}

// TestWriteFile replaces the file only when the dump is complete, and leaves
// no temporary file.
func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "dump.sql")
	if err := WriteFile(name, func(w io.Writer) error {
		_, err := io.WriteString(w, "first\n")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	failed := errors.New("failed")
	if err := WriteFile(name, func(w io.Writer) error {
		_, _ = io.WriteString(w, "second\n")
		return failed
	}); !errors.Is(err, failed) {
		t.Fatalf("WriteFile returned %v, want %v", err, failed)
	}
	if b, err := os.ReadFile(name); err != nil || string(b) != "first\n" {
		t.Errorf("the file is %q, %v, want the first dump", b, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("the directory holds %d files, want 1", len(entries))
	}
}
