// Package hive defines and registers usql's Apache Hive driver.
//
// See: https://github.com/beltran/gohive
// Group: most
package hive

import (
	_ "github.com/beltran/gohive/v2" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("hive", drivers.Driver{})
}
