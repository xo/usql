// Package sqshared contains shared types for the sqlite3 and moderncsqlite
// drivers.
package sqshared

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ConvertBytes is the byte formatter func for sqlite3 databases.
func ConvertBytes(buf []byte, tfmt string) (string, error) {
	// attempt to convert buf if it matches a time format, and if it
	// does, then return a formatted time string.
	s := string(buf)
	if s != "" && strings.TrimSpace(s) != "" {
		t := new(Time)
		if err := t.Scan(buf); err == nil {
			return time.Time(*t).Format(tfmt), nil
		}
	}
	return s, nil
}

// Time provides a type that will correctly scan the various timestamps
// values stored by the github.com/mattn/go-sqlite3 driver for time.Time
// values, as well as correctly satisfying the sql/driver/Valuer interface.
type Time time.Time

// Value satisfies the Valuer interface.
func (t *Time) Value() (driver.Value, error) {
	return t, nil
}

// Scan satisfies the Scanner interface.
func (t *Time) Scan(v interface{}) error {
	switch x := v.(type) {
	case time.Time:
		*t = Time(x)
		return nil
	case []byte:
		return t.Parse(string(x))
	case string:
		return t.Parse(x)
	}
	return fmt.Errorf("cannot convert type %T to Time", v)
}

// Parse attempts to Parse string s to t.
func (t *Time) Parse(s string) error {
	if s == "" {
		return nil
	}
	for _, f := range SQLiteTimestampFormats {
		if z, err := time.Parse(f, s); err == nil {
			*t = Time(z)
			return nil
		}
	}
	return errors.New("could not parse time")
}

// SQLiteTimestampFormats is timestamp formats understood by both this module
// and SQLite.  The first format in the slice will be used when saving time
// values into the database. When parsing a string from a timestamp or datetime
// column, the formats are tried in order.
var SQLiteTimestampFormats = []string{
	// By default, store timestamps with whatever timezone they come with.
	// When parsed, they will be returned with the same timezone.
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02T15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"2006-01-02",
}
