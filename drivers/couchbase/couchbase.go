// Package couchbase defines and registers usql's Couchbase driver.
//
// See: https://github.com/xo/dbimp
// Group: most
package couchbase

import (
	_ "github.com/xo/dbimp/couchbase" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("couchbase", drivers.Driver{})
}
