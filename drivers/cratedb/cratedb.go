// Package cratedb defines and registers usql's CrateDB driver.
//
// CrateDB speaks the PostgreSQL protocol. This package registers the driver
// of pgwire, which is pgx, or libpq in a build with no pgx, under the name
// cratedb, which dburl sets as the scheme of a cratedb:// URL.
// dbmeta reads its metadata and its version.
//
// See: https://github.com/jackc/pgx
// Group: base
package cratedb

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgwire"
)

func init() {
	drivers.Register("cratedb", pgwire.Driver())
}
