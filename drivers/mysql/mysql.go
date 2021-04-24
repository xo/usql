// Package mysql defines and registers usql's MySQL driver.
//
// See: https://github.com/go-sql-driver/mysql
package mysql

import (
	"io"
	"strconv"

	"github.com/go-sql-driver/mysql" // DRIVER: mysql
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
)

func init() {
	newReader := infos.New(
		infos.WithPlaceholder(func(int) string { return "?" }),
		infos.WithSequences(false),
		infos.WithCheckConstraints(false),
		infos.WithCustomClauses(map[infos.ClauseName]string{
			infos.ColumnsNumericPrecRadix:         "10",
			infos.FunctionColumnsNumericPrecRadix: "10",
			infos.ConstraintIsDeferrable:          "''",
			infos.ConstraintInitiallyDeferred:     "''",
			infos.ConstraintJoinCond:              "AND r.table_name = f.table_name",
		}),
		infos.WithSystemSchemas([]string{"mysql", "information_schema", "performance_schema", "sys"}),
		infos.WithCurrentSchema("COALESCE(DATABASE(), '%')"),
	)
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
		NewMetadataReader: newReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(newReader(db, opts...))(db, w)
		},
		Copy: drivers.CopyWithInsert(func(int) string { return "?" }),
	}, "memsql", "vitess", "tidb")
}
