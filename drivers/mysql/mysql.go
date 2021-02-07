// Package mysql defines and registers usql's MySQL driver.
//
// See: https://github.com/go-sql-driver/mysql
package mysql

import (
	"io"
	"log"
	"os"
	"strconv"

	"github.com/go-sql-driver/mysql" // DRIVER: mysql
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
	"github.com/xo/usql/env"
)

func init() {
	readerOpts := []infos.Option{
		infos.WithPlaceholder(func(int) string { return "?" }),
		infos.WithSequences(false),
		infos.WithCustomColumns(map[infos.ColumnName]string{
			infos.ColumnsNumericPrecRadix:         "10",
			infos.FunctionColumnsNumericPrecRadix: "10",
		}),
	}
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
		NewMetadataReader: infos.New(readerOpts...),
		NewMetadataWriter: func(db drivers.DB, w io.Writer) metadata.Writer {
			opts := append([]infos.Option{}, readerOpts...)
			// TODO if options would be common to all readers, this could be moved
			// to the caller and passed in an argument
			envs := env.All()
			if envs["ECHO_HIDDEN"] == "on" || envs["ECHO_HIDDEN"] == "noexec" {
				if envs["ECHO_HIDDEN"] == "noexec" {
					opts = append(opts, infos.WithDryRun(true))
				}
				opts = append(opts, infos.WithLogger(log.New(os.Stdout, "DEBUG: ", log.LstdFlags)))
			}
			reader := infos.New(opts...)(db)
			writerOpts := []metadata.Option{
				metadata.WithSystemSchemas([]string{"mysql", "information_schema", "performance_schema"}),
			}
			return metadata.NewDefaultWriter(reader, writerOpts...)(db, w)
		},
	}, "memsql", "vitess", "tidb")
}
