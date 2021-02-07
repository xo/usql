// Package trino defines and registers usql's Trino driver.
//
// See: https://github.com/trinodb/trino-go-client
package trino

import (
	"io"
	"log"
	"os"
	"regexp"

	_ "github.com/trinodb/trino-go-client/trino" // DRIVER: trino
	"github.com/xo/tblfmt"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
	"github.com/xo/usql/env"
)

func init() {
	endRE := regexp.MustCompile(`;?\s*$`)
	readerOpts := []infos.Option{
		infos.WithPlaceholder(func(int) string { return "?" }),
		infos.WithCustomColumns(map[infos.ColumnName]string{
			infos.ColumnsColumnSize:       "0",
			infos.ColumnsNumericScale:     "0",
			infos.ColumnsNumericPrecRadix: "0",
			infos.ColumnsCharOctetLength:  "0",
		}),
		infos.WithFunctions(false),
		infos.WithSequences(false),
		infos.WithIndexes(false),
	}
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
				metadata.WithListAllDbs(func(pattern string, verbose bool) error {
					return listAllDbs(db, w, pattern, verbose)
				}),
			}
			return metadata.NewDefaultWriter(reader, writerOpts...)(db, w)
		},
	})
}

func listAllDbs(db drivers.DB, w io.Writer, pattern string, verbose bool) error {
	rows, err := db.Query("SHOW catalogs")
	if err != nil {
		return err
	}
	defer rows.Close()

	return tblfmt.EncodeAll(w, rows, env.Pall())
}
