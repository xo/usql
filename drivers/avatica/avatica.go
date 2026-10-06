// Package avatica defines and registers usql's Apache Avatica driver, for the
// standalone Avatica server and for the Apache Phoenix Query Server.
//
// Pre-flight, checked on 2026-10-07:
//
//   - Maintained: github.com/xo/dbimp v0.10.0 added the driver in place of
//     github.com/apache/calcite-avatica-go. The repository is in the same
//     family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the JSON of the HTTP API of Avatica.
//   - Tested: the vendor's images, which dbmeta's dbrun starts as
//     avatica-1.28.0, avatica-1.29.0 and phoenix-2.0-5.0.
//
// The URL is avatica://user:password@host:port, with the keys tls and auth.
// It has no path. The old driver took an http:// URL, which this one refuses.
//
// Metadata: none. dbmeta has no model for Avatica yet, so the describe
// commands report that they are not supported (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package avatica

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/xo/dbimp"
	"github.com/xo/dbimp/avatica" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("avatica", drivers.Driver{
		// DATABASE_VERSION is a function of HSQLDB, which backs the standalone
		// server, so what it returns is the release of HSQLDB and not of
		// Avatica. The Phoenix Query Server has no statement for its version,
		// which only a call of the protocol that the driver does not send
		// reports, so there the release is unknown.
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			if err := db.QueryRowContext(ctx, `VALUES (DATABASE_VERSION())`).Scan(&ver); err != nil {
				return "Avatica <unknown>", nil
			}
			return "Avatica, HSQLDB " + ver, nil
		},
		Err: func(err error) (string, string) {
			var e *avatica.Error
			if errors.As(err, &e) {
				return strconv.Itoa(e.Code), e.Message
			}
			return "", strings.TrimPrefix(err.Error(), "avatica: ")
		},
		IsPasswordErr: func(err error) bool {
			var se *dbimp.StatusError
			if errors.As(err, &se) && se.Code == http.StatusUnauthorized {
				return true
			}
			// the server answers a wrong password as an ordinary error, with
			// the standard exception of java.sql, whatever database is behind
			// it
			var e *avatica.Error
			return errors.As(err, &e) && strings.Contains(e.Exception, "SQLInvalidAuthorizationSpecException")
		},
	})
}
