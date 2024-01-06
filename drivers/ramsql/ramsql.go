// Package ramsql defines and registers usql's RamSQL driver.
//
// See: https://github.com/proullon/ramsql
package ql

import (
	_ "github.com/proullon/ramsql/driver" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("ramsql", drivers.Driver{})
}
