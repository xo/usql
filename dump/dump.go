// Package dump writes the schema and the rows of a database as a usql script,
// which usql runs with -f to make the database again.
//
// dbmeta's dump package reads the schema and builds every statement. This
// package runs the queries that read the rows, and writes the script. It
// holds no knowledge of a database. See W77 and dbmeta's D274.
package dump

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/xo/dbmeta"
	_ "github.com/xo/dbmeta/all" // every model
	dbdump "github.com/xo/dbmeta/dump"
	"github.com/xo/dburl"
)

// Options says what a dump writes.
type Options struct {
	// Target is the dialect that the script is written for, as a scheme or
	// an alias of dburl, such as duckdb or pg. It is empty for the dialect of
	// the source.
	Target string
	// SchemaOnly writes the schema and no rows. DataOnly writes the rows and
	// no schema.
	SchemaOnly bool
	DataOnly   bool
	// Schemas and Tables are the patterns of the names to dump, and
	// ExcludeSchemas and ExcludeTables the patterns of the names to leave
	// out. A pattern is a pattern of SQL LIKE. A table pattern with a dot,
	// such as app.%, matches the schema and the table together.
	Schemas        []string
	ExcludeSchemas []string
	Tables         []string
	ExcludeTables  []string
	// Lossy allows a column type in the target that holds less than the
	// source type. Text allows the string type of the target for a type that
	// the target does not have. A dump fails on such a column when it is not
	// allowed.
	Lossy bool
	Text  bool
	// RestrictKey is the key of the \restrict line of the script. It is
	// random when it is empty.
	RestrictKey string
	// Version is the version of usql, for the header of the script.
	Version string
}

// Error is an error of a dump.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrSchemaAndDataOnly is the error for both --schema-only and
	// --data-only.
	ErrSchemaAndDataOnly Error = "schema-only and data-only cannot be used together"
	// ErrUnknownTarget is the error for a target that names no dialect.
	ErrUnknownTarget Error = "unknown target"
	// ErrCheckFailed is the error for a check that found a fault.
	ErrCheckFailed Error = "check failed"
	// ErrNotBoolean is the error for a switch that is given a value that is
	// not a Boolean.
	ErrNotBoolean Error = "not a Boolean"
)

// Plan reports what a dump with the options does with each column, before
// anything is written. A column whose type in the target is not exact has a
// note, and so does an object that is written as the text of the source.
type Plan struct {
	Source, Target dbmeta.Dialect
	dbdump.Plan
}

// Dumper writes one dump. It holds the snapshot of the schema, which it reads
// once, inside the transaction that reads the rows.
type Dumper struct {
	q      dbmeta.Queryer
	snap   *dbdump.Snapshot
	target dbmeta.Dialect
	policy dbdump.Policy
	opts   Options

	// session is whether the statements that set the session of the source
	// ran.
	session bool
}

// New reads the schema of the database into a snapshot, and makes the plan of
// the dump. q is where the queries run. It must be a transaction that sees one
// state of the database, as Begin opens one, so that the schema and the rows
// agree.
//
// New fails when the dialect cannot be dumped, or cannot be written, and says
// why. It fails with a *dbdump.PlanError, which names every column at once,
// when a column has no type in the target that the options allow.
func New(ctx context.Context, m *dbmeta.Meta, q dbmeta.Queryer, opts Options) (*Dumper, error) {
	if opts.SchemaOnly && opts.DataOnly {
		return nil, ErrSchemaAndDataOnly
	}
	source := m.Dialect()
	target := source
	if opts.Target != "" {
		var err error
		if target, err = Dialect(opts.Target); err != nil {
			return nil, err
		}
	}
	if c := dbdump.Support(source); !c.Source {
		return nil, fmt.Errorf("%s cannot be dumped, and %s can: %w", source, supported(func(c dbdump.Capability) bool { return c.Source }), c.SourceWhy)
	}
	if c := dbdump.Support(target); !c.Target {
		return nil, fmt.Errorf("a dump cannot be written for %s, and can be for %s: %w", target, supported(func(c dbdump.Capability) bool { return c.Target }), c.TargetWhy)
	}
	snap, err := dbdump.Read(ctx, m, q, dbdump.Filter{
		Schemas:        opts.Schemas,
		Tables:         opts.Tables,
		ExcludeSchemas: opts.ExcludeSchemas,
		ExcludeTables:  opts.ExcludeTables,
	})
	if err != nil {
		return nil, fmt.Errorf("reading the schema: %w", err)
	}
	d := &Dumper{
		q:      q,
		snap:   snap,
		target: target,
		policy: dbdump.Policy{Default: dbdump.Allow{Lossy: opts.Lossy, Text: opts.Text}},
		opts:   opts,
	}
	if _, err := d.Plan(); err != nil {
		return nil, err
	}
	return d, nil
}

