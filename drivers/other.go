// +build !pgx

package drivers

import (
	"database/sql"

	"github.com/knq/dburl"
)

// PgxOpen is a special PgxOpen func.
func PgxOpen(u *dburl.URL) func(string, string) (*sql.DB, error) {
	return sql.Open
}
