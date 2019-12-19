package avatica

import (
	"strconv"

	// DRIVER: avatica
	_ "github.com/apache/calcite-avatica-go/v4"

	avaticaerrors "github.com/apache/calcite-avatica-go/v4/errors"

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
