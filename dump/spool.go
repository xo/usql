package dump

import (
	"context"
	"errors"
	"io"
	"os"
	"time"
)

// Waits are the pauses before each new try of a dump that failed while it
// read the database. Oracle refuses a read-only transaction that reads a table
// made a few seconds before it started, with ORA-01466, and a new transaction
// a little later reads it.
var Waits = []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}

// retryError is an error that a new try of the dump can avoid.
type retryError struct {
	err error
}

// Error satisfies the error interface.
func (e retryError) Error() string {
	return e.err.Error()
}

// Unwrap returns the error.
func (e retryError) Unwrap() error {
	return e.err
}

// Retryable marks err as an error that a new try of the dump, in a new
// transaction, can avoid. It returns nil for nil.
func Retryable(err error) error {
	if err == nil {
		return nil
	}
	return retryError{err}
}

// Spool runs write against a temporary file, and copies the file to w only
// when write succeeds, so that w gets a complete script or nothing. When write
// fails with an error that Retryable marked, Spool calls again with the error
// and the pause, waits, and runs write again from an empty file, once for each
// pause of waits. The file can be read by its owner only, because it holds the
// rows of the database, and Spool removes it.
func Spool(ctx context.Context, w io.Writer, waits []time.Duration, write func(io.Writer) error, again func(error, time.Duration)) error {
	f, err := os.CreateTemp("", "usql-dump-*.sql")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	for i := 0; ; i++ {
		err := write(f)
		if err == nil {
			break
		}
		var re retryError
		if i == len(waits) || !errors.As(err, &re) {
			return err
		}
		if again != nil {
			again(re.err, waits[i])
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waits[i]):
		}
		if err := f.Truncate(0); err != nil {
			return err
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}
