// Package odbc defines and registers usql's ODBC driver. Requires CGO. Uses
// respective platform's standard ODBC packages.
//
// See: https://github.com/alexbrainman/odbc
// Group: all
package odbc

import (
	"regexp"
	"strings"

	"github.com/alexbrainman/odbc" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	endRE := regexp.MustCompile(`;?\s*$`)
	endAnchorRE := regexp.MustCompile(`(?i)\send\s*;\s*$`)
	drivers.Register("odbc", drivers.Driver{
		LexerName: "tsql",
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
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*odbc.Error); ok {
				msg := strings.ToLower(e.Error())
				return strings.Contains(msg, "failed") &&
					(strings.Contains(msg, "login") ||
						strings.Contains(msg, "authentication") ||
						strings.Contains(msg, "password"))
			}
			return false
		},
	})
}
