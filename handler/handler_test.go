package handler

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/rline"
	"github.com/xo/usql/feed"
)

// TestRunCommandMetaCommand checks that the text of -c runs a meta command
// that stands alone.
func TestRunCommandMetaCommand(t *testing.T) {
	// stdout captures what the meta command writes
	stdout := new(bytes.Buffer)
	l, err := rline.New(
		rline.WithStdin(strings.NewReader("")),
		rline.WithStdout(stdout),
		rline.WithStderr(io.Discard),
		rline.WithColor(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	h, err := New(l, feed.NewLines(strings.NewReader("")), &user.User{}, ".", memfs.New(), true)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.RunCommand(`\echo hello`); err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if got, want := stdout.String(), "hello\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

// TestHistory checks that \history writes the history oldest first, with the
// number of each entry, and only the last entries when it is given a count,
// and that \s writes it as psql does, to the output or to a file.
func TestHistory(t *testing.T) {
	// a meta command reads a backslash as the start of another command, so
	// the path of the file is written with forward slashes, which Windows
	// also takes
	file := filepath.ToSlash(filepath.Join(t.TempDir(), "history.txt"))
	tests := []struct {
		cmd  string
		want string
		err  string
	}{
		{`\history`, "    1  select 1;\n    2  select\n       2;\n    3  \\dt\n", ""},
		{`\history 2`, "    2  select\n       2;\n    3  \\dt\n", ""},
		{`\history 9`, "    1  select 1;\n    2  select\n       2;\n    3  \\dt\n", ""},
		{`\history 0`, "", ""},
		{`\history x`, "", `\history: "x" is not a count`},
		{`\s`, "select 1;\nselect\n2;\n\\dt\n", ""},
		{`\s ` + file, "Wrote history to file \"" + file + "\".\n", ""},
	}
	for _, test := range tests {
		t.Run(test.cmd, func(t *testing.T) {
			stdout := new(bytes.Buffer)
			l, err := rline.New(
				rline.WithStdin(strings.NewReader("")),
				rline.WithStdout(stdout),
				rline.WithStderr(io.Discard),
				rline.WithColor(false),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			for _, entry := range []string{"select 1;", "select\n2;", `\dt`} {
				l.AddHistory(entry)
			}
			h, err := New(l, feed.NewLines(strings.NewReader("")), &user.User{}, ".", memfs.New(), true)
			if err != nil {
				t.Fatal(err)
			}
			err = h.RunCommand(test.cmd)
			switch {
			case test.err != "" && (err == nil || err.Error() != test.err):
				t.Fatalf("RunCommand = %v, want %q", err, test.err)
			case test.err == "" && err != nil:
				t.Fatalf("RunCommand: %v", err)
			}
			if got := stdout.String(); got != test.want {
				t.Errorf("stdout = %q, want %q", got, test.want)
			}
		})
	}
	// \s FILE wrote the history as \s writes it to the output
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), "select 1;\nselect\n2;\n\\dt\n"; got != want {
		t.Errorf("%s = %q, want %q", file, got, want)
	}
}

// TestErrVerbose checks that \errverbose says that there is no error before a
// statement fails, and then writes the error of the last statement that
// failed, as the driver gives it.
func TestErrVerbose(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	l, err := rline.New(
		rline.WithStdin(strings.NewReader("")),
		rline.WithStdout(stdout),
		rline.WithStderr(stderr),
		rline.WithColor(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	h, err := New(l, feed.NewLines(strings.NewReader("")), &user.User{}, ".", memfs.New(), true)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Open(context.Background(), "moderncsqlite:"+filepath.Join(t.TempDir(), "e.db")); err != nil {
		t.Fatal(err)
	}
	// Windows cannot remove the file of the database while it is open
	defer h.Close()
	if err := h.RunCommand(`\errverbose`); err != nil {
		t.Fatal(err)
	}
	if got, want := stdout.String(), "There is no previous error.\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	// SQLite cannot describe a query without running it
	if err := h.RunCommand(`select 1 as a \gdesc`); err == nil || !strings.Contains(err.Error(), `\gdesc: describing a query without running it is not supported`) {
		t.Errorf("\\gdesc on SQLite = %v, want the refusal", err)
	}
	_ = h.RunCommand(`select * from nope;`)
	stderr.Reset()
	if err := h.RunCommand(`\errverbose`); err != nil {
		t.Fatal(err)
	}
	if got, want := stderr.String(), "ERROR:  1: no such table: nope\nDETAIL:  Generic error (SQLITE_ERROR)\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

// TestRestrict runs a script as pg_dump writes it. Between \restrict and
// \unrestrict, statements run and meta commands do not, and only the same key
// leaves restricted mode.
func TestRestrict(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	l, err := rline.New(
		rline.WithStdin(strings.NewReader("")),
		rline.WithStdout(stdout),
		rline.WithStderr(stderr),
		rline.WithColor(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	h, err := New(l, feed.NewLines(strings.NewReader("")), &user.User{}, ".", memfs.New(), true)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Open(context.Background(), "moderncsqlite:"+filepath.Join(t.TempDir(), "r.db")); err != nil {
		t.Fatal(err)
	}
	// Windows cannot remove the file of the database while it is open
	defer h.Close()
	script := strings.Join([]string{
		`\unrestrict abc`,
		`\restrict abc`,
		`\echo blocked`,
		`select 'ran' as statement;`,
		`\unrestrict wrong`,
		`\unrestrict abc`,
		`\echo after`,
	}, "\n")
	_ = h.IncludeReader(strings.NewReader(script), "dump.sql")
	out := stdout.String()
	for _, s := range []string{"ran", "after"} {
		if !strings.Contains(out, s) {
			t.Errorf("stdout lacks %q:\n%s", s, out)
		}
	}
	if strings.Contains(out, "blocked") {
		t.Errorf("\\echo ran in restricted mode:\n%s", out)
	}
	errs := stderr.String()
	for _, s := range []string{
		`\unrestrict: not currently in restricted mode`,
		`backslash commands are restricted: only \unrestrict is allowed`,
		`\unrestrict: wrong key`,
	} {
		if !strings.Contains(errs, s) {
			t.Errorf("stderr lacks %q:\n%s", s, errs)
		}
	}
	if h.restrictKey != "" {
		t.Errorf("restricted after \\unrestrict abc")
	}
}
