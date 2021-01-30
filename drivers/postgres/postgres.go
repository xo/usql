// Package postgres defines and registers usql's PostgreSQL driver.
//
// See: https://github.com/lib/pq
package postgres

import (
	"io"

	"github.com/lib/pq" // DRIVER: postgres
	"github.com/xo/dburl"
	"github.com/xo/tblfmt"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/drivers/metadata/informationschema"
	"github.com/xo/usql/env"
)

func init() {
	newReader := informationschema.New(
		informationschema.WithIndexes(false),
	)
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
				metadata.WithSystemSchemas([]string{"pg_catalog", "information_schema"}),
				metadata.WithListAllDbs(func(pattern string, verbose bool) error {
					return listAllDbs(db, w, pattern, verbose)
				}),
			}
			return metadata.NewDefaultWriter(reader, opts...)(db, w)
		},
	}, "cockroachdb", "redshift")
}

func listAllDbs(db drivers.DB, w io.Writer, pattern string, verbose bool) error {
	qstr := `
SELECT d.datname as "Name",
       pg_catalog.pg_get_userbyid(d.datdba) as "Owner",
       pg_catalog.pg_encoding_to_char(d.encoding) as "Encoding",
       d.datcollate as "Collate",
       d.datctype as "Ctype",
       pg_catalog.array_to_string(d.datacl, E'\n') AS "Access privileges"
FROM pg_catalog.pg_database d
ORDER BY 1;
`
	rows, err := db.Query(qstr)
	if err != nil {
		return err
	}
	defer rows.Close()

	return tblfmt.EncodeAll(w, rows, env.Pall())
}
