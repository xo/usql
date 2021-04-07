// Package firebird defines and registers usql's Firebird driver.
//
// See: https://github.com/nakagami/firebirdsql
package firebird

import (
	"context"

	_ "github.com/nakagami/firebirdsql" // DRIVER: firebirdsql
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("firebirdsql", drivers.Driver{
		AllowMultilineComments: true,
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(
				ctx,
				`SELECT rdb$get_context('SYSTEM', 'ENGINE_VERSION') FROM rdb$database;`,
			).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "Firebird " + ver, nil
		},
	})
}
