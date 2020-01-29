package ql

import (
	// DRIVER: ql
	"modernc.org/ql"

	"github.com/xo/usql/drivers"
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
