// Package snowflake defines and registers usql's Snowflake driver.
//
// Pre-flight, checked on 2026-10-10:
//
//   - Maintained: github.com/xo/dbimp v0.16.0 added the driver in place of
//     github.com/snowflakedb/gosnowflake. The repository is in the same family
//     as usql.
//   - Global state: the package's init calls sql.Register and nothing else. The
//     gosnowflake logger, which usql had to silence, is gone with it.
//   - Links: pure Go, over the SQL REST API of Snowflake, with no Arrow.
//   - Tested: not by usql, which has no Snowflake account. dbimp verified it on
//     a trial account, and dbmeta holds the credentials of its own.
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
