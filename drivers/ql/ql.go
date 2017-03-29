package ql

import (
	// DRIVER: ql
	_ "github.com/cznic/ql/driver"

	"github.com/knq/usql/drivers"
)

func init() {
	drivers.Register("ql", drivers.Driver{})
}
