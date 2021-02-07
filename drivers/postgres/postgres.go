// Package postgres defines and registers usql's PostgreSQL driver.
//
// See: https://github.com/lib/pq
package postgres

import (
	"io"
	"log"
	"os"

	"github.com/lib/pq" // DRIVER: postgres
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
	"github.com/xo/usql/env"
)

func init() {
	readerOpts := []infos.Option{
		infos.WithIndexes(false),
		infos.WithCustomColumns(map[infos.ColumnName]string{
			infos.ColumnsColumnSize:         "COALESCE(character_maximum_length, numeric_precision, datetime_precision, interval_precision, 0)",
			infos.FunctionColumnsColumnSize: "COALESCE(character_maximum_length, numeric_precision, datetime_precision, interval_precision, 0)",
		}),
	}
	drivers.Register("postgres", drivers.Driver{
		Name:                   "pq",
		AllowDollar:            true,
		AllowMultilineComments: true,
		LexerName:              "postgres",
		ForceParams: func(u *dburl.URL) {
			if u.Scheme == "cockroachdb" {
				drivers.ForceQueryParameters([]string{"sslmode", "disable"})(u)
			}
		},
		Version: func(db drivers.DB) (string, error) {
			// numeric version
			// SHOW server_version_num;
			var ver string
			err := db.QueryRow(`SHOW server_version`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "PostgreSQL " + ver, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` PASSWORD '` + newpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(*pq.Error); ok {
				return string(e.Code), e.Message
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*pq.Error); ok {
				return e.Code.Name() == "invalid_password"
			}
			return false
		},
		NewMetadataReader: func(db drivers.DB) metadata.Reader {
			return metadata.NewPluginReader(
				infos.New(readerOpts...)(db),
				&metaReader{db: db},
			)
		},
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
			reader := metadata.NewPluginReader(
				infos.New(opts...)(db),
				// TODO this reader doesn't get logger options applied
				&metaReader{db: db},
			)
			writerOpts := []metadata.Option{
				metadata.WithSystemSchemas([]string{"pg_catalog", "pg_toast", "information_schema"}),
			}
			return metadata.NewDefaultWriter(reader, writerOpts...)(db, w)
		},
	}, "cockroachdb", "redshift")
}
