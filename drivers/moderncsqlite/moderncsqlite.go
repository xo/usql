// Package moderncsqlite defines and registers usql's ModernC SQLite3 driver, a
// transpilation of SQLite3 to pure Go.
//
// See: https://gitlab.com/cznic/sqlite
package moderncsqlite

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/sqlite3/sqshared"
	"modernc.org/sqlite" // DRIVER: moderncsqlite
)

func init() {
	drivers.Register("moderncsqlite", drivers.Driver{
		AllowMultilineComments: true,
		Open: func(u *dburl.URL) (func(string, string) (*sql.DB, error), error) {
			return func(_ string, params string) (*sql.DB, error) {
				return sql.Open("sqlite", params)
			}, nil
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(ctx, `SELECT sqlite_version()`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "ModernC SQLite " + ver, nil
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(*sqlite.Error); ok {
				return strconv.Itoa(e.Code()), e.Error()
			}
			return "", err.Error()
		},
		ConvertBytes:      sqshared.ConvertBytes,
		NewMetadataReader: sqshared.NewMetadataReader,
		Copy:              drivers.CopyWithInsert(func(int) string { return "?" }),
	})
}
