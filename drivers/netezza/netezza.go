// Package netezza defines and registers usql's Netezza driver.
//
// See: https://github.com/IBM/nzgo
package netezza

import (
	"context"
	"io"

	"github.com/IBM/nzgo" // DRIVER: nzgo
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
)

func init() {
	elog := nzgo.PDALogger{
		LogLevel: "off",
	}
	elog.Initialize()

	newReader := infos.New(
		infos.WithPlaceholder(func(int) string { return "?" }),
		infos.WithIndexes(false),
		infos.WithCustomColumns(map[infos.ColumnName]string{
			infos.ColumnsColumnSize:         "COALESCE(character_maximum_length, numeric_precision, datetime_precision, interval_precision, 0)",
			infos.FunctionColumnsColumnSize: "COALESCE(character_maximum_length, numeric_precision, datetime_precision, interval_precision, 0)",
		}),
		infos.WithSystemSchemas([]string{"DEFINITION_SCHEMA", "INFORMATION_SCHEMA"}),
		infos.WithCurrentSchema("CURRENT_SCHEMA"),
	)
	drivers.Register("nzgo", drivers.Driver{
		Name:                   "nz",
		AllowDollar:            true,
		AllowMultilineComments: true,
		LexerName:              "postgres",
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(ctx, `SELECT version()`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "Netezza " + ver, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` PASSWORD '` + newpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(*nzgo.Error); ok {
				return string(e.Code), e.Message
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*nzgo.Error); ok {
				return e.Code.Name() == "invalid_password"
			}
			return false
		},
		NewMetadataReader: newReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(newReader(db, opts...))(db, w)
		},
	})
}
