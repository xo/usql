// Package cockroachdb defines and registers usql's CockroachDB driver.
//
// CockroachDB speaks the PostgreSQL protocol, and dburl opens it through pgx. This
// package registers the pgx driver under the name cockroachdb, which dburl sets as
// the scheme of a cockroachdb:// URL. dbmeta reads its metadata and its version.
//
// See: https://github.com/jackc/pgx
// Group: base
package cockroachdb

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgx"
)

func init() {
	drivers.Register("cockroachdb", pgx.Driver())
}
