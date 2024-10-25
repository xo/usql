// Package ydb defines and registers usql's YDB driver.
//
// See: https://github.com/ydb-platform/ydb-go-sdk
package ydb

import (
	"context"
	"errors"
	"strconv"

	"github.com/xo/usql/drivers"
	"github.com/ydb-platform/ydb-go-sdk/v3" // DRIVER
)

func init() {
	drivers.Register("ydb", drivers.Driver{
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			if err := db.QueryRowContext(ctx, `SELECT '<unknown>' AS version`).Scan(&ver); err != nil {
				return "", err
			}
			return "YDB " + ver, nil
		},
		Err: func(err error) (string, string) {
			var e ydb.Error
			if errors.As(err, &e) {
				return strconv.Itoa(int(e.Code())), e.Error()
			}
			return "", err.Error()
		},
	})
}
