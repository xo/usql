// +build couchbase

package drivers

import (
	_ "github.com/couchbase/go_n1ql"
)

func init() {
	Drivers["n1ql"] = "couchbase"
}
