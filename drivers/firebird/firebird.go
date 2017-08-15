package firebird

import (
	// DRIVER: firebirdsql
	_ "github.com/nakagami/firebirdsql"

	"github.com/knq/usql/drivers"
)

func init() {
	drivers.Register("firebirdsql", drivers.Driver{
		AMC: true,
	})
}
