// Package mysql defines and registers usql's MySQL driver.
//
// The memsql, tidb and vitess drivers register the same driver under their
// own names, through Driver.
//
// See: https://github.com/go-sql-driver/mysql
// Group: base
package mysql

import (
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
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*mysql.MySQLError); ok {
				return e.Number == 1045
			}
			return false
		},
	}
}
