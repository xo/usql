package yql

import (
	// DRIVER: yql
	_ "github.com/mattn/go-yql"

	"github.com/knq/usql/drivers"
)

func init() {
	drivers.Register("yql", drivers.Driver{})
}
