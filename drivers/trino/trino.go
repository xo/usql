// Package trino defines and registers usql's Trino driver.
//
// See: https://github.com/trinodb/trino-go-client
package trino

import (
	"context"
	"io"
	"regexp"

	_ "github.com/trinodb/trino-go-client/trino" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
)

func init() {
	endRE := regexp.MustCompile(`;?\s*$`)
	newReader := func(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
		ir := infos.New(
			infos.WithPlaceholder(func(int) string { return "?" }),
			infos.WithCustomClauses(map[infos.ClauseName]string{
				infos.ColumnsColumnSize:       "0",
				infos.ColumnsNumericScale:     "0",
				infos.ColumnsNumericPrecRadix: "0",
				infos.ColumnsCharOctetLength:  "0",
			}),
			infos.WithFunctions(false),
			infos.WithSequences(false),
			infos.WithIndexes(false),
			infos.WithConstraints(false),
			infos.WithColumnPrivileges(false),
			infos.WithUsagePrivileges(false),
		)(db, opts...)
		mr := &metaReader{
			LoggingReader: metadata.NewLoggingReader(db, opts...),
		}
		return metadata.NewPluginReader(ir, mr)
	}
	drivers.Register("trino", drivers.Driver{
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
			return "Trino " + ver, nil
		},
		NewMetadataReader: newReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(newReader(db, opts...))(db, w)
		},
		Copy: drivers.CopyWithInsert(func(int) string { return "?" }),
	})
}
