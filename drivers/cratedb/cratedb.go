// Package cratedb defines and registers usql's CrateDB driver.
//
// CrateDB speaks the PostgreSQL protocol, and dburl opens it through pgx. This
// package registers the pgx driver under the name cratedb, which dburl sets as
// the scheme of a cratedb:// URL. dbmeta reads its metadata and its version.
//
// See: https://github.com/jackc/pgx
// Group: base
package cratedb

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgx"
)

func init() {
	drivers.Register("cratedb", pgx.Driver())
}
