package voltdb

import (
	// DRIVER: voltdb
	_ "github.com/VoltDB/voltdb-client-go/voltdbclient"

	"github.com/knq/usql/drivers"
)

func init() {
	drivers.Register("voltdb", drivers.Driver{
		AMC: true,
	})
}
