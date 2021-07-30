package metadata

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/xo/tblfmt"
	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

// DB is the common interface for database operations, compatible with
// database/sql.DB and database/sql.Tx.
type DB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	Prepare(string) (*sql.Stmt, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

// DefaultWriter using an existing db introspector
type DefaultWriter struct {
	r             Reader
	db            DB
	w             io.Writer
	tableTypes    map[rune][]string
	funcTypes     map[rune][]string
	systemSchemas map[string]struct{}

	// custom functions for easier overloading
	listAllDbs func(string, bool) error
}

var _ Writer = &DefaultWriter{}

func NewDefaultWriter(r Reader, opts ...WriterOption) func(db DB, w io.Writer) Writer {
	defaultWriter := &DefaultWriter{
		r: r,
		tableTypes: map[rune][]string{
			't': {"TABLE", "BASE TABLE", "SYSTEM TABLE", "SYNONYM", "LOCAL TEMPORARY", "GLOBAL TEMPORARY"},
			'v': {"VIEW", "SYSTEM VIEW"},
			'm': {"MATERIALIZED VIEW"},
			's': {"SEQUENCE"},
		},
		funcTypes: map[rune][]string{
			'a': {"AGGREGATE"},
			'n': {"FUNCTION"},
			'p': {"PROCEDURE", "PACKAGE"},
			't': {"TRIGGER"},
			'w': {"WINDOW"},
		},
		systemSchemas: map[string]struct{}{
			"information_schema": {},
		},
	}
	for _, o := range opts {
		o(defaultWriter)
	}
	return func(db DB, w io.Writer) Writer {
		defaultWriter.db = db
		defaultWriter.w = w
		return defaultWriter
	}
}

// WriterOption to configure the DefaultWriter
type WriterOption func(*DefaultWriter)

// WithSystemSchemas that are ignored unless showSystem is true
func WithSystemSchemas(schemas []string) WriterOption {
	return func(w *DefaultWriter) {
		w.systemSchemas = make(map[string]struct{}, len(schemas))
		for _, s := range schemas {
			w.systemSchemas[s] = struct{}{}
		}
	}
}

// WithListAllDbs that lists all catalogs
func WithListAllDbs(f func(string, bool) error) WriterOption {
	return func(w *DefaultWriter) {
		w.listAllDbs = f
	}
}

// DescribeFunctions matching pattern
func (w DefaultWriter) DescribeFunctions(funcTypes, pattern string, verbose, showSystem bool) error {
	r, ok := w.r.(FunctionReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\df`)
	}
	types := []string{}
	for k, v := range w.funcTypes {
		if strings.ContainsRune(funcTypes, k) {
			types = append(types, v...)
		}
	}
	sp, tp, err := parsePattern(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse search pattern: %w", err)
	}
	res, err := r.Functions(Filter{Schema: sp, Name: tp, Types: types, WithSystem: showSystem})
	if err != nil {
		return fmt.Errorf("failed to list functions: %w", err)
	}
	defer res.Close()

	if !showSystem {
		// in case the reader doesn't implement WithSystem
		res.SetFilter(func(r Result) bool {
			_, ok := w.systemSchemas[r.(*Function).Schema]
			return !ok
		})
	}

	if _, ok := w.r.(FunctionColumnReader); ok {
		for res.Next() {
			f := res.Get()
			f.ArgTypes, err = w.getFunctionColumns(f.Catalog, f.Schema, f.SpecificName)
			if err != nil {
				return fmt.Errorf("failed to get columns of function %s.%s: %w", f.Schema, f.SpecificName, err)
			}
		}
		res.Reset()
	}

	columns := []string{"Schema", "Name", "Result data type", "Argument data types", "Type"}
	if verbose {
		columns = append(columns, "Volatility", "Security", "Language", "Source code")
	}
	res.SetColumns(columns)
	res.SetScanValues(func(r Result) []interface{} {
		f := r.(*Function)
		v := []interface{}{f.Schema, f.Name, f.ResultType, f.ArgTypes, f.Type}
		if verbose {
			v = append(v, f.Volatility, f.Security, f.Language, f.Source)
		}
		return v
	})
	params := env.Pall()
	params["title"] = "List of functions"
	return tblfmt.EncodeAll(w.w, res, params)
}

func (w DefaultWriter) getFunctionColumns(c, s, f string) (string, error) {
	r := w.r.(FunctionColumnReader)
	cols, err := r.FunctionColumns(Filter{Catalog: c, Schema: s, Parent: f})
	if err != nil {
		return "", err
	}
	args := []string{}
	for cols.Next() {
		c := cols.Get()
		// skip result params
		if c.OrdinalPosition == 0 {
			continue
		}
		typ := ""
		if c.Type != "IN" && c.Type != "" {
			typ = c.Type + " "
		}
		name := c.Name
		if name != "" {
			name += " "
		}
		args = append(args, fmt.Sprintf("%s%s%s", typ, name, c.DataType))
	}
	return strings.Join(args, ", "), nil
}

// DescribeTableDetails matching pattern
func (w DefaultWriter) DescribeTableDetails(pattern string, verbose, showSystem bool) error {
	sp, tp, err := parsePattern(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse search pattern: %w", err)
	}

	found := 0

	tr, isTR := w.r.(TableReader)
	_, isCR := w.r.(ColumnReader)
	if isTR && isCR {
		res, err := tr.Tables(Filter{Schema: sp, Name: tp, WithSystem: showSystem})
		if err != nil {
			return fmt.Errorf("failed to list tables: %w", err)
		}
		defer res.Close()
		if !showSystem {
			// in case the reader doesn't implement WithSystem
			res.SetFilter(func(r Result) bool {
				_, ok := w.systemSchemas[r.(*Table).Schema]
				return !ok
			})
		}
		for res.Next() {
			t := res.Get()
			err = w.describeTableDetails(t.Type, t.Schema, t.Name, verbose, showSystem)
			if err != nil {
				return fmt.Errorf("failed to describe %s %s.%s: %w", t.Type, t.Schema, t.Name, err)
			}
			found++
		}
	}

	if _, ok := w.r.(SequenceReader); ok {
		foundSeq, err := w.describeSequences(sp, tp, verbose, showSystem)
		if err != nil {
			return fmt.Errorf("failed to describe sequences: %w", err)
		}
		found += foundSeq
	}

	ir, isIR := w.r.(IndexReader)
	_, isICR := w.r.(IndexColumnReader)
	if isIR && isICR {
		res, err := ir.Indexes(Filter{Schema: sp, Name: tp, WithSystem: showSystem})
		if err != nil && err != ErrNotSupported {
			return fmt.Errorf("failed to list indexes for table %s: %w", tp, err)
		}
		if res != nil {
			defer res.Close()
			if !showSystem {
				// in case the reader doesn't implement WithSystem
				res.SetFilter(func(r Result) bool {
					_, ok := w.systemSchemas[r.(*Index).Schema]
					return !ok
				})
			}
			for res.Next() {
				i := res.Get()
				err = w.describeIndex(i)
				if err != nil {
					return fmt.Errorf("failed to describe index %s from table %s.%s: %w", i.Name, i.Schema, i.Table, err)
				}
				found++
			}
		}
	}

	if found == 0 {
		fmt.Fprintf(w.w, text.RelationNotFound, pattern)
		fmt.Fprintln(w.w)
	}
	return nil
}

func (w DefaultWriter) describeTableDetails(typ, sp, tp string, verbose, showSystem bool) error {
	r := w.r.(ColumnReader)
	res, err := r.Columns(Filter{Schema: sp, Parent: tp, WithSystem: showSystem})
	if err != nil {
		return fmt.Errorf("failed to list columns for table %s: %w", tp, err)
	}
	defer res.Close()

	columns := []string{"Name", "Type", "Nullable", "Default"}
	if verbose {
		columns = append(columns, "Size", "Decimal Digits", "Radix", "Octet Length")
	}
	res.SetColumns(columns)
	res.SetScanValues(func(r Result) []interface{} {
		f := r.(*Column)
		v := []interface{}{f.Name, f.DataType, f.IsNullable, f.Default}
		if verbose {
			v = append(v, f.ColumnSize, f.DecimalDigits, f.NumPrecRadix, f.CharOctetLength)
		}
		return v
	})
	params := env.Pall()
	params["title"] = fmt.Sprintf("%s %s\n", typ, qualifiedIdentifier(sp, tp))
	return w.encodeWithSummary(res, params, w.tableDetailsSummary(sp, tp))
}

func (w DefaultWriter) encodeWithSummary(res tblfmt.ResultSet, params map[string]string, summary func(io.Writer, int) (int, error)) error {
	newEnc, opts := tblfmt.FromMap(params)
	opts = append(opts, tblfmt.WithSummary(
		map[int]func(io.Writer, int) (int, error){
			-1: summary,
		},
	))
	enc, err := newEnc(res, opts...)
	if err != nil {
		return err
	}
	return enc.EncodeAll(w.w)
}

func (w DefaultWriter) tableDetailsSummary(sp, tp string) func(io.Writer, int) (int, error) {
	return func(out io.Writer, _ int) (int, error) {
		err := w.describeTableIndexes(out, sp, tp)
		if err != nil {
			return 0, err
		}
		err = w.describeTableConstraints(
			out,
			Filter{Schema: sp, Parent: tp},
			func(r Result) bool {
				c := r.(*Constraint)
				return c.Type == "CHECK" && c.CheckClause != "" && !strings.HasSuffix(c.CheckClause, " IS NOT NULL")
			},
			"Check constraints:",
			func(out io.Writer, c *Constraint) error {
				_, err := fmt.Fprintf(out, "  \"%s\" %s (%s)\n", c.Name, c.Type, c.CheckClause)
				return err
			},
		)
		if err != nil {
			return 0, err
		}
		err = w.describeTableConstraints(
			out,
			Filter{Schema: sp, Parent: tp},
			func(r Result) bool { return r.(*Constraint).Type == "FOREIGN KEY" },
			"Foreign-key constraints:",
			func(out io.Writer, c *Constraint) error {
				columns, foreignColumns, err := w.getConstraintColumns(c.Catalog, c.Schema, c.Table, c.Name)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(out, "  \"%s\" %s (%s) REFERENCES %s(%s) ON UPDATE %s ON DELETE %s\n",
					c.Name,
					c.Type,
					columns,
					c.ForeignTable,
					foreignColumns,
					c.UpdateRule,
					c.DeleteRule)
				return err
			},
		)
		if err != nil {
			return 0, err
		}
		err = w.describeTableConstraints(
			out,
			Filter{Schema: sp, Reference: tp},
			func(r Result) bool { return r.(*Constraint).Type == "FOREIGN KEY" },
			"Referenced by:",
			func(out io.Writer, c *Constraint) error {
				columns, foreignColumns, err := w.getConstraintColumns(c.Catalog, c.Schema, c.Table, c.Name)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(out, "  TABLE \"%s\" CONSTRAINT \"%s\" %s (%s) REFERENCES %s(%s) ON UPDATE %s ON DELETE %s\n",
					c.Table,
					c.Name,
					c.Type,
					columns,
					c.ForeignTable,
					foreignColumns,
					c.UpdateRule,
					c.DeleteRule)
				return err
			},
		)
		err = w.describeTableTriggers(out, sp, tp)
		if err != nil {
			return 0, err
		}
		return 0, err
	}
}
func (w DefaultWriter) describeTableTriggers(out io.Writer, sp, tp string) error {
	r, ok := w.r.(TriggerReader)
	if !ok {
		return nil
	}
	res, err := r.Triggers(Filter{Schema: sp, Parent: tp})
	if err != nil && err != ErrNotSupported {
		return fmt.Errorf("failed to list triggers for table %s: %w", tp, err)
	}
	if res == nil {
		return nil
	}
	defer res.Close()

	if res.Len() == 0 {
		return nil
	}
	fmt.Fprintln(out, "Triggers:")
	for res.Next() {
		t := res.Get()
		fmt.Fprintf(out, "  \"%s\" %s\n", t.Name, t.Definition)
	}
	return nil
}

func (w DefaultWriter) describeTableIndexes(out io.Writer, sp, tp string) error {
	r, ok := w.r.(IndexReader)
	if !ok {
		return nil
	}
	res, err := r.Indexes(Filter{Schema: sp, Parent: tp})
	if err != nil && err != ErrNotSupported {
		return fmt.Errorf("failed to list indexes for table %s: %w", tp, err)
	}
	if res == nil {
		return nil
	}
	defer res.Close()

	if res.Len() == 0 {
		return nil
	}
	fmt.Fprintln(out, "Indexes:")
	for res.Next() {
		i := res.Get()
		primary := ""
		unique := ""
		if i.IsPrimary == YES {
			primary = "PRIMARY_KEY, "
		}
		if i.IsUnique == YES {
			unique = "UNIQUE, "
		}
		i.Columns, err = w.getIndexColumns(i.Catalog, i.Schema, i.Table, i.Name)
		if err != nil {
			return fmt.Errorf("failed to get columns of index %s: %w", i.Name, err)
		}
		fmt.Fprintf(out, "  \"%s\" %s%s%s (%s)\n", i.Name, primary, unique, i.Type, i.Columns)
	}
	return nil
}

func (w DefaultWriter) getIndexColumns(c, s, t, i string) (string, error) {
	r := w.r.(IndexColumnReader)
	cols, err := r.IndexColumns(Filter{Catalog: c, Schema: s, Parent: t, Name: i})
	if err != nil {
		return "", err
	}
	result := []string{}
	for cols.Next() {
		result = append(result, cols.Get().Name)
	}
	return strings.Join(result, ", "), nil
}

func (w DefaultWriter) describeTableConstraints(out io.Writer, filter Filter, postFilter func(r Result) bool, label string, printer func(io.Writer, *Constraint) error) error {
	r, ok := w.r.(ConstraintReader)
	if !ok {
		return nil
	}
	res, err := r.Constraints(filter)
	if err != nil && err != ErrNotSupported {
		return fmt.Errorf("failed to list constraints: %w", err)
	}
	if res == nil {
		return nil
	}
	defer res.Close()

	res.SetFilter(postFilter)
	if res.Len() == 0 {
		return nil
	}
	fmt.Fprintln(out, label)
	for res.Next() {
		c := res.Get()
		err := printer(out, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w DefaultWriter) getConstraintColumns(c, s, t, n string) (string, string, error) {
	r := w.r.(ConstraintColumnReader)
	cols, err := r.ConstraintColumns(Filter{Catalog: c, Schema: s, Parent: t, Name: n})
	if err != nil {
		return "", "", err
	}
	columns := []string{}
	foreignColumns := []string{}
	for cols.Next() {
		columns = append(columns, cols.Get().Name)
		foreignColumns = append(foreignColumns, cols.Get().ForeignName)
	}
	return strings.Join(columns, ", "), strings.Join(foreignColumns, ", "), nil
}

func (w DefaultWriter) describeSequences(sp, tp string, verbose, showSystem bool) (int, error) {
	r := w.r.(SequenceReader)
	res, err := r.Sequences(Filter{Schema: sp, Name: tp, WithSystem: showSystem})
	if err != nil && err != ErrNotSupported {
		return 0, err
	}
	if res == nil {
		return 0, nil
	}
	defer res.Close()

	found := 0
	for res.Next() {
		s := res.Get()
		// wrap current record into a separate recordSet
		rows := NewSequenceSet([]Sequence{*s})
		params := env.Pall()
		params["footer"] = "off"
		params["title"] = fmt.Sprintf("Sequence \"%s.%s\"\n", s.Schema, s.Name)
		err = tblfmt.EncodeAll(w.w, rows, params)
		if err != nil {
			return 0, err
		}
		// TODO footer should say which table this sequence belongs to
		found++
	}

	return found, nil
}

func (w DefaultWriter) describeIndex(i *Index) error {
	r := w.r.(IndexColumnReader)
	res, err := r.IndexColumns(Filter{Schema: i.Schema, Parent: i.Table, Name: i.Name})
	if err != nil {
		return fmt.Errorf("failed to get index columns: %w", err)
	}
	defer res.Close()
	if res.Len() == 0 {
		return nil
	}

	res.SetColumns([]string{"Name", "Type"})
	res.SetScanValues(func(r Result) []interface{} {
		f := r.(*IndexColumn)
		return []interface{}{f.Name, f.DataType}
	})

	params := env.Pall()
	params["title"] = fmt.Sprintf("Index %s\n", qualifiedIdentifier(i.Schema, i.Name))
	return w.encodeWithSummary(res, params, func(out io.Writer, _ int) (int, error) {
		primary := ""
		if i.IsPrimary == YES {
			primary = "primary key, "
		}
		_, err := fmt.Fprintf(out, "%s%s, for table %s", primary, i.Type, i.Table)
		return 0, err
	})
}

// ListAllDbs matching pattern
func (w DefaultWriter) ListAllDbs(pattern string, verbose bool) error {
	if w.listAllDbs != nil {
		return w.listAllDbs(pattern, verbose)
	}
	r, ok := w.r.(CatalogReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\l`)
	}
	res, err := r.Catalogs(Filter{Name: pattern})
	if err != nil {
		return fmt.Errorf("failed to list catalogs: %w", err)
	}
	defer res.Close()

	params := env.Pall()
	params["title"] = "List of databases"
	return tblfmt.EncodeAll(w.w, res, params)
}

