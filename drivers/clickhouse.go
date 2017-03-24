// +build clickhouse

package drivers

import (
	// clickhouse driver
	_ "github.com/kshvakov/clickhouse"
)

func init() {
	Drivers["clickhouse"] = "clickhouse"
}
