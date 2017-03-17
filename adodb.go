// +build adodb

package main

import (
	_ "github.com/mattn/go-adodb"
)

func init() {
	drivers["adodb"] = "adodb"
}
