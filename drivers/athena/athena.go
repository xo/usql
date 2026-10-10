// Package athena defines and registers usql's Amazon Athena driver.
//
// Pre-flight, checked on 2026-10-11:
//
//   - Maintained: github.com/xo/dbimp added the driver in place of github.com/uber/athenadriver.
//     The repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: not by usql, which has no AWS account. dbimp measured it on a hosted account.
//
// The URL is athena://key:secret@host/database?workgroup=wg&output=s3://bucket/results/.
// The server refuses a transaction, so \copy inserts each row without one.
//
// Metadata: dbmeta's Athena model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package athena

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/xo/dbimp/athena" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("athena", drivers.Driver{
		Err: func(err error) (string, string) {
			var e *athena.Error
			if errors.As(err, &e) {
				code := e.Code
				if code == "" {
					code = e.Type
				}
				return code, e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "athena: ")
		},
		// a wrong secret is HTTP 400, and not 401 or 403, with this type
		IsPasswordErr: func(err error) bool {
			var e *athena.Error
			return errors.As(err, &e) && (strings.Contains(e.Type, "InvalidSignature") || strings.Contains(e.Type, "UnrecognizedClient"))
		},
		RowsAffected: drivers.RowsAffectedOrZero,
		// BeginTx fails with dbimp.ErrNotSupported, so each row goes in alone
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			return drivers.FlexibleCopyWithInsert(ctx, db, rows, table, func(int) string { return "?" }, false)
		},
	})
}
