// Package athena defines and registers usql's AWS Athena driver.
//
// See: https://github.com/uber/athenadriver
package athena

import (
	"context"
	"regexp"

	_ "github.com/uber/athenadriver/go" // DRIVER: awsathena
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	endRE := regexp.MustCompile(`;?\s*$`)
	drivers.Register("awsathena", drivers.Driver{
		AllowMultilineComments: true,
		Process: func(_ *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			sqlstr = endRE.ReplaceAllString(sqlstr, "")
			typ, q := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, q, nil
		},
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
