// Package exasol defines and registers usql's Exasol driver, which is dbimp's. It
// is pure Go, over the WebSocket API of Exasol, and it reports its own errors,
// row counts and refused credentials, so usql adds nothing to it.
//
// Metadata: dbmeta's Exasol model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package exasol

import (
	_ "github.com/xo/dbimp/exasol" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("exasol", drivers.Driver{})
}
