// Package databend defines and registers usql's Databend driver.
//
// The driver is dbimp's, which replaced github.com/datafuselabs/databend-go
// in dbimp v0.6.0. It refuses the keys of databend-go, such as sslmode,
// tenant and warehouse.
//
// See: https://github.com/xo/dbimp
// Group: most
package databend

import (
	_ "github.com/xo/dbimp/databend" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("databend", drivers.Driver{
		// a statement such as CREATE TABLE sends no count of rows affected,
		// which the driver reports as ErrNotSupported (dbimp D125)
	})
}
