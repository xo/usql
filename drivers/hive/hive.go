package hive

import (
	// DRIVER: hive
	_ "sqlflow.org/gohive"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("hive", drivers.Driver{})
}
