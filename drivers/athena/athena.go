// Package athena defines and registers usql's AWS Athena driver.
//
// See: https://github.com/uber/athenadriver
package athena

import (
	"context"

	_ "github.com/uber/athenadriver/go" // DRIVER: awsathena
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("awsathena", drivers.Driver{
		AllowMultilineComments: true,
		Process:                drivers.StripTrailingSemicolon,
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(
				ctx,
				`SELECT node_version FROM system.runtime.nodes LIMIT 1`,
			).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "Athena " + ver, nil
		},
	})
}
