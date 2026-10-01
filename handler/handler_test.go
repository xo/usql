package handler

import (
	"bytes"
	"io"
	"os/user"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/rline"
	"github.com/xo/usql/feed"
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
