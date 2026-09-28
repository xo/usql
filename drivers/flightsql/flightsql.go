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
	drivers.Register("flightsql", Driver())
}

// Driver returns the FlightSQL driver. dburl opens GizmoSQL through the
// flightsql driver, and usql registers this driver under that name too.
func Driver() drivers.Driver {
	return drivers.Driver{}
}
