package ql

import (
	"github.com/xo/usql/drivers"
	"modernc.org/ql" // DRIVER: ql
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
