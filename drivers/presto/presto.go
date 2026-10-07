// Package presto defines and registers usql's Presto driver.
//
// Presto speaks the same protocol as Trino, so dburl opens a presto:// URL
// through dbimp's trino driver, which replaced
// github.com/prestodb/presto-go-client. This package registers it under the
// name presto, which dburl sets as the scheme. dburl writes flavor=presto,
// so the driver does not ask the server which product it is.
//
// See: https://github.com/xo/dbimp
// Group: base
package presto

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
	drivers.Register("presto", drivers.Driver{
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
