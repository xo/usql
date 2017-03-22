// +build mymysql

package drivers

import (
	_ "github.com/ziutek/mymysql/godrv"
)

func init() {
	Drivers["mymysql"] = "mymysql"
}
