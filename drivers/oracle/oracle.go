package oracle

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	// DRIVER: ora
	_ "gopkg.in/rana/ora.v4"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
)

var allCapsRE = regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`)
var endRE = regexp.MustCompile(`;?\s*$`)

func init() {
	drivers.Register("ora", drivers.Driver{
		AllowMultilineComments: true,
		ForceParams: func(u *dburl.URL) {
			// if the service name is not specified, use the environment
			// variable if present
			if strings.TrimPrefix(u.Path, "/") == "" {
				if n := env.Getenv("ORACLE_SID", "ORASID"); n != "" {
					u.Path = "/" + n
					if u.Host == "" {
						u.Host = "localhost"
					}
				}
			}
		},
		Version: func(db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRow(`SELECT version FROM v$instance`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "Oracle " + ver, nil
		},
		User: func(db drivers.DB) (string, error) {
			var user string
			err := db.QueryRow(`SELECT user FROM dual`).Scan(&user)
			return user, err
		},
		ChangePassword: func(db drivers.DB, user, new, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` IDENTIFIED BY ` + new)
			return err
		},
		Err: func(err error) (string, string) {
			code, msg := "", err.Error()

			if e, ok := err.(interface {
				Code() int
			}); ok {
				code = fmt.Sprintf("ORA-%05d", e.Code())
			}

			if i := strings.LastIndex(msg, "ORA-"); i != -1 {
				msg = msg[i:]
				if j := strings.Index(msg, ":"); j != -1 {
					msg = msg[j+1:]
					if code == "" {
						code = msg[i:j]
					}
				}
			}

			return code, strings.TrimSpace(msg)
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(interface {
				Code() int
			}); ok {
				return e.Code() == 1017 || e.Code() == 1005
			}
			return false
		},
		Columns: func(rows *sql.Rows) ([]string, error) {
			cols, err := rows.Columns()
			if err != nil {
				return nil, err
			}

			for i, c := range cols {
				if allCapsRE.MatchString(c) {
					cols[i] = strings.ToLower(c)
				}
			}

			return cols, nil
		},
		Process: func(prefix string, sqlstr string) (string, string, bool, error) {
			sqlstr = endRE.ReplaceAllString(sqlstr, "")
			typ, q := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, q, nil
		},
	})
}
