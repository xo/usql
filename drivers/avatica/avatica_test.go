package avatica_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/avatica"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	_ "github.com/xo/usql/drivers/avatica"
)

func TestIsPasswordErr(t *testing.T) {
	t.Parallel()
	if _, ok := drivers.Available()["avatica"]; !ok {
		t.Fatal("avatica is not registered")
	}
	u, err := dburl.Parse("avatica://SA:x@localhost:8765")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		err  error
		exp  bool
	}{
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
	}
	for _, test := range tests {
		if got := drivers.IsPasswordErr(u, test.err); got != test.exp {
			t.Errorf("%s: expected %t, got %t", test.name, test.exp, got)
		}
	}
}
