// Package adodb defines and registers usql's Microsoft ADODB driver. Requires
// CGO. Windows only.
//
// Alias: oleodbc, OLE ODBC
//
// See: https://github.com/mattn/go-adodb
package adodb

import (
	"database/sql"
	"regexp"
	"strings"

	_ "github.com/mattn/go-adodb" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	endRE := regexp.MustCompile(`;?\s*$`)
	endAnchorRE := regexp.MustCompile(`(?i)\send\s*;\s*$`)
	drivers.Register("adodb", drivers.Driver{
		AllowMultilineComments: true,
		AllowCComments:         true,
		Process: func(u *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			// trim last ; but only when not END;
			if s := strings.ToLower(u.Query().Get("usql_trim")); s != "" && s != "off" && s != "0" && s != "false" {
				if !endAnchorRE.MatchString(sqlstr) {
					sqlstr = endRE.ReplaceAllString(sqlstr, "")
				}
			}
			typ, q := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, q, nil
		},
		RowsAffected: func(res sql.Result) (int64, error) {
			return 0, nil
		},
	}, "oleodbc")
}
