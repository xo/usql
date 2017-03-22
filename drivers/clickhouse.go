// +build clickhouse

package drivers

import (
	_ "github.com/kshvakov/clickhouse"
)

func init() {
	Drivers["clickhouse"] = "clickhouse"
}
