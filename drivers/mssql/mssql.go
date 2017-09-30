package mssql

import (
	"strconv"
	"strings"

	// DRIVER: mssql
	"github.com/denisenkom/go-mssqldb"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("mssql", drivers.Driver{
		AMC:   true,
		ReqPP: true,
		Syn:   "tsql",
		V: func(db drivers.DB) (string, error) {
			var ver, level, edition string
			err := db.QueryRow(
				`SELECT SERVERPROPERTY('productversion'), SERVERPROPERTY ('productlevel'), SERVERPROPERTY ('edition')`,
			).Scan(&ver, &level, &edition)
			if err != nil {
				return "", err
			}
			return "Microsoft SQL Server " + ver + ", " + level + ", " + edition, nil
		},
		ChPw: func(db drivers.DB, user, new, old string) error {
			_, err := db.Exec(`alter login ` + user + ` with password = '` + new + `' old_password = '` + old + `'`)
			return err
		},
		E: func(err error) (string, string) {
			if e, ok := err.(mssql.Error); ok {
				return strconv.Itoa(int(e.Number)), e.Message
			}

			msg := err.Error()
			if i := strings.LastIndex(msg, "mssql:"); i != -1 {
				msg = msg[i:]
			}

			return "", msg
		},
		PwErr: func(err error) bool {
			return strings.Contains(err.Error(), "Login failed for")
		},
	})
}
