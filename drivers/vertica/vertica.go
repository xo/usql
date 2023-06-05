// Package vertica defines and registers usql's Vertica driver.
//
// See: https://github.com/vertica/vertica-sql-go
package vertica

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"

	vertigo "github.com/vertica/vertica-sql-go" // DRIVER
	"github.com/vertica/vertica-sql-go/logger"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	// List of custom TLS configurations that may be applied via query in connection string.
	customTlsConfig := map[string]func(string, *tls.Config) error{
		"ca_path": func(queryValue string, c *tls.Config) error {
			rootCertPool := x509.NewCertPool()

			pem, err := os.ReadFile(queryValue)
			if err != nil {
				return err
			}

			if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
				return fmt.Errorf("error: failed to append pem to cert pool")
			}

			c.RootCAs = rootCertPool

			return nil
		},
	}

	hasCustomTlsConfig := func(queries url.Values) bool {
		for key := range customTlsConfig {
			if queries.Has(key) {
				return true
			}
		}

		return false
	}

	applyCustomTlsConfig := func(queries url.Values, tlsConfig *tls.Config) error {
		for key, configFunction := range customTlsConfig {
			if queries.Has(key) {
				if err := configFunction(queries.Get(key), tlsConfig); err != nil {
					return err
				}
			}
		}

		return nil
	}

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
			return func(_, _ string) (*sql.DB, error) {
				queries := u.Query()

				if hasCustomTlsConfig(queries) {
					if queries.Get("tlsmode") != "server-strict" {
						configNames := []string{}

						for key := range customTlsConfig {
							configNames = append(configNames, key)
						}

						return nil, fmt.Errorf(fmt.Sprintf("error: when custom tls configurations are set (%s), tlsmode must be set to server-strict", strings.Join(configNames, ",")))
					}

					tlsConfig := &tls.Config{ServerName: u.Hostname()}

					if err := applyCustomTlsConfig(queries, tlsConfig); err != nil {
						return nil, err
					}

					if err := vertigo.RegisterTLSConfig("custom_tls_config", tlsConfig); err != nil {
						return nil, err
					}

					queries.Set("tlsmode", "custom_tls_config")
				}

				dsn := url.URL{
					Scheme:   u.Driver,
					User:     u.User,
					Host:     u.Host,
					Path:     u.Path,
					RawQuery: queries.Encode(),
				}

				return sql.Open(u.Driver, dsn.String())
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
