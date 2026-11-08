package drivers

import (
	"errors"
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

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.Err
}

// chop chops off a "prefix: " prefix from a string.
func chop(s, prefix string) string {
	return strings.TrimLeftFunc(strings.TrimPrefix(strings.TrimSpace(s), prefix+":"), unicode.IsSpace)
}

// VerboseError is an error of a driver library that writes every field it
// holds, as the client of the database writes them, for \errverbose. query is
// the statement that failed, for the position of the error in it. It returns
// "" when it has nothing more than its message.
//
// A driver with no package in usql, such as Oracle's, Cassandra's or ODBC's,
// gives \errverbose its fields this way, since usql cannot name its types.
type VerboseError interface {
	error
	ErrVerbose(query string) string
}

// ErrVerbose returns err with every field that its driver gives, for
// \errverbose, as the client of the database writes them. query is the
// statement that failed. The ErrVerbose hook of the driver wins, then an
// error in the chain that is a VerboseError, and otherwise err is returned
// as usql prints it.
func ErrVerbose(err error, query string) string {
	var e *Error
	if errors.As(err, &e) {
		if d, ok := drivers[e.Driver]; ok && d.ErrVerbose != nil {
			if s := d.ErrVerbose(e.Err, query); s != "" {
				return s
			}
		}
	}
	var v VerboseError
	if errors.As(err, &v) {
		if s := v.ErrVerbose(query); s != "" {
			return s
		}
	}
	return err.Error()
}

// Verbose writes an error in psql's verbose layout, for \errverbose, so that
// every database shows its errors the same way: "ERROR:  CODE: MESSAGE" on the
// first line, then each field with a value as an upper case name, a colon, two
// spaces and the value, such as "DETAIL:  ...". A code that is empty is left
// out of the first line.
func Verbose(code, message string, fields ...[2]string) string {
	var b strings.Builder
	b.WriteString("ERROR:  ")
	if code != "" {
		b.WriteString(code + ": ")
	}
	b.WriteString(message)
	for _, f := range fields {
		if f[1] != "" {
			b.WriteString("\n" + f[0] + ":  " + f[1])
		}
	}
	return b.String()
}
