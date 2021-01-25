// Package mysql defines and registers usql's MySQL driver.
//
// See: https://github.com/go-sql-driver/mysql
package mysql

import (
	"strconv"

	"github.com/go-sql-driver/mysql" // DRIVER: mysql
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata/informationschema"
)

func init() {
	drivers.Register("mysql", drivers.Driver{
		AllowMultilineComments: true,
		AllowHashComments:      true,
		LexerName:              "mysql",
		ForceParams: drivers.ForceQueryParameters([]string{
			"parseTime", "true",
			"loc", "Local",
			"sql_mode", "ansi",
		}),
		Err: func(err error) (string, string) {
			if e, ok := err.(*mysql.MySQLError); ok {
				return strconv.Itoa(int(e.Number)), e.Message
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*mysql.MySQLError); ok {
				return e.Number == 1045
			}
			return false
		},
		NewMetadataReader: informationschema.New,
	}, "memsql", "vitess", "tidb")
}
