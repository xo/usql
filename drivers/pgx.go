// +build pgx

package drivers

import (
	// pgx driver
	"github.com/jackc/pgx/stdlib"

	"database/sql"

	"github.com/jackc/pgx"
	"github.com/knq/dburl"
)

func init() {
	Drivers["pgx"] = "pgx"
}

const (
	pgxMaxConnections = 3
)

// PgxOpen is a special pgx open func.
func PgxOpen(u *dburl.URL) func(string, string) (*sql.DB, error) {
	return func(string, string) (*sql.DB, error) {
		var err error

		u.DSN, err = dburl.GenPostgres(u)
		if err != nil {
			return nil, err
		}

		cfg, err := pgx.ParseDSN(u.DSN)
		if err != nil {
			return nil, err
		}

		pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
			ConnConfig:     cfg,
			MaxConnections: pgxMaxConnections,
		})
		if err != nil {
			return nil, err
		}

		return stdlib.OpenFromConnPool(pool)
	}
}
