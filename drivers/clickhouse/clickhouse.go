package clickhouse

import (
	// DRIVER: clickhouse
	_ "github.com/kshvakov/clickhouse"

	"github.com/knq/usql/drivers"
)

func init() {
	drivers.Register("clickhouse", drivers.Driver{
		AMC: true,
	})
}
