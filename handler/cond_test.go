package handler

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/rline"
	"github.com/xo/usql/feed"
	"github.com/xo/usql/text"
)

// TestConds checks each move of the stack of \if blocks, as psql's
// exec_command_if, exec_command_elif, exec_command_else and
// exec_command_endif make it.
func TestConds(t *testing.T) {
	t.Parallel()
	// each step is an op and its argument: if and elif take t, f or e, for
	// an expression that is true, false or an error
	tests := []struct {
		name   string
		steps  []string
		active bool
		depth  int
		err    error
		evals  int
	}{
		{"none", nil, true, 0, nil, 0},
		{"if true", []string{"if t"}, true, 1, nil, 1},
		{"if false", []string{"if f"}, false, 1, nil, 1},
		{"if error", []string{"if e"}, false, 1, errBool, 1},
		{"elif after true", []string{"if t", "elif t"}, false, 1, nil, 1},
		{"elif after false", []string{"if f", "elif t"}, true, 1, nil, 2},
		{"elif false after false", []string{"if f", "elif f"}, false, 1, nil, 2},
		{"elif error after false", []string{"if f", "elif e"}, false, 1, errBool, 2},
		{"elif after error", []string{"if e", "elif t"}, true, 1, nil, 2},
		{"elif after elif error", []string{"if f", "elif e", "elif t"}, true, 1, nil, 3},
		{"elif after taken elif", []string{"if f", "elif t", "elif t"}, false, 1, nil, 2},
		{"else after true", []string{"if t", "else"}, false, 1, nil, 1},
		{"else after false", []string{"if f", "else"}, true, 1, nil, 1},
		{"else after taken elif", []string{"if f", "elif t", "else"}, false, 1, nil, 2},
		{"else after false elif", []string{"if f", "elif f", "else"}, true, 1, nil, 2},
		{"endif", []string{"if t", "endif"}, true, 0, nil, 1},
		{"nested in false", []string{"if f", "if t"}, false, 2, nil, 1},
		{"elif nested in false", []string{"if f", "if t", "elif t"}, false, 2, nil, 1},
		{"else nested in false", []string{"if f", "if t", "else"}, false, 2, nil, 1},
		{"nested in true", []string{"if t", "if t"}, true, 2, nil, 2},
		{"nested endif", []string{"if f", "if t", "endif"}, false, 1, nil, 1},
		{"nested endif else", []string{"if f", "if t", "endif", "else"}, true, 1, nil, 1},
		{"elif no if", []string{"elif t"}, true, 0, text.ErrElifNoMatchingIf, 0},
		{"else no if", []string{"else"}, true, 0, text.ErrElseNoMatchingIf, 0},
		{"endif no if", []string{"endif"}, true, 0, text.ErrEndIfNoMatchingIf, 0},
		{"elif after else", []string{"if f", "else", "elif t"}, true, 1, text.ErrElifAfterElse, 1},
		{"else after else", []string{"if t", "else", "else"}, false, 1, text.ErrElseAfterElse, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var c conds
			var err error
			evals := 0
			for _, step := range test.steps {
				op, arg, _ := strings.Cut(step, " ")
				eval := func() (bool, error) {
					evals++
					if arg == "e" {
						return false, errBool
					}
					return arg == "t", nil
				}
				switch op {
				case "if":
					err = c.begin(eval)
				case "elif":
					err = c.elif(eval)
				case "else":
					err = c.els()
				case "endif":
					err = c.end()
				}
			}
			if !errors.Is(err, test.err) {
				t.Errorf("expected error %v, got: %v", test.err, err)
			}
			if c.active() != test.active {
				t.Errorf("expected active %t, got: %t", test.active, c.active())
			}
			if len(c) != test.depth {
				t.Errorf("expected depth %d, got: %d", test.depth, len(c))
			}
			if evals != test.evals {
				t.Errorf("expected %d evaluations, got: %d", test.evals, evals)
			}
		})
	}
}

// errBool is the error of an expression that is not a boolean.
var errBool = errors.New("not a boolean")

