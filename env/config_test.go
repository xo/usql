package env

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/xo/usql/text"
)

// TestLoadConfig reads a configuration file. Names keep their case, a value
// keeps the text that it is written with, and nothing is printed.
func TestLoadConfig(t *testing.T) {
	v := NewDefaultVars()
	name := writeConfig(t, `
usql:
  quiet: on
  fetch_count: 0010
print:
  border: 2
  null: '(null)'
  expanded: auto
vars:
  Foo: bar
  num: 0123
  flag: on
charts:
  path: graphs
env:
  MyVar: hello world
connections:
  pg: postgres://user:secret@localhost/db
  lst: [sqlite3, /tmp/x.db]
  mp:
    proto: mysql
    user: u
    pass: p
    host: localhost
aliases:
  Hi: '\echo hi'
  multi: |
    \set X 1
    \echo :X
`)
	var errs []error
	out := captureOutput(t, func() {
		errs = v.LoadConfig(name, nil)
	})
	if len(errs) != 0 {
		t.Fatalf("errors: %v", errs)
	}
	if out != "" {
		t.Errorf("LoadConfig printed %q, want nothing", out)
	}
	for _, test := range []struct {
		section, name, want string
	}{
		{SectionUsql, "quiet", "on"},
		{SectionUsql, "fetch_count", "10"},
		{SectionPrint, "border", "2"},
		{SectionPrint, "null", `"(null)"`},
		{SectionPrint, "expanded", "auto"},
		{SectionVars, "Foo", "'bar'"},
		{SectionVars, "num", "'0123'"},
		{SectionVars, "flag", "'on'"},
		{SectionCharts, "path", `"graphs"`},
		{SectionEnv, "MyVar", "'hello world'"},
		{SectionConnections, "pg", "'postgres://user:xxxxx@localhost/db'"},
		{SectionConnections, "lst", "'sqlite3 /tmp/x.db'"},
		{SectionConnections, "mp", "'mysql://u:xxxxx@localhost'"},
		{SectionAliases, "Hi", `'\\echo hi'`},
		{SectionAliases, "multi", `'\\set X 1\n\\echo :X\n'`},
	} {
		if got, ok := v.Config(test.section, test.name, false); !ok || got != test.want {
			t.Errorf("%s.%s = %q, %t, want %q", test.section, test.name, got, ok, test.want)
		}
	}
	if got, _ := v.Config(SectionConnections, "pg", true); got != "'postgres://user:secret@localhost/db'" {
		t.Errorf("connections.pg with the password = %q", got)
	}
	if _, ok := v.Config(SectionVars, "foo", false); ok {
		t.Errorf("vars.foo is set, want only vars.Foo")
	}
}

