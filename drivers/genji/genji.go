package genji

import (
	// DRIVER: genji
	_ "github.com/genjidb/genji/sql/driver"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("genji", drivers.Driver{})
}
