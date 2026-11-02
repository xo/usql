// Package oracle defines and registers usql's Oracle Database driver, which is
// github.com/xo/oracle, pure Go.
//
// The driver reports a refused credential itself, through the method
// AuthenticationRefused of its errors (W66), so usql registers it with no
// hooks.
//
// See: https://github.com/xo/oracle
// Group: base
package oracle

import (
	_ "github.com/xo/oracle" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("oracle", drivers.Driver{})
}
