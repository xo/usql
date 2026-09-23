package drivers

import (
	"database/sql"
	"reflect"
	"time"
)

// timeType is time.Time, which has no reflect.Kind of its own.
var timeType = reflect.TypeOf(time.Time{})

// nullTypes are the types that both take a NULL and are printed as the value
// they hold. A destination outside this set is given a *any instead, so that a
// value is printed as itself and a NULL is printed as the null string.
//
// The generic sql.Null[T] is absent, which MySQL asks for on any nullable
// BIGINT UNSIGNED. Passing it through would be correct with tblfmt v0.19.0 and
// later, which unwraps it, but a *any reaches the formatter as the contained
// value either way, so the shorter path is kept and this does not depend on
// the version of tblfmt.
var nullTypes = map[reflect.Type]bool{
	reflect.TypeOf(sql.NullBool{}):    true,
	reflect.TypeOf(sql.NullByte{}):    true,
	reflect.TypeOf(sql.NullInt16{}):   true,
	reflect.TypeOf(sql.NullInt32{}):   true,
	reflect.TypeOf(sql.NullInt64{}):   true,
	reflect.TypeOf(sql.NullFloat64{}): true,
	reflect.TypeOf(sql.NullString{}):  true,
	reflect.TypeOf(sql.NullTime{}):    true,
}

// NullSafeColumnType returns a scan destination for a column, chosen so that a
// NULL cannot fail the scan.
//
// A driver reports the Go type it would like to scan a column into, and usql
// uses it so that numbers are aligned as numbers and times are formatted as
// times rather than arriving as raw bytes. Scanning straight into that type is
// what `database/sql` refuses when the value turns out to be NULL:
//
//	sql: Scan error on column index 43, name "SQL_Remaining_Delay":
//	converting NULL to uint32 is unsupported
//
// The driver's own nullability report cannot be trusted to avoid that. MySQL
// describes 37 of the 56 columns of SHOW REPLICA STATUS as not nullable, four
// of them as uint32, and then sends NULL for SQL_Remaining_Delay. That is
// issues 307, 476 and 539, reported against 5.7 and 8.0 alike.
//
// So every destination here tolerates a NULL. A numeric or time column is
// given the matching sql.Null type, which keeps the type information that the
// alignment and the time format depend on. Everything else is given a *any,
// where a NULL arrives as a nil and is printed as the null string rather than
// as a zero value.
func NullSafeColumnType(ct *sql.ColumnType) (interface{}, error) {
	return destForScanType(ct.ScanType()), nil
}

// destForScanType returns the scan destination for a column whose driver asks
// to be scanned into typ. A nil typ means the driver named no type.
func destForScanType(typ reflect.Type) interface{} {
	if typ == nil {
		return new(interface{})
	}
	// A driver that already asks for a type which takes a NULL and prints as
	// its value knows the column better than this function does.
	if nullTypes[typ] {
		return reflect.New(typ).Interface()
	}
	if typ == timeType {
		return new(sql.NullTime)
	}
	switch typ.Kind() {
	case reflect.Bool:
		return new(sql.NullBool)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return new(sql.NullInt64)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		// These all fit in an int64, so the value cannot be truncated.
		return new(sql.NullInt64)
	case reflect.Float32, reflect.Float64:
		return new(sql.NullFloat64)
	case reflect.String:
		return new(sql.NullString)
	}
	// Anything else. This covers uint and uint64, which do not fit in an
	// sql.NullInt64, and every sql.Null[T]. A driver hands the value over as
	// its own type, so a number is still printed and aligned as a number, and
	// a NULL arrives as a nil.
	return new(interface{})
}
