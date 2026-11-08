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
		// the primary and the extended result codes, by number and by name
		ErrVerbose: func(err error, _ string) string {
			var e sqlite3.Error
			if !errors.As(err, &e) {
				return ""
			}
			return fmt.Sprintf("ERROR:  %d (%s), extended code %d: %s", int(e.Code), e.Code.Error(), int(e.ExtendedCode), e.Error())
		},
		ConvertBytes: sqshared.ConvertBytes,
	})
}
