// +build odbc

package main

import (
	_ "github.com/alexbrainman/odbc"
)

func init() {
	drivers["odbc"] = "odbc"
}
