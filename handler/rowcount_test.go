package handler

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xo/usql/env"
)

// TestRowCount sets ROW_COUNT after each statement, as psql does: to the
// number of rows that a query returns or that a statement changes, and to 0
// when the statement fails.
func TestRowCount(t *testing.T) {
	h, _, _ := newConfigHandler(t)
	if err := h.Open(context.Background(), "moderncsqlite:"+filepath.Join(t.TempDir(), "r.db")); err != nil {
		t.Fatal(err)
	}
	// Windows cannot remove the file of the database while it is open
	defer h.Close()
	for _, test := range []struct {
		stmt string
		want string
		err  bool
	}{
		{"create table t (a int);", "0", false},
		{"insert into t values (1), (2), (3);", "3", false},
		{"select * from t;", "3", false},
		{"select * from t where a > 1;", "2", false},
		{"update t set a = a where a = 1;", "1", false},
		{"select * from t where a > 9;", "0", false},
		{"select a as x, a as y, a as z from t \\crosstabview", "3", false},
		{"select a from t where a = 2 \\gset rc_", "1", false},
		{"select nope from t;", "0", true},
		{"select * from t;", "3", false},
		{"delete from nope;", "0", true},
	} {
		err := h.RunCommand(test.stmt)
		if (err != nil) != test.err {
			t.Errorf("%s: error %v, want error %t", test.stmt, err, test.err)
		}
		if got, _ := env.Vars().Get("ROW_COUNT"); got != test.want {
			t.Errorf("%s: ROW_COUNT = %q, want %q", test.stmt, got, test.want)
		}
	}
}
