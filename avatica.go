// +build avatica

package main

import (
	_ "github.com/Boostport/avatica"
)

func init() {
	drivers["avatica"] = "avatica"
}
