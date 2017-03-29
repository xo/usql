// +build all,!no_couchbase most,!no_couchbase couchbase,!no_couchbase

package internal

import (
	// couchbase driver
	_ "github.com/knq/usql/drivers/couchbase"
)
