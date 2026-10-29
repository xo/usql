// Package maxcompute defines and registers usql's Alibaba MaxCompute driver.
//
// Pre-flight, checked on 2026-10-30:
//
//   - Maintained: github.com/xo/dbimp v0.21.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the REST API of MaxCompute over HTTPS.
//   - Tested: MaxCompute is a hosted service with no container and no
//     emulator, so dbrun cannot start it. The dbimp tests replay recordings
//     of a real project.
//
// The driver has no parameters and no transactions. It writes each argument
// into the text of the statement. The server cuts a result at 10,000 rows
// and returns no error. See docs/MAXCOMPUTE.md in dbimp.
//
// See: https://github.com/xo/dbimp
// Group: most
package maxcompute

import (
	_ "github.com/xo/dbimp/maxcompute" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("maxcompute", drivers.Driver{})
}
