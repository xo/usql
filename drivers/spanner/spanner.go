package spanner

import (
	// DRIVER: spanner
	_ "github.com/rakyll/go-sql-driver-spanner"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("spanner", drivers.Driver{})
}
