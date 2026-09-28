// Package influxql defines and registers usql's InfluxQL driver, which speaks
// InfluxQL to InfluxDB 1, 2, 3 and later.
//
// dburl opens an influxql:// URL through the influxdb driver, with
// sqlmode=disable, so this package registers the influxdb driver under the
// name influxql.
//
// See: https://github.com/xo/dbimp
// Group: most
package influxql

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/influxdb"
)

func init() {
	drivers.Register("influxql", influxdb.Driver())
}
