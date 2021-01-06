package h2

import (
	_ "github.com/jmrobles/h2go" // DRIVER: h2
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("h2", drivers.Driver{
		AllowDollar:            true,
		AllowMultilineComments: true,
		AllowCComments:         true,
	})
}
