package ignite

import (
	"strconv"

	// DRIVER: ignite
	_ "github.com/amsokol/ignite-go-client/sql"

	"github.com/amsokol/ignite-go-client/binary/errors"

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
