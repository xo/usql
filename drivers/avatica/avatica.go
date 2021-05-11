// Package avatica defines and registers usql's Apache Avatica driver.
//
// See: https://github.com/apache/calcite-avatica-go
package avatica

import (
	"strconv"

	_ "github.com/apache/calcite-avatica-go/v5" // DRIVER
	avaticaerrors "github.com/apache/calcite-avatica-go/v5/errors"
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("avatica", drivers.Driver{
		AllowMultilineComments: true,
		AllowCComments:         true,
		Err: func(err error) (string, string) {
			if e, ok := err.(avaticaerrors.ResponseError); ok {
				return strconv.Itoa(int(e.ErrorCode)), e.ErrorMessage
			}
			return "", err.Error()
		},
	})
}
