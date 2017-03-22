// +build saphana

package drivers

import (
	_ "github.com/SAP/go-hdb/driver"
)

func init() {
	Drivers["hdb"] = "saphana"
}
