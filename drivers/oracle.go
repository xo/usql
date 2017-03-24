// +build oracle

package drivers

import (
	// oracle driver
	_ "gopkg.in/rana/ora.v4"
)

func init() {
	Drivers["ora"] = "oracle"
}
