package describe

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/xo/dbmeta"
	"github.com/xo/usql/text"
)

// Formatter formats the answers of the describe commands, such as \d and \dt,
// the way psql formats them. It reads each answer from dbmeta, and writes the
// result to its output.
type Formatter struct {
	meta *dbmeta.Meta
	db   dbmeta.Queryer
	out  io.Writer
	// name is the name of the driver, for the message when a command is not
	// supported.
	name string

	// The options of the session, such as ECHO_HIDDEN.
	logger  Logger
	dryRun  bool
	timeout time.Duration
}

// NewFormatter returns a Formatter that reads the database db through m, and
// writes to out. name names the driver in the message for a command that the
// database cannot answer.
func NewFormatter(m *dbmeta.Meta, db dbmeta.Queryer, out io.Writer, name string, opts ...Option) *Formatter {
	f := &Formatter{meta: m, db: db, out: out, name: name}
	for _, o := range opts {
		o(f)
	}
	return f
}

// notSupported returns the error for a command that the database cannot
// answer.
func (f *Formatter) notSupported(cmd string) error {
	return fmt.Errorf(text.NotSupportedByDriver, cmd, f.name)
}

// check turns ErrNotSupported into the message for cmd.
func (f *Formatter) check(cmd string, err error) error {
	if errors.Is(err, ErrNotSupported) {
		return f.notSupported(cmd)
	}
	return err
}

// The classes of relation that \d lists, keyed by the letter of the command:
// \dt tables, \dv views, \dm materialized views, \ds sequences and \dE
// foreign tables.
const (
	classTable        = 't'
	classView         = 'v'
	classMaterialized = 'm'
	classSequence     = 's'
	classForeign      = 'E'
)

// classOf returns the class of a relation from its type, which each database
// spells its own way, such as "BASE TABLE" or "materialized view".
func classOf(typ string) rune {
	switch t := strings.ToLower(typ); {
	case strings.Contains(t, "materialized"):
		return classMaterialized
	case strings.Contains(t, "sequence"):
		return classSequence
	case strings.Contains(t, "foreign"):
		return classForeign
	case strings.Contains(t, "view"):
		return classView
	}
	return classTable
}

// className returns the name that psql prints for a class.
func className(class rune) string {
	switch class {
	case classView:
		return "view"
	case classMaterialized:
		return "materialized view"
	case classSequence:
		return "sequence"
	case classForeign:
		return "foreign table"
	}
	return "table"
}

// classes returns the classes named by the letters of types, such as "tv".
// No letter means every class.
func classes(types string) map[rune]bool {
	m := make(map[rune]bool)
	for _, c := range types {
		switch c {
		case classTable, classView, classMaterialized, classSequence, classForeign:
			m[c] = true
		}
	}
	if len(m) == 0 {
		for _, c := range "tvmsE" {
			m[c] = true
		}
	}
	return m
}

// relations returns the relations that match pattern, sorted by schema and
// name. When sequences are asked for and the database lists them apart from
// its tables, they are added.
func (f *Formatter) relations(ctx context.Context, pattern string, withSystem bool, want map[rune]bool) ([]dbmeta.Table, error) {
	schema, name := f.meta.Dialect().Pattern(pattern)
	tables, err := all(ctx, f, dbmeta.Tables, dbmeta.Args{Schema: schema, Name: name, WithSystem: withSystem})
	if err != nil {
		return nil, err
	}
	var res []dbmeta.Table
	seen := make(map[string]bool)
	for _, t := range tables {
		if want[classOf(t.Type)] {
			res = append(res, t)
			seen[t.Schema+"."+t.Name] = true
		}
	}
	if want[classSequence] && f.supports(dbmeta.Sequences) {
		sequences, err := all(ctx, f, dbmeta.Sequences, dbmeta.Args{Schema: schema, Name: name, WithSystem: withSystem})
		if err != nil {
			return nil, err
		}
		for _, s := range sequences {
			if !seen[s.Schema+"."+s.Name] {
				res = append(res, dbmeta.Table{Schema: s.Schema, Name: s.Name, Type: "sequence", Comment: s.Comment})
			}
		}
	}
	slices.SortStableFunc(res, func(a, b dbmeta.Table) int {
		return cmp.Or(cmp.Compare(a.Schema, b.Schema), cmp.Compare(a.Name, b.Name))
	})
	return res, nil
}

