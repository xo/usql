package odbc_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/xo/dburl"
	"github.com/xo/odbc"
	"github.com/xo/usql/drivers"
	_ "github.com/xo/usql/drivers/odbc"
)

// TestIsPasswordErr checks that usql prompts for the password on the errors
// that the driver reports as a refused credential, with no hook of usql's own.
func TestIsPasswordErr(t *testing.T) {
	t.Parallel()
	if _, ok := drivers.Available()["odbc"]; !ok {
		t.Fatal("odbc is not registered")
	}
	u, err := dburl.Parse("odbc+postgres://user:x@localhost/db")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		err  error
		exp  bool
	}{
		{"wrong password of MySQL", &odbc.Error{Op: "SQLDriverConnect", State: "28000", Native: 1045}, true},
		{"wrapped", fmt.Errorf("opening: %w", &odbc.Error{Op: "SQLDriverConnect", State: "28000"}), true},
		{"class 28 after a warning", &odbc.Error{Op: "SQLDriverConnect", State: "01000", Next: []odbc.Error{
			{Op: "SQLDriverConnect", State: "28000"},
		}}, true},
		{"no privilege", &odbc.Error{Op: "SQLExecute", State: "42501"}, false},
		{"connection refused", &odbc.Error{Op: "SQLDriverConnect", State: "08001", Message: "Connection refused"}, false},
		{"other", errors.New("boom"), false},
	}
	for _, test := range tests {
		if got := drivers.IsPasswordErr(u, test.err); got != test.exp {
			t.Errorf("%s: expected %t, got %t", test.name, test.exp, got)
		}
	}
}
