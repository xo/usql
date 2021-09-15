// Package spanner defines and registers usql's Google Spanner driver.
//
// See: https://github.com/cloudspannerecosystem/go-sql-spanner
package spanner

import (
	_ "github.com/cloudspannerecosystem/go-sql-spanner" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("spanner", drivers.Driver{})
}
