package handler

import (
	"errors"
	"strings"
	"unicode"

	"github.com/knq/usql/text"
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
)

// Error is a wrapper to standardize errors.
type Error struct {
	Driver string
	Err    error
}

// Error satisfies the error interface.
func (e *Error) Error() string {
	n := text.CommandName

	s := e.Err.Error()

	if e.Driver != "" {
		n = e.Driver
		s = strings.TrimLeftFunc(strings.TrimPrefix(strings.TrimSpace(s), e.Driver+":"), unicode.IsSpace)

		switch e.Driver {
		case "ora", "oracle":
			if i := strings.Index(s, "ORA-"); i != -1 {
				s = s[i:]
			}

		case "mysql":
			s = strings.TrimSpace(strings.TrimPrefix(s, "Error "))
		}
	}

	return n + ": " + s
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
