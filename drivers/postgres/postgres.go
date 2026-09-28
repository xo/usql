// Package postgres defines and registers usql's PostgreSQL driver.
//
// dburl opens postgres:// URLs through the pgx driver, and this package
// registers it under the name postgres, which dburl sets as the scheme.
//
// See: https://github.com/jackc/pgx
// Group: base
package postgres

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgx"
)

func init() {
	drivers.Register("postgres", pgx.Driver())
}
