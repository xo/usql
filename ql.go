// +build ql

package main

import (
	_ "github.com/cznic/ql/driver"
)

func init() {
	drivers["ql"] = "ql"
}
