package mssql

import (
	"database/sql"
	"strconv"
	"strings"

	// DRIVER: mssql
	"github.com/denisenkom/go-mssqldb"

	"github.com/knq/usql/drivers"
)

func init() {
	drivers.Register("mssql", drivers.Driver{
		V: func(db *sql.DB) (string, error) {
			var ver, level, edition string
			err := db.QueryRow(
				`SELECT SERVERPROPERTY('productversion'), SERVERPROPERTY ('productlevel'), SERVERPROPERTY ('edition')`,
			).Scan(&ver, &level, &edition)
			if err != nil {
				return "", err
			}
			return "Microsoft SQL Server " + ver + ", " + level + ", " + edition, nil
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
