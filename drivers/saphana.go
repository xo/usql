// +build saphana

package drivers

import (
	// saphana driver
	_ "github.com/SAP/go-hdb/driver"
)

func init() {
	Drivers["hdb"] = "saphana"
}
