// Package questdb defines and registers usql's QuestDB driver.
//
// QuestDB speaks the PostgreSQL protocol, and dburl opens it through pgx. This
// package registers the pgx driver under the name questdb, which dburl sets as
// the scheme of a questdb:// URL. dbmeta reads its metadata and its version.
//
// See: https://github.com/jackc/pgx
// Group: most
package questdb

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgx"
)

func init() {
	drivers.Register("questdb", pgx.Driver())
}
