// Package couchbase defines and registers usql's Couchbase driver.
//
// See: https://github.com/couchbase/go_n1ql
package couchbase

import (
	"context"
	"strconv"
	"strings"

	_ "github.com/couchbase/go_n1ql" // DRIVER: n1ql
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("n1ql", drivers.Driver{
		AllowMultilineComments: true,
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			ver := "<unknown>"
			/*
				var buf []byte
				if err := db.QueryRowContext(ctx, `SELECT ds_version() AS version`).Scan(&buf); err == nil {
					var m map[string]string
					if err := json.Unmarshal(buf, &m); err == nil {
						if s, ok := m["version"]; ok {
							ver = s
						}
					}
				}
			*/
			var v string
			if err := db.QueryRowContext(ctx, `SELECT RAW ds_version()`).Scan(&v); err == nil {
				if s, err := strconv.Unquote(v); err == nil {
					ver = s
				}
			}
			return "Couchbase " + ver, nil
		},
		Err: func(err error) (string, string) {
			return "", strings.TrimPrefix(err.Error(), "N1QL: ")
		},
	})
}
