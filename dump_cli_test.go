package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDumpCLI runs usql --dump as a script runs it: it dumps a SQLite
// database to a file with -o, loads the file into a new database with -f,
// and dumps the copy to standard output. The two dumps are the same text.
func TestDumpCLI(t *testing.T) {
	if usqlPath == "" {
		t.Skip("TestMain builds usql on Linux and macOS only")
	}
	dir := t.TempDir()
	src, dst := "sqlite3:"+filepath.Join(dir, "src.db"), "sqlite3:"+filepath.Join(dir, "dst.db")
	fixture, err := filepath.Abs(filepath.Join("drivers", "testdata", "dump", "sqlite3.sql"))
	if err != nil {
		t.Fatal(err)
	}
	usql := func(t *testing.T, args ...string) (string, string, error) {
		t.Helper()
		cmd := exec.Command(usqlPath, args...)
		// no configuration, history or variable of the user applies
		cmd.Env = []string{"HOME=" + dir, "PATH=" + os.Getenv("PATH"), "XDG_CONFIG_HOME=" + filepath.Join(dir, "config")}
		cmd.Dir = dir
		stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
		cmd.Stdout, cmd.Stderr = stdout, stderr
		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	}
	if _, stderr, err := usql(t, "-X", "-q", src, "-f", fixture); err != nil {
		t.Fatalf("loading the fixture: %v\n%s", err, stderr)
	}
	out := filepath.Join(dir, "dump.sql")
	if _, stderr, err := usql(t, src, "--dump", "--restrict-key", "k", "-o", out); err != nil {
		t.Fatalf("usql --dump -o: %v\n%s", err, stderr)
	}
	want, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{"-- usql dump", `\uset on_error_stop on`, `\restrict k`, `\unrestrict k`, `CREATE TABLE "authors"`, `INSERT INTO "books"`} {
		if !strings.Contains(string(want), line+"\n") && !strings.Contains(string(want), line+" ") {
			t.Errorf("the dump has no line %q", line)
		}
	}
	if _, stderr, err := usql(t, "-X", "-q", dst, "-f", out); err != nil {
		t.Fatalf("loading the dump: %v\n%s", err, stderr)
	}
	got, stderr, err := usql(t, dst, "-D", "--restrict-key", "k")
	if err != nil {
		t.Fatalf("usql -D: %v\n%s", err, stderr)
	}
	if got != string(want) {
		t.Errorf("the dump of the copy differs from the dump of the source:\n--- source\n%s\n--- copy\n%s", want, got)
	}

	// the options of a dump
	if got, _, err := usql(t, src, "--dump", "--schema-only", "--restrict-key", "k"); err != nil || strings.Contains(got, "INSERT INTO") {
		t.Errorf("--schema-only returned %v, and wrote rows: %t", err, strings.Contains(got, "INSERT INTO"))
	}
	if got, _, err := usql(t, src, "--dump", "--data-only", "--restrict-key", "k"); err != nil || strings.Contains(got, "CREATE TABLE") {
		t.Errorf("--data-only returned %v, and wrote tables: %t", err, strings.Contains(got, "CREATE TABLE"))
	}
	if got, _, err := usql(t, src, "--dump", "--table", "books", "--restrict-key", "k"); err != nil || strings.Contains(got, `"authors"`) || !strings.Contains(got, `"books"`) {
		t.Errorf("--table books returned %v, and wrote:\n%s", err, got)
	}

	// \dump takes the options of --dump, as -NAME or --NAME, and writes the
	// same script to the output, to a file and to a command
	cli, _, err := usql(t, src, "--dump", "--table", "books", "--restrict-key", "k")
	if err != nil {
		t.Fatal(err)
	}
	metaFile, pipeFile := filepath.Join(dir, "meta.sql"), filepath.Join(dir, "pipe.sql")
	for _, cmd := range []string{
		`\dump -table books -restrict-key k`,
		`\dump --table books --restrict-key k -`,
		`\dump -table books -restrict-key k ` + metaFile,
		`\dump -table books -restrict-key k |cat > ` + pipeFile,
	} {
		stdout, stderr, err := usql(t, "-X", src, "-c", cmd)
		if err != nil {
			t.Fatalf("%s: %v\n%s", cmd, err, stderr)
		}
		got := stdout
		switch {
		case strings.Contains(cmd, metaFile):
			got = readFile(t, metaFile)
		case strings.Contains(cmd, pipeFile):
			got = readFile(t, pipeFile)
		}
		if got != cli {
			t.Errorf("%s wrote a script that differs from --dump:\n--- --dump\n%s\n--- \\dump\n%s", cmd, cli, got)
		}
	}
	if fi, err := os.Stat(metaFile); err != nil || fi.Mode().Perm() != 0o600 {
		t.Errorf("the file of \\dump has mode %v, %v, want 0600", fi.Mode().Perm(), err)
	}
	for cmd, want := range map[string]string{
		`\dump -bogus`:                         `invalid option "bogus"`,
		`\dump -table`:                         "-table needs a value: PATTERN",
		`\dump -schema-only -data-only ` + out: "schema-only and data-only cannot be used together",
	} {
		if _, stderr, _ := usql(t, "-X", src, "-c", cmd); !strings.Contains(stderr, want) {
			t.Errorf("%s wrote %q, want an error with %q", cmd, stderr, want)
		}
	}

	// what usql refuses, and a failed dump leaves the file as it was
	for _, test := range []struct {
		args []string
		want string
	}{
		{[]string{src, "--schema-only"}, "the options of a dump need --dump"},
		{[]string{src, "--dump", "-c", "select 1"}, "--dump cannot be used with -c or -f"},
		{[]string{src, "--dump", "-1"}, "--dump cannot be used with --single-transaction"},
		{[]string{src, "--dump", "--schema-only", "--data-only", "-o", out}, "schema-only and data-only cannot be used together"},
		{[]string{src, "--dump", "--target", "nosuchdb", "-o", out}, "unknown target: nosuchdb"},
	} {
		_, stderr, err := usql(t, test.args...)
		if err == nil || !strings.Contains(stderr, test.want) {
			t.Errorf("usql %s returned %v and wrote %q, want an error with %q", strings.Join(test.args[1:], " "), err, stderr, test.want)
		}
	}
	if b, err := os.ReadFile(out); err != nil || !bytes.Equal(b, want) {
		t.Errorf("a failed dump changed the file of -o")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".dump.sql.") {
			t.Errorf("a failed dump left the temporary file %s", e.Name())
		}
	}
}
