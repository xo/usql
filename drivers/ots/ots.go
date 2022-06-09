// Package ots defines and registers usql's Alibaba Tablestore driver.
//
// See: https://github.com/aliyun/aliyun-tablestore-go-sql-driver
package ots

import (
	_ "github.com/aliyun/aliyun-tablestore-go-sql-driver" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("ots", drivers.Driver{})
}
