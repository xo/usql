// Package pinot defines and registers usql's Apache Pinot driver.
//
// Pre-flight, checked on 2026-09-30:
//
//   - Maintained: github.com/xo/dbimp v0.7.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the query API of the Broker, which is HTTP.
//   - Tested: the vendor's image, which dbmeta's dbrun starts as pinot-1.4.0
//     and pinot-1.5.1.
//
// The driver runs queries and takes no write, because the Broker refuses a
// write of rows. A write fails with the error of the server.
//
// Metadata: none. All metadata is moving into dbmeta, so usql writes no
// reader for Pinot (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package pinot

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/pinot" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("pinot", drivers.Driver{
		// no statement on the Broker returns the version, and SELECT
		// version() fails with code 700, so none is run
		Version: func(context.Context, drivers.DB) (string, error) {
			return "Apache Pinot", nil
		},
		Err: func(err error) (string, string) {
			var e *pinot.Error
			if errors.As(err, &e) {
				return strconv.Itoa(e.Code), e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "pinot: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
	})
}
