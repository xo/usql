package postgres

import (
	"database/sql"

	// DRIVER: postgres
	"github.com/lib/pq"

	"github.com/knq/usql/drivers"
)

func init() {
	drivers.Register("postgres", drivers.Driver{
		N: "pq",
		V: func(db *sql.DB) (string, error) {
			var ver string
			err := db.QueryRow(`show server_version`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "PostgreSQL " + ver, nil
		},
		E: func(err error) (string, string) {
			if e, ok := err.(*pq.Error); ok {
				return string(e.Code), e.Message
			}
			return "", err.Error()
		},
		PwErr: func(err error) bool {
			if e, ok := err.(*pq.Error); ok {
				return e.Code.Name() == "invalid_password"
			}
			return false
		},
	}, "cockroachdb")
}
