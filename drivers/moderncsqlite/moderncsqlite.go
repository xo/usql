// Package moderncsqlite defines and registers usql's ModernC SQLite3 driver, a
// transpilation of SQLite3 to pure Go.
//
// See: https://gitlab.com/cznic/sqlite
package moderncsqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"modernc.org/sqlite" // DRIVER: moderncsqlite
)

func init() {
	drivers.Register("moderncsqlite", drivers.Driver{
		AllowMultilineComments: true,
		Open: func(u *dburl.URL) (func(string, string) (*sql.DB, error), error) {
			return func(_ string, params string) (*sql.DB, error) {
				return sql.Open("sqlite", params)
			}, nil
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(ctx, `SELECT sqlite_version()`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "ModernC SQLite " + ver, nil
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(*sqlite.Error); ok {
				return strconv.Itoa(e.Code()), e.Error()
			}
			return "", err.Error()
		},
		ConvertBytes: func(buf []byte, tfmt string) (string, error) {
			// attempt to convert buf if it matches a time format, and if it
			// does, then return a formatted time string.
			s := string(buf)
			if s != "" && strings.TrimSpace(s) != "" {
				t := new(sqTime)
				if err := t.Scan(buf); err == nil {
					return t.Format(tfmt), nil
				}
			}
			return s, nil
		},
	})
}

// sqTime provides a type that will correctly scan the various timestamps
// values stored by the github.com/mattn/go-sqlite3 driver for time.Time
// values, as well as correctly satisfying the sql/driver/Valuer interface.
type sqTime struct {
	time.Time
}

// Value satisfies the Valuer interface.
func (t sqTime) Value() (driver.Value, error) {
	return t.Time, nil
}

// Scan satisfies the Scanner interface.
func (t *sqTime) Scan(v interface{}) error {
	switch x := v.(type) {
	case time.Time:
		t.Time = x
		return nil
	case []byte:
		return t.parse(string(x))
	case string:
		return t.parse(x)
	}
	return fmt.Errorf("cannot convert type %T to time.Time", v)
}

// parse attempts to parse string s to t.
func (t *sqTime) parse(s string) error {
	if s == "" {
		return nil
	}
	for _, f := range SQLiteTimestampFormats {
		z, err := time.Parse(f, s)
		if err == nil {
			t.Time = z
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
