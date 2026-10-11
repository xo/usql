// Package presto defines and registers usql's Presto driver.
//
// Presto speaks the same protocol as Trino, so dburl opens a presto:// URL
// through dbimp's trino driver, which replaced
// github.com/prestodb/presto-go-client. This package registers it under the
// name presto, which dburl sets as the scheme. dburl writes flavor=presto,
// so the driver does not ask the server which product it is.
//
// See: https://github.com/xo/dbimp
// Group: base
package presto

import (
	_ "github.com/xo/dbimp/trino" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("presto", drivers.Driver{})
}
