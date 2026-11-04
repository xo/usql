// Package chai defines and registers usql's ChaiSQL driver.
//
// See: https://github.com/chaisql/chai
// Group: most
package chai

import (
	_ "github.com/chaisql/chai" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("chai", drivers.Driver{})
}
