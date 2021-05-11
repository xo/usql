// Package impala defines and registers usql's Apache Impala driver.
//
// See: https://github.com/bippio/go-impala
package impala

import (
	_ "github.com/bippio/go-impala" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("impala", drivers.Driver{})
}
