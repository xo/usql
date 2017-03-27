// +build mymysql

package drivers

import (
	// mymysql driver

	_ "github.com/ziutek/mymysql/godrv"
	"github.com/ziutek/mymysql/mysql"
)

func init() {
	Drivers["mymysql"] = "mymysql"

	pwErr["mymysql"] = func(err error) bool {
		if e, ok := err.(*mysql.Error); ok {
			return e.Code == 1045
		}
		return false
	}
}
