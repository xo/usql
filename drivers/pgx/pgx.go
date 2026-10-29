// Package pgx defines and registers usql's PostgreSQL PGX driver.
//
// dburl opens pgx:// and postgres:// URLs with it, so it registers under both
// schemes. When the build has no pgx, internal gives the postgres scheme to
// the libpq driver. The redshift, cockroachdb, cratedb and questdb drivers
// register the same driver under their own names, through pgwire.
//
// See: https://github.com/jackc/pgx
// Group: base
// Tags: postgres
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
	"github.com/xo/usql/drivers/pgwire/pgshared"
	"github.com/xo/usql/text"
)

func init() {
	drivers.Register("pgx", Driver())
	drivers.Register("postgres", Driver())
}

// Driver returns the pgx driver. dburl opens a database that speaks the
// PostgreSQL protocol, such as CockroachDB, through pgx, which it names in
// URL.Driver, and usql registers this driver under that database's own name.
func Driver() drivers.Driver {
	return drivers.Driver{
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
		ColumnTypes: pgshared.Dest,
		Err: func(err error) (string, string) {
			var e *pgconn.PgError
			if errors.As(err, &e) {
				return e.Code, e.Message
			}
			return "", err.Error()
		},
		// class 28 is a refused authorization: PostgreSQL says 28P01, and Redshift
		// says 28000
		IsPasswordErr: func(err error) bool {
			var e *pgconn.PgError
			return errors.As(err, &e) && strings.HasPrefix(e.Code, "28")
		},
		// pgx copies through CopyFrom, on the connection that it holds
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			schema, name, columns, err := drivers.CopyTarget(ctx, db, table)
			if err != nil {
				return 0, err
			}
			ident := pgx.Identifier{name}
			if schema != "" {
				ident = pgx.Identifier{schema, name}
			}
			conn, err := db.Conn(ctx)
			if err != nil {
				return 0, fmt.Errorf("failed to get a connection from pool: %w", err)
			}
			defer conn.Close()
			var n int64
			err = conn.Raw(func(driverConn any) error {
				var err error
				n, err = driverConn.(*stdlib.Conn).Conn().CopyFrom(ctx, ident, columns, drivers.NewCopyRows(rows, len(columns)))
				return err
			})
			return n, err
		},
	}
}
