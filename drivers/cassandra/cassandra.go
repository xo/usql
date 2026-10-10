// Package cassandra defines and registers usql's Cassandra driver, which is
// github.com/xo/cassandra. The driver reports its own errors and row counts, so
// usql adds only a longer timeout of a request and the check for a wrong
// password.
//
// A batch from BEGIN BATCH to APPLY BATCH is one statement, which the grammar
// of usql reads for every database that dbmeta gives batches.
//
// See: https://github.com/xo/cassandra
// Group: most
package cassandra

import (
	"strings"

	_ "github.com/xo/cassandra" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("cassandra", drivers.Driver{
		// a statement can run long, so each request has five minutes unless the
		// URL sets the timeout key
		ForceParams: func(u *dburl.URL) {
			if q := u.Query(); !q.Has("timeout") {
				q.Set("timeout", "300s")
				u.RawQuery = q.Encode()
			}
		},
		// the driver has no type for it, and gocql words it this way
		IsPasswordErr: func(err error) bool {
			return strings.Contains(err.Error(), "password are incorrect")
		},
	})
}
