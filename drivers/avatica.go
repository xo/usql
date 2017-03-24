// +build avatica

package drivers

import (
	// avatica driver
	_ "github.com/Boostport/avatica"
)

func init() {
	Drivers["avatica"] = "avatica"
}
