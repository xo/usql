// Package cratedb defines and registers usql's CrateDB driver.
//
// CrateDB speaks the PostgreSQL protocol, and dburl opens it through pgx by
// its GoDriver. This package registers the pgx driver under the name
// cratedb, which dburl sets on a cratedb:// URL.
//
// Metadata: none. The pgx driver's reader is PostgreSQL's, and CrateDB
// refuses its queries, so this driver drops it. All metadata is moving into
// dbmeta (D3).
//
// See: https://github.com/jackc/pgx
// Group: base
package cratedb

import (
	"context"
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgx"
)

func init() {
	d := pgx.Driver()
	d.NewMetadataReader, d.NewMetadataWriter = nil, nil
	// SHOW server_version gives the PostgreSQL release that CrateDB is
	// compatible with, not its own
	d.Version = func(ctx context.Context, db drivers.DB) (string, error) {
		var ver string
		if err := db.QueryRowContext(ctx, `SELECT version()`).Scan(&ver); err != nil {
			return "", err
		}
		// the text after the number names the build and the platform
		if i := strings.Index(ver, " ("); i != -1 {
			ver = ver[:i]
		}
		return ver, nil
	}
	drivers.Register("cratedb", d)
}
