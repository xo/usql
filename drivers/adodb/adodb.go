// Package adodb defines and registers usql's ADODB driver. Requires CGO.
//
// See: https://github.com/mattn/go-adodb
package adodb

import (
	"database/sql"

	_ "github.com/mattn/go-adodb" // DRIVER: adodb
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("adodb", drivers.Driver{
		AllowMultilineComments: true,
		AllowCComments:         true,
		RowsAffected: func(res sql.Result) (int64, error) {
			return 0, nil
		},
	}, "oleodbc")
}
