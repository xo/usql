// Package libpq defines and registers usql's PostgreSQL lib/pq driver.
//
// It registers as pq, the scheme that dburl gives pq:// and libpq:// URLs.
// lib/pq itself registers as postgres in database/sql, which is the driver
// dburl opens. postgres:// URLs go to the pgx driver.
//
// See: https://github.com/lib/pq
// Group: most
package libpq

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/lib/pq" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

func init() {
	openConn := func(stdout, stderr func() io.Writer, dsn string) (*sql.DB, error) {
		conn, err := pq.NewConnector(dsn)
		if err != nil {
			return nil, err
		}
		noticeConn := pq.ConnectorWithNoticeHandler(conn, func(notice *pq.Error) {
			out := stderr()
			fmt.Fprintln(out, notice.Severity+": ", notice.Message)
			if notice.Hint != "" {
				fmt.Fprintln(out, "HINT: ", notice.Hint)
			}
		})
		notificationConn := pq.ConnectorWithNotificationHandler(noticeConn, func(notification *pq.Notification) {
			var payload string
			if notification.Extra != "" {
				payload = fmt.Sprintf(text.NotificationPayload, notification.Extra)
			}
			fmt.Fprintln(stdout(), fmt.Sprintf(text.NotificationReceived, notification.Channel, payload, notification.BePid))
		})
		return sql.OpenDB(notificationConn), nil
	}
	drivers.Register("pq", drivers.Driver{
		Name: "pq",
		Open: func(ctx context.Context, u *dburl.URL, stdout, stderr func() io.Writer) (func(string, string) (*sql.DB, error), error) {
			return func(_, dsn string) (*sql.DB, error) {
				conn, err := openConn(stdout, stderr, dsn)
				if err != nil {
					return nil, err
				}
				// special retry handling case, since there's no lib/pq retry mode
				if env.Setting("sslmode") == "retry" && !u.Query().Has("sslmode") {
					switch err = conn.PingContext(ctx); {
					case errors.Is(err, pq.ErrSSLNotSupported):
						s := "sslmode=disable " + dsn
						conn, err = openConn(stdout, stderr, s)
						if err != nil {
							return nil, err
						}
						u.DSN = s
					case err != nil:
						return nil, err
					}
				}
				return conn, nil
			}, nil
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(*pq.Error); ok {
				return string(e.Code), e.Message
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*pq.Error); ok {
				return e.Code.Name() == "invalid_password"
			}
			return false
		},
		// lib/pq copies through a prepared COPY statement, which takes a row for
		// each Exec, and an Exec with no arguments ends the stream
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			schema, name, columns, err := drivers.CopyTarget(ctx, db, table)
			if err != nil {
				return 0, err
			}
			target := pq.QuoteIdentifier(name)
			if schema != "" {
				target = pq.QuoteIdentifier(schema) + "." + target
			}
			quoted := make([]string, len(columns))
			for i, c := range columns {
				quoted[i] = pq.QuoteIdentifier(c)
			}
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return 0, fmt.Errorf("failed to begin transaction: %w", err)
			}
			defer tx.Rollback()
			stmt, err := tx.PrepareContext(ctx, "COPY "+target+" ("+strings.Join(quoted, ", ")+") FROM STDIN")
			if err != nil {
				return 0, fmt.Errorf("failed to prepare copy: %w", err)
			}
			defer stmt.Close()
			types, err := rows.ColumnTypes()
			if err != nil {
				return 0, fmt.Errorf("failed to fetch source column types: %w", err)
			}
			src := drivers.NewCopyRows(rows, len(columns))
			for src.Next() {
				values, err := src.Values()
				if err != nil {
					return 0, err
				}
				// lib/pq reads most types as bytes, which a copy writes as bytea
				for i, v := range values {
					if b, ok := v.([]byte); ok && types[i].DatabaseTypeName() != "BYTEA" {
						values[i] = string(b)
					}
				}
				if _, err := stmt.ExecContext(ctx, values...); err != nil {
					return 0, fmt.Errorf("failed to exec copy: %w", err)
				}
			}
			if err := src.Err(); err != nil {
				return 0, err
			}
			res, err := stmt.ExecContext(ctx)
			if err != nil {
				return 0, fmt.Errorf("failed to end copy: %w", err)
			}
			n, err := res.RowsAffected()
			if err != nil {
				return 0, fmt.Errorf("failed to check rows affected: %w", err)
			}
			if err := tx.Commit(); err != nil {
				return n, fmt.Errorf("failed to commit transaction: %w", err)
			}
			return n, nil
		},
	})
}
