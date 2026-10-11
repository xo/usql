// Package druid defines and registers usql's Apache Druid driver.
//
// Pre-flight, checked on 2026-10-08:
//
//   - Maintained: github.com/xo/dbimp v0.14.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the SQL API of the Router or the Broker, which is
//     HTTP.
//   - Tested: the vendor's images, which dbmeta's dbrun starts as druid-37.0.0
//     and druid-38.0.0.
//
// The driver runs queries and takes no write, because the SQL API of Druid
// refuses INSERT, REPLACE, UPDATE and DELETE. A write fails with the error of
// the server.
//
// Metadata: dbmeta's Druid model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package druid

import (
	_ "github.com/xo/dbimp/druid" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("druid", drivers.Driver{})
}
