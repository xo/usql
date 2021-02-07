// Package godror defines and registers usql's GO DRiver for ORacle. Requires
// CGO. Uses Oracle's ODPI-C (instant client) library.
//
// See: https://github.com/godror/godror
package godror

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	_ "github.com/godror/godror" // DRIVER: godror
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	orameta "github.com/xo/usql/drivers/metadata/oracle"
	"github.com/xo/usql/env"
	"golang.org/x/xerrors"
)

func init() {
	allCapsRE := regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`)
	endRE := regexp.MustCompile(`;?\s*$`)
	drivers.Register("godror", drivers.Driver{
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
			return "Oracle Database " + ver, nil
		},
		User: func(db drivers.DB) (string, error) {
			var user string
			err := db.QueryRow(`SELECT user FROM dual`).Scan(&user)
			return user, err
		},
		ChangePassword: func(db drivers.DB, user, newpw, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` IDENTIFIED BY ` + newpw)
			return err
		},
		Err: func(err error) (string, string) {
			if e := xerrors.Unwrap(err); e != nil {
				err = e
			}
			code, msg := "", err.Error()
			if e, ok := err.(interface {
				Code() int
			}); ok {
				code = fmt.Sprintf("ORA-%05d", e.Code())
			}
			if e, ok := err.(interface {
				Message() string
			}); ok {
				msg = e.Message()
			}
			if i := strings.LastIndex(msg, "ORA-"); msg == "" && i != -1 {
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
			if e := xerrors.Unwrap(err); e != nil {
				err = e
			}
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
		NewMetadataReader: orameta.NewReader(),
		NewMetadataWriter: func(db drivers.DB, w io.Writer) metadata.Writer {
			// TODO if options would be common to all readers, this could be moved
			// to the caller and passed in an argument
			envs := env.All()
			opts := []orameta.Option{}
			if envs["ECHO_HIDDEN"] == "on" || envs["ECHO_HIDDEN"] == "noexec" {
				if envs["ECHO_HIDDEN"] == "noexec" {
					opts = append(opts, orameta.WithDryRun(true))
				}
				opts = append(opts, orameta.WithLogger(log.New(os.Stdout, "DEBUG: ", log.LstdFlags)))
			}
			newReader := orameta.NewReader(opts...)
			writerOpts := []metadata.Option{
				metadata.WithSystemSchemas([]string{"ctxsys", "flows_files", "mdsys", "outln", "sys", "system", "xdb", "xs$null"}),
			}
			return metadata.NewDefaultWriter(newReader(db), writerOpts...)(db, w)
		},
	})
}
