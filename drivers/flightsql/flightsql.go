// Package flightsql defines and registers usql's FlightSQL driver.
//
// See: https://github.com/apache/arrow/tree/main/go/arrow/flight/flightsql/driver
package flightsql

import (
	"context"

	_ "github.com/apache/arrow/go/v12/arrow/flight/flightsql/driver" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("flightsql", drivers.Driver{
		AllowMultilineComments: true,
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(
				ctx,
				`SELECT version()`,
			).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "FlightSQL " + ver, nil
		},
	})
}
