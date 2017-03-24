// +build mymysql

package drivers

import (
	// mymysql driver
	_ "github.com/ziutek/mymysql/godrv"
)

func init() {
	Drivers["mymysql"] = "mymysql"
}
