package handler

import (
	"errors"
	"os"
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

	// ErrNoSuchFileOrDirectory is the no such file or directory error.
	ErrNoSuchFileOrDirectory = errors.New("no such file or directory")

	// ErrCannotIncludeDirectories is the cannot include directories error.
	ErrCannotIncludeDirectories = errors.New("cannot include directories")

	// ErrNoEditorDefined is the no editor defined error.
	ErrNoEditorDefined = errors.New("no editor defined")
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

// pop pops the top item off of a if it is present, returning the value and the
// new slice. if a is empty, then v will be the returned value.
func pop(a []string, v string) ([]string, string) {
	if len(a) != 0 {
		return a[1:], a[0]
	}
	return a, v
}

// getenv tries retrieving successive keys from os environment variables.
func getenv(keys ...string) string {
	for _, key := range keys {
		if s := os.Getenv(key); s != "" {
			return s
		}
	}

	return ""
}
