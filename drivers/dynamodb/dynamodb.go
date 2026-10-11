// Package dynamodb defines and registers usql's Amazon DynamoDB driver.
//
// Pre-flight, checked on 2026-10-07:
//
//   - Maintained: github.com/xo/dbimp v0.14.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the vendor's images, which dbmeta's dbrun starts as dynamodb-3.2.0,
//     dynamodb-3.3.1, alternator-2025.1 and alternator-2026.3.
//
// The URL is dynamodb://key:secret@host:port?region=us-east-1. It needs the
// access key and the secret key, and a missing one is dburl.ErrMissingUser.
// The driver replaced github.com/btnguyen2k/godynamo, and its scheme is
// dynamodb in place of godynamo.
//
// Metadata: none. dbmeta has no model for DynamoDB yet, so the describe
// commands report that they are not supported (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package dynamodb

import (
	_ "github.com/xo/dbimp/dynamodb" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("dynamodb", drivers.Driver{})
}
