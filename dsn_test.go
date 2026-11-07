package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestDSNFromEnv runs usql with USQL_DSN set. The variable opens a connection
// when the command line names none, a DSN on the command line wins, and a
// shell that usql starts does not see the variable.
func TestDSNFromEnv(t *testing.T) {
	if usqlPath == "" {
		t.Skipf("TestMain builds usql only on linux and darwin, not on %s", runtime.GOOS)
	}
	dir := t.TempDir()
	envDB, argDB := filepath.Join(dir, "env.db"), filepath.Join(dir, "arg.db")
	tests := []struct {
		name    string
		dsn     string
		args    []string
		want    string
		notWant string
	}{
		{"env opens", "sq:" + envDB, []string{"-c", `\conninfo`}, "env.db", "arg.db"},
		{"argument wins", "sq:" + envDB, []string{"-c", `\conninfo`, "sq:" + argDB}, "arg.db", "env.db"},
		{"blank is unset", "  ", []string{"-c", `\conninfo`}, "not connected", "env.db"},
		{"shell does not see it", "sq:" + envDB, []string{"-c", `\! env`}, "HOME=", "USQL_DSN"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := exec.Command(usqlPath, append([]string{"-X"}, test.args...)...)
			cmd.Env = append(os.Environ(), "HOME="+dir, "USQL_DSN="+test.dsn)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("running usql: %v\n%s", err, out)
			}
			if s := string(out); !strings.Contains(s, test.want) || strings.Contains(s, test.notWant) {
				t.Errorf("want %q and not %q in the output, got:\n%s", test.want, test.notWant, s)
			}
		})
	}
}
