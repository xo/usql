// Package cosmos defines and registers usql's Azure Cosmos DB driver.
//
// Pre-flight, checked on 2026-10-11:
//
//   - Maintained: github.com/xo/dbimp added the driver in place of github.com/btnguyen2k/gocosmos.
//     The repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the emulator and a hosted account, by dbimp. usql has no Azure account.
//
// The driver reads only: Exec fails with dbimp.ErrNotSupported, and so does BeginTx.
// A wrong key is HTTP 401 on a hosted account, and the emulator does not check it.
//
// Metadata: dbmeta's Cosmos DB model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package cosmos

import (
	_ "github.com/xo/dbimp/cosmos" // DRIVER

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("cosmos", drivers.Driver{})
}
