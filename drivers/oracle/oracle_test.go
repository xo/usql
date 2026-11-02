package oracle_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/xo/dburl"
	"github.com/xo/oracle"
	"github.com/xo/usql/drivers"
	_ "github.com/xo/usql/drivers/oracle"
)

// TestIsPasswordErr checks that usql prompts for the password on the errors
// that the driver reports as a refused credential, with no hook of usql's own.
func TestIsPasswordErr(t *testing.T) {
	t.Parallel()
	if _, ok := drivers.Available()["oracle"]; !ok {
		t.Fatal("oracle is not registered")
	}
	u, err := dburl.Parse("oracle://user:x@localhost/service")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		err  error
		exp  bool
	}{
		{"wrong user or password", &oracle.ServerError{ErrCode: 1017}, true},
		{"no password", &oracle.ServerError{ErrCode: 1005}, true},
		{"wrapped", fmt.Errorf("connecting: %w", &oracle.ServerError{ErrCode: 1017}), true},
		{"sentinel", oracle.ErrAuthentication, true},
		{"no privilege", &oracle.ServerError{ErrCode: 1031}, false},
		{"locked account", &oracle.ServerError{ErrCode: 28000}, false},
		{"expired password", &oracle.ServerError{ErrCode: 28001}, false},
		{"other", errors.New("boom"), false},
	}
	for _, test := range tests {
		if got := drivers.IsPasswordErr(u, test.err); got != test.exp {
			t.Errorf("%s: expected %t, got %t", test.name, test.exp, got)
		}
	}
}
