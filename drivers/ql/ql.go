// Package ql defines and registers usql's Cznic QL driver.
//
// See: https://gitlab.com/cznic/ql
package ql

import (
	"github.com/xo/usql/drivers"
	"modernc.org/ql" // DRIVER
)

func init() {
	ql.RegisterDriver()
	// ql.RegisterMemDriver()
	drivers.Register("ql", drivers.Driver{
		AllowMultilineComments: true,
		AllowCComments:         true,
		BatchQueryPrefixes: map[string]string{
			"BEGIN TRANSACTION": "COMMIT",
		},
		BatchAsTransaction: true,
	})
}
