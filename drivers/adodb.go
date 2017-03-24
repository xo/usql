// +build adodb

package drivers

import (
	// adodb driver
	_ "github.com/mattn/go-adodb"
)

func init() {
	Drivers["adodb"] = "adodb"
	Drivers["oleodbc"] = "oleodbc"
}
