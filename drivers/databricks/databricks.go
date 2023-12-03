// Package databricks defines and registers usql's Databricks driver.
//
// See: https://github.com/databricks/databricks-sql-go
package databricks

import (
	"errors"

	_ "github.com/databricks/databricks-sql-go" // DRIVER
	dberrs "github.com/databricks/databricks-sql-go/errors"
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("databricks", drivers.Driver{
		Err: func(err error) (string, string) {
			var e dberrs.DBExecutionError
			if errors.As(err, &e) {
				return e.SqlState(), e.Error()
			}
			return "", err.Error()
		},
	})
}
