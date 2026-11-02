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
// The driver reports a refused credential itself, through the method
// AuthenticationRefused of its errors (W66), so usql registers it with no
// hooks.
//
// Metadata: dbmeta has no model for ODBC, which is a bridge to any database and
// has no dialect of its own, so the describe commands report that they are not
// supported (D3).
//
// See: https://github.com/xo/odbc
// Group: most
package odbc

import (
	_ "github.com/xo/odbc" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("odbc", drivers.Driver{})
}
