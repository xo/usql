package avatica_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/avatica"
	"github.com/xo/usql/drivers"
	_ "github.com/xo/usql/drivers/avatica"
)

func TestIsPasswordErr(t *testing.T) {
	t.Parallel()
	d, ok := drivers.Available()["avatica"]
	if !ok {
		t.Fatal("avatica is not registered")
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
		if got := d.IsPasswordErr(test.err); got != test.exp {
			t.Errorf("%s: expected %t, got %t", test.name, test.exp, got)
		}
	}
}
