// +build yql

package drivers

import (
	// yql driver
	_ "github.com/mattn/go-yql"
)

func init() {
	Drivers["yql"] = "yql"
}
