// Package questdb defines and registers usql's QuestDB driver.
//
// QuestDB speaks the PostgreSQL protocol, and dburl opens it through pgx.
// This package registers the pgx driver under the name questdb.
//
// Metadata: none. The pgx driver's reader is PostgreSQL's, so this driver
// drops it. All metadata is moving into dbmeta (D3).
//
// See: https://github.com/jackc/pgx
// Group: most
package questdb

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgx"
)

func init() {
	d := pgx.Driver()
	d.NewMetadataReader, d.NewMetadataWriter = nil, nil
	drivers.Register("questdb", d)
}
