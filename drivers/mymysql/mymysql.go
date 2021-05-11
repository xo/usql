// Package mymysql defines and registers usql's MySQL MyMySQL driver.
//
// See: https://github.com/ziutek/mymysql
package mymysql

import (
	"strconv"

	"github.com/xo/usql/drivers"
	_ "github.com/ziutek/mymysql/godrv" // DRIVER
	"github.com/ziutek/mymysql/mysql"
)

func init() {
	drivers.Register("mymysql", drivers.Driver{
		AllowMultilineComments: true,
		AllowHashComments:      true,
		LexerName:              "mysql",
		UseColumnTypes:         true,
		Err: func(err error) (string, string) {
			if e, ok := err.(*mysql.Error); ok {
				return strconv.Itoa(int(e.Code)), string(e.Msg)
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*mysql.Error); ok {
				return e.Code == mysql.ER_ACCESS_DENIED_ERROR
			}
			return false
		},
		Copy: drivers.CopyWithInsert(func(int) string { return "?" }),
	})
}