// notFound prints the line that psql prints when nothing matches.
func (f *Formatter) notFound(pattern string) error {
	if pattern == "" {
		_, err := fmt.Fprintln(f.out, text.NoRelationsFound)
		return err
	}
	_, err := fmt.Fprintf(f.out, text.RelationNotFound+"\n", pattern)
	return err
}

// ListTables prints the relations of the classes named by types that match
// pattern, for \dt, \dv, \dm, \ds and \d with no pattern.
func (f *Formatter) ListTables(ctx context.Context, types, pattern string, verbose, showSystem bool) error {
	tables, err := f.relations(ctx, pattern, showSystem, classes(types))
	if err != nil {
		return f.check(`\dt`, err)
	}
	if len(tables) == 0 {
		return f.notFound(pattern)
	}
	r := &rows{columns: []string{"Schema", "Name", "Type"}}
	if verbose {
		r.columns = append(r.columns, "Description")
	}
	for _, t := range tables {
		v := []any{t.Schema, t.Name, className(classOf(t.Type))}
		if verbose {
			v = append(v, null(t.Comment))
		}
		r.add(v...)
	}
	return encode(f.out, r, text.ListOfRelations, nil, nil)
}

// DescribeTableDetails prints each relation and each index that matches
// pattern, for \d NAME.
func (f *Formatter) DescribeTableDetails(ctx context.Context, pattern string, verbose, showSystem bool) error {
	tables, err := f.relations(ctx, pattern, showSystem, classes(""))
	if err != nil {
		return f.check(`\d`, err)
	}
	var found bool
	for _, t := range tables {
		found = true
		if classOf(t.Type) == classSequence {
			err = f.describeSequence(ctx, t)
		} else {
			err = f.describeRelation(ctx, t, verbose)
		}
		if err != nil {
			return err
		}
	}
	n, err := f.describeIndexes(ctx, pattern, showSystem)
	if err != nil {
		return err
	}
	if !found && n == 0 {
		return f.notFound(pattern)
	}
	return nil
}

// title returns the title of a relation, such as `Table "public.film"`.
func title(class rune, schema, name string) string {
	s := className(class)
	return strings.ToUpper(s[:1]) + s[1:] + " " + qualified(schema, name)
}

// describeRelation prints the columns of a table or a view, and below them its
// indexes, constraints, references and triggers.
func (f *Formatter) describeRelation(ctx context.Context, t dbmeta.Table, verbose bool) error {
	columns, err := all(ctx, f, dbmeta.Columns, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true})
	if err != nil {
		return f.check(`\d`, err)
	}
	slices.SortStableFunc(columns, func(a, b dbmeta.Column) int {
		return cmp.Compare(a.Ordinal, b.Ordinal)
	})
	r := &rows{columns: []string{"Column", "Type", "Nullable", "Default"}}
	if verbose {
		r.columns = append(r.columns, "Description")
	}
	for _, c := range columns {
		// the columns of another table whose name the pattern also matches
		if c.Table != t.Name {
			continue
		}
		nullable := ""
		if !c.Nullable {
			nullable = "not null"
		}
		v := []any{c.Name, c.DataType, nullable, null(c.Default)}
		if verbose {
			v = append(v, null(c.Comment))
		}
		r.add(v...)
	}
	footer, err := f.relationFooter(ctx, t)
	if err != nil {
		return err
	}
	return encode(f.out, r, title(classOf(t.Type), t.Schema, t.Name), footer, nil)
}

// relationFooter returns the lines that psql prints below the columns of a
// table: its indexes, check constraints, foreign keys, the foreign keys that
// refer to it, and its triggers. A section the database cannot answer is left
// out.
func (f *Formatter) relationFooter(ctx context.Context, t dbmeta.Table) (func(io.Writer) error, error) {
	var lines []string
	section := func(header string, items []string) {
		if len(items) == 0 {
			return
		}
		lines = append(lines, header+":")
		for _, s := range items {
			lines = append(lines, "    "+s)
		}
	}
	indexes, err := f.tableIndexes(ctx, t)
	if err != nil {
		return nil, err
	}
	section("Indexes", indexes)
	checks, foreign, err := f.tableConstraints(ctx, t)
	if err != nil {
		return nil, err
	}
	section("Check constraints", checks)
	section("Foreign-key constraints", foreign)
	referenced, err := f.referencedBy(ctx, t)
	if err != nil {
		return nil, err
	}
	section("Referenced by", referenced)
	triggers, err := f.tableTriggers(ctx, t)
	if err != nil {
		return nil, err
	}
	section("Triggers", triggers)
	if len(lines) == 0 {
		return noFooter, nil
	}
	return func(out io.Writer) error {
		_, err := fmt.Fprintln(out, strings.Join(lines, "\n"))
		return err
	}, nil
}

