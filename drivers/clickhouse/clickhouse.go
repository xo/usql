// Package clickhouse defines and registers usql's ClickHouse driver.
//
// Group: base
// See: https://github.com/ClickHouse/clickhouse-go
package clickhouse

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("clickhouse", drivers.Driver{
		AllowMultilineComments: true,
		RowsAffected: func(sql.Result) (int64, error) {
			return 0, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, oldpw string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` IDENTIFIED BY '` + newpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(*clickhouse.Exception); ok {
				return strconv.Itoa(int(e.Code)), strings.TrimPrefix(e.Message, "clickhouse: ")
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*clickhouse.Exception); ok {
				return e.Code == 516
			}
			return false
		},
		Copy:              drivers.CopyWithInsert(func(int) string { return "?" }),
		NewMetadataReader: NewMetadataReader,
	})
}
