// Package trino defines and registers usql's Trino driver.
//
// See: https://github.com/trinodb/trino-go-client
package trino

import (
	"regexp"

	_ "github.com/trinodb/trino-go-client/trino" // DRIVER: trino
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/informationschema"
)

func init() {
	endRE := regexp.MustCompile(`;?\s*$`)
	drivers.Register("trino", drivers.Driver{
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
			return "Trino " + ver, nil
		},
		NewMetadataReader: informationschema.New,
	})
}