// Dialect returns the dialect that a scheme or an alias of dburl names, such
// as duckdb, pg or sqlite.
func Dialect(name string) (dbmeta.Dialect, error) {
	scheme, _ := dburl.SchemeNameAndAliases(strings.ToLower(strings.TrimSuffix(name, ":")))
	for _, s := range dburl.BaseSchemes() {
		if s.Name != scheme || scheme == "" {
			continue
		}
		d := dbmeta.Dialect(s.Dialect)
		if d == "" {
			d = dbmeta.Dialect(s.Name)
		}
		if _, ok := d.Info(); ok {
			return d, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrUnknownTarget, name)
}

// Snapshot returns the snapshot of the schema.
func (d *Dumper) Snapshot() *dbdump.Snapshot {
	return d.snap
}

// Plan returns what the dump does with each column.
func (d *Dumper) Plan() (Plan, error) {
	p, err := dbdump.MakePlan(d.snap, d.target, d.policy)
	if err != nil {
		return Plan{}, err
	}
	return Plan{Source: dbmeta.Dialect(d.snap.Dialect), Target: d.target, Plan: p}, nil
}

// Warnings returns a line for each column whose type in the target is not
// exact, each object that is written as the text of the source, each
// collation that is left out, and each object that the reader skipped. The
// lines are what a user needs to know before they load the script.
func (d *Dumper) Warnings() []string {
	var lines []string
	p, err := d.Plan()
	if err != nil {
		return nil
	}
	for _, c := range p.Columns {
		if c.Level != dbdump.Exact {
			lines = append(lines, fmt.Sprintf("column %s: %s is written as %s (%s): %s", c.Name, typeText(c.Source), c.SQL, c.Level, c.Note))
		}
	}
	for _, o := range p.Untranslated {
		lines = append(lines, fmt.Sprintf("%s %s: the %s statement is not translated, and the script holds it as a comment", o.Kind, o.Name, p.Source))
	}
	for _, c := range p.DroppedCollations {
		lines = append(lines, fmt.Sprintf("column %s: the collation is left out, which can change how values compare and sort", c))
	}
	for _, s := range d.snap.Skipped {
		name := s.Name
		if s.Schema != "" {
			name = s.Schema + "." + s.Name
		}
		lines = append(lines, fmt.Sprintf("%s %s is not dumped: %s", s.Kind, name, s.Reason))
	}
	return lines
}

// typeText returns the text of a type, as the source wrote it when it did.
func typeText(t dbdump.Type) string {
	if t.Raw != "" {
		return t.Raw
	}
	return string(t.Kind)
}

// Write writes the script to w: the header, the statements of the schema
// before the rows, the rows of each table, and the statements after the rows.
// It reads the rows with q, a table at a time and a row at a time, so a dump
// holds one batch of rows in memory.
func (d *Dumper) Write(ctx context.Context, w io.Writer) error {
	key := d.opts.RestrictKey
	if key == "" {
		key = randomKey()
	}
	var stmts []dbdump.Statement
	for st, err := range dbdump.Script(d.snap, d.target, d.policy) {
		if err != nil {
			return err
		}
		stmts = append(stmts, st)
	}
	bw := &errWriter{w: w}
	d.header(bw, key)
	rowsDone := false
	for _, st := range stmts {
		if st.Stage == dbdump.Post && !rowsDone {
			if err := d.rows(ctx, bw, stmts); err != nil {
				return err
			}
			rowsDone = true
		}
		if d.skip(st) {
			continue
		}
		d.statement(bw, st)
		if bw.err != nil {
			return bw.err
		}
	}
	if !rowsDone {
		if err := d.rows(ctx, bw, stmts); err != nil {
			return err
		}
	}
	fmt.Fprintf(bw, "\n\\unrestrict %s\n", key)
	return bw.err
}

// header writes what the script says about itself, and the lines that make
// usql stop at the first error and refuse a backslash command in the rest of
// the script, as pg_dump's \restrict does.
func (d *Dumper) header(w io.Writer, key string) {
	fmt.Fprintln(w, "--")
	fmt.Fprintln(w, "-- usql dump")
	fmt.Fprintln(w, "--")
	if d.opts.Version != "" {
		fmt.Fprintf(w, "-- Dumped by usql %s\n", d.opts.Version)
	}
	fmt.Fprintf(w, "-- Source dialect: %s\n", d.snap.Dialect)
	if d.snap.Server != "" {
		fmt.Fprintf(w, "-- Source version: %s\n", oneLine(d.snap.Server))
	}
	fmt.Fprintf(w, "-- Target dialect: %s\n", d.target)
	switch {
	case d.opts.SchemaOnly:
		fmt.Fprintln(w, "-- Contents: schema only")
	case d.opts.DataOnly:
		fmt.Fprintln(w, "-- Contents: data only")
	}
	fmt.Fprintln(w, "--")
	fmt.Fprintln(w)
	fmt.Fprintln(w, `\uset on_error_stop on`)
	fmt.Fprintf(w, "\\restrict %s\n", key)
}

// skip reports whether the options leave a statement out. The statements
// about the rows are the settings around them, the checks of the rows, and
// the next values of the sequences. The rest is the schema.
func (d *Dumper) skip(st dbdump.Statement) bool {
	data := st.Stage == dbdump.Data
	switch st.Kind {
	case "setting", "check", "sequence value":
		data = true
	}
	switch {
	case d.opts.SchemaOnly:
		return data
	case d.opts.DataOnly:
		return !data
	}
	return false
}

// statement writes one statement, with a comment that names it. A statement
// in the dialect of the source for another dialect is written as a comment,
// because it was not translated and would stop the script.
func (d *Dumper) statement(w io.Writer, st dbdump.Statement) {
	fmt.Fprintln(w)
	name := st.Name
	if name == "" {
		name = "-"
	}
	fmt.Fprintf(w, "-- Name: %s; Type: %s; Stage: %s\n", name, st.Kind, st.Stage)
	if st.Note != "" {
		fmt.Fprintf(w, "-- Note: %s\n", oneLine(st.Note))
	}
	switch {
	case st.SQL == "":
	case st.Untranslated:
		fmt.Fprintf(w, "-- Not translated from %s:\n", d.snap.Dialect)
		for line := range strings.SplitSeq(st.SQL, "\n") {
			fmt.Fprintf(w, "-- %s\n", line)
		}
	default:
		fmt.Fprintln(w, terminate(st.SQL))
	}
}

// rows writes the rows of each table, in the order of the CREATE TABLE
// statements of the script. A product that checks a foreign key on each row,
// and cannot add one to a table that exists, then finds the rows of a parent
// before the rows of a child.
func (d *Dumper) rows(ctx context.Context, w *errWriter, stmts []dbdump.Statement) error {
	if d.opts.SchemaOnly {
		return nil
	}
	tables := make(map[string][2]string)
	var order [][2]string
	for _, sc := range d.snap.Schemas {
		for _, t := range sc.Tables {
			// a query of the parent returns the rows of its partitions, so the
			// rows of a partition are read only once
			if t.IsPartition() {
				continue
			}
			k := [2]string{sc.Name, t.Name}
			tables[sc.Name+"."+t.Name] = k
			order = append(order, k)
		}
	}
	// the order of the script, then any table that it does not create, as in
	// a dump of data only
	var seq [][2]string
	seen := make(map[[2]string]bool)
	for _, st := range stmts {
		if k, ok := tables[st.Name]; ok && st.Kind == "table" && !seen[k] {
			seq, seen[k] = append(seq, k), true
		}
	}
	for _, k := range order {
		if !seen[k] {
			seq = append(seq, k)
		}
	}
	for _, k := range seq {
		if err := d.table(ctx, w, k[0], k[1]); err != nil {
			return fmt.Errorf("dumping the rows of %s.%s: %w", k[0], k[1], err)
		}
	}
	return nil
}

// table writes the rows of one table, a batch to each INSERT, between the
// statements that the target runs before and after them.
func (d *Dumper) table(ctx context.Context, w *errWriter, schema, table string) error {
	l, err := dbdump.NewLoader(d.snap, schema, table, d.target, d.policy)
	if err != nil {
		return err
	}
	if err := d.sourceSession(ctx, l); err != nil {
		return err
	}
	rows, err := d.q.QueryContext(ctx, l.SelectAll())
	if err != nil {
		return err
	}
	defer rows.Close()
	n := len(l.Columns())
	size := max(l.BatchSize(), 1)
	batch := make([][]any, 0, size)
	first := true
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if first {
			fmt.Fprintf(w, "\n-- Name: %s.%s; Type: data; Stage: data\n", schema, table)
			if !l.Ordered() {
				fmt.Fprintln(w, "-- Note: the table has no primary key, so the order of the rows is the order that the source gave")
			}
			// SQL Server takes a value for an identity column only after SET
			// IDENTITY_INSERT
			if s := l.Prologue(); s != "" {
				fmt.Fprintln(w, terminate(s))
			}
			first = false
		}
		s, err := l.Insert(batch)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, terminate(s))
		batch = batch[:0]
		return w.err
	}
	for rows.Next() {
		vals, ptrs := make([]any, n), make([]any, n)
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		if batch = append(batch, vals); len(batch) == size {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := flush(); err != nil {
		return err
	}
	if s := l.Epilogue(); s != "" && !first {
		fmt.Fprintln(w, terminate(s))
	}
	return w.err
}

// sourceSession runs the statements that dbmeta gives for the session of the
// source, once, before the first table is read. MySQL sets the zone of the
// session to UTC, so that a TIMESTAMP is read as the loader writes it. The
// setting stays on the connection after the dump.
func (d *Dumper) sourceSession(ctx context.Context, l *dbdump.Loader) error {
	if d.session {
		return nil
	}
	d.session = true
	for _, stmt := range l.SourceSession() {
		rows, err := d.q.QueryContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("setting the session of the source: %w", err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("setting the session of the source: %w", err)
		}
	}
	return nil
}

// supported returns the dialects whose capability for a dump has ok, in a
// sorted list such as "duckdb, mysql and postgres", so that a refusal names the
// dialects that work.
func supported(ok func(dbdump.Capability) bool) string {
	var names []string
	for _, d := range dbmeta.Dialects() {
		if ok(dbdump.Support(d)) {
			names = append(names, string(d))
		}
	}
	slices.Sort(names)
	switch len(names) {
	case 0:
		return "no other"
	case 1:
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// terminate ends a statement with a semicolon. The semicolon goes on a line of
// its own when the last line holds a line comment, because a semicolon at the
// end of the comment is part of the comment. The text of a view of SQL Server
// can end in a comment such as "--ORDER BY City;". When the "--" is inside a
// string instead, the extra semicolon is an empty statement, which usql skips.
func terminate(s string) string {
	s = strings.TrimRight(s, " \t\r\n")
	last := s[strings.LastIndexByte(s, '\n')+1:]
	switch {
	case strings.Contains(last, "--"):
		return s + "\n;"
	case strings.HasSuffix(s, ";"):
		return s
	}
	return s + ";"
}

// oneLine joins the lines of s, for a comment.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// randomKey returns a key for \restrict, of 32 letters and digits, as pg_dump
// writes one.
func randomKey() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// errWriter keeps the first error of a writer, so that a run of writes checks
// it once.
type errWriter struct {
	w   io.Writer
	err error
}

// Write satisfies io.Writer.
func (e *errWriter) Write(p []byte) (int, error) {
	if e.err != nil {
		return 0, e.err
	}
	n, err := e.w.Write(p)
	e.err = err
	return n, err
}

// Begin opens a transaction that reads one state of the database, so that the
// schema and the rows of a dump agree. It asks for a read only transaction at
// the isolation level repeatable read, which is a snapshot in PostgreSQL and
// MySQL. A driver that refuses the level, or a read only transaction, gets the
// strongest that it takes.
func Begin(ctx context.Context, db interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
},
) (*sql.Tx, error) {
	var errs []error
	for _, opts := range []*sql.TxOptions{
		{Isolation: sql.LevelRepeatableRead, ReadOnly: true},
		{Isolation: sql.LevelSnapshot, ReadOnly: true},
		{ReadOnly: true},
		nil,
	} {
		tx, err := db.BeginTx(ctx, opts)
		if err == nil {
			return tx, nil
		}
		errs = append(errs, err)
	}
	return nil, errors.Join(errs...)
}
