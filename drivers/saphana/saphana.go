// Package saphana defines and registers usql's SAP HANA driver.
//
// See: https://github.com/SAP/go-hdb
package saphana

import (
	"context"
	"strconv"

	_ "github.com/SAP/go-hdb/driver" // DRIVER: hdb
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("hdb", drivers.Driver{
		AllowMultilineComments: true,
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			if err := db.QueryRowContext(ctx, `SELECT version FROM m_database`).Scan(&ver); err != nil {
				return "", err
			}
			return "SAP HANA " + ver, nil
		},
		Err: func(err error) (string, string) {
			code, msg := "", err.Error()
			if e, ok := err.(interface {
				Code() int
			}); ok {
				code = strconv.Itoa(e.Code())
			}
			if e, ok := err.(interface {
				Text() string
			}); ok {
				msg = e.Text()
			}
			return code, msg
		},
	})
}
