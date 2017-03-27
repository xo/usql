// +build odbc

package drivers

import (
	"strings"

	"github.com/alexbrainman/odbc"
)

func init() {
	Drivers["odbc"] = "odbc"

	pwErr["odbc"] = func(err error) bool {
		if e, ok := err.(*odbc.Error); ok {
			msg := strings.ToLower(e.Error())
			return strings.Contains(msg, "failed") &&
				(strings.Contains(msg, "login") ||
					strings.Contains(msg, "authentication") ||
					strings.Contains(msg, "password"))
		}
		return false
	}
}
