// +build oracle

package drivers

import (
	_ "gopkg.in/rana/ora.v4"
)

func init() {
	Drivers["ora"] = "oracle"
}
