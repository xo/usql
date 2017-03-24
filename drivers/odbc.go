// +build odbc

package drivers

import (
	// odbc driver
	_ "github.com/alexbrainman/odbc"
)

func init() {
	Drivers["odbc"] = "odbc"
}
