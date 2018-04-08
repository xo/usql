package clickhouse

import (
	"regexp"

	// DRIVER: presto
	_ "github.com/prestodb/presto-go-client/presto"

	"github.com/xo/usql/drivers"
)

var endRE = regexp.MustCompile(`;?\s*$`)

func init() {
	drivers.Register("presto", drivers.Driver{
		AllowMultilineComments: true,
		Process: func(prefix string, sqlstr string) (string, string, bool, error) {
			sqlstr = endRE.ReplaceAllString(sqlstr, "")
			typ, q := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, q, nil
		},
	})
}
