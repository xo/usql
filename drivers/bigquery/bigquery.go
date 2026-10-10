// Package bigquery defines and registers usql's Google BigQuery driver.
//
// Pre-flight, checked on 2026-10-11:
//
//   - Maintained: github.com/xo/dbimp added the driver in place of gorm.io/driver/bigquery.
//     The repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the emulator, which dbmeta's dbrun starts, by dbimp. usql has no Google account.
//
// The user of the URL is ignored. A wrong token is HTTP 401. The service refuses a
// transaction, so \copy inserts each row without one.
//
// Metadata: dbmeta's BigQuery model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package bigquery

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/bigquery" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("bigquery", drivers.Driver{
		Err: func(err error) (string, string) {
			var e *bigquery.Error
			if errors.As(err, &e) {
				return e.Reason, e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "bigquery: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
		RowsAffected: drivers.RowsAffectedOrZero,
		// BeginTx fails with dbimp.ErrNotSupported, so each row goes in alone
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			return drivers.FlexibleCopyWithInsert(ctx, db, rows, table, func(int) string { return "?" }, false)
		},
	})
}
