package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/xo/usql/internal/capture"
)

// ansiRE matches the escape sequences the line editor writes while redrawing.
// A recording is full of them and none of them are the completion.
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

// completionTags are the tags the recorded binary is built with. duckdb needs
// cgo and prebuilt archives, so this is slow, but it is the driver that found
// the bug this test exists for.
const completionTags = "duckdb no_base"

// TestCompletion checks that pressing tab offers the names of the tables in
// the database.
//
// Nothing else covers completion. It only runs when usql owns a terminal, so
// piping input into the binary exercises none of it, and the driver bug this
// test was written for reached a release because of that: duckdb was
// registered with MySQL's completer, so every completion ran a MySQL function
// against duckdb and failed with an error the user saw and no test did.
//
// This covers one driver. Completion is per driver, because each supplies its
// own metadata reader and some supply their own completer, so a pass here says
// nothing about the others. Widening it is a backlog item.
func TestCompletion(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skipf("recording a terminal session needs a pseudo-terminal, which %s does not provide", runtime.GOOS)
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "usql")
	build := exec.Command("go", "build", "-tags", completionTags, "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Skipf("cannot build usql with %q, so completion cannot be recorded: %v\n%s", completionTags, err, out)
	}

	db := filepath.Join(dir, "completion.duckdb")
	create := exec.Command(bin, "duckdb:"+db, "-c", "create table albums (id int, title varchar)")
	if out, err := create.CombinedOutput(); err != nil {
		t.Fatalf("creating the database: %v\n%s", err, out)
	}

	tr, err := capture.Record(context.Background(), bin, capture.Session{
		Name:  "completion",
		About: "tab offers the tables in the database",
		Args:  []string{"--pset=pager=off", "duckdb:" + db},
		Steps: []capture.Step{
			// The metadata query runs on the first tab, so allow for it.
			{Send: "select * from alb\t", Wait: 10 * time.Second},
			{Send: capture.CtrlU},
			{Send: `\q` + capture.KeyEnter},
		},
	})
	if err != nil {
		t.Fatalf("recording: %v", err)
	}
	got := ansiRE.ReplaceAllString(string(tr.Bytes()), "")

	// The completer logs a failed metadata query and carries on with no
	// candidates, so an empty completion and a broken one look the same on
	// screen. Check for the error as well as for the name.
	if strings.Contains(got, "Error getting selectables") {
		t.Errorf("the completer could not read the metadata:\n%s", lastLines(got, 4))
	}
	if !strings.Contains(got, "albums") {
		t.Errorf("tab did not offer the table name:\n%s", lastLines(got, 4))
	}
}

// lastLines returns the last n non-empty lines of s, for a failure message.
func lastLines(s string, n int) string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, "    "+line)
		}
	}
	if len(out) > n {
		out = out[len(out)-n:]
	}
	return strings.Join(out, "\n")
}
