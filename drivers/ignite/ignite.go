// Package ignite defines and registers usql's Apache Ignite driver.
//
// See: https://github.com/amsokol/ignite-go-client
package ignite

import (
	"strconv"

	"github.com/amsokol/ignite-go-client/binary/errors"
	_ "github.com/amsokol/ignite-go-client/sql" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("ignite", drivers.Driver{
		Err: func(err error) (string, string) {
			if e, ok := err.(*errors.IgniteError); ok {
				return strconv.Itoa(int(e.IgniteStatus)), e.IgniteMessage
			}
			return "", err.Error()
		},
	})
}