// ListTables matching pattern
func (w DefaultWriter) ListTables(tableTypes, pattern string, verbose, showSystem bool) error {
	r, ok := w.r.(TableReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\dt`)
	}
	types := []string{}
	for k, v := range w.tableTypes {
		if strings.ContainsRune(tableTypes, k) {
			types = append(types, v...)
		}
	}
	sp, tp, err := parsePattern(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse search pattern: %w", err)
	}
	res, err := r.Tables(Filter{Schema: sp, Name: tp, Types: types, WithSystem: showSystem})
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}
	defer res.Close()
	if !showSystem {
		// in case the reader doesn't implement WithSystem
		res.SetFilter(func(r Result) bool {
			_, ok := w.systemSchemas[r.(*Table).Schema]
			return !ok
		})
	}
	if res.Len() == 0 {
		fmt.Fprintf(w.w, text.RelationNotFound, pattern)
		fmt.Fprintln(w.w)
		return nil
	}
	columns := []string{"Schema", "Name", "Type"}
	if verbose {
		columns = append(columns, "Size", "Comment")
	}
	res.SetColumns(columns)
	res.SetScanValues(func(r Result) []interface{} {
		f := r.(*Table)
		v := []interface{}{f.Schema, f.Name, f.Type}
		if verbose {
			v = append(v, f.Size, f.Comment)
		}
		return v
	})

	params := env.Pall()
	params["title"] = "List of relations"
	return tblfmt.EncodeAll(w.w, res, params)
}

// ListSchemas matching pattern
func (w DefaultWriter) ListSchemas(pattern string, verbose, showSystem bool) error {
	r, ok := w.r.(SchemaReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\d`)
	}
	res, err := r.Schemas(Filter{Name: pattern, WithSystem: showSystem})
	if err != nil {
		return fmt.Errorf("failed to list schemas: %w", err)
	}
	defer res.Close()

	if !showSystem {
		// in case the reader doesn't implement WithSystem
		res.SetFilter(func(r Result) bool {
			_, ok := w.systemSchemas[r.(*Schema).Schema]
			return !ok
		})
	}
	params := env.Pall()
	params["title"] = "List of schemas"
	return tblfmt.EncodeAll(w.w, res, params)
}

