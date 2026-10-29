// Package questdb defines and registers usql's QuestDB driver.
//
// QuestDB speaks the PostgreSQL protocol. This package registers the driver
// of pgwire, which is pgx, or libpq in a build with no pgx, under the name
// questdb, which dburl sets as the scheme of a questdb:// URL.
// dbmeta reads its metadata and its version.
//
// See: https://github.com/jackc/pgx
// Group: most
package questdb

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgwire"
)

func init() {
	drivers.Register("questdb", pgwire.Driver())
}
