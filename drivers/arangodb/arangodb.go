// Package arangodb defines and registers usql's ArangoDB driver.
//
// Pre-flight, checked on 2026-09-29:
//
//   - Maintained: github.com/xo/dbimp v0.5.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the cursor API of HTTP.
//   - Tested: the vendor's image, which dbmeta's dbrun starts as
//     arangodb-3.12.12.
//
// Metadata: none. All metadata is moving into dbmeta, so usql writes no
// reader for ArangoDB (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package arangodb

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/arangodb" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("arangodb", drivers.Driver{
		AllowMultilineComments: true,
		Process: func(_ *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			// an AQL statement such as FOR ... RETURN is not one that usql
			// knows returns rows, so each statement is run as a query
			typ, _ := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, true, nil
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			if err := db.QueryRowContext(ctx, `RETURN VERSION()`).Scan(&ver); err != nil {
				return "", err
			}
			return "ArangoDB " + ver, nil
		},
		Err: func(err error) (string, string) {
			var e *arangodb.Error
			if errors.As(err, &e) && e.Num != 0 {
				return strconv.Itoa(e.Num), e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "arangodb: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
	})
}
