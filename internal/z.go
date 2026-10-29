package internal

import (
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

// init gives the scheme of a default driver, and its aliases, to the other
// driver for the same database when the build has only the other one.
func init() {
	fallback("sqlite3", "moderncsqlite")
	fallback("postgres", "pq")
}

// fallback makes the scheme name, and its aliases, aliases of the scheme to
// when the driver for name is not registered and the driver for to is.
func fallback(name, to string) {
	if drivers.Registered(to) && !drivers.Registered(name) {
		old := dburl.Unregister(name)
		dburl.RegisterAlias(to, name)
		for _, alias := range old.Aliases {
			dburl.RegisterAlias(to, alias)
		}
	}
}
