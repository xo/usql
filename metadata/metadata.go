// Package metadata reads the structure of a database through
// github.com/xo/dbmeta, and prints it the way psql does.
//
// dbmeta reads, tblfmt renders, and usql decides what to print. usql writes no
// query of its own here (D3). A database that dbmeta has no model for answers
// none of the describe commands.
package metadata

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

// Logger prints a query before it runs, for ECHO_HIDDEN.
type Logger interface {
	Println(...any)
}

// Option is an option for New.
type Option func(*Meta)

// WithLogger prints every query before it runs.
func WithLogger(l Logger) Option {
	return func(m *Meta) {
		m.logger = l
	}
}

// WithDryRun prints the queries without running them.
func WithDryRun(dryRun bool) Option {
	return func(m *Meta) {
		m.dryRun = dryRun
	}
}

// WithTimeout limits how long one query may run.
func WithTimeout(timeout time.Duration) Option {
	return func(m *Meta) {
		m.timeout = timeout
	}
}

// Meta reads the structure of one database.
type Meta struct {
	meta    *dbmeta.Meta
	db      dbmeta.Queryer
	logger  Logger
	dryRun  bool
	timeout time.Duration
}

// New returns a Meta for the database db, whose dialect is dialect, the
// URL.Dialect that dburl sets. It reads the version of the server once,
// because every query is chosen by it. It returns ErrNotSupported when
// dbmeta has no model for the dialect.
func New(ctx context.Context, dialect string, db dbmeta.Queryer, opts ...Option) (*Meta, error) {
	m := &Meta{db: db}
	for _, o := range opts {
		o(m)
	}
	d := dbmeta.Dialect(dialect)
	if _, ok := d.Info(); !ok {
		return nil, ErrNotSupported
	}
	// an unknown version selects the newest form of every query, so a version
	// that cannot be read is not an error
	versions, err := d.Version(ctx, m.queryer())
	if err != nil {
		versions = dbmeta.VersionSet{}
	}
	if m.meta, err = dbmeta.New(d, versions); err != nil {
		return nil, err
	}
	return m, nil
}

// String returns the display line of the server, such as "PostgreSQL 18.1".
func (m *Meta) String() string {
	return m.meta.String()
}

// Dialect returns the dialect.
func (m *Meta) Dialect() string {
	return string(m.meta.Dialect())
}

// Supports reports whether the database answers the query q.
func (m *Meta) Supports(q dbmeta.AnyQuery) bool {
	return q.Support(m.meta) == dbmeta.Supported
}

// queryer returns the handle that the queries run through, which prints each
// query first and runs none of them for a dry run.
func (m *Meta) queryer() dbmeta.Queryer {
	if m.logger == nil && !m.dryRun {
		return m.db
	}
	return logging{db: m.db, logger: m.logger, dryRun: m.dryRun}
}

// args are the arguments of one query. Schema, Name and Parent are SQL LIKE
// patterns, and an empty one matches everything.
type args struct {
	Catalog      string
	Schema       string
	Parent       string
	Name         string
	AccessMethod string
	Server       string
	Database     string
	WithSystem   bool
}

// all runs q and returns every row. It passes only the arguments that q takes
// on this database, and it returns ErrNotSupported when the database cannot
// answer q.
func all[T any](ctx context.Context, m *Meta, q *dbmeta.Query[T], a args) ([]T, error) {
	if !m.Supports(q) {
		return nil, ErrNotSupported
	}
	params, err := q.Params(m.meta)
	if err != nil {
		return nil, err
	}
	takes := make(map[string]bool, len(params))
	for _, p := range params {
		takes[p.Name] = true
	}
	values := make(map[string]any)
	set := func(name, value string) {
		if value != "" && takes[name] {
			values[name] = value
		}
	}
	set("catalog", a.Catalog)
	set("schema", a.Schema)
	set("parent", a.Parent)
	set("name", a.Name)
	set("access_method", a.AccessMethod)
	set("server", a.Server)
	set("database", a.Database)
	if a.WithSystem && takes["with_system"] {
		values["with_system"] = true
	}
	if m.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.timeout)
		defer cancel()
	}
	var rows []T
	for row, err := range q.All(ctx, m.meta, m.queryer(), values) {
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

// Pattern splits a psql pattern such as public.film* into a schema and a
// name, each an SQL LIKE pattern. Each part is folded as the database folds a
// name that is not quoted, as psql does, so \d book finds BOOK on Oracle. A
// part between double quotes keeps its case.
func (m *Meta) Pattern(pattern string) (string, string) {
	schema, name := "", pattern
	if i := dot(pattern); i != -1 {
		schema, name = pattern[:i], pattern[i+1:]
	}
	d := m.meta.Dialect()
	return like(d.FoldIdentifier(schema)), like(d.FoldIdentifier(name))
}

// dot returns the index of the first '.' in s that is not between double
// quotes, or -1.
func dot(s string) int {
	var quoted bool
	for i, c := range s {
		switch {
		case c == '"':
			quoted = !quoted
		case c == '.' && !quoted:
			return i
		}
	}
	return -1
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

// Has reports whether dbmeta has a model for dialect.
func Has(dialect string) bool {
	_, ok := dbmeta.Dialect(dialect).Info()
	return ok
}

// With returns a copy of m with opts applied. The options of a session, such
// as ECHO_HIDDEN, can change between commands, so each command applies them.
func (m *Meta) With(opts ...Option) *Meta {
	c := *m
	for _, o := range opts {
		o(&c)
	}
	return &c
}

// CurrentUser returns the name of the current user.
func (m *Meta) CurrentUser(ctx context.Context) (string, error) {
	if !m.Supports(dbmeta.CurrentUser) {
		return "", ErrNotSupported
	}
	u, ok, err := dbmeta.First(dbmeta.CurrentUser.All(ctx, m.meta, m.queryer(), nil))
	if err != nil || !ok {
		return "", err
	}
	return u.Name, nil
}

// CanChangePassword reports whether dbmeta builds the statement that changes a
// password for dialect.
func CanChangePassword(dialect string) bool {
	info, ok := dbmeta.Dialect(dialect).Info()
	return ok && info.ChangePassword != nil
}

// ChangePassword returns the statement that changes the password of user on
// db. old is the password being replaced, which only some databases take. The
// statement holds the password, so it must not be logged.
func ChangePassword(ctx context.Context, dialect string, db dbmeta.Queryer, user, password, old string) (string, error) {
	d := dbmeta.Dialect(dialect)
	// how a string is quoted can depend on a setting of the server
	q, err := d.Quoting(ctx, db)
	if err != nil {
		return "", err
	}
	return d.ChangePassword(dbmeta.PasswordChange{User: user, Password: password, Old: old}, q)
}
