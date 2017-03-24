// +build firebird

package drivers

import (
	// firebird driver
	_ "github.com/nakagami/firebirdsql"
)

func init() {
	Drivers["firebirdsql"] = "firebird"
}
