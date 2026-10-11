// Package spanner defines and registers usql's Google Cloud Spanner driver.
//
// Pre-flight, checked on 2026-10-11:
//
//   - Maintained: github.com/xo/dbimp added the driver in place of github.com/googleapis/go-sql-spanner.
//     The repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the emulator, which dbmeta's dbrun starts, by dbimp. usql has no Google account.
//
// The emulator accepts a wrong token, so a wrong password is not detected there.
//
// Metadata: none. dbmeta has no model for Spanner, so the describe commands report
// that they are not supported (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package spanner

import (
	_ "github.com/xo/dbimp/spanner" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("spanner", drivers.Driver{})
}
