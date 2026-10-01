// Package firebird defines and registers usql's Firebird driver.
//
// See: https://github.com/nakagami/firebirdsql
package firebird

import (
	_ "github.com/nakagami/firebirdsql" // DRIVER: firebirdsql
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("firebirdsql", drivers.Driver{})
}
