// Package pgshared gives the scan destination of a PostgreSQL column, so that
// usql prints a value of the column as psql does. The pgx and the libpq
// drivers share it.
package pgshared

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Dest returns a new scan destination for a value of col. It satisfies the
// ColumnTypes field of drivers.Driver. A column that needs no type of its own
// scans into a *any.
//
// pgx names a type that it has no name for by its OID, such as 1266 for timetz,
// so both spellings are matched.
func Dest(col *sql.ColumnType) (any, error) {
	switch col.DatabaseTypeName() {
	case "FLOAT4":
		// pgx reads a float4 as a float64, which then prints with the digits
		// of a float64
		return new(sql.Null[float32]), nil
	case "OID", "XID", "CID":
		return new(sql.Null[uint32]), nil
	case "XID8":
		return new(sql.Null[uint64]), nil
	case "TIME":
		return new(sql.Null[clock]), nil
	case "TIMETZ", "1266":
		return new(sql.Null[clockZone]), nil
	}
	return new(any), nil
}

// clock is a time of day. lib/pq reads a time as a time.Time on 1 January of
// year 0, which would print as a timestamp, so clock prints it as psql does.
// pgx reads a time as text, which clock keeps.
type clock string

// Scan satisfies the [database/sql.Scanner] interface.
func (c *clock) Scan(src any) error {
	s, err := scanClock(src, false)
	*c = clock(s)
	return err
}

// String satisfies the [fmt.Stringer] interface.
func (c clock) String() string {
	return string(c)
}

// clockZone is a time of day with a time zone, which prints as psql prints
// it.
type clockZone string

// Scan satisfies the [database/sql.Scanner] interface.
func (c *clockZone) Scan(src any) error {
	s, err := scanClock(src, true)
	*c = clockZone(s)
	return err
}

// String satisfies the [fmt.Stringer] interface.
func (c clockZone) String() string {
	return string(c)
}

// scanClock returns the text of a time of day, with its offset when zone is
// true.
func scanClock(src any, zone bool) (string, error) {
	switch v := src.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case time.Time:
		return formatClock(v, zone), nil
	}
	return "", fmt.Errorf("scanning a time of day: unexpected type %T", src)
}

// formatClock writes t as psql writes a time: the fraction of a second has no
// zeros at its end, and the offset has its minutes and its seconds only when
// they are not zero. 24:00:00 is a time of PostgreSQL, and lib/pq reads it as
// the midnight of the next day.
func formatClock(t time.Time, zone bool) string {
	var b strings.Builder
	h := t.Hour()
	if t.YearDay() > 1 && h == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0 {
		h = 24
	}
	fmt.Fprintf(&b, "%02d:%02d:%02d", h, t.Minute(), t.Second())
	if us := t.Nanosecond() / 1000; us != 0 {
		b.WriteString(strings.TrimRight(fmt.Sprintf(".%06d", us), "0"))
	}
	if !zone {
		return b.String()
	}
	_, off := t.Zone()
	sign := byte('+')
	if off < 0 {
		sign, off = '-', -off
	}
	fmt.Fprintf(&b, "%c%02d", sign, off/3600)
	if m, s := off/60%60, off%60; m != 0 || s != 0 {
		fmt.Fprintf(&b, ":%02d", m)
		if s != 0 {
			fmt.Fprintf(&b, ":%02d", s)
		}
	}
	return b.String()
}
