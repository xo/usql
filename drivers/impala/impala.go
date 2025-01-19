// Package impala defines and registers usql's Apache Impala driver.
//
// See: https://github.com/bippio/go-impala
package impala

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sclgo/impala-go" // DRIVER
	"github.com/xo/usql/drivers"
	meta "github.com/xo/usql/drivers/metadata/impala"
)

func init() {
	drivers.Register("impala", drivers.Driver{
		NewMetadataReader: meta.New,
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			placeholder := func(int) string {
				return "?"
			}
			return drivers.FlexibleCopyWithInsert(ctx, db, rows, table, placeholder, false)
		},
		IsPasswordErr: func(err error) bool {
			var authError *impala.AuthError
			return errors.As(err, &authError)
		},
	})
}
