// +build !go1.8

package handler

import (
	"database/sql"
)

// NextResultSet is a wrapper around the go1.8 introduced
// sql.Rows.NextResultSet call.
func NextResultSet(q *sql.Rows) bool {
	return false
}
