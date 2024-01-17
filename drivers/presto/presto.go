// Package presto defines and registers usql's Presto driver.
//
// See: https://github.com/prestodb/presto-go-client
package presto

import (
	"context"
	"regexp"

	_ "github.com/prestodb/presto-go-client/presto" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	endRE := regexp.MustCompile(`;?\s*$`)
	drivers.Register("presto", drivers.Driver{
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
			return "Presto " + ver, nil
		},
	})
}
