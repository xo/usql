// +build couchbase

package main

import (
	_ "github.com/couchbase/go_n1ql"
)

func init() {
	drivers["n1ql"] = "couchbase"
}
