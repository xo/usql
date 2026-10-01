// Package couchbase defines and registers usql's Couchbase driver.
//
// See: https://github.com/xo/dbimp
// Group: most
package couchbase

import (
	"errors"
	"strconv"
	"strings"

	"github.com/xo/dbimp/couchbase" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("couchbase", drivers.Driver{
		ForceParams: func(u *dburl.URL) {
			// the server ends a transaction after 15 seconds by default,
			// which is too short for a person typing into one
			if q := u.Query(); q.Get("txtimeout") == "" {
				q.Set("txtimeout", "30m")
				u.RawQuery = q.Encode()
			}
		},
		Err: func(err error) (string, string) {
			// a response can hold several errors, and only a single one is
			// reported by its code
			var re *couchbase.ResponseError
			if errors.As(err, &re) && len(re.Errs) == 1 {
				return strconv.Itoa(re.Errs[0].Code), re.Errs[0].Msg
			}
			var e couchbase.Error
			if re == nil && errors.As(err, &e) {
				return strconv.Itoa(e.Code), e.Msg
			}
			return "", strings.TrimPrefix(err.Error(), "couchbase: ")
		},
	})
}
