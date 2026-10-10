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
// Metadata: none. dbmeta has no model for Cosmos DB, so the describe commands report
// that they are not supported (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package cosmos

import (
	"errors"
	"net/http"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/cosmos" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("cosmos", drivers.Driver{
		Err: func(err error) (string, string) {
			var e *cosmos.Error
			if errors.As(err, &e) {
				return e.Code, e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "cosmos: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
	})
}
