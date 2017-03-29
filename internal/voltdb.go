// +build all,!no_voltdb most,!no_voltdb voltdb,!no_voltdb

package internal

import (
	// voltdb driver
	_ "github.com/knq/usql/drivers/voltdb"
)
