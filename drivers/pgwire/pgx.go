//go:build (!no_base || pgx || postgres) && !no_pgx && !no_postgres

package pgwire

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/pgx"
)

// Driver returns the pgx driver.
func Driver() drivers.Driver {
	return pgx.Driver()
}
