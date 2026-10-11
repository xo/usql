// Package drill defines and registers usql's Apache Drill driver.
//
// Pre-flight, checked on 2026-10-07:
//
//   - Maintained: github.com/xo/dbimp v0.14.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the vendor's images, which dbmeta's dbrun starts as drill-1.21.2 and
//     drill-1.22.0.
//
// Metadata: dbmeta's Drill model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package drill

import (
	_ "github.com/xo/dbimp/drill" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("drill", drivers.Driver{})
}
