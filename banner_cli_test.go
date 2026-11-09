//go:build linux || darwin

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/xo/usql/internal/capture"
)

// TestNoBannerInOutput runs usql with a terminal as its standard input and a
// file as its standard output, as `usql -D URL > dump.sql` runs at a shell.
// The output holds only what was asked for: no line says which driver
// connected, as psql writes no banner then. Before, the banner was the first
// line of a dump, and the dump could not be loaded.
func TestNoBannerInOutput(t *testing.T) {
	if usqlPath == "" {
		t.Skip("TestMain builds usql on Linux and macOS only")
	}
	dir := t.TempDir()
	src := "sqlite3:" + filepath.Join(dir, "src.db")
	script := filepath.Join(dir, "script.sql")
	if err := os.WriteFile(script, []byte("select 2 as two;\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, stderr, err := runUsql(t, dir, nil, "-X", "-q", src, "-c", "create table t (a integer); insert into t values (1);"); err != nil {
		t.Fatalf("creating the database: %v\n%s", err, stderr)
	}
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"dump", []string{"-X", src, "-D", "--restrict-key", "k"}, "--\n-- usql dump\n"},
		{"command", []string{"-X", src, "-c", "select 1 as one"}, " one \n-----\n   1\n(1 row)\n"},
		{"file", []string{"-X", src, "-f", script}, " two \n-----\n   2\n(1 row)\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr := runWithTerminalInput(t, dir, test.args...)
			if strings.Contains(stdout, "Connected with driver") {
				t.Errorf("stdout has the banner:\n%s", stdout)
			}
			if !strings.HasPrefix(stdout, test.want) {
				t.Errorf("stdout starts with %q, want %q", firstLines(stdout, 3), test.want)
			}
			if strings.Contains(stderr, "Connected with driver") {
				t.Errorf("stderr has the banner: %q", stderr)
			}
		})
	}
}

// runWithTerminalInput runs usql in dir with a pseudo-terminal as its
// standard input and its controlling terminal, and buffers as its standard
// output and standard error. It returns what usql wrote to each.
func runWithTerminalInput(t *testing.T, dir string, args ...string) (string, string) {
	t.Helper()
	leader, follower, err := capture.OpenPTY()
	if err != nil {
		t.Fatal(err)
	}
	defer leader.Close()
	cmd := exec.Command(usqlPath, args...)
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = follower, stdout, stderr
	cmd.Dir = dir
	cmd.Env = []string{"HOME=" + dir, "PATH=" + os.Getenv("PATH"), "TERM=xterm-256color", "XDG_CONFIG_HOME=" + filepath.Join(dir, "xdg")}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	err = cmd.Run()
	_ = follower.Close()
	if err != nil {
		t.Fatalf("usql %s: %v\n%s", strings.Join(args, " "), err, stderr)
	}
	return stdout.String(), stderr.String()
}

// firstLines returns the first n lines of s.
func firstLines(s string, n int) string {
	lines := strings.SplitN(s, "\n", n+1)
	return strings.Join(lines[:min(n, len(lines))], "\n")
}
