// Package impala defines and registers usql's Apache Impala driver.
//
// See: https://github.com/sclgo/impala-go
package impala

import (
	"errors"

	"github.com/sclgo/impala-go" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("impala", drivers.Driver{
		IsPasswordErr: func(err error) bool {
			var authError *impala.AuthError
			return errors.As(err, &authError)
		},
	})
}
