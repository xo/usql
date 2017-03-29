// +build all,!no_mymysql most,!no_mymysql mymysql,!no_mymysql

package internal

import (
	// mymysql driver
	_ "github.com/knq/usql/drivers/mymysql"
)