// optional returns rows, or none when the database cannot answer the query,
// so that a section it cannot answer is left out rather than failing \d.
func optional[T any](rows []T, err error) ([]T, error) {
	if errors.Is(err, ErrNotSupported) {
		return nil, nil
	}
	return rows, err
}

// indexColumns returns the columns of each index of a table, keyed by the
// name of the index, in their order in the index.
func (f *Formatter) indexColumns(ctx context.Context, schema, table string) (map[string][]string, error) {
	cols, err := optional(all(ctx, f, dbmeta.IndexColumns, dbmeta.Args{Schema: schema, Parent: table, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(cols, func(a, b dbmeta.IndexColumn) int {
		return cmp.Compare(a.Ordinal, b.Ordinal)
	})
	m := make(map[string][]string)
	for _, c := range cols {
		if c.Table != table {
			continue
		}
		name := null(c.Name)
		if name == "" {
			name = null(c.Expression)
		}
		if c.Descending {
			name += " DESC"
		}
		m[c.Index] = append(m[c.Index], name)
	}
	return m, nil
}

// method returns the access method of an index in lower case, such as btree.
// SQLite reports how the index came to be instead, such as "primary key",
// which repeats what the line already says, so that is left out.
func method(i dbmeta.Index) string {
	switch t := strings.ToLower(strings.TrimSpace(i.Type)); t {
	case "primary key", "unique", "pk", "u":
		return ""
	default:
		return t
	}
}

// indexLine returns the line of one index, the way psql writes it, such as
// `"film_pkey" PRIMARY KEY, btree (film_id)`.
func indexLine(i dbmeta.Index, cols []string) string {
	var parts []string
	switch {
	case i.Primary:
		parts = append(parts, "PRIMARY KEY")
	case i.Unique:
		parts = append(parts, "UNIQUE")
	}
	if t := method(i); t != "" {
		parts = append(parts, t)
	}
	s := `"` + i.Name + `"`
	if len(parts) != 0 {
		s += " " + strings.Join(parts, ", ")
	}
	return s + " (" + strings.Join(cols, ", ") + ")"
}

// tableIndexes returns a line for each index of a table.
func (f *Formatter) tableIndexes(ctx context.Context, t dbmeta.Table) ([]string, error) {
	indexes, err := optional(all(ctx, f, dbmeta.Indexes, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}))
	if err != nil || len(indexes) == 0 {
		return nil, err
	}
	cols, err := f.indexColumns(ctx, t.Schema, t.Name)
	if err != nil {
		return nil, err
	}
	// psql lists the primary key first, then the rest by name
	slices.SortStableFunc(indexes, func(a, b dbmeta.Index) int {
		if a.Primary != b.Primary {
			if a.Primary {
				return -1
			}
			return 1
		}
		return cmp.Compare(a.Name, b.Name)
	})
	var lines []string
	for _, i := range indexes {
		if i.Table == t.Name {
			lines = append(lines, indexLine(i, cols[i.Name]))
		}
	}
	return lines, nil
}

// kind returns the kind of a constraint in lower case, such as "foreign key".
func kind(typ string) string {
	return strings.ToLower(strings.TrimSpace(typ))
}

// constraintColumns returns the columns of each constraint of a table, and the
// columns and table that each one refers to, keyed by the name of the
// constraint.
func (f *Formatter) constraintColumns(ctx context.Context, a dbmeta.Args) (map[string]*reference, error) {
	cols, err := optional(all(ctx, f, dbmeta.ConstraintColumns, a))
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(cols, func(a, b dbmeta.ConstraintColumn) int {
		return cmp.Compare(a.Ordinal, b.Ordinal)
	})
	m := make(map[string]*reference)
	for _, c := range cols {
		key := c.Table + "." + c.Constraint
		r, ok := m[key]
		if !ok {
			r = &reference{table: c.Table, name: c.Constraint, foreignSchema: null(c.ForeignSchema), foreignTable: null(c.ForeignTable)}
			m[key] = r
		}
		r.columns = append(r.columns, c.Name)
		if n := null(c.ForeignName); n != "" {
			r.foreign = append(r.foreign, n)
		}
	}
	return m, nil
}

// reference is a foreign key, read from its columns.
type reference struct {
	table         string
	name          string
	columns       []string
	foreignSchema string
	foreignTable  string
	foreign       []string
}

// definition returns the definition of a foreign key built from its columns,
// for a database that records no definition of its own.
func (r *reference) definition() string {
	return fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)",
		strings.Join(r.columns, ", "), r.foreignTable, strings.Join(r.foreign, ", "))
}

