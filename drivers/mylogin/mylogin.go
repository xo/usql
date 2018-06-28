package mysql

import (
	"strconv"

	// DRIVER: mylogin
	_ "github.com/dolmen-go/mylogin-driver/register"

	"github.com/go-sql-driver/mysql"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("mylogin", drivers.Driver{
		// Duplicated from "mysql" -- BEGIN
		AllowMultilineComments: true,
		AllowHashComments:      true,
		LexerName:              "mysql",
		ForceParams: drivers.ForceQueryParameters([]string{
			"parseTime", "true",
			"loc", "Local",
			"sql_mode", "ansi",
		}),
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
		// Duplicated from "mysql" -- END
	})
}
