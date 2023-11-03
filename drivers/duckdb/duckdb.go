// Package duckdb defines and registers usql's DuckDB driver. Requires CGO.
//
// See: https://github.com/marcboeker/go-duckdb
package duckdb

import (
	"context"

	_ "github.com/marcboeker/go-duckdb" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("duckdb", drivers.Driver{
		AllowMultilineComments: true,
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(ctx, `SELECT library_version FROM pragma_version()`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "DuckDB " + ver, nil
		},
		Copy: drivers.CopyWithInsert(func(int) string { return "?" }),
	})
}
