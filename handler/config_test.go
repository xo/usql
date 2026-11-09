package handler

import (
	"bytes"
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/rline"
	"github.com/xo/usql/env"
	"github.com/xo/usql/feed"
	"github.com/xo/usql/text"
)

// TestConfig sets and shows settings with \config, and with \pset, \uset,
// \set, \cset and \setenv, which set its sections.
func TestConfig(t *testing.T) {
	h, stdout, stderr := newConfigHandler(t)
	t.Cleanup(func() {
		for _, opt := range []string{"print.border=1", "print.footer=on", "print.expanded=off", "usql.fetch_count=0"} {
			_ = env.Vars().SetOption(opt)
		}
	})
	for _, test := range []struct {
		cmd, want string
	}{
		{`\config print.border 2`, "Border style is 2.\n"},
		{`\config print.border`, "print.border 2\n"},
		{`\pset border 3`, "Border style is 3.\n"},
		{`\config print.border`, "print.border 3\n"},
		{`\pset footer`, "Default footer is off.\n"},
		{`\x`, "Expanded display is on.\n"},
		{`\config print.expanded`, "print.expanded on\n"},
		{`\uset fetch_count 5`, "fetch_count is 5\n"},
		{`\config usql.fetch_count 6`, "usql.fetch_count is 6\n"},
		{`\uset fetch_count`, "fetch_count is 6\n"},
		{`\set w80_a 1`, ""},
		{`\config vars.w80_a`, "vars.w80_a '1'\n"},
		{`\config vars.w80_a 2`, ""},
		{`\echo :w80_a`, "2\n"},
		{`\unset w80_a`, ""},
		{`\cset w80_c pg://u:p@h/d`, ""},
		{`\config connections.w80_c`, "connections.w80_c 'pg://u:xxxxx@h/d'\n"},
		{`\config -password connections.w80_c`, "connections.w80_c 'pg://u:p@h/d'\n"},
		{`\cset`, "w80_c 'pg://u:xxxxx@h/d'\n"},
		{`\setenv W80_E x`, ""},
		{`\config env.W80_E`, "env.W80_E 'x'\n"},
		{`\config env`, "env.W80_E 'x'\n"},
		{`\setenv W80_E`, ""},
		{`\config env`, ""},
	} {
		stdout.Reset()
		if err := h.RunCommand(test.cmd); err != nil {
			t.Fatalf("%s: %v\n%s", test.cmd, err, stderr)
		}
		if got := stdout.String(); got != test.want {
			t.Errorf("%s = %q, want %q", test.cmd, got, test.want)
		}
	}
	// \set shows the variables by their names, without the section
	stdout.Reset()
	if err := h.RunCommand(`\set w80_b 3`); err != nil {
		t.Fatal(err)
	}
	if err := h.RunCommand(`\set`); err != nil {
		t.Fatal(err)
	}
	if s := stdout.String(); !strings.Contains(s, "w80_b = '3'\n") || strings.Contains(s, "w80_a") || strings.Contains(s, "vars.") {
		t.Errorf(`\set = %q, want w80_b = '3', and no w80_a`, s)
	}
	// \pset and \uset show their names without the section
	for _, cmd := range []string{`\pset`, `\uset`} {
		stdout.Reset()
		if err := h.RunCommand(cmd); err != nil {
			t.Fatal(err)
		}
		if s := stdout.String(); strings.Contains(s, "print.") || strings.Contains(s, "usql.") || s == "" {
			t.Errorf("%s = %q, want names without the section", cmd, s)
		}
	}
	// \config shows only usql and print, and -all shows every section
	stdout.Reset()
	if err := h.RunCommand(`\config`); err != nil {
		t.Fatal(err)
	}
	if s := stdout.String(); !strings.Contains(s, "print.border ") || !strings.Contains(s, "usql.quiet ") || strings.Contains(s, "connections.") || strings.Contains(s, "vars.") || strings.Contains(s, "charts.") {
		t.Errorf(`\config = %q, want only the sections usql and print`, s)
	}
	stdout.Reset()
	if err := h.RunCommand(`\config -all`); err != nil {
		t.Fatal(err)
	}
	if s := stdout.String(); !strings.Contains(s, "connections.w80_c ") || !strings.Contains(s, "vars.w80_b ") || !strings.Contains(s, "charts.path ") {
		t.Errorf(`\config -all = %q, want every section`, s)
	}
	for cmd, want := range map[string]error{
		`\config aliases.a b`: text.ErrAliasConfigOnly,
		`\config border 2`:    nil,
		`\config -x`:          nil,
		`\config vars.w80_no`: nil,
	} {
		err := h.RunCommand(cmd)
		switch {
		case err == nil:
			t.Errorf("%s: no error", cmd)
		case want != nil && !errors.Is(err, want):
			t.Errorf("%s: error %v, want %v", cmd, err, want)
		}
	}
}

