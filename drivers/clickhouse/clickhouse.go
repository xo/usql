// Package clickhouse defines and registers usql's ClickHouse driver.
//
// See: https://github.com/ClickHouse/clickhouse-go
package clickhouse

import (
	"database/sql"

	_ "github.com/ClickHouse/clickhouse-go" // DRIVER: clickhouse
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/informationschema"
)

func init() {
	drivers.Register("clickhouse", drivers.Driver{
		AllowMultilineComments: true,
		RowsAffected: func(sql.Result) (int64, error) {
			return 0, nil
		},
		NewMetadataReader: informationschema.New,
	})
}
