package bigquery

import (
	// DRIVER: bigquery
	_ "gorm.io/driver/bigquery/driver"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("bigquery", drivers.Driver{})
}
