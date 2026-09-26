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
			// The driver panics on an auth value it does not recognize, and it
			// does not default an empty one, so an absent auth panics as surely
			// as a wrong one. Always supply a value.
			//
			// NONE is the value that means username and password over SASL.
			// The driver's own PLAIN is the SASL mechanism it uses for NONE,
			// LDAP and CUSTOM, and is not accepted here.
			if u.Query().Get("auth") == "" {
				drivers.ForceQueryParameters([]string{"auth", "NONE"})(u)
			}
		},
	})
}
