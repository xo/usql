// Package elasticsearch defines and registers usql's Elasticsearch driver.
//
// Pre-flight, checked on 2026-10-07:
//
//   - Maintained: github.com/xo/dbimp v0.14.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the vendor's images, which dbmeta's dbrun starts as
//     elasticsearch-8.19.22, elasticsearch-9.4.6 and elasticsearch-9.5.3.
//
// The driver runs queries and takes no write, because the SQL of Elasticsearch
// reads only.
//
// Metadata: dbmeta's Elasticsearch model. usql writes no reader of its own (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package elasticsearch

import (
	"errors"
	"net/http"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/elasticsearch" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("elasticsearch", drivers.Driver{
		Err: func(err error) (string, string) {
			var e *elasticsearch.Error
			if errors.As(err, &e) {
				return e.Type, e.Reason
			}
			return "", strings.TrimPrefix(err.Error(), "elasticsearch: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
		RowsAffected: drivers.RowsAffectedOrZero,
	})
}
