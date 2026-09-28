// Package redshift defines and registers usql's Amazon Redshift driver.
//
// Redshift speaks the PostgreSQL protocol, and dburl opens it through pgx.
// This package registers the pgx driver under the name redshift.
//
// See: https://github.com/jackc/pgx
// Group: base
package redshift

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgx"
)

func init() {
	drivers.Register("redshift", pgx.Driver())
}
