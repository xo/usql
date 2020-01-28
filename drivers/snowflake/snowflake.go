package snowflake

import (
	"database/sql"
	"strconv"
	"strings"

	// DRIVER: snowflake
	"github.com/snowflakedb/gosnowflake"

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
		RowsAffected: func(res sql.Result) (int64, error) {
			count, err := res.RowsAffected()
			switch {
			case err != nil && strings.TrimSpace(err.Error()) == "no RowsAffected available after DDL statement":
				return 0, nil
			case err != nil:
				return 0, err
			}
			return count, nil
		},
	})
}
