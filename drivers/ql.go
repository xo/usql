// +build ql

package drivers

import (
	// ql driver
	_ "github.com/cznic/ql/driver"
)

func init() {
	Drivers["ql"] = "ql"
}
