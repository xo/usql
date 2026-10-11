// Package clickhouse defines and registers usql's ClickHouse driver.
//
// The driver is dbimp's, which speaks HTTP only, on port 8123 or, with
// tls=true, 8443. It replaced github.com/ClickHouse/clickhouse-go in W42, at
// Ken's decision, and takes the native protocol and its port 9000 with it.
// It refuses every key but tls, such as the secure=true of clickhouse-go.
//
// ClickHouse has no transactions over HTTP, so \copy inserts each row
// without one.
//
// See: https://github.com/xo/dbimp
// Group: base
package clickhouse

import (
	_ "github.com/xo/dbimp/clickhouse" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("clickhouse", drivers.Driver{})
}
