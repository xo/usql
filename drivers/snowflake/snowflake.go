// Package snowflake defines and registers usql's Snowflake driver.
//
// The driver is dbimp's: pure Go, over the SQL REST API of Snowflake, with no
// Arrow. It sets no global state, and its init calls sql.Register and nothing
// else.
//
// The URL is snowflake://user:key@host/database/schema?role=r. The password is
// the private key, as base64url PKCS8 DER, and the query takes role, warehouse,
// timeout with a unit, and timezone. dburl adds .snowflakecomputing.com to an
// account name such as org-account.
//
// The server has no transaction across requests, so BeginTx fails and \copy
// inserts each row without one.
//
// Metadata: dbmeta's Snowflake model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package snowflake

import (
	_ "github.com/xo/dbimp/snowflake" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("snowflake", drivers.Driver{})
}
