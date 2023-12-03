// Package dynamodb defines and registers usql's DynamoDb driver.
//
// See: https://github.com/btnguyen2k/godynamo
package dynamodb

import (
	_ "github.com/btnguyen2k/godynamo" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("godynamo", drivers.Driver{})
}
