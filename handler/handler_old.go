// +build !go1.8

package handler

import (
	"database/sql"
)

// nextResultSet is a wrapper around the go1.8 introduced
// sql.Rows.NextResultSet call.
func nextResultSet(q *sql.Rows) bool {
	return false
}
