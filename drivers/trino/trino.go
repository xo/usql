// Package trino defines and registers usql's Trino driver.
//
// See: https://github.com/trinodb/trino-go-client
package trino

import (
	_ "github.com/trinodb/trino-go-client/trino" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("trino", drivers.Driver{
		Copy: drivers.CopyWithInsert(func(int) string { return "?" }),
	})
}
