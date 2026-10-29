// Package tablestore defines and registers usql's Alibaba Tablestore driver.
//
// Pre-flight, checked on 2026-10-30:
//
//   - Maintained: github.com/xo/dbimp v0.21.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the API of Tablestore over HTTPS.
//   - Tested: Tablestore is a hosted service with no container and no
//     emulator, so dbrun cannot start it. The dbimp tests replay recordings
//     of a real instance.
//
// The driver has no parameters and no transactions. It sends one statement in
// each request, and holds the whole result in memory. A DATETIME is a
// dbimp.LocalDateTime in the zone Asia/Shanghai. See docs/TABLESTORE.md in
// dbimp.
//
// See: https://github.com/xo/dbimp
// Group: most
package tablestore

import (
	_ "github.com/xo/dbimp/tablestore" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("tablestore", drivers.Driver{})
}
