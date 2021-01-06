package snowflake

import (
	"strconv"

	"github.com/snowflakedb/gosnowflake" // DRIVER: snowflake
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("snowflake", drivers.Driver{
		Err: func(err error) (string, string) {
			if e, ok := err.(*gosnowflake.SnowflakeError); ok {
				return strconv.Itoa(e.Number), e.Message
			}
			return "", err.Error()
		},
	})
}
