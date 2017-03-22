// +build adodb

package drivers

import (
	_ "github.com/mattn/go-adodb"
)

func init() {
	Drivers["adodb"] = "adodb"
	Drivers["oleodbc"] = "oleodbc"
}
