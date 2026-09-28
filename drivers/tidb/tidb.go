// Package tidb defines and registers usql's TiDB driver.
//
// TiDB speaks the MySQL protocol, and dburl opens it through the mysql
// driver. This package registers that driver under the name tidb.
//
// See: https://github.com/go-sql-driver/mysql
// Group: base
package tidb

import (
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/mysql"
)

func init() {
	drivers.Register("tidb", mysql.Driver())
}
