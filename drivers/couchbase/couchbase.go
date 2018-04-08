package couchbase

import (
	"strings"

	// DRIVER: n1ql
	_ "github.com/couchbase/go_n1ql"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("n1ql", drivers.Driver{
		AllowMultilineComments: true,
		Err: func(err error) (string, string) {
			return "", strings.TrimPrefix(err.Error(), "N1QL: ")
		},
	})
}
