// Package odbc defines and registers usql's ODBC driver, which uses the ODBC
// driver manager of the system.
//
// The driver is github.com/xo/odbc, pure Go on purego, so it needs no cgo. It
// loads the driver manager when a connection opens, and reports an error when
// there is none.
//
// The URL is odbc+<driver>://user:pass@host:port/db, and a setting of the ODBC
// driver goes in the query.
//
// Metadata: dbmeta has no model for ODBC, which is a bridge to any database and
// has no dialect of its own, so the describe commands report that they are not
// supported (D3).
//
// See: https://github.com/xo/odbc
// Group: most
package odbc

import (
	"errors"
	"strings"

	"github.com/xo/odbc" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("odbc", drivers.Driver{
		// the standard class 28 of SQLSTATE is a failed authorization, in
		// every database
		IsPasswordErr: func(err error) bool {
			var e *odbc.Error
			if !errors.As(err, &e) {
				return false
			}
			if strings.HasPrefix(e.State, "28") {
				return true
			}
			for _, n := range e.Next {
				if strings.HasPrefix(n.State, "28") {
					return true
				}
			}
			return false
		},
	})
}