// ListIndexes matching pattern
func (w DefaultWriter) ListIndexes(pattern string, verbose, showSystem bool) error {
	r, ok := w.r.(IndexReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\di`)
	}
	sp, tp, err := parsePattern(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse search pattern: %w", err)
	}
	res, err := r.Indexes(Filter{Schema: sp, Name: tp, WithSystem: showSystem})
	if err != nil {
		return fmt.Errorf("failed to list indexes: %w", err)
	}
	defer res.Close()

	if !showSystem {
		// in case the reader doesn't implement WithSystem
		res.SetFilter(func(r Result) bool {
			_, ok := w.systemSchemas[r.(*Index).Schema]
			return !ok
		})
	}
	if res.Len() == 0 {
		fmt.Fprintf(w.w, text.RelationNotFound, pattern)
		fmt.Fprintln(w.w)
		return nil
	}

	columns := []string{"Schema", "Name", "Type", "Table"}
	if verbose {
		columns = append(columns, "Primary?", "Unique?")
	}
	res.SetColumns(columns)
	res.SetScanValues(func(r Result) []interface{} {
		f := r.(*Index)
		v := []interface{}{f.Schema, f.Name, f.Type, f.Table}
		if verbose {
			v = append(v, f.IsPrimary, f.IsUnique)
		}
		return v
	})

	params := env.Pall()
	params["title"] = "List of indexes"
	return tblfmt.EncodeAll(w.w, res, params)
}

func parsePattern(pattern string) (string, string, error) {
	// TODO do proper escaping, quoting etc
	if strings.ContainsRune(pattern, '.') {
		parts := strings.SplitN(pattern, ".", 2)
		return strings.ReplaceAll(parts[0], "*", "%"), strings.ReplaceAll(parts[1], "*", "%"), nil
	}
	return "", strings.ReplaceAll(pattern, "*", "%"), nil
}

func qualifiedIdentifier(schema, name string) string {
	if schema == "" {
		return fmt.Sprintf("\"%s\"", name)
	}
	return fmt.Sprintf("\"%s.%s\"", schema, name)
}
