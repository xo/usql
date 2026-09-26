// Package hive defines and registers usql's Apache Hive driver.
//
// See: https://github.com/beltran/gohive
// Group: most
package hive

import (
	_ "github.com/beltran/gohive/v2" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("hive", drivers.Driver{
		ForceParams: func(u *dburl.URL) {
			if u.User != nil && u.Query().Get("auth") == "" {
				drivers.ForceQueryParameters([]string{"auth", "PLAIN"})(u)
			}
		},
	})
}
