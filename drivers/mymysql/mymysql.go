package mymysql

import (
	"strconv"

	// DRIVER: mymysql
	_ "github.com/ziutek/mymysql/godrv"

	"github.com/ziutek/mymysql/mysql"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("mymysql", drivers.Driver{
		AllowMultilineComments: true,
		AllowHashComments:      true,
		LexerName:              "mysql",
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
	})
}
