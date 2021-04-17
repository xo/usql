// Package postgres defines and registers usql's PostgreSQL driver.
//
// See: https://github.com/lib/pq
package postgres

import (
	"context"
	"io"

	"github.com/lib/pq" // DRIVER: postgres
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	pgmeta "github.com/xo/usql/drivers/metadata/postgres"
)

func init() {
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
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			// numeric version
			// SHOW server_version_num;
			var ver string
			err := db.QueryRowContext(ctx, `SHOW server_version`).Scan(&ver)
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
		NewMetadataReader: pgmeta.NewReader(),
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(pgmeta.NewReader()(db, opts...))(db, w)
		},
	}, "cockroachdb", "redshift")
}
