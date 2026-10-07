// Package trino defines and registers usql's Trino driver.
//
// The driver is dbimp's, which replaced github.com/trinodb/trino-go-client.
// It serves Presto too. The URL is trino://user:password@host:port/catalog/schema,
// and dburl writes the user "user" when the URL has none, because the
// servers need one.
//
// See: https://github.com/xo/dbimp
// Group: base
package trino

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/trino" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("trino", drivers.Driver{
		// a statement such as CREATE TABLE sends no count of rows affected,
		// which the driver reports as ErrNotSupported
		RowsAffected: drivers.RowsAffectedOrZero,
		Err: func(err error) (string, string) {
			var e *trino.Error
			if errors.As(err, &e) && e.Code != 0 {
				return strconv.Itoa(e.Code), e.Message
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
		// many connectors take a write only with autocommit, so \copy inserts
		// each row without a transaction
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			return drivers.FlexibleCopyWithInsert(ctx, db, rows, table, func(int) string { return "?" }, false)
		},
	})
}
