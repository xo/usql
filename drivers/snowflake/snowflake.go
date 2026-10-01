// Package snowflake defines and registers usql's Snowflake driver.
//
// See: https://github.com/snowflakedb/gosnowflake
package snowflake

import (
	"io"
	"strconv"

	"github.com/snowflakedb/gosnowflake/v2" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	gosnowflake.GetLogger().SetOutput(io.Discard)
	drivers.Register("snowflake", drivers.Driver{
		Err: func(err error) (string, string) {
			if e, ok := err.(*gosnowflake.SnowflakeError); ok {
				return strconv.Itoa(e.Number), e.Message
			}
			return "", err.Error()
		},
	})
}
