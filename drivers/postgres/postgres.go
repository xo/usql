// Package postgres defines and registers usql's PostgreSQL driver.
//
// See: https://github.com/lib/pq
package postgres

import (
	"io"

	"github.com/lib/pq" // DRIVER: postgres
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/drivers/metadata/informationschema"
)

func init() {
	newReader := func(db drivers.DB) metadata.Reader {
		ir := informationschema.New(
			informationschema.WithIndexes(false),
		)(db)
		mr := &metaReader{
			db: db,
		}
		return metadata.NewPluginReader(ir, mr)
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
		NewMetadataReader: newReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer) metadata.Writer {
			reader := newReader(db)
			opts := []metadata.Option{
				metadata.WithSystemSchemas([]string{"pg_catalog", "pg_toast", "information_schema"}),
			}
			return metadata.NewDefaultWriter(reader, opts...)(db, w)
		},
	}, "cockroachdb", "redshift")
}