// TestConditional runs scripts with \if blocks, as -f runs them. It does not
// run in parallel, because a script sets variables, which are global. The
// handler has no database, so a statement that ran would print the error that
// usql is not connected.
//
// The expected output was read from psql 18.6, which ran each script with -f.
// psql writes "psql:FILE:LINE: error: " before an error, where usql writes
// "error: ". psql writes "invalid command \foo" for an unknown command, where
// usql writes its own text.
func TestConditional(t *testing.T) {
	tests := []struct {
		name   string
		script string
		stdout string
		stderr string
	}{
		{
			"branches",
			`\set v 1
\if :v
\echo a
\elif 1
\echo b
\else
\echo c
\endif
\if false
\echo :v
\foo
\i nosuchfile.sql
select 1;
\elif TR
\echo elif-tr
\endif
\if o
\echo o-true
\else
\echo o-else
\endif
\if
\echo empty-true
\endif
\if true false
\echo two-true
\endif
\if Of
\echo of
\else
\echo of-else
\endif
\if 'ye'
\echo ye
\endif
\else
\elif x
\endif
\if 1
\else
\else
\elif 1
\endif
\if 1
\if 0
\echo no
\if :undefined
\echo no2
\endif
\elif bad
\echo no3
\else
\echo nested-else
\endif
\endif
\echo end
\if 1
`,
			"a\nelif-tr\no-else\nof-else\nye\nnested-else\nend\n",
			`Invalid command \foo. Try \? for help.
error: unrecognized value "o" for "\if expression": Boolean expected
error: unrecognized value "" for "\if expression": Boolean expected
error: unrecognized value "true false" for "\if expression": Boolean expected
error: \else: no matching \if
error: \elif: no matching \if
error: \endif: no matching \if
error: \else: cannot occur after \else
error: \elif: cannot occur after \else
error: unrecognized value "bad" for "\elif expression": Boolean expected
error: reached EOF without finding closing \endif(s)
`,
		},
		{
			"buffer and variables",
			`\set v 1
select :v as a,
\if false
\set v 2
select 2 as b;
:v as b,
\elif true
3 as c,
\else
4 as d,
\endif
5 as e
\p
\r
\echo :v
\if false
\if :undefined
\elif ` + "`echo true`" + `
\endif
\endif
\if true
\echo yes
\elif ` + "`echo true`" + `
\echo no
\endif
\if f
\elif t
\echo elif-t
\endif
`,
			"select 1 as a,\n3 as c,\n5 as e\nQuery buffer reset (cleared).\n1\nyes\nelif-t\n",
			"",
		},
		{
			// psql does not send the buffer at the end of the input in a
			// branch that does not run
			"unsent at end",
			"select 1\n\\if false\n",
			"",
			"error: reached EOF without finding closing \\endif(s)\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			name := filepath.Join(dir, "script.sql")
			if err := os.WriteFile(name, []byte(test.script), 0o644); err != nil {
				t.Fatal(err)
			}
			h, stdout, stderr := newTestHandler(t, feed.NewLines(strings.NewReader("")))
			if err := h.Include(name, false); err != nil && test.stderr == "" {
				t.Errorf("expected no error, got: %v", err)
			}
			if got := stdout.String(); got != test.stdout {
				t.Errorf("stdout:\n%s\nexpected:\n%s", got, test.stdout)
			}
			if got := stderr.String(); got != test.stderr {
				t.Errorf("stderr:\n%s\nexpected:\n%s", got, test.stderr)
			}
		})
	}
}

