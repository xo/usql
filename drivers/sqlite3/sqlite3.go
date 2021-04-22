// Package sqlite3 defines and registers usql's SQLite3 driver. Requires CGO.
//
// See: https://github.com/mattn/go-sqlite3
package sqlite3

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3" // DRIVER: sqlite3
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

func init() {
	newReader := func(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
		return &metaReader{
			LoggingReader: metadata.NewLoggingReader(db, opts...),
		}
	}
	drivers.Register("sqlite3", drivers.Driver{
		AllowMultilineComments: true,
		ForceParams: drivers.ForceQueryParameters([]string{
			"loc", "auto",
		}),
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(ctx, `SELECT sqlite_version()`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "SQLite3 " + ver, nil
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(sqlite3.Error); ok {
				return strconv.Itoa(int(e.Code)), e.Error()
			}
			code, msg := "", err.Error()
			if e, ok := err.(sqlite3.ErrNo); ok {
				code = strconv.Itoa(int(e))
			}
			return code, msg
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
		NewMetadataReader: newReader,
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
