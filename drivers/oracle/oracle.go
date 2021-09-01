// Package oracle defines and registers usql's Oracle Database driver.
//
// See: https://github.com/sijms/go-ora
// Group: base
package oracle

import (
	"errors"
	"fmt"
	"strings"

	_ "github.com/sijms/go-ora/v2" // DRIVER
	"github.com/xo/usql/drivers/oracle/orshared"
)

func init() {
	orshared.Register(
		"oracle",
		// unwrap error
		func(err error) (string, string) {
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
		func(err error) bool {
			if e := errors.Unwrap(err); e != nil {
				err = e
			}
			return strings.Contains(err.Error(), "empty password")
		},
	)
}
