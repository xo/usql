// Package cockroachdb defines and registers usql's CockroachDB driver.
//
// CockroachDB speaks the PostgreSQL protocol. This package registers the driver
// of pgwire, which is pgx, or libpq in a build with no pgx, under the name
// cockroachdb, which dburl sets as the scheme of a cockroachdb:// URL.
// dbmeta reads its metadata and its version.
//
// See: https://github.com/jackc/pgx
// Group: base
package cockroachdb

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgwire"
)

func init() {
	drivers.Register("cockroachdb", pgwire.Driver())
}
