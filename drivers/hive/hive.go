// Package hive defines and registers usql's Apache Hive driver.
//
// See: https://github.com/sql-machine-learning/gohive
package hive

import (
	"github.com/xo/usql/drivers"
	_ "sqlflow.org/gohive" // DRIVER
)

func init() {
	drivers.Register("hive", drivers.Driver{})
}
