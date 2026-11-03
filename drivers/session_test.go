package drivers

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
)

// sessionConnector opens the connections of sessionConn, and counts them.
type sessionConnector struct {
	opened int
	resets int
}

func (c *sessionConnector) Connect(context.Context) (driver.Conn, error) {
	c.opened++
	return &sessionConn{id: c.opened, c: c}, nil
}

func (c *sessionConnector) Driver() driver.Driver { return nil }

// sessionConn is a connection that runs any statement, and counts the resets of
// its session, as go-mssqldb resets a session that database/sql takes back.
type sessionConn struct {
	id int
	c  *sessionConnector
}

func (c *sessionConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("no prepare") }
func (c *sessionConn) Close() error                        { return nil }
func (c *sessionConn) Begin() (driver.Tx, error)           { return nil, errors.New("no transactions") }

func (c *sessionConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}

func (c *sessionConn) ResetSession(context.Context) error {
	c.c.resets++
	return nil
}

// TestSession checks that the statements of a session run on one connection,
// which database/sql never takes back, so that its state is never reset.
func TestSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// the pool takes its connection back after each statement, and resets it
	pool := &sessionConnector{}
	db := sql.OpenDB(pool)
	defer db.Close()
	for range 3 {
		if _, err := db.Exec("set x = 1"); err != nil {
			t.Fatal(err)
		}
	}
	if pool.resets == 0 {
		t.Fatal("expected the pool to reset its connection between statements")
	}
	c := &sessionConnector{}
	db = sql.OpenDB(c)
	defer db.Close()
	s, err := NewSession(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for range 3 {
		if _, err := s.Exec("set x = 1"); err != nil {
			t.Fatal(err)
		}
	}
	if c.opened != 1 || c.resets != 0 {
		t.Errorf("expected 1 connection and no resets, got %d connections and %d resets", c.opened, c.resets)
	}
	// Raw reaches the connection of the session
	var id int
	ok, err := Raw(ctx, s, func(dc any) error {
		id = dc.(*sessionConn).id
		return nil
	})
	if !ok || err != nil || id != 1 {
		t.Errorf("expected the connection of the session, got %d, %t, %v", id, ok, err)
	}
}

func TestIsConnLost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err error
		exp bool
	}{
		{driver.ErrBadConn, true},
		{fmt.Errorf("querying: %w", driver.ErrBadConn), true},
		{sql.ErrConnDone, true},
		{errors.New("syntax error"), false},
		{nil, false},
	}
	for _, test := range tests {
		if got := IsConnLost(test.err); got != test.exp {
			t.Errorf("IsConnLost(%v): expected %t, got %t", test.err, test.exp, got)
		}
	}
}
