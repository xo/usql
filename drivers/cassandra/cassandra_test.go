package cassandra_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/xo/cassandra"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	_ "github.com/xo/usql/drivers/cassandra"
)

// TestIsPasswordErr checks that usql prompts for the password on the errors
// that the driver reports as a refused credential, with no hook of usql's own.
func TestIsPasswordErr(t *testing.T) {
	t.Parallel()
	if _, ok := drivers.Available()["cassandra"]; !ok {
		t.Fatal("cassandra is not registered")
	}
	u, err := dburl.Parse("cassandra://user:x@localhost/keyspace")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		err  error
		exp  bool
	}{
		{"sentinel", cassandra.ErrAuthentication, true},
		{"wrapped", fmt.Errorf("connecting: %w", cassandra.ErrAuthentication), true},
		{"other error of the driver", cassandra.ErrConnectorClosed, false},
		{"other", errors.New("Provided username and/or password are incorrect"), false},
	}
	for _, test := range tests {
		if got := drivers.IsPasswordErr(u, test.err); got != test.exp {
			t.Errorf("%s: expected %t, got %t", test.name, test.exp, got)
		}
	}
}
