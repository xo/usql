// Package bigquery defines and registers usql's Google BigQuery driver.
//
// See: https://github.com/go-gorm/gorm
package bigquery

import (
	"github.com/xo/usql/drivers"
	_ "gorm.io/driver/bigquery/driver" // DRIVER
)

func init() {
	drivers.Register("bigquery", drivers.Driver{})
}
