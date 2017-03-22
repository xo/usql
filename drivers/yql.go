// +build yql

package drivers

import (
	_ "github.com/mattn/go-yql"
)

func init() {
	Drivers["yql"] = "yql"
}
