// Package pgx defines and registers usql's PostgreSQL PGX driver.
//
// See: https://github.com/jackc/pgx
package pgx

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib" // DRIVER: pgx
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("pgx", drivers.Driver{
		AllowDollar:            true,
		AllowMultilineComments: true,
		LexerName:              "postgres",
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
	})
}
