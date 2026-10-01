// Package rqlite defines and registers usql's rqlite driver.
//
// Pre-flight, checked on 2026-10-01:
//
//   - Maintained: github.com/xo/dbimp v0.8.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the HTTP API of rqlite.
//   - Tested: the vendor's image, which dbmeta's dbrun starts as
//     rqlite-10.3.6.
//
// rqlite shares one write connection among every client, so the driver has
// no transactions. \copy inserts each row without one.
//
// Metadata: dbmeta's rqlite model, which shares the statements of its sqlite3
// model. dbmeta reads the version too, which is the release of SQLite that
// the server runs, because no statement names the release of rqlite.
//
// See: https://github.com/xo/dbimp
// Group: most
package rqlite

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/rqlite" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("rqlite", drivers.Driver{
		Err: func(err error) (string, string) {
			var e *rqlite.Error
			if errors.As(err, &e) {
				return "", e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "rqlite: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			placeholder := func(int) string {
				return "?"
			}
			return drivers.FlexibleCopyWithInsert(ctx, db, rows, table, placeholder, false)
		},
	})
}
