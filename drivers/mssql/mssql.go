// Package mssql defines and registers usql's Microsoft SQL Server driver.
//
// See: https://github.com/denisenkom/go-mssqldb
package mssql

import (
	"strconv"
	"strings"

	mssql "github.com/denisenkom/go-mssqldb" // DRIVER: mssql
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("mssql", drivers.Driver{
		AllowMultilineComments:  true,
		RequirePreviousPassword: true,
		LexerName:               "tsql",
		Version: func(db drivers.DB) (string, error) {
			var ver, level, edition string
			err := db.QueryRow(
				`SELECT SERVERPROPERTY('productversion'), SERVERPROPERTY ('productlevel'), SERVERPROPERTY ('edition')`,
			).Scan(&ver, &level, &edition)
			if err != nil {
				return "", err
			}
			return "Microsoft SQL Server " + ver + ", " + level + ", " + edition, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, oldpw string) error {
			_, err := db.Exec(`ALTER LOGIN ` + user + ` WITH password = '` + newpw + `' old_password = '` + oldpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(mssql.Error); ok {
				return strconv.Itoa(int(e.Number)), e.Message
			}
			msg := err.Error()
			if i := strings.LastIndex(msg, "mssql:"); i != -1 {
				msg = msg[i:]
			}
			return "", msg
		},
		IsPasswordErr: func(err error) bool {
			return strings.Contains(err.Error(), "Login failed for")
		},
	})
}
