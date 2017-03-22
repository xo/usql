// +build avatica

package drivers

import (
	_ "github.com/Boostport/avatica"
)

func init() {
	Drivers["avatica"] = "avatica"
}
