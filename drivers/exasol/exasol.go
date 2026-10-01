// Package exasol defines and registers usql's Exasol driver.
//
// See: https://github.com/exasol/exasol-driver-go
package exasol

import (
	"regexp"

	_ "github.com/exasol/exasol-driver-go" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	errCodeRE := regexp.MustCompile(`^\[([0-9]+)]\s+`)
	drivers.Register("exasol", drivers.Driver{
		Copy: drivers.CopyWithInsert(func(int) string { return "?" }),
		Err: func(err error) (string, string) {
			code, msg := "", err.Error()
			if m := errCodeRE.FindStringSubmatch(msg); m != nil {
				code, msg = m[1], errCodeRE.ReplaceAllString(msg, "")
			}
			return code, msg
		},
	})
}
