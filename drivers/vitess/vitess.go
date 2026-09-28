// Package vitess defines and registers usql's Vitess driver.
//
// Vitess speaks the MySQL protocol, and dburl opens it through the mysql
// driver. This package registers that driver under the name vitess.
//
// See: https://github.com/go-sql-driver/mysql
// Group: base
package vitess

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/mysql"
)

func init() {
	drivers.Register("vitess", mysql.Driver())
}
