// +build all,!no_pgx most,!no_pgx pgx,!no_pgx

package internal

import (
	// pgx driver
	_ "github.com/knq/usql/drivers/pgx"
)
