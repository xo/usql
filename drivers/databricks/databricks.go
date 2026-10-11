// Package databricks defines and registers usql's Databricks driver.
//
// Pre-flight, checked on 2026-10-11:
//
//   - Maintained: github.com/xo/dbimp, in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: through usql on 2026-10-11, on the hosted service or its emulator.
//
// A wrong token is HTTP 403 with the message "Invalid access token", and not 401.
// The API has no session, so \copy inserts each row without a transaction.
//
// Metadata: dbmeta's Databricks model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package databricks

import (
	_ "github.com/xo/dbimp/databricks" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("databricks", drivers.Driver{})
}
