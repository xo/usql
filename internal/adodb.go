// +build all,!no_adodb most,!no_adodb adodb,!no_adodb

package internal

import (
	// adodb driver
	_ "github.com/knq/usql/drivers/adodb"
)
