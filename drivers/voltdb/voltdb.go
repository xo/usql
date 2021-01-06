package voltdb

import (
	_ "github.com/VoltDB/voltdb-client-go/voltdbclient" // DRIVER: voltdb
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("voltdb", drivers.Driver{
		AllowMultilineComments: true,
	})
}
