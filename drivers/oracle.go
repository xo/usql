// +build oracle

package drivers

import (
	_ "gopkg.in/rana/ora.v4"
)

func init() {
	Drivers["ora"] = "oracle"
	pwErr["ora"] = func(err error) bool {
		if e, ok := err.(interface {
			Code() int
		}); ok {
			return e.Code() == 1017
		}
		return false
	}
}
