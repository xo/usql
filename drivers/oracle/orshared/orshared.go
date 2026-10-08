// Package orshared contains shared a shared driver implementation for the
// Oracle Database. Used by the Oracle driver.
package orshared

import (
	"fmt"
	"strings"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
)

// Register registers an oracle driver.
func Register(name string, err func(error) (string, string), isPasswordErr func(error) bool) {
	drivers.Register(name, drivers.Driver{
		ForceParams: func(u *dburl.URL) {
			// if the service name is not specified, use the environment
			// variable if present
			if strings.TrimPrefix(u.Path, "/") == "" {
				if n, ok := env.Getenv("ORACLE_SID", "ORASID"); ok && n != "" {
					u.Path = "/" + n
					if u.Host == "" {
						u.Host = "localhost"
					}
				}
			}
		},
		Err:           err,
		IsPasswordErr: isPasswordErr,
		Copy: drivers.CopyWithInsert(func(n int) string {
			return fmt.Sprintf(":%d", n)
		}),
	})
}
