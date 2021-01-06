package maxcompute

import (
	"github.com/xo/usql/drivers"
	_ "sqlflow.org/gomaxcompute" // DRIVER: maxcompute
)

func init() {
	drivers.Register("maxcompute", drivers.Driver{})
}