// TestLoadConfigErrors reports each setting that cannot be set, with the
// file, the line and the full name, and sets the others.
func TestLoadConfigErrors(t *testing.T) {
	v := NewDefaultVars()
	name := writeConfig(t, `print:
  border: 2
  format: nope
  nope: 1
usql:
  pager: [a, b]
aliases:
  d: '\echo d'
  ok: '\echo ok'
quiet: true
charts_path: charts
init: select 1
bogus: 1
`)
	errs := v.LoadConfig(name, func(s string) bool { return s == "d" })
	var got []string
	for _, err := range errs {
		var e *ConfigError
		if !errors.As(err, &e) || e.File != name {
			t.Errorf("error %v does not name the file", err)
			continue
		}
		got = append(got, strings.TrimPrefix(err.Error(), name))
	}
	want := []string{
		":3: print.format: " + text.ErrInvalidFormatType.Error(),
		":4: print.nope: unknown option: nope",
		":6: usql.pager: " + text.ErrConfigNotAValue.Error(),
		":8: aliases.d: \\d is a meta command of usql, so an alias cannot take the name",
		":10: quiet: move it to usql.quiet",
		":11: charts_path: move it to charts.path",
		":12: init: ",
		":13: bogus: unknown configuration section: bogus",
	}
	if len(got) != len(want) {
		t.Fatalf("errors:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	for i := range want {
		if !strings.HasPrefix(got[i], want[i]) {
			t.Errorf("error %d = %q, want %q", i, got[i], want[i])
		}
	}
	// the other settings apply
	if got, _ := v.Config(SectionPrint, "border", false); got != "2" {
		t.Errorf("print.border = %q, want 2", got)
	}
	if _, ok := v.Alias("ok"); !ok {
		t.Errorf("alias ok is not set")
	}
	if _, ok := v.Alias("d"); ok {
		t.Errorf("alias d is set, but \\d is a command")
	}
	if got, _ := v.Setting("quiet"); got != "off" {
		t.Errorf("quiet = %q, want off: a key at the top of the file sets nothing", got)
	}
}

// TestSetOption applies the options of the command line.
func TestSetOption(t *testing.T) {
	v := NewDefaultVars()
	for _, opt := range []string{
		"print.border=3",
		`print.fieldsep="\t"`,
		"print.footer",
		"usql.quiet",
		"vars.A=1",
		"vars.B=2",
		"vars.B",
		"env.E=x",
		"connections.c=pg://u:p@h/d",
		"charts.path=c",
	} {
		if err := v.SetOption(opt); err != nil {
			t.Fatalf("%s: %v", opt, err)
		}
	}
	for _, test := range []struct {
		section, name, want string
	}{
		{SectionPrint, "border", "3"},
		{SectionPrint, "fieldsep", `"\t"`},
		{SectionPrint, "footer", "off"},
		{SectionUsql, "quiet", "on"},
		{SectionVars, "A", "'1'"},
		{SectionEnv, "E", "'x'"},
		{SectionConnections, "c", "'pg://u:xxxxx@h/d'"},
		{SectionCharts, "path", `"c"`},
	} {
		if got, ok := v.Config(test.section, test.name, false); !ok || got != test.want {
			t.Errorf("%s.%s = %q, %t, want %q", test.section, test.name, got, ok, test.want)
		}
	}
	if _, ok := v.Get("B"); ok {
		t.Errorf("vars.B is set, want it unset")
	}
	for opt, want := range map[string]error{
		"aliases.a=b": text.ErrAliasConfigOnly,
		"border=1":    nil,
		"nope.a=b":    nil,
		"print=1":     nil,
	} {
		err := v.SetOption(opt)
		switch {
		case err == nil:
			t.Errorf("%s: no error", opt)
		case want != nil && !errors.Is(err, want):
			t.Errorf("%s: error %v, want %v", opt, err, want)
		}
	}
}

// TestEnviron gives a program that usql starts the settings of env, and does
// not change the environment of usql.
func TestEnviron(t *testing.T) {
	t.Setenv("W80_KEEP", "keep")
	t.Setenv("W80_DROP", "drop")
	v := NewDefaultVars()
	if err := v.SetEnv("W80_NEW", "new"); err != nil {
		t.Fatal(err)
	}
	if err := v.SetEnv("W80_KEEP", "changed"); err != nil {
		t.Fatal(err)
	}
	if err := v.UnsetEnv("W80_DROP"); err != nil {
		t.Fatal(err)
	}
	environ := v.Environ()
	for _, s := range []string{"W80_NEW=new", "W80_KEEP=changed"} {
		if !slices.Contains(environ, s) {
			t.Errorf("the environment has no %s", s)
		}
	}
	for _, s := range []string{"W80_KEEP=keep", "W80_DROP=drop"} {
		if slices.Contains(environ, s) {
			t.Errorf("the environment has %s", s)
		}
	}
	if s := os.Getenv("W80_KEEP"); s != "keep" {
		t.Errorf("the environment of usql changed: W80_KEEP=%q", s)
	}
	if s, ok := v.Getenv("W80_DROP"); ok {
		t.Errorf("Getenv(W80_DROP) = %q, want it unset", s)
	}
	if s, _ := v.Getenv("W80_NEW"); s != "new" {
		t.Errorf("Getenv(W80_NEW) = %q, want new", s)
	}
}

// TestDumpConfig writes the sections usql and print, unless the sections are
// named.
func TestDumpConfig(t *testing.T) {
	v := NewVars()
	v.uset["quiet"] = "off"
	v.prnt["border"] = "1"
	_ = v.Set("x", "1")
	_ = v.SetEnv("E", "e")
	_ = v.SetConn("c", "pg://u:p@h/d")
	_ = v.SetAlias("a", `\echo a`)
	var b bytes.Buffer
	if err := v.DumpConfig(&b, false); err != nil {
		t.Fatal(err)
	}
	if want := "usql.quiet   off\nprint.border 1\n"; b.String() != want {
		t.Errorf("DumpConfig = %q, want %q", b.String(), want)
	}
	b.Reset()
	if err := v.DumpConfig(&b, false, Sections...); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"vars.x        '1'", "env.E         'e'", `charts.path   "charts"`, "connections.c 'pg://u:xxxxx@h/d'", `aliases.a     '\\echo a'`} {
		if !strings.Contains(b.String(), s+"\n") {
			t.Errorf("DumpConfig of every section has no line %q:\n%s", s, b.String())
		}
	}
}

// writeConfig writes a configuration file, and returns its name.
func writeConfig(t *testing.T, s string) string {
	t.Helper()
	name := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(name, []byte(s), 0o600); err != nil {
		t.Fatal(err)
	}
	return name
}

// captureOutput returns what f writes to standard output and standard error.
func captureOutput(t *testing.T, f func()) string {
	t.Helper()
	tmp, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	stdout, stderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = tmp, tmp
	defer func() { os.Stdout, os.Stderr = stdout, stderr }()
	f()
	// Windows cannot remove a file that is open
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}
	buf, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(buf)
}
