package handler

import (
	"bytes"
	"io"
	"os/user"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/xo/usql/rline"
)

// TestRunSingleLineMetaCommand verifies that -c dispatches a standalone meta command.
func TestRunSingleLineMetaCommand(t *testing.T) {
	// stdout captures the user-visible result of the meta command.
	stdout := new(bytes.Buffer)
	// line supplies non-interactive IO without reading from the terminal.
	line := &rline.Rline{Out: stdout, Err: io.Discard}
	// commandHandler runs the same single-line path used by the -c flag.
	commandHandler := New(line, &user.User{}, ".", memfs.New(), true)
	// Enable the non-interactive execution mode selected by the CLI's -c flag.
	commandHandler.SetSingleLineMode(true)
	// Seed the command buffer with the standalone meta command being reproduced.
	commandHandler.Reset([]rune(`\echo hello`))

	// Running the handler must execute the meta command before reaching EOF.
	if err := commandHandler.Run(); err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	// The echo result proves that single-line mode did not mask command dispatch.
	if got, want := stdout.String(), "hello\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}
