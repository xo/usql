// +build saphana

package main

import (
	_ "github.com/SAP/go-hdb/driver"
)

func init() {
	drivers["hdb"] = "saphana"
}
