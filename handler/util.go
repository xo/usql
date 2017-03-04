package handler

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/knq/xoutil"
)

var (
	// ErrDriverNotAvailable is the driver not available error.
	ErrDriverNotAvailable = errors.New("driver not available")

	// ErrNotConnected is the not connected error.
	ErrNotConnected = errors.New("not connected")
)

// Error is a wrapper to standardize errors.
type Error struct {
	Driver string
	Err    error
}

// Error satisfies the error interface.
func (e *Error) Error() string {
	n := "usql"
	s := e.Err.Error()

	if e.Driver != "" {
		n = e.Driver
		s = strings.TrimLeftFunc(strings.TrimPrefix(strings.TrimSpace(s), e.Driver+":"), unicode.IsSpace)
	}

	return n + ": " + s
}

// sqlite3Parse will convert buf matching a time format to a time, and will
// format it according to the handler time settings.
//
// TODO: only do this if the type of the column is a timestamp type.
func sqlite3Parse(buf []byte) string {
	s := string(buf)
	if s != "" && strings.TrimSpace(s) != "" {
		t := &xoutil.SqTime{}
		err := t.Scan(buf)
		if err == nil {
			return t.Format(time.RFC3339Nano)
		}
	}

	return s
}

// max returns the maximum of a, b.
func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// min returns the maximum of a, b.
func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// addQueryParam conditionally adds a ?name=val style query parameter to the
// end of urlstr if a == b, when urlstr does not already contain name=.
func addQueryParam(a, b, urlstr, name, val string) string {
	if a == b && !strings.Contains(urlstr, name+"=") {
		s := "?"
		if strings.Contains(urlstr, "?") {
			s = "&"
		}
		return urlstr + s + name + "=" + val
	}

	return urlstr
}

var drivers map[string]string

// SetAvailableDrivers sets the known available drivers.
func SetAvailableDrivers(m map[string]string) {
	drivers = m
}
