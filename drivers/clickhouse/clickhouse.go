// Package clickhouse defines and registers usql's ClickHouse driver.
//
// The driver is dbimp's, which speaks HTTP only, on port 8123 or, with
// tls=true, 8443. It replaced github.com/ClickHouse/clickhouse-go in W42, at
// Ken's decision, and takes the native protocol and its port 9000 with it.
// It refuses every key but tls, such as the secure=true of clickhouse-go.
//
// ClickHouse has no transactions over HTTP, so \copy inserts each row
// without one.
//
// See: https://github.com/xo/dbimp
// Group: base
package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/clickhouse" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("clickhouse", drivers.Driver{
		// a statement such as CREATE TABLE sends no count of rows affected,
		// which the driver reports as ErrNotSupported
		RowsAffected: drivers.RowsAffectedOrZero,
		Err: func(err error) (string, string) {
			var e *clickhouse.Error
			if errors.As(err, &e) && e.Code != 0 {
				return strconv.Itoa(e.Code), e.Message
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			var e *clickhouse.Error
			if errors.As(err, &e) && (e.Code == 194 || e.Code == 516) {
				return true
			}
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			return drivers.FlexibleCopyWithInsert(ctx, db, rows, table, func(int) string { return "?" }, false)
		},
	})
}
