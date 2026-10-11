package internal

import (
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	if drivers.Registered("moderncsqlite") && !drivers.Registered("sqlite3") {
		old := dburl.Unregister("sqlite3")
		dburl.RegisterAlias("moderncsqlite", "sqlite3")
		for _, alias := range old.Aliases {
			dburl.RegisterAlias("moderncsqlite", alias)
		}
	}
}
