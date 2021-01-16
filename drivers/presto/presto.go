package presto

import (
	"regexp"

	_ "github.com/prestodb/presto-go-client/presto" // DRIVER: presto
	"github.com/xo/usql/drivers"
)

func init() {
	endRE := regexp.MustCompile(`;?\s*$`)
	drivers.Register("presto", drivers.Driver{
		AllowMultilineComments: true,
		Process: func(prefix string, sqlstr string) (string, string, bool, error) {
			sqlstr = endRE.ReplaceAllString(sqlstr, "")
			typ, q := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, q, nil
		},
		Version: func(db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRow(
				`SELECT node_version FROM system.runtime.nodes LIMIT 1`,
			).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "Presto " + ver, nil
		},
	})
}
