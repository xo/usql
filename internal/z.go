package internal

import (
	"runtime"

	"github.com/xo/dburl"

	"github.com/xo/usql/drivers"
)

func init() {
	if runtime.GOOS == "windows" {
		// if no odbc driver, but we have adodb, add 'odbc' (and related
		// aliases) as alias for oleodbc
		if drivers.Registered("adodb") && !drivers.Registered("odbc") {
			old := dburl.Unregister("odbc")
			dburl.RegisterAlias("oleodbc", "odbc")
			for _, alias := range old.Aliases {
				dburl.RegisterAlias("oleodbc", alias)
			}
		}
	}
}
