// Package spanner defines and registers usql's Google Spanner driver.
//
// See: https://github.com/rakyll/go-sql-driver-spanner
package spanner

import (
	_ "github.com/rakyll/go-sql-driver-spanner" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("spanner", drivers.Driver{})
}
