package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfigPrecedence sets the pager from each source in turn. A source wins
// over the ones before it: the default from PAGER, config.yaml, USQL_PAGER,
// the command line in the order given, init.sql, then the session.
func TestConfigPrecedence(t *testing.T) {
	if usqlPath == "" {
		t.Skip("TestMain builds usql on Linux and macOS only")
	}
	dir := t.TempDir()
	config := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(config, []byte("usql:\n  pager: config\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	init := filepath.Join(dir, "init.sql")
	if err := os.WriteFile(init, []byte("\\uset pager init\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	show := []string{"-c", `\config usql.pager`}
	for _, test := range []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{"default", []string{"PAGER=default"}, append([]string{"-X"}, show...), "default"},
		{"config.yaml", []string{"PAGER=default", "USQL_CONFIG=" + config}, append([]string{"-X"}, show...), "config"},
		{"USQL_PAGER", []string{"PAGER=default", "USQL_CONFIG=" + config, "USQL_PAGER=environ"}, append([]string{"-X"}, show...), "environ"},
		{"-O", []string{"USQL_CONFIG=" + config, "USQL_PAGER=environ"}, append([]string{"-X", "-O", "usql.pager=option"}, show...), "option"},
		{"-O then -U", []string{"USQL_CONFIG=" + config}, append([]string{"-X", "-O", "usql.pager=option", "-U", "pager=uset"}, show...), "uset"},
		{"-U then -O", []string{"USQL_CONFIG=" + config}, append([]string{"-X", "-U", "pager=uset", "-O", "usql.pager=option"}, show...), "option"},
		{"init.sql", []string{"USQL_CONFIG=" + config, "USQLRC=" + init}, append([]string{"-O", "usql.pager=option"}, show...), "init"},
		{"session", []string{"USQL_CONFIG=" + config, "USQLRC=" + init}, append([]string{"-c", `\uset pager session`}, show...), "session"},
	} {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, err := runUsql(t, dir, test.env, test.args...)
			if err != nil {
				t.Fatalf("%v\n%s", err, stderr)
			}
			// \uset writes what it set, and the last line is \config
			lines := strings.Split(strings.TrimSpace(stdout), "\n")
			if got, want := lines[len(lines)-1], `usql.pager "`+test.want+`"`; got != want {
				t.Errorf("stdout = %q, want %q", got, want)
			}
			if stderr != "" {
				t.Errorf("stderr = %q, want nothing", stderr)
			}
		})
	}
}

// TestConfigOption refuses an alias on the command line, reports the errors
// of config.yaml and goes on, and runs an alias of config.yaml.
func TestConfigOption(t *testing.T) {
	if usqlPath == "" {
		t.Skip("TestMain builds usql on Linux and macOS only")
	}
	dir := t.TempDir()
	config := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(config, []byte("print:\n  border: x2\naliases:\n  hello: '\\echo hello :W'\nquiet: on\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	environ := []string{"USQL_CONFIG=" + config}
	if _, stderr, err := runUsql(t, dir, environ, "-X", "-O", "aliases.a=b", "-c", `\q`); err == nil || !strings.Contains(stderr, "only the configuration file sets an alias") {
		t.Errorf("-O aliases.a=b: %v, stderr %q, want an error", err, stderr)
	}
	stdout, stderr, err := runUsql(t, dir, environ, "-X", "-v", "W=world", "-c", `\hello`)
	if err != nil {
		t.Fatalf("%v\n%s", err, stderr)
	}
	if stdout != "hello world\n" {
		t.Errorf("stdout = %q, want %q", stdout, "hello world\n")
	}
	if !strings.Contains(stderr, config+":5: quiet: move it to usql.quiet") {
		t.Errorf("stderr = %q, want the error of quiet", stderr)
	}
	if _, stderr, err := runUsql(t, dir, []string{"USQL_CONFIG=" + filepath.Join(dir, "none.yaml")}, "-X", "-c", `\q`); err == nil {
		t.Errorf("a missing USQL_CONFIG: no error, stderr %q", stderr)
	}
}

// runUsql runs usql in dir with no configuration, history or variable of the
// user, and the variables of environ.
func runUsql(t *testing.T, dir string, environ []string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(usqlPath, args...)
	cmd.Env = append([]string{"HOME=" + dir, "PATH=" + os.Getenv("PATH"), "XDG_CONFIG_HOME=" + filepath.Join(dir, "xdg")}, environ...)
	cmd.Dir = dir
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
