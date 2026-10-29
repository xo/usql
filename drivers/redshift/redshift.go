// Package redshift defines and registers usql's Amazon Redshift driver.
//
// Redshift speaks the PostgreSQL protocol. This package registers the driver
// of pgwire, which is pgx, or libpq in a build with no pgx, under the name
// redshift. It leaves out the native copy, because Redshift has no COPY FROM
// STDIN: \copy inserts each row.
//
// See: https://github.com/jackc/pgx
// Group: base
package redshift

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgwire"
)

func init() {
	d := pgwire.Driver()
	d.Copy = nil
	drivers.Register("redshift", d)
}
