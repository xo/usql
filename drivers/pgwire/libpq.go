//go:build !((!no_drivers || pgx || postgres) && !no_pgx && !no_postgres)

package pgwire

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/libpq"
)

// Driver returns the libpq driver, because the build has no pgx.
func Driver() drivers.Driver {
	return libpq.Driver()
}
