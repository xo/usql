// Package orshared contains shared a shared driver implementation for the
// Oracle Database. Used by Oracle and Godror drivers.
package orshared

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	orameta "github.com/xo/usql/drivers/metadata/oracle"
	"github.com/xo/usql/env"
)

// Register registers an oracle driver.
func Register(name string, err func(error) (string, string), isPasswordErr func(error) bool) {
	endRE := regexp.MustCompile(`;?\s*$`)
	endAnchorRE := regexp.MustCompile(`(?i)\send\s*;\s*$`)
	drivers.Register(name, drivers.Driver{
		AllowMultilineComments: true,
		LowerColumnNames:       true,
		ForceParams: func(u *dburl.URL) {
			// if the service name is not specified, use the environment
			// variable if present
			if strings.TrimPrefix(u.Path, "/") == "" {
				if n, ok := env.Getenv("ORACLE_SID", "ORASID"); ok && n != "" {
					u.Path = "/" + n
					if u.Host == "" {
						u.Host = "localhost"
					}
				}
			}
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			if err := db.QueryRowContext(ctx, `SELECT version FROM v$instance`).Scan(&ver); err != nil {
				return "", err
			}
			return "Oracle Database " + ver, nil
		},
		User: func(ctx context.Context, db drivers.DB) (string, error) {
			var user string
			if err := db.QueryRowContext(ctx, `SELECT user FROM dual`).Scan(&user); err != nil {
				return "", err
			}
			return user, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` IDENTIFIED BY ` + newpw)
			return err
		},
		Err:           err,
		IsPasswordErr: isPasswordErr,
		Process: func(_ *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			if !endAnchorRE.MatchString(sqlstr) {
				// trim last ; but only when not END;
				sqlstr = endRE.ReplaceAllString(sqlstr, "")
			}
			typ, q := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, q, nil
		},
		NewMetadataReader: orameta.NewReader(),
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(orameta.NewReader()(db, opts...))(db, w)
		},
		Copy: drivers.CopyWithInsert(func(n int) string {
			return fmt.Sprintf(":%d", n)
		}),
	})
}
