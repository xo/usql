// Package vertica defines and registers usql's Vertica driver.
//
// See: https://github.com/vertica/vertica-sql-go
package vertica

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"

	vertica "github.com/vertica/vertica-sql-go" // DRIVER
	"github.com/vertica/vertica-sql-go/logger"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	// turn off logging
	if os.Getenv("VERTICA_SQL_GO_LOG_LEVEL") == "" {
		logger.SetLogLevel(logger.NONE)
	}

	errCodeRE := regexp.MustCompile(`(?i)^\[([0-9a-z]+)\]\s+(.+)`)
	drivers.Register("vertica", drivers.Driver{
		AllowDollar:            true,
		AllowMultilineComments: true,
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			if err := db.QueryRowContext(ctx, `SELECT version()`).Scan(&ver); err != nil {
				return "", err
			}
			return ver, nil
		},
		Open: func(_ context.Context, u *dburl.URL, stdout, stderr func() io.Writer) (func(string, string) (*sql.DB, error), error) {
			return func(driver, dsn string) (*sql.DB, error) {
				u, err := url.Parse(dsn)
				if err != nil {
					return nil, err
				}
				q := u.Query()
				if name := q.Get("ca_path"); name != "" {
					if q.Get("tlsmode") != "server-strict" {
						return nil, errors.New("tlsmode must be set to server-strict: ca_path is set")
					}
					cfg := &tls.Config{
						ServerName: u.Hostname(),
					}
					if err := addCA(name, cfg); err != nil {
						return nil, err
					}
					if err := vertica.RegisterTLSConfig("custom_tls_config", cfg); err != nil {
						return nil, err
					}
					q.Set("tlsmode", "custom_tls_config")
				}
				return sql.Open(driver, u.String())
			}, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` IDENTIFIED BY '` + newpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			msg := strings.TrimSpace(strings.TrimPrefix(err.Error(), "Error:"))
			if m := errCodeRE.FindAllStringSubmatch(msg, -1); m != nil {
				return m[0][1], strings.TrimSpace(m[0][2])
			}
			return "", msg
		},
		IsPasswordErr: func(err error) bool {
			return strings.HasSuffix(strings.TrimSpace(err.Error()), "Invalid username or password")
		},
	})
}

// addCA adds the specified file name as a ca to the tls config.
func addCA(name string, cfg *tls.Config) error {
	pool := x509.NewCertPool()
	switch pem, err := os.ReadFile(name); {
	case err != nil:
		return err
	case !pool.AppendCertsFromPEM(pem):
		return errors.New("failed to append pem to cert pool")
	}
	cfg.RootCAs = pool
	return nil
}
