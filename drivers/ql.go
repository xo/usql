// +build ql

package drivers

import (
	_ "github.com/cznic/ql/driver"
)

func init() {
	Drivers["ql"] = "ql"
}
