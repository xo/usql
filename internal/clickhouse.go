// +build all,!no_clickhouse most,!no_clickhouse clickhouse,!no_clickhouse

package internal

import (
	// clickhouse driver
	_ "github.com/knq/usql/drivers/clickhouse"
)
