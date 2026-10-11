// Package opensearch defines and registers usql's OpenSearch driver.
//
// Pre-flight, checked on 2026-10-07:
//
//   - Maintained: github.com/xo/dbimp v0.14.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the vendor's images, which dbmeta's dbrun starts as opensearch-2.19.6
//     and opensearch-3.9.0.
//
// OpenSearch has its own driver, apart from Elasticsearch (dbimp's D162).
//
// Metadata: dbmeta's OpenSearch model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package opensearch

import (
	_ "github.com/xo/dbimp/opensearch" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("opensearch", drivers.Driver{})
}
