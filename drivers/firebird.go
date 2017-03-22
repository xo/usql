// +build firebird

package drivers

import (
	_ "github.com/nakagami/firebirdsql"
)

func init() {
	Drivers["firebirdsql"] = "firebird"
}
