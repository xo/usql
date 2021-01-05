package impala

import (
	// DRIVER: impala
	_ "github.com/bippio/go-impala"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("impala", drivers.Driver{})
}
