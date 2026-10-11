// Package describe formats the answers of usql's describe commands, such as
// \d, \dt and \df, the way psql formats them.
//
// dbmeta reads, tblfmt renders, and this package decides what to print. It
// writes no query of its own (D3). A database that dbmeta has no model for
// answers none of the describe commands.
package describe

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xo/dbmeta"
	_ "github.com/xo/dbmeta/all" // every model
	"github.com/xo/usql/text"
)

// ErrNotSupported is returned when the database cannot answer a command.
var ErrNotSupported = text.ErrNotSupported

// errDryRun stops a query from running when a dry run is asked for.
var errDryRun = errors.New("dry run")

// Logger prints a query before it runs, for echo_hidden.
type Logger interface {
	Println(...any)
}

// Option is an option of a Formatter.
type Option func(*Formatter)

// WithLogger prints every query before it runs.
func WithLogger(l Logger) Option {
	return func(f *Formatter) {
		f.logger = l
	}
}

// WithDryRun prints the queries without running them.
func WithDryRun(dryRun bool) Option {
	return func(f *Formatter) {
		f.dryRun = dryRun
	}
}

// WithTimeout limits how long one query may run.
func WithTimeout(timeout time.Duration) Option {
	return func(f *Formatter) {
		f.timeout = timeout
	}
}

// supports reports whether the database answers the query q.
func (f *Formatter) supports(q dbmeta.AnyQuery) bool {
	return q.Support(f.meta) == dbmeta.Supported
}

// queryer returns the handle that the queries run through, which prints each
// query first and runs none of them for a dry run.
func (f *Formatter) queryer() dbmeta.Queryer {
	if f.logger == nil && !f.dryRun {
		return f.db
	}
	return logging{db: f.db, logger: f.logger, dryRun: f.dryRun}
}

// all runs q and returns every row. dbmeta passes only the arguments that q
// takes. It returns ErrNotSupported when the database cannot answer q, and
// no rows for a dry run.
func all[T any](ctx context.Context, f *Formatter, q *dbmeta.Query[T], a dbmeta.Args) ([]T, error) {
	if !f.supports(q) {
		return nil, ErrNotSupported
	}
	if f.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, f.timeout)
		defer cancel()
	}
	var rows []T
	for row, err := range q.Each(ctx, f.meta, f.queryer(), a) {
		switch {
		case errors.Is(err, errDryRun):
			return nil, nil
		case err != nil:
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// logging prints each query before it runs.
type logging struct {
	db     dbmeta.Queryer
	logger Logger
	dryRun bool
}

// QueryContext satisfies dbmeta.Queryer.
func (l logging) QueryContext(ctx context.Context, query string, v ...any) (*sql.Rows, error) {
	if l.logger != nil {
		l.logger.Println(query)
		if len(v) != 0 {
			l.logger.Println(v...)
		}
	}
	if l.dryRun {
		return nil, errDryRun
	}
	return l.db.QueryContext(ctx, query, v...)
}

// like turns a psql wildcard pattern into an SQL LIKE pattern.
func like(s string) string {
	return strings.NewReplacer("*", "%", "?", "_").Replace(s)
}

// null returns the value of v, or an empty string when v is NULL.
func null[T any](v sql.Null[T]) string {
	if !v.Valid {
		return ""
	}
	return fmt.Sprint(v.V)
}

// qualified returns schema.name, quoted as psql quotes it.
func qualified(schema, name string) string {
	if schema == "" {
		return `"` + name + `"`
	}
	return `"` + schema + "." + name + `"`
}
