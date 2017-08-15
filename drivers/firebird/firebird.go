package firebird

import (
	// DRIVER: firebirdsql
	_ "github.com/nakagami/firebirdsql"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("firebirdsql", drivers.Driver{
		AMC: true,
	})
}
