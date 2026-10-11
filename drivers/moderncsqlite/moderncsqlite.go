// Package moderncsqlite defines and registers usql's ModernC SQLite3 driver.
// Transpilation of SQLite3 to Go.
//
// See: https://gitlab.com/cznic/sqlite
package moderncsqlite

import (
	"context"
	"database/sql"
	"io"
	"regexp"
	"strconv"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/sqlite3/sqshared"
	"modernc.org/sqlite" // DRIVER
)

// errorRE splits the text of an error of modernc.org/sqlite.
var errorRE = regexp.MustCompile(`^[^:]+: (.*) \(\d+\)$`)

func init() {
	drivers.Register("moderncsqlite", drivers.Driver{
		Open: func(_ context.Context, u *dburl.URL, stdout, stderr func() io.Writer) (func(string, string) (*sql.DB, error), error) {
			return func(_ string, params string) (*sql.DB, error) {
				return sql.Open("sqlite", params)
			}, nil
		},
		// modernc words an error as "SQL logic error: message (1)", where
		// go-sqlite3 says "message", so the code and the name of the code are
		// taken off the text, and the two drivers print the same
		Err: func(err error) (string, string) {
			if e, ok := err.(*sqlite.Error); ok {
				msg := e.Error()
				if m := errorRE.FindStringSubmatch(msg); m != nil {
					msg = m[1]
				}
				return strconv.Itoa(e.Code()), msg
			}
			return "", err.Error()
		},
		ConvertBytes: sqshared.ConvertBytes,
	})
}
