// Package trino defines and registers usql's Trino driver.
//
// The driver is dbimp's, which replaced github.com/trinodb/trino-go-client.
// It serves Presto too. The URL is trino://user:password@host:port/catalog/schema,
// and dburl writes the user "user" when the URL has none, because the
// servers need one.
//
// See: https://github.com/xo/dbimp
// Group: base
package trino

import (
	_ "github.com/xo/dbimp/trino" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("trino", drivers.Driver{})
}
