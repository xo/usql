// Package csvq defines and registers usql's CSVQ driver.
//
// See: https://github.com/mithrandie/csvq-driver
// Group: base
package csvq

import (
	"context"
	"os"
	"strings"

	"github.com/mithrandie/csvq-driver" // DRIVER
	"github.com/mithrandie/csvq/lib/query"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	csvq.SetStdout(query.NewDiscard())
	drivers.Register("csvq", drivers.Driver{
		AllowMultilineComments: true,
		Process: func(_ *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			typ, q := drivers.QueryExecType(prefix, sqlstr)
			if strings.HasPrefix(prefix, "SHOW") {
				csvq.SetStdout(os.Stdout)
				q = false
			}
			return typ, sqlstr, q, nil
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			if err := db.QueryRowContext(ctx, `SELECT @#VERSION`).Scan(&ver); err != nil {
				return "", err
			}
			return "CSVQ " + ver, nil
		},
		Copy: drivers.CopyWithInsert(func(int) string { return "?" }),
	})
}
