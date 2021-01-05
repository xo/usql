package h2

import (
	// DRIVER: h2
	_ "github.com/jmrobles/h2go"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("h2", drivers.Driver{
		AllowDollar:            true,
		AllowMultilineComments: true,
		AllowCComments:         true,
	})
}
