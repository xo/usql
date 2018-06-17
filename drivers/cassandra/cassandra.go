package cassandra

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	// DRIVER: cql
	cql "github.com/MichaelS11/go-cql-driver"

	"github.com/gocql/gocql"
	"github.com/xo/dburl"

	"github.com/xo/usql/drivers"
)

// logger is a null logger that satisfies the gocql.StdLogger and the io.Writer
// interfaces in order to capture the last error issued by the cql/gocql
// packages, since the cql package does not (at this time) return any error
// other than sql.ErrBadConn.
type logger struct {
	debug bool
	last  string
}

func (l *logger) Print(v ...interface{}) {
	if l.debug {
		log.Print(v...)
	}
}
func (l *logger) Printf(s string, v ...interface{}) {
	if l.debug {
		log.Printf(s, v...)
	}
}
func (l *logger) Println(v ...interface{}) {
	if l.debug {
		log.Println(v...)
	}
}
func (l *logger) Write(buf []byte) (int, error) {
	if l.debug {
		log.Printf("WRITE: %s", string(buf))
	}
	l.last = string(buf)
	return len(buf), nil
}

func init() {
	var debug bool
	if s := os.Getenv("CQL_DEBUG"); s != "" {
		log.Printf("ENABLING DEBUGGING FOR CQL")
		debug = true
	}

	// error regexp's
	authReqRE := regexp.MustCompile(`authentication required`)
	passwordErrRE := regexp.MustCompile(`Provided username (.*)and/or password are incorrect`)

	var l *logger
	drivers.Register("cql", drivers.Driver{
		AllowDollar:            true,
		AllowMultilineComments: true,
		AllowCComments:         true,
		LexerName:              "cql",
		ForceParams: func(u *dburl.URL) {
			if q := u.Query(); q.Get("timeout") == "" {
				q.Set("timeout", "300s")
				u.RawQuery = q.Encode()
			}
		},
		Open: func(*dburl.URL) (func(string, string) (*sql.DB, error), error) {
			// override cql and gocql loggers
			l = &logger{debug: debug}
			gocql.Logger, cql.CqlDriver.Logger = l, log.New(l, "", 0)
			return sql.Open, nil
		},
		Version: func(db drivers.DB) (string, error) {
			var release, protocol, cql string
			err := db.QueryRow(
				`SELECT release_version, cql_version, native_protocol_version FROM system.local WHERE key = 'local'`,
			).Scan(&release, &cql, &protocol)
			if err != nil {
				return "", err
			}
			return "Cassandra " + release + ", CQL " + cql + ", Protocol v" + protocol, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, _ string) error {
			_, err := db.Exec(`ALTER ROLE ` + user + ` WITH PASSWORD = '` + newpw + `'`)
			return err
		},
		IsPasswordErr: func(err error) bool {
			return passwordErrRE.MatchString(l.last)
		},
		Err: func(err error) (string, string) {
			if authReqRE.MatchString(l.last) {
				return "", "authentication required"
			}
			if m := passwordErrRE.FindStringSubmatch(l.last); m != nil {
				return "", fmt.Sprintf("invalid username %sor password", m[1])
			}
			return "", strings.TrimPrefix(strings.TrimPrefix(err.Error(), "driver: "), "gocql: ")
		},
		RowsAffected: func(sql.Result) (int64, error) {
			return 0, nil
		},
		ConvertDefault: func(v interface{}) (string, error) {
			buf, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(buf), nil
		},
		BatchQueryPrefixes: map[string]string{
			"BEGIN BATCH": "APPLY BATCH",
		},
	})
}
