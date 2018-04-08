package drivers

import (
	"strings"
	"unicode"
)

// Error is a wrapper to standardize errors.
type Error struct {
	Driver string
	Err    error
}

// WrapErr wraps an error using the specified driver when err is not nil.
func WrapErr(driver string, err error) error {
	if err == nil {
		return nil
	}

	// avoid double wrapping error
	if _, ok := err.(*Error); ok {
		return err
	}

	return &Error{driver, err}
}

// Error satisfies the error interface, returning simple information about the
// wrapped error in standardized way.
func (e *Error) Error() string {
	if d, ok := drivers[e.Driver]; ok {
		n := e.Driver
		if d.Name != "" {
			n = d.Name
		}
		s := n

		var msg string
		if d.Err != nil {
			var code string
			code, msg = d.Err(e.Err)
			if code != "" {
				s += ": " + code
			}
		} else {
			msg = e.Err.Error()
		}

		return s + ": " + chop(msg, n)
	}

	return e.Driver + ": " + chop(e.Err.Error(), e.Driver)
}

// chop chops off a "prefix: " prefix from a string.
func chop(s, prefix string) string {
	return strings.TrimLeftFunc(strings.TrimPrefix(strings.TrimSpace(s), prefix+":"), unicode.IsSpace)
}
