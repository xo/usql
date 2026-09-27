// Package flightsql defines and registers usql's FlightSQL driver.
//
// See: https://github.com/apache/arrow-go/tree/main/arrow/flight/flightsql/driver
// Group: most
package flightsql

import (
	_ "github.com/apache/arrow-go/v18/arrow/flight/flightsql/driver" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("flightsql", drivers.Driver{})
}
