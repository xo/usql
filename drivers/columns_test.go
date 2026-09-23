package drivers

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"reflect"
	"testing"
	"time"
)

// TestNullSafeColumnType checks that a NULL can be scanned out of a column the
// driver described as not nullable, which is what MySQL does for
// SHOW REPLICA STATUS. See issues 307, 476 and 539.
func TestNullSafeColumnType(t *testing.T) {
	sql.Register("nullsafetest", fakeDriver{})
	db, err := sql.Open("nullsafetest", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	cts, err := rows.ColumnTypes()
	if err != nil {
		t.Fatal(err)
	}
	dest := make([]interface{}, len(cts))
	for i, ct := range cts {
		if dest[i], err = NullSafeColumnType(ct); err != nil {
			t.Fatalf("building a destination for %s: %v", ct.Name(), err)
		}
	}
	if !rows.Next() {
		t.Fatal("the fake driver returned no rows")
	}
	// Every value in the row is NULL. Before the fix this failed on the first
	// column with "converting NULL to uint32 is unsupported".
	if err := rows.Scan(dest...); err != nil {
		t.Fatalf("scanning a row of NULLs: %v", err)
	}
	// A NULL must arrive as something that prints as the null string rather
	// than as a zero value.
	for i, d := range dest {
		if isNull(d) {
			continue
		}
		t.Errorf("column %s (%s): a NULL was not preserved, got %#v",
			cts[i].Name(), cts[i].ScanType(), d)
	}
}

// isNull reports whether a scanned destination holds a NULL.
func isNull(d interface{}) bool {
	switch v := d.(type) {
	case *interface{}:
		return *v == nil
	case *sql.NullBool:
		return !v.Valid
	case *sql.NullInt64:
		return !v.Valid
	case *sql.NullFloat64:
		return !v.Valid
	case *sql.NullString:
		return !v.Valid
	case *sql.NullTime:
		return !v.Valid
	}
	return false
}

// TestNullSafeColumnTypeKeepsTypes checks that the mapping does not throw away
// the type information that the column alignment and the time format need.
func TestNullSafeColumnTypeKeepsTypes(t *testing.T) {
	for _, test := range []struct {
		scan reflect.Type
		want interface{}
	}{
		{reflect.TypeOf(uint32(0)), new(sql.NullInt64)},
		{reflect.TypeOf(int64(0)), new(sql.NullInt64)},
		{reflect.TypeOf(float64(0)), new(sql.NullFloat64)},
		{reflect.TypeOf(""), new(sql.NullString)},
		{reflect.TypeOf(true), new(sql.NullBool)},
		{reflect.TypeOf(time.Time{}), new(sql.NullTime)},
		{reflect.TypeOf(sql.NullInt64{}), new(sql.NullInt64)},
		// uint64 does not fit in an int64, and sql.Null[uint64] is printed as
		// the JSON of the struct, so both fall back to an any.
		{reflect.TypeOf(uint64(0)), new(interface{})},
		{reflect.TypeOf(sql.Null[uint64]{}), new(interface{})},
		{reflect.TypeOf([]byte(nil)), new(interface{})},
	} {
		got := destForScanType(test.scan)
		if reflect.TypeOf(got) != reflect.TypeOf(test.want) {
			t.Errorf("a %s column was given a %T, want %T", test.scan, got, test.want)
		}
	}
}

// fakeDriver reports every column as not nullable and then returns NULL for
// all of them, the way MySQL does for SHOW REPLICA STATUS.
type fakeDriver struct{}

func (fakeDriver) Open(string) (driver.Conn, error) { return fakeConn{}, nil }

type fakeConn struct{}

func (fakeConn) Prepare(string) (driver.Stmt, error) { return fakeStmt{}, nil }
func (fakeConn) Close() error                        { return nil }
func (fakeConn) Begin() (driver.Tx, error)           { return nil, io.EOF }

type fakeStmt struct{}

func (fakeStmt) Close() error                               { return nil }
func (fakeStmt) NumInput() int                              { return 0 }
func (fakeStmt) Exec([]driver.Value) (driver.Result, error) { return nil, io.EOF }
func (fakeStmt) Query([]driver.Value) (driver.Rows, error)  { return &fakeRows{}, nil }

// fakeCols are the shapes MySQL reports for SHOW REPLICA STATUS: a uint32, a
// uint16, a uint64 and a string, none of them marked nullable.
var fakeCols = []struct {
	name string
	typ  reflect.Type
}{
	{"SQL_Remaining_Delay", reflect.TypeOf(uint32(0))},
	{"Source_Port", reflect.TypeOf(uint16(0))},
	{"Relay_Log_Space", reflect.TypeOf(uint64(0))},
	{"Replica_IO_State", reflect.TypeOf("")},
	{"Seconds_Behind_Source", reflect.TypeOf(int64(0))},
	{"Some_Time", reflect.TypeOf(time.Time{})},
}

type fakeRows struct{ done bool }

func (r *fakeRows) Columns() []string {
	out := make([]string, len(fakeCols))
	for i, c := range fakeCols {
		out[i] = c.name
	}
	return out
}

func (r *fakeRows) Close() error { return nil }

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	for i := range dest {
		dest[i] = nil // every column is NULL
	}
	return nil
}

func (r *fakeRows) ColumnTypeScanType(i int) reflect.Type { return fakeCols[i].typ }

func (r *fakeRows) ColumnTypeNullable(i int) (nullable, ok bool) { return false, true }
