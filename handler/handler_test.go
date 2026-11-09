package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/rline"
	"github.com/xo/usql/env"
	"github.com/xo/usql/feed"
	"github.com/xo/usql/text"
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

// TestHistory checks that \hist writes the history oldest first, with the
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
		{`\hist`, "    1  select 1;\n    2  select\n       2;\n    3  \\dt\n", ""},
		{`\hist 2`, "    2  select\n       2;\n    3  \\dt\n", ""},
		{`\history 9`, "    1  select 1;\n    2  select\n       2;\n    3  \\dt\n", ""},
		{`\history 0`, "", ""},
		{`\history x`, "", `\history: "x" is not a count`},
		{`\hist x`, "", `\hist: "x" is not a count`},
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

// TestGsetRows checks that \gset refuses a query that returns no row or more
// than one, as psql does, and sets the variables from a single row.
func TestGsetRows(t *testing.T) {
	l, err := rline.New(
		rline.WithStdin(strings.NewReader("")),
		rline.WithStdout(new(bytes.Buffer)),
		rline.WithStderr(new(bytes.Buffer)),
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
	if err := h.Open(context.Background(), "moderncsqlite:"+filepath.Join(t.TempDir(), "g.db")); err != nil {
		t.Fatal(err)
	}
	// Windows cannot remove the file of the database while it is open
	defer h.Close()
	if err := h.RunCommand(`select 1 as gset_a where 1 = 0 \gset`); !errors.Is(err, text.ErrNoRowsGset) {
		t.Errorf("\\gset with no rows = %v, want %v", err, text.ErrNoRowsGset)
	}
	if err := h.RunCommand(`select 1 as gset_a union all select 2 \gset`); !errors.Is(err, text.ErrTooManyRows) {
		t.Errorf("\\gset with two rows = %v, want %v", err, text.ErrTooManyRows)
	}
	if err := h.RunCommand(`select 7 as gset_a \gset`); err != nil {
		t.Fatal(err)
	}
	if got, _ := env.Vars().Get("gset_a"); got != "7" {
		t.Errorf("gset_a = %q, want 7", got)
	}
	_ = env.Vars().Unset("gset_a")
}

// TestQueryBufferRaw checks that \p and \w write the query buffer with its
// variables replaced, as it is sent, and as it was written with -raw. \raw is
// \p -raw, \exec is \p, and -exec is not an option.
func TestQueryBufferRaw(t *testing.T) {
	stdout := new(bytes.Buffer)
	l, err := rline.New(
		rline.WithStdin(strings.NewReader("")),
		rline.WithStdout(stdout),
		rline.WithStderr(new(bytes.Buffer)),
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
	dir := t.TempDir()
	if err := h.Open(context.Background(), "moderncsqlite:"+filepath.Join(dir, "q.db")); err != nil {
		t.Fatal(err)
	}
	// Windows cannot remove the file of the database while it is open
	defer h.Close()
	if err := env.Vars().Set("qbuf_x", "42"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = env.Vars().Unset("qbuf_x") }()
	// a meta command reads a backslash as the start of another command, so
	// the paths are written with forward slashes, which Windows also takes
	exp, raw := filepath.ToSlash(filepath.Join(dir, "exp.sql")), filepath.ToSlash(filepath.Join(dir, "raw.sql"))
	// the buffer holds the query
	if err := h.RunCommand(`select :qbuf_x as v \p \p -raw \raw \exec \w ` + exp + ` \w -raw ` + raw); err != nil {
		t.Fatal(err)
	}
	if want := "select 42 as v\nselect :qbuf_x as v\nselect :qbuf_x as v\nselect 42 as v\n"; !strings.HasPrefix(stdout.String(), want) {
		t.Errorf("stdout = %q, want it to start with %q", stdout.String(), want)
	}
	for file, want := range map[string]string{exp: "select 42 as v\n", raw: "select :qbuf_x as v\n"} {
		if b, err := os.ReadFile(file); err != nil || string(b) != want {
			t.Errorf("%s = %q, %v, want %q", file, b, err, want)
		}
	}
	// the buffer is empty, and the last query is written
	stdout.Reset()
	if err := h.RunCommand(`\p \p -raw`); err != nil {
		t.Fatal(err)
	}
	if got, want := stdout.String(), "select 42 as v\nselect :qbuf_x as v\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	for _, cmd := range []string{`\p -exec`, `\w -exec ` + exp, `\e -exec`} {
		if err := h.RunCommand(cmd); err == nil || err.Error() != `invalid option "exec"` {
			t.Errorf("%s = %v, want invalid option", cmd, err)
		}
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
