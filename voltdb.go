// +build voltdb

package main

import (
	_ "github.com/VoltDB/voltdb-client-go/voltdbclient"
)

func init() {
	drivers["voltdb"] = "voltdb"
}
