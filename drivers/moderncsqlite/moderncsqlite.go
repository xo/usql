package moderncsqlite

import (
	"database/sql"
	"strconv"
	"strings"

	// DRIVER: moderncsqlite
	"modernc.org/sqlite"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/xoutil"
)

func init() {
	drivers.Register("moderncsqlite", drivers.Driver{
		AllowMultilineComments: true,
		Open: func(u *dburl.URL) (func(string, string) (*sql.DB, error), error) {
			return func(_ string, params string) (*sql.DB, error) {
				return sql.Open("sqlite", params)
			}, nil
		},
		Version: func(db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRow(`SELECT sqlite_version()`).Scan(&ver)
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
