// Package presto defines and registers usql's Presto driver.
//
// See: https://github.com/prestodb/presto-go-client
package presto

import (
	_ "github.com/prestodb/presto-go-client/v2" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("presto", drivers.Driver{})
}
