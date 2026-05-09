// Package impala defines and registers usql's Apache Impala driver.
//
// See: https://github.com/bippio/go-impala
package impala

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/sclgo/impala-go" // DRIVER
	"github.com/xo/dburl"
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
		Process: func(url *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			prefix = strings.ToUpper(prefix)
			if strings.HasPrefix(prefix, "SET") && len(strings.Split(prefix, " ")) < 3 {
				// SET and SET ALL are queries that list option values for current session
				return "SET", sqlstr, true, nil
			}
			// fallback to regular drivers.Process
			typ, isQuery := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, isQuery, nil
		},
		ForceParams: func(u *dburl.URL) {
			if strings.ToLower(u.Query().Get("reuse-session")) != "false" {
				drivers.ForceQueryParameters([]string{"reuse-session", "true"})(u)
			}
		},
	})
}
