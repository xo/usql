package impala

import (
	_ "github.com/bippio/go-impala" // DRIVER: impala
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("impala", drivers.Driver{})
}
