// +build clickhouse

package main

import (
	_ "github.com/kshvakov/clickhouse"
)

func init() {
	drivers["clickhouse"] = "clickhouse"
}
