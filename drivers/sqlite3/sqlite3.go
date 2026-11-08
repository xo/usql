// Package sqlite3 defines and registers usql's SQLite3 driver. Requires CGO.
//
// See: https://github.com/mattn/go-sqlite3
package sqlite3

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mattn/go-sqlite3" // DRIVER
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/sqwire/sqshared"
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
		// psql's layout, with the primary result code as the code, and its
		// name and the extended code in DETAIL
		ErrVerbose: func(err error, _ string) string {
			var e sqlite3.Error
			if !errors.As(err, &e) {
				return ""
			}
			detail := fmt.Sprintf("%s, extended code %d", e.Code.Error(), int(e.ExtendedCode))
			return drivers.Verbose(strconv.Itoa(int(e.Code)), e.Error(), [2]string{"DETAIL", detail})
		},
		ConvertBytes: sqshared.ConvertBytes,
	})
}
