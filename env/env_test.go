package env

import (
	"os"
	"os/user"
	"runtime"
	"testing"
)

// TestParseVariableBool checks that a value reads as a boolean as psql's
// ParseVariableBool reads it. Each case was run as \if in psql 18.6.
func TestParseVariableBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s     string
		v, ok bool
	}{
		{"true", true, true},
		{"TR", true, true},
		{"t", true, true},
		{"yes", true, true},
		{"Y", true, true},
		{"on", true, true},
		{"ON", true, true},
		{"1", true, true},
		{"false", false, true},
		{"f", false, true},
		{"no", false, true},
		{"N", false, true},
		{"of", false, true},
		{"Of", false, true},
		{"off", false, true},
		{"0", false, true},
		{"o", false, false},
		{"", false, false},
		{"truex", false, false},
		{"onn", false, false},
		{"true false", false, false},
		{"01", false, false},
		{" true", false, false},
		{"yeſ", false, false},
		{"bad", false, false},
	}
	for _, test := range tests {
		v, ok := ParseVariableBool(test.s)
		if v != test.v || ok != test.ok {
			t.Errorf("%q: expected %t, %t, got: %t, %t", test.s, test.v, test.ok, v, ok)
		}
	}
}

// TestUnquote checks that a single-quoted argument decodes as psql 18 decodes
// the argument of a meta command. Each expected value was read from psql.
func TestUnquote(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s, exp string
	}{
		{`'a\tb'`, "a\tb"},
		{`'a''b'`, "a'b"},
		{`'\q'`, "q"},
		{`'\61'`, "1"},
		{`'\101'`, "A"},
		{`'\x41'`, "A"},
		{`'\x4'`, "\x04"},
		{`'\xff'`, "\xff"},
		{"'\\" + "u0041'", "u0041"},
		{`'\\'`, `\`},
		{`'\''`, "'"},
		{`''`, ""},
		{`"a b"`, "a b"},
	}
	for _, test := range tests {
		got, err := Unquote(test.s)
		if err != nil {
			t.Errorf("Unquote(%s): %v", test.s, err)
			continue
		}
		if got != test.exp {
			t.Errorf("Unquote(%s): expected %q, got %q", test.s, test.exp, got)
		}
	}
	for _, s := range []string{`'a`, `'a'b'`} {
		if _, err := Unquote(s); err == nil {
			t.Errorf("Unquote(%s): expected an error", s)
		}
	}
}

// TestStatePaths checks where the history and init.sql are on Linux and
// other Unix systems: in the directories of the XDG Base Directory
// specification, or in the files that USQL_HISTORY and USQLRC name.
func TestStatePaths(t *testing.T) {
	switch runtime.GOOS {
	case "windows", "darwin", "ios", "plan9":
		t.Skip("the XDG directories apply to Linux and other Unix systems")
	}
	u := &user.User{HomeDir: "/home/u"}
	t.Setenv("HOME", "/home/u")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("USQL_HISTORY", "")
	t.Setenv("USQLRC", "")
	os.Unsetenv("USQL_HISTORY")
	os.Unsetenv("USQLRC")
	if got, exp := HistoryFile(u), "/home/u/.local/state/usql/history"; got != exp {
		t.Errorf("history: got %q, want %q", got, exp)
	}
	if got, exp := InitFile(u), "/home/u/.config/usql/init.sql"; got != exp {
		t.Errorf("init.sql: got %q, want %q", got, exp)
	}
	// a relative XDG_STATE_HOME is ignored, as the specification says
	t.Setenv("XDG_STATE_HOME", "state")
	if got, exp := HistoryFile(u), "/home/u/.local/state/usql/history"; got != exp {
		t.Errorf("history with a relative XDG_STATE_HOME: got %q, want %q", got, exp)
	}
	t.Setenv("XDG_STATE_HOME", "/x/state")
	t.Setenv("XDG_CONFIG_HOME", "/x/config")
	if got, exp := HistoryFile(u), "/x/state/usql/history"; got != exp {
		t.Errorf("history with XDG_STATE_HOME: got %q, want %q", got, exp)
	}
	if got, exp := InitFile(u), "/x/config/usql/init.sql"; got != exp {
		t.Errorf("init.sql with XDG_CONFIG_HOME: got %q, want %q", got, exp)
	}
	t.Setenv("USQL_HISTORY", "~/h")
	t.Setenv("USQLRC", "/r.sql")
	if got, exp := HistoryFile(u), "/home/u/h"; got != exp {
		t.Errorf("history with USQL_HISTORY: got %q, want %q", got, exp)
	}
	if got, exp := InitFile(u), "/r.sql"; got != exp {
		t.Errorf("init.sql with USQLRC: got %q, want %q", got, exp)
	}
}
