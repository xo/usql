// Package godror defines and registers usql's GO DRiver for ORacle driver.
// Requires CGO. Uses Oracle's ODPI-C (instant client) library.
//
// See: https://github.com/godror/godror
// Group: all
package godror

import (
	"errors"
	"fmt"
	"strings"

	_ "github.com/godror/godror" // DRIVER
	"github.com/xo/usql/drivers/oracle/orshared"
)

func init() {
	orshared.Register(
		"godror",
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
			if e, ok := err.(interface {
				Code() int
			}); ok {
				return e.Code() == 1017 || e.Code() == 1005
			}
			return false
		},
	)
}