// TestConditionalInclude checks that each file has its own stack of \if
// blocks, and that a file included in a branch that does not run is not read.
// The expected output was read from psql 18.6, as in TestConditional.
func TestConditionalInclude(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"main.sql": `\if true
\ir inc.sql
\echo after-inc
\endif
\if false
\ir inc.sql
\ir nosuchfile.sql
\endif
\echo top
`,
		"inc.sql": `\echo in-inc
\endif
\if true
\echo inc-if
`,
	}
	for name, s := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	h, stdout, stderr := newTestHandler(t, feed.NewLines(strings.NewReader("")))
	if err := h.Include(filepath.Join(dir, "main.sql"), true); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got, exp := stdout.String(), "in-inc\ninc-if\nafter-inc\ntop\n"; got != exp {
		t.Errorf("stdout:\n%s\nexpected:\n%s", got, exp)
	}
	exp := "error: \\endif: no matching \\if\nerror: reached EOF without finding closing \\endif(s)\n"
	if got := stderr.String(); got != exp {
		t.Errorf("stderr:\n%s\nexpected:\n%s", got, exp)
	}
}

// TestConditionalCommand checks that the text of -c has a stack of its own,
// and that it does not change the stack of the input that runs it.
func TestConditionalCommand(t *testing.T) {
	h, stdout, stderr := newTestHandler(t, feed.NewLines(strings.NewReader("")))
	h.cond = conds{branchFalse}
	if err := h.RunCommand("\\echo a\n\\if false\n\\echo b\n"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got, exp := stdout.String(), "a\n"; got != exp {
		t.Errorf("stdout = %q, expected %q", got, exp)
	}
	if got, exp := stderr.String(), "error: reached EOF without finding closing \\endif(s)\n"; got != exp {
		t.Errorf("stderr = %q, expected %q", got, exp)
	}
	if len(h.cond) != 1 || h.cond.active() {
		t.Errorf("expected the stack of the handler to stay as it was, got: %v", h.cond)
	}
}

// TestConditionalInterrupt checks that an interrupt leaves the innermost \if
// block, as Ctrl-C does in psql. psql 18.6 at a terminal printed
// "\if: escaped".
func TestConditionalInterrupt(t *testing.T) {
	r := &interruptLines{lines: []string{`\if false`, `\if true`, "", "", `\echo after`}}
	h, stdout, stderr := newTestHandler(t, r)
	if err := h.Run(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got, exp := stdout.String(), "after\n"; got != exp {
		t.Errorf("stdout = %q, expected %q", got, exp)
	}
	if got, exp := stderr.String(), "error: \\if: escaped\nerror: \\if: escaped\n"; got != exp {
		t.Errorf("stderr = %q, expected %q", got, exp)
	}
}

// interruptLines is a feed.LineReader of lines, where an empty line is an
// interrupt.
type interruptLines struct {
	lines []string
}

// ReadLine returns the next line.
func (r *interruptLines) ReadLine(string) (string, error) {
	if len(r.lines) == 0 {
		return "", io.EOF
	}
	line := r.lines[0]
	r.lines = r.lines[1:]
	if line == "" {
		return "", rline.ErrInterrupted
	}
	return line, nil
}

// TestPromptInactive checks that %R is @ in a branch that does not run, as
// in psql, while the query buffer is empty and the entry is new.
func TestPromptInactive(t *testing.T) {
	t.Parallel()
	h, _, _ := newTestHandler(t, feed.NewLines(strings.NewReader("")))
	if got := h.render("%R", false); got != "=" {
		t.Errorf("active: expected =, got %q", got)
	}
	h.cond = conds{branchTrue, branchIgnored}
	if got := h.render("%R", false); got != "@" {
		t.Errorf("inactive: expected @, got %q", got)
	}
	if got := h.render("%R", true); got != "-" {
		t.Errorf("inactive, continued: expected -, got %q", got)
	}
	h.buf.Add(feed.Item{Kind: feed.Statement, Text: "select 1"})
	if got := h.render("%R", false); got != "-" {
		t.Errorf("inactive, with a buffer: expected -, got %q", got)
	}
}

// newTestHandler returns a handler that reads r and has no database, and
// the buffers that hold what it writes to standard output and standard
// error.
func newTestHandler(t *testing.T, r feed.LineReader) (*Handler, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
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
	t.Cleanup(func() { l.Close() })
	h, err := New(l, r, &user.User{}, ".", memfs.New(), true)
	if err != nil {
		t.Fatal(err)
	}
	return h, stdout, stderr
}
