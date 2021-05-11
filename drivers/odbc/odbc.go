// Package odbc defines and registers usql's ODBC driver. Requires CGO. Uses
// respective platform's standard ODBC packages.
//
// See: https://github.com/alexbrainman/odbc
package odbc

import (
	"strings"

	"github.com/alexbrainman/odbc" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("odbc", drivers.Driver{
		LexerName: "tsql",
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*odbc.Error); ok {
				msg := strings.ToLower(e.Error())
				return strings.Contains(msg, "failed") &&
					(strings.Contains(msg, "login") ||
						strings.Contains(msg, "authentication") ||
						strings.Contains(msg, "password"))
			}
			return false
		},
	})
}
