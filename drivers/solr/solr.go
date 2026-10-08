// Package solr defines and registers usql's Solr driver.
//
// Pre-flight, checked on 2026-10-07:
//
//   - Maintained: github.com/xo/dbimp v0.14.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the vendor's images, which dbmeta's dbrun starts as solr-9.9.0,
//     solr-9.10.1 and solr-10.0.0.
//
// The driver runs queries and takes no write, because the SQL of Solr has none.
//
// Metadata: dbmeta's Solr model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package solr

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/solr" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("solr", drivers.Driver{
		Err: func(err error) (string, string) {
			var e *solr.Error
			if errors.As(err, &e) {
				// the SQL handler answers a failed statement with HTTP 200
				if e.HTTPStatus == http.StatusOK {
					return "", e.Message
				}
				return strconv.Itoa(e.HTTPStatus), e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "solr: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
		RowsAffected: drivers.RowsAffectedOrZero,
	})
}
