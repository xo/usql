package drivers_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/xo/cassandra"
	"github.com/xo/dbimp"
	"github.com/xo/dbimp/avatica"
	"github.com/xo/dburl"
	"github.com/xo/odbc"
	"github.com/xo/oracle"
	"github.com/xo/usql/drivers"
)

// refusedTest is an error of a driver, and whether usql reads it as a refused
// credential.
type refusedTest struct {
	name string
	err  error
	exp  bool
}

// TestRefusedCredential checks that usql prompts for the password on the
// errors that a driver reports as a refused credential, with no hook of usql's
// own. These drivers have no package under drivers, and gen.go registers each
// with an empty drivers.Driver, so the method AuthenticationRefused of their
// errors is the only thing that decides.
func TestRefusedCredential(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url   string
		tests []refusedTest
	}{
		{"avatica://SA:x@localhost:8765", []refusedTest{
			{"unauthorized", &dbimp.StatusError{Code: http.StatusUnauthorized}, true},
			{"wrapped unauthorized", fmt.Errorf("opening: %w", &dbimp.StatusError{Code: http.StatusUnauthorized}), true},
			{"wrong password of HSQLDB", &avatica.Error{
				HTTPStatus: http.StatusInternalServerError,
				Code:       -1,
				Exception:  "java.sql.SQLInvalidAuthorizationSpecException: invalid authorization specification",
			}, true},
			{"syntax error", &avatica.Error{
				HTTPStatus: http.StatusInternalServerError,
				Code:       -1,
				Exception:  "java.sql.SQLSyntaxErrorException: unexpected token: SELEC",
			}, false},
			{"server error", &dbimp.StatusError{Code: http.StatusInternalServerError}, false},
			{"other", errors.New("boom"), false},
		}},
		{"cassandra://user:x@localhost/keyspace", []refusedTest{
			{"sentinel", cassandra.ErrAuthentication, true},
			{"wrapped", fmt.Errorf("connecting: %w", cassandra.ErrAuthentication), true},
			{"other error of the driver", cassandra.ErrConnectorClosed, false},
			{"other", errors.New("Provided username and/or password are incorrect"), false},
		}},
		{"odbc+postgres://user:x@localhost/db", []refusedTest{
			{"wrong password of MySQL", &odbc.Error{Op: "SQLDriverConnect", State: "28000", Native: 1045}, true},
			{"wrapped", fmt.Errorf("opening: %w", &odbc.Error{Op: "SQLDriverConnect", State: "28000"}), true},
			{"class 28 after a warning", &odbc.Error{Op: "SQLDriverConnect", State: "01000", Next: []odbc.Error{
				{Op: "SQLDriverConnect", State: "28000"},
			}}, true},
			{"no privilege", &odbc.Error{Op: "SQLExecute", State: "42501"}, false},
			{"connection refused", &odbc.Error{Op: "SQLDriverConnect", State: "08001", Message: "Connection refused"}, false},
			{"other", errors.New("boom"), false},
		}},
		{"oracle://user:x@localhost/service", []refusedTest{
			{"wrong user or password", &oracle.ServerError{ErrCode: 1017}, true},
			{"no password", &oracle.ServerError{ErrCode: 1005}, true},
			{"wrapped", fmt.Errorf("connecting: %w", &oracle.ServerError{ErrCode: 1017}), true},
			{"sentinel", oracle.ErrAuthentication, true},
			{"no privilege", &oracle.ServerError{ErrCode: 1031}, false},
			{"locked account", &oracle.ServerError{ErrCode: 28000}, false},
			{"expired password", &oracle.ServerError{ErrCode: 28001}, false},
			{"other", errors.New("boom"), false},
		}},
	}
	for _, test := range tests {
		u, err := dburl.Parse(test.url)
		if err != nil {
			t.Fatal(err)
		}
		for _, tt := range test.tests {
			if got := drivers.IsPasswordErr(u, tt.err); got != tt.exp {
				t.Errorf("%s: %s: expected %t, got %t", u.SchemeName, tt.name, tt.exp, got)
			}
		}
	}
}
