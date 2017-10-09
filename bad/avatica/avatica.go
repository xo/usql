package avatica

import (
	"strconv"

	// DRIVER: avatica
	"github.com/Boostport/avatica"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("avatica", drivers.Driver{
		AMC: true,
		ACC: true,
		E: func(err error) (string, string) {
			if e, ok := err.(avatica.ResponseError); ok {
				return strconv.Itoa(int(e.ErrorCode)), e.ErrorMessage
			}
			return "", err.Error()
		},
	})
}
