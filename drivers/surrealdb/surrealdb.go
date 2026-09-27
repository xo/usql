// Package surrealdb defines and registers usql's SurrealDB driver.
//
// Pre-flight, checked on 2026-09-27:
//
//   - Maintained: github.com/xo/dbimp v0.2.0 was tagged on 2026-09-27, and
//     its last commit is from the same day. The repository is in the same
//     family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//     Its dependencies set no logger, flag, signal, default HTTP client or
//     random seed.
//   - Links: pure Go. It adds github.com/cockroachdb/apd/v3, which the
//     Couchbase driver already brings.
//   - Tested: the vendor's docker.io/surrealdb/surrealdb image, which
//     dbmeta's dbrun starts as surrealdb-2.7.0, -3.1.6, -3.2.4 and -3.3.0.
//
// Metadata: no reader yet. SurrealDB has namespaces and databases, which
// INFO FOR ROOT and INFO FOR NS list, and users and access methods with
// roles, which INFO FOR NS and INFO FOR DB list. Each returns one object
// rather than rows, so a CatalogReader and a PrivilegeSummaryReader need a
// reader of their own. Neither is written.
//
// See: https://github.com/xo/dbimp
// Group: most
package surrealdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/surrealdb" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("surrealdb", drivers.Driver{
		AllowMultilineComments: true,
		Process: func(_ *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			// every statement of SurrealQL returns a result, even CREATE and
			// DEFINE, and the server sends no count of rows affected, so each
			// statement is run as a query
			typ, _ := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, true, nil
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			// no statement of SurrealQL returns the version, so it is read
			// through the driver's connection
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
				ver, err = surrealdb.Version(ctx, dc)
				return err
			})
			if err != nil {
				return "", err
			}
			return "SurrealDB " + strings.TrimPrefix(ver, "surrealdb-"), nil
		},
		Err: func(err error) (string, string) {
			// a response can hold several errors, and only a single one is
			// reported by its code
			var re *surrealdb.ResponseError
			if errors.As(err, &re) && len(re.Errs) == 1 {
				e := re.Errs[0]
				switch {
				case e.Kind != "":
					return e.Kind, e.Msg
				case e.Code != 0:
					return strconv.Itoa(e.Code), e.Msg
				}
				return "", e.Msg
			}
			return "", strings.TrimPrefix(err.Error(), "surrealdb: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			return errors.As(err, &se) && se.Code == http.StatusUnauthorized
		},
	})
}
