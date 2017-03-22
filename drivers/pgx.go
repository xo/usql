// +build pgx

package drivers

import (
	_ "github.com/jackc/pgx/stdlib"
)

func init() {
	Drivers["pgx"] = "pgx"
}
