package hive

import (
	"github.com/xo/usql/drivers"
	_ "sqlflow.org/gohive" // DRIVER: hive
)

func init() {
	drivers.Register("hive", drivers.Driver{})
}
