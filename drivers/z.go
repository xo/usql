package drivers

import (
	"runtime"

	"github.com/knq/dburl"
)

func init() {
	if runtime.GOOS == "windows" {
		// if no odbc driver, but we have adodb, add 'odbc' (and related
		// aliases) as oleodbc alias
		_, haveADODB := Drivers["adodb"]
		_, haveODBC := Drivers["odbc"]
		if haveADODB && !haveODBC {
			old := dburl.Unregister("odbc")
			dburl.RegisterAlias("oleodbc", "odbc")
			for _, alias := range old.Aliases {
				dburl.RegisterAlias("oleodbc", alias)
			}
		}
	}
}
