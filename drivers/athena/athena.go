// Package athena defines and registers usql's Amazon Athena driver.
//
// Pre-flight, checked on 2026-10-11:
//
//   - Maintained: github.com/xo/dbimp added the driver in place of github.com/uber/athenadriver.
//     The repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: not by usql, which has no AWS account. dbimp measured it on a hosted account.
//
// The URL is athena://key:secret@host/database?workgroup=wg&output=s3://bucket/results/.
// The server refuses a transaction, so \copy inserts each row without one.
//
// Metadata: dbmeta's Athena model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package athena

import (
	_ "github.com/xo/dbimp/athena" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("athena", drivers.Driver{})
}
