// Package moderncsqlite defines and registers usql's ModernC SQLite3 driver.
// Transpilation of SQLite3 to Go.
//
// See: https://gitlab.com/cznic/sqlite
package moderncsqlite

import (
	"regexp"
	"strconv"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/sqwire/sqshared"
	"modernc.org/sqlite" // DRIVER
)

// errorRE splits the text of an error of modernc.org/sqlite.
var errorRE = regexp.MustCompile(`^[^:]+: (.*) \(\d+\)$`)

func init() {
	drivers.Register("moderncsqlite", drivers.Driver{
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
