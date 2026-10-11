// Package csvq defines and registers usql's CSVQ driver.
//
// See: https://github.com/mithrandie/csvq-driver
// Group: most
package csvq

import (
	"context"

	"github.com/mithrandie/csvq-driver" // DRIVER
	"github.com/mithrandie/csvq/lib/query"
	"github.com/xo/usql/drivers"
)

func init() {
	csvq.SetStdout(query.NewDiscard())
	drivers.Register("csvq", drivers.Driver{
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			if err := db.QueryRowContext(ctx, `SELECT @#VERSION`).Scan(&ver); err != nil {
				return "", err
			}
			return "CSVQ " + ver, nil
		},
	})
}
