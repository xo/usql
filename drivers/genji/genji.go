// Package genji defines and registers usql's Genji driver.
//
// See: https://github.com/genjidb/genji
package genji

import (
	_ "github.com/genjidb/genji/sql/driver" // DRIVER: genji
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("genji", drivers.Driver{})
}
