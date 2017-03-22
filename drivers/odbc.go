// +build odbc

package drivers

import (
	_ "github.com/alexbrainman/odbc"
)

func init() {
	Drivers["odbc"] = "odbc"
}
