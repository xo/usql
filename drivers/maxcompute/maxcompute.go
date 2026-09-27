// Package maxcompute defines and registers usql's Alibaba MaxCompute driver.
//
// See: https://github.com/aliyun/aliyun-odps-go-sdk
// Group: most
package maxcompute

import (
	"database/sql"

	"github.com/aliyun/aliyun-odps-go-sdk/sqldriver" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	// sqldriver registers itself as odps. dburl names the scheme maxcompute,
	// so it is registered under that name as well.
	sql.Register("maxcompute", sqldriver.OdpsDriver{})
	drivers.Register("maxcompute", drivers.Driver{})
}
