package adodb

import (
	"database/sql"

	// DRIVER: adodb
	_ "github.com/mattn/go-adodb"

	"github.com/knq/usql/drivers"
)

func init() {
	drivers.Register("adodb", drivers.Driver{
		A: func(res sql.Result) (int64, error) {
			return 0, nil
		},
	}, "oleodbc")
}
