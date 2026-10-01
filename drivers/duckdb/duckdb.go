// Package duckdb defines and registers usql's DuckDB driver. Requires CGO.
//
// See: https://github.com/duckdb/duckdb-go
// Group: base
// Build: !(linux && arm)
package duckdb

import (
	_ "github.com/duckdb/duckdb-go/v2" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("duckdb", drivers.Driver{
		Copy: drivers.CopyWithInsert(func(int) string { return "?" }),
	})
}
