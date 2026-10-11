// Package arangodb defines and registers usql's ArangoDB driver.
//
// Pre-flight, checked on 2026-09-29:
//
//   - Maintained: github.com/xo/dbimp v0.5.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the cursor API of HTTP.
//   - Tested: the vendor's image, which dbmeta's dbrun starts as
//     arangodb-3.12.12.
//
// Metadata: dbmeta's ArangoDB model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package arangodb

import (
	_ "github.com/xo/dbimp/arangodb" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("arangodb", drivers.Driver{})
}
