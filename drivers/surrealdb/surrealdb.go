// Package surrealdb defines and registers usql's SurrealDB driver.
//
// Pre-flight, checked on 2026-09-27:
//
//   - Maintained: github.com/xo/dbimp v0.2.0 was tagged on 2026-09-27, and
//     its last commit is from the same day. The repository is in the same
//     family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//     Its dependencies set no logger, flag, signal, default HTTP client or
//     random seed.
//   - Links: pure Go. It adds github.com/cockroachdb/apd/v3, which the
//     Couchbase driver already brings.
//   - Tested: the vendor's docker.io/surrealdb/surrealdb image, which
//     dbmeta's dbrun starts as surrealdb-2.7.0, -3.1.6, -3.2.4 and -3.3.0.
//
// Metadata: dbmeta's SurrealDB model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package surrealdb

import (
	_ "github.com/xo/dbimp/surrealdb" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("surrealdb", drivers.Driver{})
}