// isForeignKey reports whether def is the definition of a foreign key written
// as SQL, the way psql prints one. A database that records no definition, such
// as SQLite, gets one built from the columns of the key.
func isForeignKey(def string) bool {
	_, ok := cutPrefixFold(def, "FOREIGN KEY")
	return ok
}

// tableConstraints returns a line for each check constraint and each foreign
// key of a table.
func (f *Formatter) tableConstraints(ctx context.Context, t dbmeta.Table) ([]string, []string, error) {
	constraints, err := optional(all(ctx, f, dbmeta.Constraints, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}))
	if err != nil || len(constraints) == 0 {
		return nil, nil, err
	}
	var refs map[string]*reference
	slices.SortStableFunc(constraints, func(a, b dbmeta.Constraint) int {
		return cmp.Compare(a.Name, b.Name)
	})
	var checks, foreign []string
	for _, c := range constraints {
		if c.Table != t.Name {
			continue
		}
		def := null(c.Definition)
		switch kind(c.Type) {
		case "check":
			// information_schema records only the expression
			if !strings.HasPrefix(strings.ToUpper(def), "CHECK") {
				def = "CHECK (" + def + ")"
			}
			checks = append(checks, `"`+c.Name+`" `+def)
		case "foreign key":
			if !isForeignKey(def) {
				if refs == nil {
					if refs, err = f.constraintColumns(ctx, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}); err != nil {
						return nil, nil, err
					}
				}
				if r, ok := refs[t.Name+"."+c.Name]; ok {
					def = r.definition()
				}
			}
			foreign = append(foreign, `"`+c.Name+`" `+def)
		}
	}
	return checks, foreign, nil
}

