package drivers

import (
	"errors"
	"fmt"
	"testing"
)

// verboseErr is an error of a driver library that writes its fields.
type verboseErr struct{ code int }

func (e *verboseErr) Error() string { return "failed" }

func (e *verboseErr) ErrVerbose(query string) string {
	return fmt.Sprintf("ERROR %d: failed in %s", e.code, query)
}

// TestErrVerbose checks that an error of a driver with no package in usql
// writes its own fields through VerboseError, wrapped or not, and that any
// other error is written as usql prints it.
func TestErrVerbose(t *testing.T) {
	tests := []struct {
		err error
		exp string
	}{
		{WrapErr("oracle", &verboseErr{942}), "ERROR 942: failed in select 1"},
		{fmt.Errorf("running: %w", &verboseErr{1}), "ERROR 1: failed in select 1"},
		{errors.New("plain"), "plain"},
	}
	for _, test := range tests {
		if got := ErrVerbose(test.err, "select 1"); got != test.exp {
			t.Errorf("ErrVerbose(%v) = %q, want %q", test.err, got, test.exp)
		}
	}
}
