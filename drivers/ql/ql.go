package ql

import (
	"database/sql"

	// DRIVER: ql
	"github.com/cznic/ql"

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
		RowsAffected: func(res sql.Result) (int64, error) {
			return 0, nil
		},
	})
}
