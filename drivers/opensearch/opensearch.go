// Package opensearch defines and registers usql's OpenSearch driver.
//
// Pre-flight, checked on 2026-10-07:
//
//   - Maintained: github.com/xo/dbimp v0.14.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the vendor's images, which dbmeta's dbrun starts as opensearch-2.19.6
//     and opensearch-3.9.0.
//
// OpenSearch has its own driver, apart from Elasticsearch (dbimp's D162).
//
// Metadata: none. dbmeta has no model for OpenSearch yet, so the describe commands
// report that they are not supported (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package opensearch

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/opensearch" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("opensearch", drivers.Driver{
		// SELECT VERSION() fails on every release, and GET / is refused to the ordinary user
		Version: func(context.Context, drivers.DB) (string, error) {
			return "OpenSearch", nil
		},
		Err: func(err error) (string, string) {
			var e *opensearch.Error
			if errors.As(err, &e) {
				return e.Type, e.Reason
			}
			return "", strings.TrimPrefix(err.Error(), "opensearch: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
		RowsAffected: drivers.RowsAffectedOrZero,
	})
}
