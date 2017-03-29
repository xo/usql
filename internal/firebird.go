// +build all,!no_firebird most,!no_firebird firebird,!no_firebird

package internal

import (
	// firebird driver
	_ "github.com/knq/usql/drivers/firebird"
)
