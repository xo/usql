// Package cockroachdb defines and registers usql's CockroachDB driver.
//
// CockroachDB speaks the PostgreSQL protocol, and dburl opens it through pgx
// by its GoDriver. This package registers the pgx driver under the name
// cockroachdb, which dburl sets on a cockroachdb:// URL.
//
// See: https://github.com/jackc/pgx
// Group: base
package cockroachdb

import (
	"context"
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgx"
)

func init() {
	d := pgx.Driver()
	// SHOW server_version gives the PostgreSQL release that CockroachDB is
	// compatible with, not its own
	d.Version = func(ctx context.Context, db drivers.DB) (string, error) {
		var ver string
		if err := db.QueryRowContext(ctx, `SELECT version()`).Scan(&ver); err != nil {
			return "", err
		}
		// the text after the number names the platform and the build
		if i := strings.Index(ver, " ("); i != -1 {
			ver = ver[:i]
		}
		return ver, nil
	}
	drivers.Register("cockroachdb", d)
}
