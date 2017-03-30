package sqlite3

import (
	"strconv"
	"strings"
	"time"

	// DRIVER: sqlite3
	"github.com/mattn/go-sqlite3"

	"github.com/knq/xoutil"

	"github.com/knq/usql/drivers"
)

func init() {
	drivers.Register("sqlite3", drivers.Driver{
		V: func(db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRow(`select sqlite_version()`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "SQLite3 " + ver, nil
		},
		E: func(err error) (string, string) {
			if e, ok := err.(sqlite3.Error); ok {
				return strconv.Itoa(int(e.Code)), e.Error()
			}

			code, msg := "", err.Error()
			if e, ok := err.(sqlite3.ErrNo); ok {
				code = strconv.Itoa(int(e))
			}

			return code, msg
		},
		Cb: func(buf []byte) string {
			// attempt to convert buf if it matches a time format, and if it
			// does, then return a formatted time string.
			s := string(buf)
			if s != "" && strings.TrimSpace(s) != "" {
				t := &xoutil.SqTime{}
				err := t.Scan(buf)
				if err == nil {
					return t.Format(time.RFC3339Nano)
				}
			}

			return s
		},
	})
}
