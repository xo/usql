// Package cassandra defines and registers usql's Cassandra driver.
//
// See: https://github.com/xo/cql
// Group: most
package cassandra

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/xo/cql" // DRIVER: cql
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	var debug bool
	if s := os.Getenv("CQL_DEBUG"); s != "" {
		log.Printf("ENABLING DEBUGGING FOR CQL")
		debug = true
	}
	// error regexp's
	authReqRE := regexp.MustCompile(`authentication required`)
	passwordErrRE := regexp.MustCompile(`Provided username (.*)and/or password are incorrect`)
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
		Open: func(_ context.Context, u *dburl.URL, stdout, stderr func() io.Writer) (func(string, string) (*sql.DB, error), error) {
			return func(_, dsn string) (*sql.DB, error) {
				cfg, err := cql.ParseDSN(dsn)
				if err != nil {
					return nil, err
				}
				// gocql logs nothing unless it is given a logger
				if debug {
					cfg.Logger = gocql.NewLogger(gocql.LogLevelDebug)
				}
				return sql.OpenDB(cql.NewConnector(cfg)), nil
			}, nil
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var release, protocol, cql string
			err := db.QueryRowContext(
				ctx,
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
			return passwordErrRE.MatchString(err.Error())
		},
		Err: func(err error) (string, string) {
			if authReqRE.MatchString(err.Error()) {
				return "", "authentication required"
			}
			if m := passwordErrRE.FindStringSubmatch(err.Error()); m != nil {
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
