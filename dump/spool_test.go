package dump

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	dbdump "github.com/xo/dbmeta/dump"
)

// TestSpool checks that w gets only the script of the try that succeeded, so
// that a dump that starts again writes nothing twice, and that an error that
// is not marked, or one more failure than the pauses, stops the dump with
// nothing written.
func TestSpool(t *testing.T) {
	errRead := errors.New("ORA-01466: unable to read data - table definition has changed")
	waits := []time.Duration{time.Millisecond, time.Millisecond}
	for _, test := range []struct {
		name  string
		fails []error
		exp   string
		err   error
		again int
	}{
		{"first try", nil, "try 1", nil, 0},
		{"second try", []error{Retryable(errRead)}, "try 2", nil, 1},
		{"third try", []error{Retryable(errRead), Retryable(errRead)}, "try 3", nil, 2},
		{"too many tries", []error{Retryable(errRead), Retryable(errRead), Retryable(errRead)}, "", errRead, 2},
		{"not marked", []error{errRead}, "", errRead, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			var b strings.Builder
			n, again := 0, 0
			err := Spool(context.Background(), &b, waits, func(w io.Writer) error {
				n++
				// a failed try writes part of a script first
				if _, err := io.WriteString(w, "-- part of try\n"); err != nil {
					return err
				}
				if n <= len(test.fails) {
					return test.fails[n-1]
				}
				_, err := io.WriteString(w, "try "+string(rune('0'+n)))
				return err
			}, func(err error, wait time.Duration) {
				again++
				if !errors.Is(err, errRead) || wait != time.Millisecond {
					t.Errorf("again got %v and %v", err, wait)
				}
			})
			if !errors.Is(err, test.err) || (err == nil) != (test.err == nil) {
				t.Fatalf("Spool returned %v, want %v", err, test.err)
			}
			if want := test.exp; want != "" {
				want = "-- part of try\n" + want
				if b.String() != want {
					t.Errorf("w got %q, want %q", b.String(), want)
				}
			} else if b.Len() != 0 {
				t.Errorf("w got %q, want nothing", b.String())
			}
			if again != test.again {
				t.Errorf("again was called %d times, want %d", again, test.again)
			}
		})
	}
}

// TestSpoolCanceled checks that a canceled context stops the pause.
func TestSpoolCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Spool(ctx, io.Discard, []time.Duration{time.Hour}, func(io.Writer) error {
		return Retryable(errors.New("read"))
	}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Spool returned %v, want %v", err, context.Canceled)
	}
}

// failQueryer is a dbmeta.Queryer whose every query fails.
type failQueryer struct {
	err error
}

// QueryContext satisfies the dbmeta.Queryer interface.
func (q failQueryer) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, q.err
}

// TestTableRetryable checks that a dump marks an error of the database while
// it reads the rows as one that a new try can avoid, so that Spool starts the
// dump again only for such an error, and not for a value that has no literal.
func TestTableRetryable(t *testing.T) {
	errRead := errors.New("ORA-01466: unable to read data - table definition has changed")
	d := &Dumper{
		q: failQueryer{errRead},
		snap: &dbdump.Snapshot{
			Dialect: "sqlite3",
			Schemas: []dbdump.Schema{{
				Name: "main",
				Tables: []dbdump.Table{{
					Name:    "t",
					Columns: []dbdump.Column{{Name: "id", Type: dbdump.Type{Kind: dbdump.Integer}}},
				}},
			}},
		},
		target: "sqlite3",
	}
	err := d.table(context.Background(), &errWriter{w: io.Discard}, "main", "t")
	var re retryError
	if !errors.Is(err, errRead) || !errors.As(err, &re) {
		t.Errorf("table returned %v, want %v marked as retryable", err, errRead)
	}
}
