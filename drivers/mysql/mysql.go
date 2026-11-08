// Package mysql defines and registers usql's MySQL driver.
//
// The memsql, tidb and vitess drivers register the same driver under their
// own names, through Driver.
//
// See: https://github.com/go-sql-driver/mysql
package mysql

import (
	"errors"
	"strconv"

	"github.com/go-sql-driver/mysql" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("mysql", Driver())
}

// Driver returns the MySQL driver. dburl opens a database that speaks the
// MySQL protocol, such as TiDB, through the mysql driver, and usql registers
// this driver under that database's own name.
func Driver() drivers.Driver {
	return drivers.Driver{
		Err: func(err error) (string, string) {
			if e, ok := err.(*mysql.MySQLError); ok {
				return strconv.Itoa(int(e.Number)), e.Message
			}
			return "", err.Error()
		},
		// psql's layout, with the SQLSTATE as the code, and MySQL's own
		// number of the error in DETAIL
		ErrVerbose: func(err error, _ string) string {
			var e *mysql.MySQLError
			if !errors.As(err, &e) {
				return ""
			}
			var code string
			if e.SQLState != [5]byte{} {
				code = string(e.SQLState[:])
			}
			return drivers.Verbose(code, e.Message, [2]string{"DETAIL", "error " + strconv.Itoa(int(e.Number))})
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*mysql.MySQLError); ok {
				return e.Number == 1045
			}
			return false
		},
	}
}
