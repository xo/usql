// Package clickhouse defines and registers usql's ClickHouse driver.
//
// The driver is dbimp's, which speaks HTTP only, on port 8123 or, with tls=true,
// 8443. It has no native protocol or port 9000, and it refuses every key of the
// URL but tls.
//
// ClickHouse has no transactions over HTTP, so \copy inserts each row without
// one.
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
