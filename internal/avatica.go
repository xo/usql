// +build all,!no_avatica most,!no_avatica avatica,!no_avatica

package internal

import (
	// avatica driver
	_ "github.com/knq/usql/drivers/avatica"
)
