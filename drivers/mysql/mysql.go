package mysql

import (
	"strconv"

	// DRIVER: mysql
	"github.com/go-sql-driver/mysql"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("mysql", drivers.Driver{
		AMC: true,
		AHC: true,
		E: func(err error) (string, string) {
			if e, ok := err.(*mysql.MySQLError); ok {
				return strconv.Itoa(int(e.Number)), e.Message
			}

			return "", err.Error()
		},
		PwErr: func(err error) bool {
			if e, ok := err.(*mysql.MySQLError); ok {
				return e.Number == 1045
			}
			return false
		},
	}, "memsql", "vitess", "tidb")
}