// referencedBy returns a line for each foreign key of another table that
// refers to a table.
func (f *Formatter) referencedBy(ctx context.Context, t dbmeta.Table) ([]string, error) {
	if !f.supports(dbmeta.ConstraintColumns) {
		return nil, nil
	}
	refs, err := f.constraintColumns(ctx, dbmeta.Args{Schema: t.Schema, WithSystem: true})
	if err != nil {
		return nil, err
	}
	var matched []*reference
	for _, r := range refs {
		if r.foreignTable == t.Name && (r.foreignSchema == "" || t.Schema == "" || r.foreignSchema == t.Schema) {
			matched = append(matched, r)
		}
	}
	if len(matched) == 0 {
		return nil, nil
	}
	// the definitions come from the constraints of the tables that refer
	defs := make(map[string]string)
	constraints, err := optional(all(ctx, f, dbmeta.Constraints, dbmeta.Args{Schema: t.Schema, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	for _, c := range constraints {
		if kind(c.Type) == "foreign key" {
			defs[c.Table+"."+c.Name] = null(c.Definition)
		}
	}
	slices.SortStableFunc(matched, func(a, b *reference) int {
		return cmp.Or(cmp.Compare(a.table, b.table), cmp.Compare(a.name, b.name))
	})
	var lines []string
	for _, r := range matched {
		def := defs[r.table+"."+r.name]
		if !isForeignKey(def) {
			def = r.definition()
		}
		lines = append(lines, fmt.Sprintf(`TABLE "%s" CONSTRAINT "%s" %s`, r.table, r.name, def))
	}
	return lines, nil
}

// tableTriggers returns a line for each trigger of a table.
func (f *Formatter) tableTriggers(ctx context.Context, t dbmeta.Table) ([]string, error) {
	triggers, err := optional(all(ctx, f, dbmeta.Triggers, dbmeta.Args{Schema: t.Schema, Parent: t.Name, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(triggers, func(a, b dbmeta.Trigger) int {
		return cmp.Compare(a.Name, b.Name)
	})
	var lines []string
	for _, tr := range triggers {
		if tr.Table != t.Name {
			continue
		}
		def := strings.TrimSpace(tr.Definition)
		// psql prints what follows CREATE TRIGGER
		if s, ok := cutPrefixFold(def, "CREATE TRIGGER "); ok {
			def = s
		}
		// a definition that does not name the trigger, as MySQL's, gets its name
		if !strings.Contains(def, tr.Name) {
			def = strings.TrimSpace(tr.Name + " " + def)
		}
		lines = append(lines, def)
	}
	return lines, nil
}

// cutPrefixFold is strings.CutPrefix, ignoring case.
func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):], true
	}
	return s, false
}

// describeSequence prints one sequence.
func (f *Formatter) describeSequence(ctx context.Context, t dbmeta.Table) error {
	sequences, err := optional(all(ctx, f, dbmeta.Sequences, dbmeta.Args{Schema: t.Schema, Name: t.Name, WithSystem: true}))
	if err != nil {
		return err
	}
	r := &rows{columns: []string{"Type", "Start", "Minimum", "Maximum", "Increment", "Cycles?"}}
	var ownedBy string
	for _, s := range sequences {
		if s.Name != t.Name || s.Schema != t.Schema {
			continue
		}
		cycles := ""
		if s.Cycles.Valid {
			cycles = "no"
			if s.Cycles.V {
				cycles = "yes"
			}
		}
		r.add(null(s.DataType), null(s.Start), null(s.Minimum), null(s.Maximum), null(s.Increment), cycles)
		ownedBy = s.OwnedBy
		break
	}
	footer := noFooter
	if ownedBy != "" {
		footer = func(out io.Writer) error {
			_, err := fmt.Fprintln(out, "Owned by: "+ownedBy)
			return err
		}
	}
	return encode(f.out, r, title(classSequence, t.Schema, t.Name), footer, nil)
}

// describeIndexes prints each index whose name matches pattern, and returns
// how many it printed.
func (f *Formatter) describeIndexes(ctx context.Context, pattern string, showSystem bool) (int, error) {
	if !f.supports(dbmeta.Indexes) {
		return 0, nil
	}
	schema, name := f.meta.Dialect().Pattern(pattern)
	indexes, err := all(ctx, f, dbmeta.Indexes, dbmeta.Args{Schema: schema, Name: name, WithSystem: showSystem})
	if err != nil {
		return 0, f.check(`\d`, err)
	}
	slices.SortStableFunc(indexes, func(a, b dbmeta.Index) int {
		return cmp.Or(cmp.Compare(a.Schema, b.Schema), cmp.Compare(a.Name, b.Name))
	})
	var n int
	for _, i := range indexes {
		cols, err := f.indexColumns(ctx, i.Schema, i.Table)
		if err != nil {
			return n, err
		}
		r := &rows{columns: []string{"Column"}}
		for _, c := range cols[i.Name] {
			r.add(c)
		}
		var parts []string
		switch {
		case i.Primary:
			parts = append(parts, "primary key")
		case i.Unique:
			parts = append(parts, "unique")
		}
		if t := method(i); t != "" {
			parts = append(parts, t)
		}
		parts = append(parts, "for table "+qualified(i.Schema, i.Table))
		footer := func(out io.Writer) error {
			_, err := fmt.Fprintln(out, strings.Join(parts, ", "))
			return err
		}
		if err := encode(f.out, r, "Index "+qualified(i.Schema, i.Name), footer, nil); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// functionKind returns the name that psql prints for the kind of a function,
// and the letter of \df that selects it: a aggregate, n normal, p procedure,
// t trigger and w window.
func functionKind(kind string) (string, rune) {
	switch k := strings.ToLower(kind); {
	case strings.HasPrefix(k, "agg"):
		return "agg", 'a'
	case strings.HasPrefix(k, "proc"):
		return "proc", 'p'
	case strings.HasPrefix(k, "trigger"):
		return "trigger", 't'
	case strings.HasPrefix(k, "window"):
		return "window", 'w'
	}
	return "func", 'n'
}

// DescribeFunctions prints the functions that match pattern, for \df and
// \da. funcTypes is the name of the command, and a letter after its first
// two, as in dfn, selects the kinds of function.
func (f *Formatter) DescribeFunctions(ctx context.Context, funcTypes, pattern string, verbose, showSystem bool) error {
	cmd := `\` + funcTypes
	aggregates := strings.HasPrefix(funcTypes, "da")
	want := make(map[rune]bool)
	if len(funcTypes) > 2 {
		for _, c := range funcTypes[2:] {
			want[c] = true
		}
	}
	q := dbmeta.Functions
	if aggregates {
		if f.supports(dbmeta.Aggregates) {
			q = dbmeta.Aggregates
		} else {
			want = map[rune]bool{'a': true}
		}
	}
	schema, name := f.meta.Dialect().Pattern(pattern)
	funcs, err := all(ctx, f, q, dbmeta.Args{Schema: schema, Name: name, WithSystem: showSystem})
	if err != nil {
		return f.check(cmd, err)
	}
	slices.SortStableFunc(funcs, func(a, b dbmeta.Function) int {
		return cmp.Or(cmp.Compare(a.Schema, b.Schema), cmp.Compare(a.Name, b.Name), cmp.Compare(null(a.ArgTypes), null(b.ArgTypes)))
	})
	r := &rows{columns: []string{"Schema", "Name", "Result data type", "Argument data types"}}
	switch {
	case aggregates:
		r.columns = append(r.columns, "Description")
	case verbose:
		r.columns = append(r.columns, "Type", "Volatility", "Parallel", "Owner", "Security", "Access privileges", "Language", "Source code", "Description")
	default:
		r.columns = append(r.columns, "Type")
	}
	for _, fn := range funcs {
		typ, letter := functionKind(fn.Kind)
		if len(want) != 0 && !want[letter] {
			continue
		}
		v := []any{fn.Schema, fn.Name, null(fn.ResultType), null(fn.ArgTypes)}
		switch {
		case aggregates:
			v = append(v, null(fn.Comment))
		case verbose:
			v = append(v, typ, fn.Volatility, fn.Parallel, null(fn.Owner), fn.Security, null(fn.Access), fn.Language, null(fn.Source), null(fn.Comment))
		default:
			v = append(v, typ)
		}
		r.add(v...)
	}
	title := text.ListOfFunctions
	if aggregates {
		title = text.ListOfAggregates
	}
	return encode(f.out, r, title, nil, nil)
}

// ListAllDbs prints the databases that match pattern, for \l.
func (f *Formatter) ListAllDbs(ctx context.Context, pattern string, verbose bool) error {
	dbs, err := all(ctx, f, dbmeta.Databases, dbmeta.Args{Name: like(pattern)})
	if err != nil {
		return f.check(`\l`, err)
	}
	slices.SortStableFunc(dbs, func(a, b dbmeta.Database) int {
		return cmp.Compare(a.Name, b.Name)
	})
	r := &rows{columns: []string{"Name", "Owner", "Encoding", "Collate", "Ctype", "Access privileges"}}
	if verbose {
		r.columns = append(r.columns, "Size", "Tablespace", "Description")
	}
	for _, d := range dbs {
		v := []any{d.Name, d.Owner, d.Encoding, d.Collate, d.CType, null(d.Access)}
		if verbose {
			v = append(v, null(d.Size), null(d.Tablespace), null(d.Comment))
		}
		r.add(v...)
	}
	return encode(f.out, r, text.ListOfDatabases, nil, nil)
}

// ListSchemas prints the schemas that match pattern, for \dn.
func (f *Formatter) ListSchemas(ctx context.Context, pattern string, verbose, showSystem bool) error {
	schemas, err := all(ctx, f, dbmeta.Schemas, dbmeta.Args{Name: like(pattern), WithSystem: showSystem})
	if err != nil {
		return f.check(`\dn`, err)
	}
	slices.SortStableFunc(schemas, func(a, b dbmeta.Schema) int {
		return cmp.Compare(a.Name, b.Name)
	})
	r := &rows{columns: []string{"Name", "Owner"}}
	if verbose {
		r.columns = append(r.columns, "Description")
	}
	for _, s := range schemas {
		v := []any{s.Name, s.Owner}
		if verbose {
			v = append(v, null(s.Comment))
		}
		r.add(v...)
	}
	return encode(f.out, r, text.ListOfSchemas, nil, nil)
}

// ListIndexes prints the indexes that match pattern, for \di.
func (f *Formatter) ListIndexes(ctx context.Context, pattern string, verbose, showSystem bool) error {
	schema, name := f.meta.Dialect().Pattern(pattern)
	indexes, err := all(ctx, f, dbmeta.Indexes, dbmeta.Args{Schema: schema, Name: name, WithSystem: showSystem})
	if err != nil {
		return f.check(`\di`, err)
	}
	slices.SortStableFunc(indexes, func(a, b dbmeta.Index) int {
		return cmp.Or(cmp.Compare(a.Schema, b.Schema), cmp.Compare(a.Name, b.Name))
	})
	r := &rows{columns: []string{"Schema", "Name", "Type", "Table"}}
	if verbose {
		r.columns = append(r.columns, "Access method", "Description")
	}
	for _, i := range indexes {
		v := []any{i.Schema, i.Name, "index", i.Table}
		if verbose {
			v = append(v, method(i), null(i.Comment))
		}
		r.add(v...)
	}
	if len(r.values) == 0 {
		return f.notFound(pattern)
	}
	return encode(f.out, r, text.ListOfRelations, nil, nil)
}

// ShowStats prints the statistics of the columns of the tables that match
// pattern, for \ss. k limits the number of common values shown.
func (f *Formatter) ShowStats(ctx context.Context, pattern string, verbose bool, k int) error {
	schema, name := f.meta.Dialect().Pattern(pattern)
	stats, err := all(ctx, f, dbmeta.ColumnStats, dbmeta.Args{Schema: schema, Parent: name})
	if err != nil {
		return f.check(`\ss`, err)
	}
	if len(stats) == 0 {
		return f.notFound(pattern)
	}
	r := &rows{columns: []string{"Schema", "Table", "Name", "Average width", "Nulls fraction", "Distinct values"}}
	if verbose {
		r.columns = append(r.columns, "Minimum value", "Maximum value", "Mean value", "Top N common values", "Top N values freqs")
	}
	for _, s := range stats {
		v := []any{s.Schema, s.Table, s.Name, null(s.AvgWidth), float(s.NullFrac), float(s.Distinct)}
		if verbose {
			v = append(v, null(s.Min), null(s.Max), null(s.Mean), top(null(s.TopN), k), top(null(s.TopNFreqs), k))
		}
		r.add(v...)
	}
	return encode(f.out, r, text.ColumnStats, nil, nil)
}

// float returns a fraction as psql prints it, or an empty string for NULL.
func float(v sql.Null[float64]) string {
	if !v.Valid {
		return ""
	}
	return fmt.Sprintf("%.4f", v.V)
}

// top returns the first k of a list of values separated by commas, or all of
// them when k is 0.
func top(s string, k int) string {
	if k <= 0 || s == "" {
		return s
	}
	parts := strings.Split(s, ",")
	if len(parts) > k {
		parts = parts[:k]
	}
	return strings.Join(parts, ",")
}

// ListPrivilegeSummaries prints the access privileges of the relations that
// match pattern, for \dp.
func (f *Formatter) ListPrivilegeSummaries(ctx context.Context, pattern string, showSystem bool) error {
	schema, name := f.meta.Dialect().Pattern(pattern)
	privs, err := all(ctx, f, dbmeta.Privileges, dbmeta.Args{Schema: schema, Name: name, WithSystem: showSystem})
	if err != nil {
		return f.check(`\dp`, err)
	}
	slices.SortStableFunc(privs, func(a, b dbmeta.Privilege) int {
		return cmp.Or(cmp.Compare(null(a.Schema), null(b.Schema)), cmp.Compare(a.Name, b.Name))
	})
	r := &rows{columns: []string{"Schema", "Name", "Type", "Access privileges", "Column privileges", "Policies"}}
	for _, p := range privs {
		r.add(null(p.Schema), p.Name, strings.ToLower(p.Type), null(p.Access), null(p.ColumnAccess), null(p.Policies))
	}
	return encode(f.out, r, text.AccessPrivileges, nil, nil)
}
