// Package sqlite3 defines and registers usql's SQLite3 driver. Requires CGO.
//
// See: https://github.com/mattn/go-sqlite3
// Group: base
package sqlite3

import (
	"strconv"

	"github.com/mattn/go-sqlite3" // DRIVER
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/sqlite3/sqshared"
)

func init() {
	drivers.Register("sqlite3", drivers.Driver{
		Err: func(err error) (string, string) {
			if e, ok := err.(sqlite3.Error); ok {
				return strconv.Itoa(int(e.Code)), e.Error()
			}
			code, msg := "", err.Error()
			if e, ok := err.(sqlite3.ErrNo); ok {
				code = strconv.Itoa(int(e))
			}
			return code, msg
		},
		ConvertBytes: sqshared.ConvertBytes,
	})
}
