// +build voltdb

package drivers

import (
	_ "github.com/VoltDB/voltdb-client-go/voltdbclient"
)

func init() {
	Drivers["voltdb"] = "voltdb"
}
