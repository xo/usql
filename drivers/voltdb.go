// +build voltdb

package drivers

import (
	// voltdb driver
	_ "github.com/VoltDB/voltdb-client-go/voltdbclient"
)

func init() {
	Drivers["voltdb"] = "voltdb"
}
