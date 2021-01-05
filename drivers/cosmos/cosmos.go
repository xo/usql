package cosmos

import (
	// DRIVER: cosmos
	_ "github.com/btnguyen2k/gocosmos"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("cosmos", drivers.Driver{})
}
