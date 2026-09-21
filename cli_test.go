package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/xo/usql/internal/capture"
)

// update rewrites the golden files instead of comparing against them.
//
// Read the diff before committing a rewrite. A golden here records what usql
// does, so rewriting one accepts whatever it does now as correct.
var update = flag.Bool("update", false, "rewrite the golden transcripts")

// usqlPath is the binary the sessions drive. TestMain builds it.
var usqlPath string

// buildTags are the tags the recorded binary is built with. Only the drivers
// the sessions use are needed, and a smaller set builds much faster than the
// full one.
const buildTags = "sqlite3 sqlite_app_armor sqlite_fts5 sqlite_introspect sqlite_json1 sqlite_math_functions sqlite_stat4 sqlite_vtable no_base"

func TestMain(m *testing.M) {
	flag.Parse()
	code, err := runMain(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

// runMain builds usql into a temporary directory and runs the tests.
func runMain(m *testing.M) (int, error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		// Nothing to build, because Record cannot run here. The tests
		// themselves report the skip, so that the reason is visible.
		return m.Run(), nil
	}
	dir, err := os.MkdirTemp("", "usql-cli-")
	if err != nil {
		return 0, fmt.Errorf("creating build directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	usqlPath = filepath.Join(dir, "usql")
	cmd := exec.Command("go", "build", "-tags", buildTags, "-o", usqlPath, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		return 0, fmt.Errorf("building usql: %w\n%s", err, out)
	}
	return m.Run(), nil
}

// scrubbers replace the parts of the output that change between runs. Without
// them a golden churns on every release and on every machine.
var scrubbers = []capture.Scrubber{
	// The banner carries the version of usql itself.
	{Match: regexp.MustCompile(`usql, the universal command-line interface for SQL databases, version [^\r\n]+`), With: `usql, the universal command-line interface for SQL databases, version VERSION`},
	{Match: regexp.MustCompile(`usql \d+\.\d+\.\d+[^\r\n]*`), With: `usql VERSION`},
	// The connection banner carries the SQLite version.
	{Match: regexp.MustCompile(`\([^)\r\n]*\d+\.\d+\.\d+[^)\r\n]*\)`), With: `(VERSION)`},
	// Query timings.
	{Match: regexp.MustCompile(`\d+\.\d+ ?ms`), With: `TIME`},
	// The temporary home directory appears in error messages and in \conninfo.
	{Match: regexp.MustCompile(`/tmp/usql-capture-[A-Za-z0-9]+`), With: `HOME`},
	{Match: regexp.MustCompile(`/var/folders/[^\s"']+`), With: `HOME`},
}

// sessions are the recorded terminal sessions.
//
// Every one of these uses SQLite, so they need no container and run wherever
// usql builds. A session that needed a server would make the whole file depend
// on one.
func sessions() []capture.Session {
	// memory is an in-memory SQLite database, so nothing is written to disk
	// and each session starts from the same empty state.
	const memory = "sqlite3://:memory:"
	// noPager keeps the output in the terminal. Without it usql hands long
	// output to a pager and the recording captures the pager instead.
	noPager := []string{"--pset=pager=off"}
	args := func(extra ...string) []string {
		return append(append([]string{}, noPager...), extra...)
	}
	return []capture.Session{
		{
			Name:  "banner-and-quit",
			About: "usql prints its banner and a prompt, and \\q leaves",
			Args:  args(memory),
			Steps: []capture.Step{
				{Send: `\q` + capture.KeyEnter},
			},
		},
		{
			Name:  "query",
			About: "a query prints a result table and a row count",
			Args:  args(memory),
			Steps: []capture.Step{
				{Send: `select 1 as n, 'two' as s;` + capture.KeyEnter},
				{Send: `\q` + capture.KeyEnter},
			},
		},
		{
			Name:  "multiline-continuation",
			About: "an unterminated statement prompts for more input",
			Args:  args(memory),
			Steps: []capture.Step{
				{Send: `select` + capture.KeyEnter},
				{Send: `  1 as n` + capture.KeyEnter},
				{Send: `;` + capture.KeyEnter},
				{Send: `\q` + capture.KeyEnter},
			},
		},
		{
			Name:  "variables",
			About: `\set defines a variable, :name interpolates it, \unset removes it`,
			Args:  args(memory),
			Steps: []capture.Step{
				{Send: `\set n 42` + capture.KeyEnter},
				{Send: `\set` + capture.KeyEnter},
				{Send: `select :n as answer;` + capture.KeyEnter},
				{Send: `\unset n` + capture.KeyEnter},
				{Send: `\set` + capture.KeyEnter},
				{Send: `\q` + capture.KeyEnter},
			},
		},
		{
			Name:  "quoted-interpolation",
			About: `:'name' quotes as a literal and :"name" quotes as an identifier`,
			Args:  args(memory),
			Steps: []capture.Step{
				{Send: `create table t (a int);` + capture.KeyEnter},
				{Send: `\set tbl t` + capture.KeyEnter},
				{Send: `\set word hello` + capture.KeyEnter},
				{Send: `select :'word' as literal;` + capture.KeyEnter},
				{Send: `select count(*) as n from :"tbl";` + capture.KeyEnter},
				{Send: `\q` + capture.KeyEnter},
			},
		},
		{
			Name:  "error-recovery",
			About: "a failed statement reports an error and the prompt returns",
			Args:  args(memory),
			Steps: []capture.Step{
				{Send: `select * from nonexistent;` + capture.KeyEnter},
				{Send: `select 1 as still_works;` + capture.KeyEnter},
				{Send: `\q` + capture.KeyEnter},
			},
		},
		{
			Name:  "describe",
			About: `\d lists relations and \d name describes one`,
			Args:  args(memory),
			Steps: []capture.Step{
				{Send: `create table films (id int primary key, title text);` + capture.KeyEnter},
				{Send: `\dt` + capture.KeyEnter},
				{Send: `\d films` + capture.KeyEnter},
				{Send: `\q` + capture.KeyEnter},
			},
		},
		{
			Name:  "line-editing",
			About: "the line editor handles backspace and start of line",
			Args:  args(memory),
			Steps: []capture.Step{
				{Send: `select 11 as nn`},
				{Send: capture.KeyBackspace + capture.KeyBackspace},
				{Send: `x`},
				{Send: capture.CtrlA},
				{Send: capture.KeyEnter},
				{Send: capture.CtrlU},
				{Send: `select 2 as n;` + capture.KeyEnter},
				{Send: `\q` + capture.KeyEnter},
			},
		},
		{
			Name:  "no-connection",
			About: "usql starts without a database and reports the lack of one",
			Args:  args(),
			Steps: []capture.Step{
				{Send: `select 1;` + capture.KeyEnter},
				{Send: `\q` + capture.KeyEnter},
			},
		},
	}
}

func TestCLI(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skipf("recording a terminal session needs a pseudo-terminal, which %s does not provide", runtime.GOOS)
	}
	for _, s := range sessions() {
		t.Run(s.Name, func(t *testing.T) {
			tr, err := capture.Record(context.Background(), usqlPath, s)
			if err != nil {
				t.Fatalf("recording %s: %v", s.Name, err)
			}
			tr.Apply(scrubbers)
			got := tr.Encode()
			golden := filepath.Join("testdata", "cli", s.Name+".txt")
			if *update {
				if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
					t.Fatalf("creating %s: %v", filepath.Dir(golden), err)
				}
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatalf("writing %s: %v", golden, err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			switch {
			case errors.Is(err, os.ErrNotExist):
				t.Fatalf("%s does not exist. Record it with: go test -run TestCLI -update .", golden)
			case err != nil:
				t.Fatalf("reading %s: %v", golden, err)
			}
			if got != string(want) {
				t.Errorf("%s does not match the recording.\nRe-record with: go test -run TestCLI -update .\n%s", golden, firstDifference(string(want), got))
			}
		})
	}
}

// firstDifference describes where two transcripts stop matching, because a
// whole transcript is too long to read in a failure message.
func firstDifference(want, got string) string {
	wantLines, gotLines := splitLines(want), splitLines(got)
	for i := 0; i < len(wantLines) && i < len(gotLines); i++ {
		if wantLines[i] != gotLines[i] {
			return fmt.Sprintf("line %d:\n  want: %q\n  got:  %q", i+1, wantLines[i], gotLines[i])
		}
	}
	return fmt.Sprintf("the recordings agree for %d lines, then the lengths differ: want %d lines, got %d",
		min(len(wantLines), len(gotLines)), len(wantLines), len(gotLines))
}

// splitLines splits s on newlines without dropping a trailing empty line.
func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

// TestCLIRecordsSomething checks that the harness itself works, so that an
// empty recording cannot pass every golden comparison quietly.
func TestCLIRecordsSomething(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skipf("recording a terminal session needs a pseudo-terminal, which %s does not provide", runtime.GOOS)
	}
	tr, err := capture.Record(context.Background(), usqlPath, capture.Session{
		Name:  "smoke",
		Args:  []string{"--pset=pager=off", "sqlite3://:memory:"},
		Steps: []capture.Step{{Send: `\q` + capture.KeyEnter, Wait: 2 * time.Second}},
	})
	if err != nil {
		t.Fatalf("recording: %v", err)
	}
	if len(tr.Bytes()) == 0 {
		t.Fatal("the recording is empty, so every golden comparison would be meaningless")
	}
}
