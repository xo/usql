// Package gizmosql defines and registers usql's GizmoSQL driver.
//
// GizmoSQL speaks Arrow Flight SQL, and dburl opens it through the
// flightsql driver. This package registers that driver under the name
// gizmosql.
//
// It is in the bad group. The flightsql driver never keeps the session that
// GizmoSQL issues, so a connection opens and every query fails with "No
// session ID in request context". It needs a driver that keeps the session.
//
// See: https://github.com/apache/arrow-go/tree/main/arrow/flight/flightsql/driver
// Group: bad
package gizmosql

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/flightsql"
)

func init() {
	drivers.Register("gizmosql", flightsql.Driver())
}
