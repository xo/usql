// Package voltdb defines and registers usql's VoltDB driver.
//
// See: https://github.com/VoltDB/voltdb-client-go
// Group: most
package voltdb

import (
	_ "github.com/VoltDB/voltdb-client-go/voltdbclient" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("voltdb", drivers.Driver{})
}
