package handler

// Error wraps handler errors
type Error struct {
	Buf string
	Err error
}

// WrapErr wraps an error using the specified driver when err is not nil.
func WrapErr(buf string, err error) error {
	if err == nil {
		return nil
	}
	// avoid double wrapping error
	if _, ok := err.(*Error); ok {
		return err
	}
	return &Error{buf, err}
}

// Error satisfies the error interface, returning the original error message
func (e *Error) Error() string { return e.Err.Error() }

// Unwrap returns the original error
func (e *Error) Unwrap() error { return e.Err }
