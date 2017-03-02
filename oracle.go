// +build oracle

package main

import (
	_ "gopkg.in/rana/ora.v4"
)

func init() {
	drivers["ora"] = true
}
