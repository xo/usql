package drivers

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
)

// Session is the one connection of a database that every statement of a
// session of usql runs on, as psql runs a session on one connection. A SET,
// a temporary table or a variable of the session then holds until the
// connection ends. Without it, database/sql hands each statement whichever
// connection of its pool is free, and go-mssqldb resets the state of a
// connection each time database/sql takes it back into the pool.
type Session struct {
	*sql.Conn
}

// NewSession takes one connection out of the pool of db for a session.
func NewSession(ctx context.Context, db *sql.DB) (*Session, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return &Session{Conn: conn}, nil
}

// Exec runs a statement that returns no rows.
func (s *Session) Exec(query string, args ...any) (sql.Result, error) {
	return s.ExecContext(context.Background(), query, args...)
}

// Query runs a statement that returns rows.
func (s *Session) Query(query string, args ...any) (*sql.Rows, error) {
	return s.QueryContext(context.Background(), query, args...)
}

// QueryRow runs a statement that returns at most one row.
func (s *Session) QueryRow(query string, args ...any) *sql.Row {
	return s.QueryRowContext(context.Background(), query, args...)
}

// Prepare prepares a statement.
func (s *Session) Prepare(query string) (*sql.Stmt, error) {
	return s.PrepareContext(context.Background(), query)
}

// IsConnLost reports whether err says that the connection of a session is
// gone, after which the session takes a new connection.
func IsConnLost(err error) bool {
	return errors.Is(err, driver.ErrBadConn) || errors.Is(err, sql.ErrConnDone)
}

// Raw calls f with the connection of the driver under db: the connection of
// the session, or one connection of a pool. It returns false for a handle
// that has no such connection, such as a transaction.
func Raw(ctx context.Context, db DB, f func(driverConn any) error) (bool, error) {
	switch v := db.(type) {
	case *Session:
		return true, v.Conn.Raw(f)
	case *sql.DB:
		conn, err := v.Conn(ctx)
		if err != nil {
			return true, err
		}
		defer conn.Close()
		return true, conn.Raw(f)
	}
	return false, nil
}
