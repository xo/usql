// Package pgx defines and registers usql's PostgreSQL PGX driver.
//
// See: https://github.com/jackc/pgx
package pgx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	pgmeta "github.com/xo/usql/drivers/metadata/postgres"
	"github.com/xo/usql/text"
)

func init() {
	drivers.Register("pgx", drivers.Driver{
		AllowDollar:            true,
		AllowMultilineComments: true,
		LexerName:              "postgres",
		Open: func(ctx context.Context, u *dburl.URL, stdout, stderr func() io.Writer) (func(string, string) (*sql.DB, error), error) {
			return func(_, dsn string) (*sql.DB, error) {
				config, err := pgx.ParseConfig(dsn)
				if err != nil {
					return nil, err
				}
				config.OnNotice = func(_ *pgconn.PgConn, notice *pgconn.Notice) {
					out := stderr()
					fmt.Fprintln(out, notice.Severity+": ", notice.Message)
					if notice.Hint != "" {
						fmt.Fprintln(out, "HINT: ", notice.Hint)
					}
				}
				config.OnNotification = func(_ *pgconn.PgConn, notification *pgconn.Notification) {
					var payload string
					if notification.Payload != "" {
						payload = fmt.Sprintf(text.NotificationPayload, notification.Payload)
					}
					fmt.Fprintln(stdout(), fmt.Sprintf(text.NotificationReceived, notification.Channel, payload, notification.PID))
				}
				// NOTE: as opposed to the github.com/lib/pq driver, this
				// NOTE: driver has a "prefer" mode that is enabled by default.
				// NOTE: as such there is no logic here to try to reconnect as
				// NOTE: in the postgres driver.
				return stdlib.OpenDB(*config), nil
			}, nil
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(ctx, `SHOW server_version`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "PostgreSQL " + ver, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` PASSWORD '` + newpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			var e *pgconn.PgError
			if errors.As(err, &e) {
				return e.Code, e.Message
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			var e *pgconn.PgError
			if errors.As(err, &e) {
				return e.Code == "28P01"
			}
			return false
		},
		NewMetadataReader: pgmeta.NewReader(),
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(pgmeta.NewReader()(db, opts...))(db, w)
		},
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			conn, err := db.Conn(context.Background())
			if err != nil {
				return 0, fmt.Errorf("failed to get a connection from pool: %w", err)
			}

			leftParen := strings.IndexRune(table, '(')
			colQuery := "SELECT * FROM " + table + " WHERE 1=0"
			if leftParen != -1 {
				// pgx's CopyFrom needs a slice of column names and splitting them by a comma is unreliable
				// so evaluate the possible expressions against the target table
				colQuery = "SELECT " + table[leftParen+1:len(table)-1] + " FROM " + table[:leftParen] + " WHERE 1=0"
				table = table[:leftParen]
			}
			colStmt, err := db.PrepareContext(ctx, colQuery)
			if err != nil {
				return 0, fmt.Errorf("failed to prepare query to determine target table columns: %w", err)
			}
			colRows, err := colStmt.QueryContext(ctx)
			if err != nil {
				return 0, fmt.Errorf("failed to execute query to determine target table columns: %w", err)
			}
			columns, err := colRows.Columns()
			if err != nil {
				return 0, fmt.Errorf("failed to fetch target table columns: %w", err)
			}
			clen := len(columns)

			crows := &copyRows{
				rows:   rows,
				values: make([]interface{}, clen),
			}
			for i := 0; i < clen; i++ {
				crows.values[i] = new(interface{})
			}

			var n int64
			err = conn.Raw(func(driverConn interface{}) error {
				conn := driverConn.(*stdlib.Conn).Conn()
				n, err = conn.CopyFrom(ctx, pgx.Identifier{table}, columns, crows)
				return err
			})
			return n, err
		},
	})
}

type copyRows struct {
	rows   *sql.Rows
	values []interface{}
}

func (r *copyRows) Next() bool {
	return r.rows.Next()
}

func (r *copyRows) Values() ([]interface{}, error) {
	err := r.rows.Scan(r.values...)
	return r.values, err
}

func (r *copyRows) Err() error {
	return r.rows.Err()
}
