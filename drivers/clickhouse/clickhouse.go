package clickhouse

import (
	"database/sql"

	// DRIVER: clickhouse
	_ "github.com/kshvakov/clickhouse"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("clickhouse", drivers.Driver{
		AllowMultilineComments: true,
		RowsAffected: func(sql.Result) (int64, error) {
			return 0, nil
		},
	})
}
