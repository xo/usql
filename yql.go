// +build yql

package main

import (
	_ "github.com/mattn/go-yql"
)

func init() {
	drivers["yql"] = "yql"
}
