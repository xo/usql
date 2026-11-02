package drivers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/xo/dburl"
)

// refused is an error of a driver that has the method AuthenticationRefused.
type refused bool

func (e refused) Error() string { return fmt.Sprintf("refused %t", bool(e)) }

func (e refused) AuthenticationRefused() bool { return bool(e) }

// TestIsPasswordErrMethod checks that IsPasswordErr finds an error that says
// that the server refused the credential, anywhere in the tree of the error.
func TestIsPasswordErrMethod(t *testing.T) {
	t.Parallel()
	u := &dburl.URL{SchemeName: "none"}
	tests := []struct {
		name string
		err  error
		exp  bool
	}{
		{"refused", refused(true), true},
		{"not refused", refused(false), false},
		{"no method", errors.New("refused"), false},
		{"wrapped", fmt.Errorf("connecting: %w", refused(true)), true},
		{"in an Error of usql", &Error{Driver: "none", Err: fmt.Errorf("a: %w", refused(true))}, true},
		// errors.As would stop at the first error, which returns false
		{"after a false one", fmt.Errorf("%w: %w", refused(false), refused(true)), true},
		{"joined after a false one", errors.Join(refused(false), fmt.Errorf("b: %w", refused(true))), true},
		{"joined, none refused", errors.Join(refused(false), errors.New("c")), false},
		{"lost by %v", fmt.Errorf("d: %v", refused(true)), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsPasswordErr(u, test.err); got != test.exp {
				t.Errorf("expected %t, got: %t", test.exp, got)
			}
		})
	}
}
