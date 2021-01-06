package spanner

import (
	_ "github.com/rakyll/go-sql-driver-spanner" // DRIVER: spanner
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("spanner", drivers.Driver{})
}
