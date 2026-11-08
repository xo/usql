package mysql

import (
	"testing"

	"github.com/go-sql-driver/mysql"
)

// TestErrVerbose checks that \errverbose writes a MySQL error in psql's
// layout: the SQLSTATE as the code, and MySQL's number in DETAIL.
func TestErrVerbose(t *testing.T) {
	err := &mysql.MySQLError{Number: 1146, SQLState: [5]byte{'4', '2', 'S', '0', '2'}, Message: "Table 'db.nope' doesn't exist"}
	got := Driver().ErrVerbose(err, "select * from nope")
	want := "ERROR:  42S02: Table 'db.nope' doesn't exist\nDETAIL:  error 1146"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
