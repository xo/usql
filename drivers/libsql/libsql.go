// Package libsql defines and registers usql's libSQL driver, for sqld and
// Turso Cloud.
//
// Pre-flight, checked on 2026-10-01:
//
//   - Maintained: github.com/xo/dbimp v0.9.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the Hrana protocol on HTTP.
//   - Tested: the vendor's image, which dbmeta's dbrun starts as
//     libsql-0.24.33.
//
// TLS is on unless the URL sets tls=false. The token is the password of the
// URL. A transaction ends when no statement runs on it for 10 seconds, and
// the server then rolls it back.
//
// Metadata: dbmeta's libSQL model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package libsql

import (
	"errors"
	"net/http"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/libsql" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("libsql", drivers.Driver{
		Err: func(err error) (string, string) {
			var e *libsql.Error
			if errors.As(err, &e) {
				return e.Code, e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "libsql: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
		Copy: drivers.CopyWithInsert(func(int) string { return "?" }),
	})
}
