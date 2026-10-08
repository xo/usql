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

	// path is the schemas that a name resolves in, read when a name is first
	// printed
	path     map[string]bool
	pathRead bool
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
func (f *Formatter) relations(ctx context.Context, pattern string, withSystem bool, want map[rune]bool, composite bool) ([]dbmeta.Table, error) {
	schema, name := f.meta.Dialect().Pattern(pattern)
	tables, err := all(ctx, f, dbmeta.Tables, dbmeta.Args{Schema: schema, Name: name, WithSystem: withSystem})
	if err != nil {
		return nil, err
	}
	// \d NAME also describes a composite type, which no list of tables holds
	if composite && pattern != "" {
		more, err := all(ctx, f, dbmeta.Tables, dbmeta.Args{Schema: schema, Name: name, Types: []string{"composite type"}, WithSystem: withSystem})
		if err == nil {
			tables = append(tables, more...)
		}
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
	tables, err := f.relations(ctx, pattern, showSystem, classes(types), false)
	if err != nil {
		return f.check(`\dt`, err)
	}
	if len(tables) == 0 {
		return f.notFound(pattern)
	}
	// a column that no row has a value for is one the database cannot give
	owner := some(tables, func(t dbmeta.Table) bool { return t.Owner.Valid })
	persistence := verbose && some(tables, func(t dbmeta.Table) bool { return t.Persistence.Valid })
	method := verbose && some(tables, func(t dbmeta.Table) bool { return t.AccessMethod.Valid })
	sizes := verbose && some(tables, func(t dbmeta.Table) bool { return t.Size.Valid })
	counts := verbose && some(tables, func(t dbmeta.Table) bool { return t.Rows.Valid && t.Rows.V >= 0 })
	r := &rows{columns: []string{"Schema", "Name", "Type"}}
	if owner {
		r.columns = append(r.columns, "Owner")
	}
	if persistence {
		r.columns = append(r.columns, "Persistence")
	}
	if method {
		r.columns = append(r.columns, "Access method")
	}
	if sizes {
		r.columns = append(r.columns, "Size")
	}
	if counts {
		r.columns = append(r.columns, "Rows")
	}
	if verbose {
		r.columns = append(r.columns, "Description")
	}
	for _, t := range tables {
		v := []any{t.Schema, t.Name, relationTypeName(t)}
		if owner {
			v = append(v, null(t.Owner))
		}
		if persistence {
			v = append(v, null(t.Persistence))
		}
		if method {
			v = append(v, null(t.AccessMethod))
		}
		if sizes {
			v = append(v, size(t.Size))
		}
		if counts {
			v = append(v, rowCount(t.Rows))
		}
		if verbose {
			v = append(v, null(t.Comment))
		}
		r.add(v...)
	}
	return encode(f.out, r, relationsTitle(classes(types)), nil, nil)
}

// relationsTitle returns the title that psql gives a list of the classes in
// want: the class for one, as in "List of tables", and "List of relations"
// for several.
func relationsTitle(want map[rune]bool) string {
	if len(want) != 1 {
		return text.ListOfRelations
	}
	for class := range want {
		switch class {
		case classTable:
			return "List of tables"
		case classView:
			return "List of views"
		case classMaterialized:
			return "List of materialized views"
		case classSequence:
			return "List of sequences"
		case classForeign:
			return text.ListOfForeignTables
		}
	}
	return text.ListOfRelations
}

// DescribeTableDetails prints each relation and each index that matches
// pattern, for \d NAME.
func (f *Formatter) DescribeTableDetails(ctx context.Context, pattern string, verbose, showSystem bool) error {
	tables, err := f.relations(ctx, pattern, showSystem, classes(""), true)
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
func title(t dbmeta.Table) string {
	name := relationTypeName(t)
	return prefixed(t.Persistence, name) + " " + qualified(t.Schema, t.Name)
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
	// the columns of another table whose name the pattern also matches
	columns = slices.DeleteFunc(columns, func(c dbmeta.Column) bool {
		return c.Table != t.Name
	})
	// psql always lists the collation of a column of PostgreSQL, so a database
	// that gives the storage of a column gives its collation as well
	collation := some(columns, func(c dbmeta.Column) bool { return c.Collation.Valid || c.Storage.Valid })
	storage := verbose && some(columns, func(c dbmeta.Column) bool { return c.Storage.Valid })
	class := classOf(t.Type)
	// psql shows the compression and the statistics target of the columns of
	// a table, and the statistics target of a foreign table, but not of a view
	compression := storage && (class == classTable || class == classMaterialized) && !strings.EqualFold(t.Type, "composite type")
	stats := storage && class != classView && class != classSequence && !strings.EqualFold(t.Type, "composite type")
	r := &rows{columns: []string{"Column", "Type"}}
	if collation {
		r.columns = append(r.columns, "Collation")
	}
	r.columns = append(r.columns, "Nullable", "Default")
	// a foreign table has the options of each column too, which no database
	// gives yet
	foreign := storage && class == classForeign
	if foreign {
		r.columns = append(r.columns, "FDW options")
	}
	if storage {
		r.columns = append(r.columns, "Storage")
	}
	if compression {
		r.columns = append(r.columns, "Compression")
	}
	if stats {
		r.columns = append(r.columns, "Stats target")
	}
	if verbose {
		r.columns = append(r.columns, "Description")
	}
	for _, c := range columns {
		nullable := ""
		if !c.Nullable {
			nullable = "not null"
		}
		v := []any{c.Name, c.DataType}
		if collation {
			v = append(v, columnCollation(c))
		}
		v = append(v, nullable, columnDefault(c))
		if foreign {
			v = append(v, "")
		}
		if storage {
			v = append(v, null(c.Storage))
		}
		if compression {
			v = append(v, null(c.Compression))
		}
		if stats {
			v = append(v, null(c.StatsTarget))
		}
		if verbose {
			v = append(v, null(c.Comment))
		}
		r.add(v...)
	}
	footer, err := f.relationFooter(ctx, t, verbose)
	if err != nil {
		return err
	}
	return encode(f.out, r, title(t), footer, nil)
}

// relationFooter returns the lines that psql prints below the columns of a
// table: its indexes, check constraints, foreign keys, the foreign keys that
// refer to it, and its triggers. A section the database cannot answer is left
// out.
func (f *Formatter) relationFooter(ctx context.Context, t dbmeta.Table, verbose bool) (func(io.Writer) error, error) {
	var lines []string
	block := func(header string, items []string) {
		if len(items) == 0 {
			return
		}
		lines = append(lines, header+":")
		for _, s := range items {
			lines = append(lines, "    "+s)
		}
	}
	add := func(s section) {
		lines = append(lines, s.indented()...)
	}
	before, partitions, err := f.partitionSections(ctx, t, verbose)
	if err != nil {
		return nil, err
	}
	lines = append(lines, before...)
	indexes, err := f.tableIndexes(ctx, t)
	if err != nil {
		return nil, err
	}
	block("Indexes", indexes)
	checks, foreign, err := f.tableConstraints(ctx, t)
	if err != nil {
		return nil, err
	}
	block("Check constraints", checks)
	block("Foreign-key constraints", foreign)
	referenced, err := f.referencedBy(ctx, t)
	if err != nil {
		return nil, err
	}
	block("Referenced by", referenced)
	definition, err := f.viewDefinition(ctx, t)
	if err != nil {
		return nil, err
	}
	lines = append(lines, definition...)
	policies, err := f.policySection(ctx, t)
	if err != nil {
		return nil, err
	}
	add(policies)
	statistics, err := f.statisticsSection(ctx, t)
	if err != nil {
		return nil, err
	}
	add(statistics)
	rules, err := f.ruleSections(ctx, t)
	if err != nil {
		return nil, err
	}
	for _, r := range rules {
		add(r)
	}
	publications, err := f.publicationSection(ctx, t)
	if err != nil {
		return nil, err
	}
	add(publications)
	notNulls, err := f.notNullSection(ctx, t)
	if err != nil {
		return nil, err
	}
	add(notNulls)
	triggers, err := f.tableTriggers(ctx, t)
	if err != nil {
		return nil, err
	}
	block("Triggers", triggers)
	inherits, err := f.inheritanceSections(ctx, t, verbose)
	if err != nil {
		return nil, err
	}
	lines = append(lines, inherits...)
	lines = append(lines, partitions...)
	if classOf(t.Type) == classForeign {
		foreign, err := f.foreignTableLines(ctx, t)
		if err != nil {
			return nil, err
		}
		lines = append(lines, foreign...)
	}
	if verbose {
		if m := null(t.AccessMethod); m != "" {
			lines = append(lines, "Access method: "+m)
		}
		if o := null(t.Options); o != "" {
			lines = append(lines, "Options: "+o)
		}
	}
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
func (f *Formatter) indexColumns(ctx context.Context, schema, table string) (map[string]indexCols, error) {
	cols, err := optional(all(ctx, f, dbmeta.IndexColumns, dbmeta.Args{Schema: schema, Parent: table, WithSystem: true}))
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(cols, func(a, b dbmeta.IndexColumn) int {
		return cmp.Compare(a.Ordinal, b.Ordinal)
	})
	m := make(map[string]indexCols)
	for _, c := range cols {
		if c.Table != table {
			continue
		}
		name := null(c.Name)
		if name == "" {
			name = null(c.Expression)
		}
		if c.Descending.Valid && c.Descending.V {
			name += " DESC"
		}
		ic := m[c.Index]
		if c.Include {
			ic.include = append(ic.include, name)
		} else {
			ic.keys = append(ic.keys, name)
		}
		m[c.Index] = ic
	}
	return m, nil
}

// indexCols are the columns of an index: the keys, and the columns that an
// INCLUDE adds, which the index holds without ordering by them.
type indexCols struct {
	keys, include []string
}

// all returns the keys and then the included columns.
func (c indexCols) all() []string {
	return append(slices.Clone(c.keys), c.include...)
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
// `"film_pkey" PRIMARY KEY, btree (film_id)`. A database that gives the
// definition of the index gets it as it wrote it, and any other gets one built
// from its columns.
func indexLine(i dbmeta.Index, cols indexCols) string {
	s := `"` + i.Name + `"`
	switch ct := null(i.ConstraintType); {
	case ct == "x" && i.ConstraintDefinition.Valid:
		// an exclusion constraint is written as its constraint
		s += " " + i.ConstraintDefinition.V
	case i.Using.Valid:
		switch {
		case i.Primary:
			s += " PRIMARY KEY, "
		case ct == "u":
			s += " UNIQUE CONSTRAINT, "
		case i.Unique:
			s += " UNIQUE, "
		default:
			s += " "
		}
		s += i.Using.V
	default:
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
		if len(parts) != 0 {
			s += " " + strings.Join(parts, ", ")
		}
		s += " (" + strings.Join(cols.keys, ", ") + ")"
		if len(cols.include) != 0 {
			s += " INCLUDE (" + strings.Join(cols.include, ", ") + ")"
		}
		if p := null(i.Predicate); p != "" {
			s += " WHERE " + p
		}
	}
	// the flags that psql writes after the definition
	if i.Valid.Valid && !i.Valid.V {
		s += " INVALID"
	}
	if i.Clustered.Valid && i.Clustered.V {
		s += " CLUSTER"
	}
	if i.Deferrable.Valid && i.Deferrable.V {
		s += " DEFERRABLE"
	}
	if i.InitiallyDeferred.Valid && i.InitiallyDeferred.V {
		s += " INITIALLY DEFERRED"
	}
	if i.ReplicaIdentity.Valid && i.ReplicaIdentity.V {
		s += " REPLICA IDENTITY"
	}
	return s
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
	cache := some(sequences, func(s dbmeta.Sequence) bool { return s.CacheSize.Valid })
	if cache {
		r.columns = append(r.columns, "Cache")
	}
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
		v := []any{null(s.DataType), null(s.Start), null(s.Minimum), null(s.Maximum), null(s.Increment), cycles}
		if cache {
			v = append(v, null(s.CacheSize))
		}
		r.add(v...)
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
	return encode(f.out, r, title(t), footer, nil)
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
		cols, err := optional(all(ctx, f, dbmeta.IndexColumns, dbmeta.Args{Schema: i.Schema, Parent: i.Table, WithSystem: true}))
		if err != nil {
			return n, err
		}
		cols = slices.DeleteFunc(cols, func(c dbmeta.IndexColumn) bool { return c.Table != i.Table || c.Index != i.Name })
		slices.SortStableFunc(cols, func(a, b dbmeta.IndexColumn) int { return cmp.Compare(a.Ordinal, b.Ordinal) })
		// the type of a column of the index is the type of the column of the
		// table that it names
		types := make(map[string]string)
		if tcols, err := optional(all(ctx, f, dbmeta.Columns, dbmeta.Args{Schema: i.Schema, Parent: i.Table, WithSystem: true})); err == nil {
			for _, c := range tcols {
				if c.Table == i.Table {
					types[c.Name] = c.DataType
				}
			}
		}
		r := &rows{columns: []string{"Column", "Type", "Key?", "Definition"}}
		for _, c := range cols {
			def := null(c.Name)
			if def == "" {
				def = null(c.Expression)
			}
			key := "yes"
			if c.Include {
				key = "no"
			}
			name := null(c.Name)
			if name == "" {
				name = def
			}
			r.add(name, types[null(c.Name)], key, def)
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
		if p := null(i.Predicate); p != "" {
			parts = append(parts, "predicate ("+p+")")
		}
		if i.Valid.Valid && !i.Valid.V {
			parts = append(parts, "invalid")
		}
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
		// psql 18 lists a trigger function as a function
		return "func", 't'
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
	// a database that gives no source gives no leakproof flag either
	sources := verbose && some(funcs, func(fn dbmeta.Function) bool { return fn.Prosrc.Valid })
	leakproof := sources
	r := &rows{columns: []string{"Schema", "Name", "Result data type", "Argument data types"}}
	switch {
	case aggregates:
		r.columns = append(r.columns, "Description")
	case verbose:
		r.columns = append(r.columns, "Type", "Volatility", "Parallel", "Owner", "Security")
		if leakproof {
			r.columns = append(r.columns, "Leakproof?")
		}
		r.columns = append(r.columns, "Access privileges", "Language", "Internal name", "Description")
		if sources {
			r.columns = append(r.columns, "Source code")
		}
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
			v = append(v, typ, fn.Volatility, fn.Parallel, null(fn.Owner), fn.Security)
			if leakproof {
				v = append(v, yesNo(fn.Leakproof))
			}
			v = append(v, null(fn.Access), fn.Language, null(fn.Source), null(fn.Comment))
			if sources {
				v = append(v, null(fn.Prosrc))
			}
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
	provider := some(dbs, func(d dbmeta.Database) bool { return d.LocaleProvider.Valid })
	// a database that gives the provider of its locale gives the locale and
	// the ICU rules, and leaves them empty where it has none
	locale, rules := provider, provider
	r := &rows{columns: []string{"Name", "Owner", "Encoding"}}
	if provider {
		r.columns = append(r.columns, "Locale Provider")
	}
	r.columns = append(r.columns, "Collate", "Ctype")
	if locale {
		r.columns = append(r.columns, "Locale")
	}
	if rules {
		r.columns = append(r.columns, "ICU Rules")
	}
	r.columns = append(r.columns, "Access privileges")
	if verbose {
		r.columns = append(r.columns, "Size", "Tablespace", "Description")
	}
	for _, d := range dbs {
		v := []any{d.Name, d.Owner, d.Encoding}
		if provider {
			v = append(v, null(d.LocaleProvider))
		}
		v = append(v, d.Collate, d.CType)
		if locale {
			v = append(v, null(d.Locale))
		}
		if rules {
			v = append(v, null(d.ICURules))
		}
		v = append(v, null(d.Access))
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
	access := verbose && some(schemas, func(s dbmeta.Schema) bool { return s.Access.Valid })
	r := &rows{columns: []string{"Name", "Owner"}}
	if access {
		r.columns = append(r.columns, "Access privileges")
	}
	if verbose {
		r.columns = append(r.columns, "Description")
	}
	for _, s := range schemas {
		v := []any{s.Name, s.Owner}
		if access {
			v = append(v, null(s.Access))
		}
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
	owner := some(indexes, func(i dbmeta.Index) bool { return i.Owner.Valid })
	persistence := verbose && some(indexes, func(i dbmeta.Index) bool { return i.Persistence.Valid })
	sizes := verbose && some(indexes, func(i dbmeta.Index) bool { return i.Size.Valid })
	r := &rows{columns: []string{"Schema", "Name", "Type"}}
	if owner {
		r.columns = append(r.columns, "Owner")
	}
	r.columns = append(r.columns, "Table")
	if persistence {
		r.columns = append(r.columns, "Persistence")
	}
	if verbose {
		r.columns = append(r.columns, "Access method")
	}
	if sizes {
		r.columns = append(r.columns, "Size")
	}
	if verbose {
		r.columns = append(r.columns, "Description")
	}
	for _, i := range indexes {
		v := []any{i.Schema, i.Name, "index"}
		if owner {
			v = append(v, null(i.Owner))
		}
		v = append(v, i.Table)
		if persistence {
			v = append(v, null(i.Persistence))
		}
		if verbose {
			v = append(v, method(i))
		}
		if sizes {
			v = append(v, size(i.Size))
		}
		if verbose {
			v = append(v, null(i.Comment))
		}
		r.add(v...)
	}
	if len(r.values) == 0 {
		return f.notFound(pattern)
	}
	return encode(f.out, r, "List of indexes", nil, nil)
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
	columnAccess, err := f.columnPrivileges(ctx)
	if err != nil {
		return err
	}
	policies, err := f.policySummaries(ctx)
	if err != nil {
		return err
	}
	r := &rows{columns: []string{"Schema", "Name", "Type", "Access privileges", "Column privileges", "Policies"}}
	for _, p := range privs {
		key := null(p.Schema) + "\x00" + p.Name
		cols, pols := null(p.ColumnAccess), null(p.Policies)
		// the kinds give the text that psql prints, and the flat text of the
		// privilege is the answer of a database that has not got them
		if columnAccess != nil {
			cols = columnAccess[key]
		}
		if policies != nil {
			pols = policies[key]
		}
		r.add(null(p.Schema), p.Name, strings.ToLower(p.Type), null(p.Access), cols, pols)
	}
	return encode(f.out, r, text.AccessPrivileges, nil, nil)
}
