// Package neo4j defines and registers usql's Neo4j driver.
//
// Pre-flight, checked on 2026-09-29:
//
//   - Maintained: github.com/xo/dbimp v0.4.0 holds the driver, which v0.3.0
//     added on 2026-09-28. The repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the Query API of HTTP.
//   - Tested: the vendor's image, which dbmeta's dbrun starts as
//     neo4j-2026.09.0.
//
// Metadata: dbmeta's Neo4j model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package neo4j

import (
	_ "github.com/xo/dbimp/neo4j" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("neo4j", drivers.Driver{})
}
