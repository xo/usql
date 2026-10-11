// Package oracle defines and registers usql's Oracle Database driver.
//
// See: https://github.com/sijms/go-ora
// Group: base
// Build: !(linux && arm)
package oracle

import (
	"errors"
	"fmt"
	"strings"

	_ "github.com/sijms/go-ora/v3" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
)

func init() {
	drivers.Register("oracle", drivers.Driver{
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
		// unwrap error
		Err: func(err error) (string, string) {
			if e := errors.Unwrap(err); e != nil {
				err = e
			}
			code, msg := "", err.Error()
			if e, ok := err.(interface {
				Code() int
			}); ok {
				code = fmt.Sprintf("ORA-%05d", e.Code())
			}
			if e, ok := err.(interface {
				Message() string
			}); ok {
				msg = e.Message()
			}
			if i := strings.LastIndex(msg, "ORA-"); msg == "" && i != -1 {
				msg = msg[i:]
				if j := strings.Index(msg, ":"); j != -1 {
					msg = msg[j+1:]
					if code == "" {
						code = msg[i:j]
					}
				}
			}
			return code, strings.TrimSpace(msg)
		},
		// is password error
		IsPasswordErr: func(err error) bool {
			if e := errors.Unwrap(err); e != nil {
				err = e
			}
			return strings.Contains(err.Error(), "empty password")
		},
	})
}
