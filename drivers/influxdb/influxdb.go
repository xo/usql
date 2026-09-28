// Package influxdb defines and registers usql's InfluxDB driver, which speaks
// SQL to InfluxDB 3 and later.
//
// Pre-flight, checked on 2026-09-29:
//
//   - Maintained: github.com/xo/dbimp v0.4.0 added the driver. The
//     repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over HTTP.
//   - Tested: the vendor's image, which dbmeta's dbrun starts as
//     influxdb-3.11.5.
//
// The influxql driver registers the same driver under the name influxql,
// through Driver. Metadata: none, because all metadata is moving into dbmeta
// (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package influxdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/influxdb" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("influxdb", Driver())
}

// Driver returns the InfluxDB driver. dburl opens an influxql:// URL through
// the same driver, and the influxql package registers it under the name
// influxql.
func Driver() drivers.Driver {
	return drivers.Driver{
		AllowMultilineComments: true,
		Process: func(_ *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			// InfluxDB sends no count of rows affected, and InfluxQL has
			// statements such as SHOW MEASUREMENTS that return rows, so each
			// statement is run as a query
			typ, _ := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, true, nil
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			// the release comes from GET /ping, through the driver's
			// connection, and no statement returns it
			rc, ok := db.(interface {
				Conn(context.Context) (*sql.Conn, error)
			})
			if !ok {
				return "", fmt.Errorf("reading the version through %T", db)
			}
			conn, err := rc.Conn(ctx)
			if err != nil {
				return "", err
			}
			defer conn.Close()
			var ver string
			err = conn.Raw(func(dc any) error {
				ver, err = influxdb.Version(ctx, dc)
				return err
			})
			if err != nil {
				return "", err
			}
			return "InfluxDB " + ver, nil
		},
		Err: func(err error) (string, string) {
			var e *influxdb.Error
			if errors.As(err, &e) {
				return "", e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "influxdb: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
	}
}
