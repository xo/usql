// Package spanner defines and registers usql's Google Cloud Spanner driver.
//
// Pre-flight, checked on 2026-10-11:
//
//   - Maintained: github.com/xo/dbimp, in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: through usql on 2026-10-11, on the hosted service or its emulator.
//
// The emulator accepts a wrong token, so a wrong password is not detected there.
//
// Metadata: dbmeta's Spanner model. usql writes no reader of its own (D3).
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
