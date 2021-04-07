// Package moderncsqlite defines and registers usql's ModernC SQLite3 driver, a
// transpilation of SQLite3 to pure Go.
//
// See: https://gitlab.com/cznic/sqlite
package moderncsqlite

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/xoutil"
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
		ConvertBytes: func(buf []byte, tfmt string) (string, error) {
			// attempt to convert buf if it matches a time format, and if it
			// does, then return a formatted time string.
			s := string(buf)
			if s != "" && strings.TrimSpace(s) != "" {
				t := new(xoutil.SqTime)
				if err := t.Scan(buf); err == nil {
					return t.Format(tfmt), nil
				}
			}
			return s, nil
		},
	})
}
