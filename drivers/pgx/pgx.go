package pgx

import (
	// DRIVER: pgx
	_ "github.com/jackc/pgx/stdlib"

	"github.com/jackc/pgx"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("pgx", drivers.Driver{
		AD:  true,
		AMC: true,
		Syn: "postgres",
		V: func(db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRow(`SHOW server_version`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "PostgreSQL " + ver, nil
		},
		ChPw: func(db drivers.DB, user, new, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` PASSWORD '` + new + `'`)
			return err
		},
		E: func(err error) (string, string) {
			if e, ok := err.(pgx.PgError); ok {
				return e.Code, e.Message
			}
			return "", err.Error()
		},
		PwErr: func(err error) bool {
			if e, ok := err.(pgx.PgError); ok {
				return e.Code == "28P01"
			}
			return false
		},
	})
}
