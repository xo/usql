package vertica

import (
	"database/sql"
	"os"
	"regexp"
	"strings"

	// DRIVER: vertica
	_ "github.com/vertica/vertica-sql-go"

	"github.com/vertica/vertica-sql-go/logger"
	"github.com/xo/usql/drivers"
)

func init() {
	// turn off logging
	if os.Getenv("VERTICA_SQL_GO_LOG_LEVEL") == "" {
		logger.SetLogLevel(logger.NONE)
	}

	codeRE := regexp.MustCompile(`(?i)^\[([0-9a-z]+)\]\s+(.+)`)

	drivers.Register("vertica", drivers.Driver{
		AllowDollar:            true,
		AllowMultilineComments: true,
		Version: func(db drivers.DB) (string, error) {
			var ver string
			if err := db.QueryRow(`SELECT version()`).Scan(&ver); err != nil {
				return "", err
			}
			return ver, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` IDENTIFIED BY '` + newpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			msg := strings.TrimSpace(strings.TrimPrefix(err.Error(), "Error:"))
			if m := codeRE.FindAllStringSubmatch(msg, -1); m != nil {
				return m[0][1], m[0][2]
			}
			return "", msg
		},
		IsPasswordErr: func(err error) bool {
			return strings.HasSuffix(strings.TrimSpace(err.Error()), "Invalid username or password")
		},
		RowsAffected: func(sql.Result) (int64, error) {
			return 0, nil
		},
	})
}
