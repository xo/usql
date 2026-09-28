// Package memsql defines and registers usql's SingleStore MemSQL driver.
//
// SingleStore speaks the MySQL protocol, and dburl opens it through the
// mysql driver. This package registers that driver under the name memsql.
//
// See: https://github.com/go-sql-driver/mysql
// Group: base
package memsql

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/mysql"
)

func init() {
	drivers.Register("memsql", mysql.Driver())
}
