package dump

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	dbdump "github.com/xo/dbmeta/dump"
)

// TestTableLeadingColumns checks that a dump drops the values that the driver
// of the source puts before the columns of each row. The InfluxQL driver gives
// the name of the series first. Without the drop, the scan of a row fails,
// because the row holds one value more than the table has columns.
func TestTableLeadingColumns(t *testing.T) {
	db := sql.OpenDB(seriesConnector{rows: [][]driver.Value{
		{"cpu", int64(1), "idle"},
		{"cpu", int64(2), "busy"},
	}})
	defer db.Close()
	d := &Dumper{
		q: db,
		snap: &dbdump.Snapshot{
			Dialect: "influxql",
			Schemas: []dbdump.Schema{{
				Name: "*",
				Tables: []dbdump.Table{{
					Name: "cpu",
					Columns: []dbdump.Column{
						{Name: "time", Type: dbdump.Type{Kind: dbdump.Integer}},
						{Name: "state", Type: dbdump.Type{Kind: dbdump.String}},
					},
				}},
			}},
		},
		target: "sqlite3",
	}
	var b strings.Builder
	if err := d.table(context.Background(), &errWriter{w: &b}, "*", "cpu"); err != nil {
		t.Fatalf("table returned %v", err)
	}
	out := b.String()
	for _, want := range []string{"(1, 'idle')", "(2, 'busy')"} {
		if !strings.Contains(out, want) {
			t.Errorf("the dump does not hold the row %s:\n%s", want, out)
		}
	}
	if strings.Contains(out, "'cpu'") {
		t.Errorf("the dump holds the name of the series as a value:\n%s", out)
	}
}

// seriesConnector is a database/sql connector whose every query returns the
// same rows, with the name of the series before the columns, as the InfluxQL
// driver does.
type seriesConnector struct {
	rows [][]driver.Value
}

// Connect satisfies the driver.Connector interface.
func (c seriesConnector) Connect(context.Context) (driver.Conn, error) {
	return seriesConn(c), nil
}

// Driver satisfies the driver.Connector interface.
func (c seriesConnector) Driver() driver.Driver {
	return nil
}

// seriesConn is a connection of seriesConnector.
type seriesConn struct {
	rows [][]driver.Value
}

// Prepare satisfies the driver.Conn interface. The test runs no prepared
// statement.
func (seriesConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not supported")
}

// Close satisfies the driver.Conn interface.
func (seriesConn) Close() error {
	return nil
}

// Begin satisfies the driver.Conn interface. The test runs no transaction.
func (seriesConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not supported")
}

// QueryContext satisfies the driver.QueryerContext interface.
func (c seriesConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &seriesRows{rows: c.rows}, nil
}

// seriesRows are the rows of a query of seriesConn.
type seriesRows struct {
	rows [][]driver.Value
}

// Columns satisfies the driver.Rows interface.
func (r *seriesRows) Columns() []string {
	return []string{"name", "time", "state"}
}

// Close satisfies the driver.Rows interface.
func (r *seriesRows) Close() error {
	return nil
}

// Next satisfies the driver.Rows interface.
func (r *seriesRows) Next(dest []driver.Value) error {
	if len(r.rows) == 0 {
		return io.EOF
	}
	copy(dest, r.rows[0])
	r.rows = r.rows[1:]
	return nil
}
