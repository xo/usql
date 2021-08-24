// Package clickhouse defines and registers usql's ClickHouse driver.
//
// See: https://github.com/ClickHouse/clickhouse-go
package clickhouse

import (
	"database/sql"

	_ "github.com/ClickHouse/clickhouse-go" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("clickhouse", drivers.Driver{
		AllowMultilineComments: true,
		RowsAffected: func(sql.Result) (int64, error) {
			return 0, nil
		},
		Copy:              drivers.CopyWithInsert(func(int) string { return "?" }),
		NewMetadataReader: NewMetadataReader,
	})
}