// TestAlias runs the aliases of the configuration file.
func TestAlias(t *testing.T) {
	h, stdout, stderr := newConfigHandler(t)
	for name, body := range map[string]string{
		"w80_hi":   `\echo hi`,
		"w80_two":  "\\set w80_x 1\n\\echo x is :w80_x\n",
		"w80_sel":  "select 1",
		"w80_loop": `\w80_loop`,
		"w80_bye":  "\\echo bye\n\\q\n",
	} {
		if err := env.Vars().SetAlias(name, body); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		cmd, want string
		err       error
	}{
		{`\w80_hi`, "hi\n", nil},
		{`\w80_two`, "x is 1\n", nil},
		// a statement that an alias leaves in the buffer stays there
		{"\\w80_sel\n\\p\n\\r", "select 1\nQuery buffer reset (cleared).\n", nil},
		{"\\w80_bye\n\\echo after", "bye\n", nil},
		{`\w80_hi there`, "", text.ErrAliasArguments},
		{`\w80_loop`, "", text.ErrAliasTooDeep},
	} {
		stdout.Reset()
		stderr.Reset()
		err := h.RunCommand(test.cmd)
		switch {
		case test.err == nil && err != nil:
			t.Errorf("%s: %v", test.cmd, err)
		case test.err != nil && !errors.Is(err, test.err):
			t.Errorf("%s: error %v, want %v", test.cmd, err, test.err)
		case test.err != nil && strings.Count(stderr.String(), test.err.Error()) != 1:
			t.Errorf("%s: stderr = %q, want the error once", test.cmd, stderr)
		}
		if got := stdout.String(); got != test.want {
			t.Errorf("%s = %q, want %q", test.cmd, got, test.want)
		}
	}
}

// TestEnvChild gives the settings of env to a backtick and to \!, and not to
// usql itself.
func TestEnvChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the commands are written for sh")
	}
	h, _, stderr := newConfigHandler(t)
	file := filepath.Join(t.TempDir(), "out")
	t.Cleanup(func() { _ = env.Vars().UnsetEnv("W80_CHILD") })
	for _, cmd := range []string{
		`\setenv W80_CHILD hello`,
		"\\set w80_got `echo $W80_CHILD`",
		`\! echo $W80_CHILD > ` + file,
	} {
		if err := h.RunCommand(cmd); err != nil {
			t.Fatalf("%s: %v\n%s", cmd, err, stderr)
		}
	}
	if got, _ := env.Vars().Get("w80_got"); got != "hello" {
		t.Errorf("a backtick read W80_CHILD=%q, want hello", got)
	}
	if b, err := os.ReadFile(file); err != nil || string(b) != "hello\n" {
		t.Errorf(`\! wrote %q, %v, want "hello\n"`, b, err)
	}
	if s, ok := os.LookupEnv("W80_CHILD"); ok {
		t.Errorf("the environment of usql has W80_CHILD=%q", s)
	}
}

// newConfigHandler returns a handler that is not connected, and its output.
func newConfigHandler(t *testing.T) (*Handler, *bytes.Buffer, *bytes.Buffer) {
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
	h, err := New(l, feed.NewLines(strings.NewReader("")), &user.User{}, ".", memfs.New(), true)
	if err != nil {
		t.Fatal(err)
	}
	return h, stdout, stderr
}
