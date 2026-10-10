// Package databricks defines and registers usql's Databricks driver.
//
// Pre-flight, checked on 2026-10-11:
//
//   - Maintained: github.com/xo/dbimp added the driver in place of github.com/databricks/databricks-sql-go.
//     The repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: not by usql, which has no Databricks account. dbimp measured it on a hosted account.
//
// A wrong token is HTTP 403 with the message "Invalid access token", and not 401.
// The API has no session, so \copy inserts each row without a transaction.
//
// Metadata: dbmeta's Databricks model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package databricks

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/databricks" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("databricks", drivers.Driver{
		Err: func(err error) (string, string) {
			var e *databricks.Error
			if errors.As(err, &e) {
				if e.SQLState != "" {
					return e.SQLState, e.Message
				}
				return e.Code, e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "databricks: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			if errors.As(err, &se) && se.Code == http.StatusUnauthorized {
				return true
			}
			var e *databricks.Error
			return errors.As(err, &e) && strings.Contains(e.Message, "Invalid access token")
		},
		RowsAffected: drivers.RowsAffectedOrZero,
		// BeginTx fails with dbimp.ErrNotSupported, so each row goes in alone
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			return drivers.FlexibleCopyWithInsert(ctx, db, rows, table, func(int) string { return "?" }, false)
		},
	})
}
