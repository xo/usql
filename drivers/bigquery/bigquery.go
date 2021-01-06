package bigquery

import (
	"github.com/xo/usql/drivers"
	_ "gorm.io/driver/bigquery/driver" // DRIVER: bigquery
)

func init() {
	drivers.Register("bigquery", drivers.Driver{})
}
