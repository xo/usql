package maxcompute

import (
	// DRIVER: maxcompute
	_ "sqlflow.org/gomaxcompute"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("maxcompute", drivers.Driver{})
}
