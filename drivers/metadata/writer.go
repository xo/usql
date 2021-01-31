package metadata

import (
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
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	Prepare(string) (*sql.Stmt, error)
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

func NewDefaultWriter(r Reader, opts ...Option) func(db DB, w io.Writer) Writer {
	defaultWriter := &DefaultWriter{
		r: r,
		tableTypes: map[rune][]string{
			't': {"TABLE", "BASE TABLE"},
			'v': {"VIEW"},
			'm': {"MATERIALIZED VIEW"},
			's': {"SEQUENCE"},
		},
		funcTypes: map[rune][]string{
			'a': {"AGGREGATE"},
			'f': {"FUNCTION"},
			'n': {"FUNCTION"},
			'p': {"PROCEDURE"},
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

// Option to configure the DefaultWriter
type Option func(*DefaultWriter)

// WithSystemSchemas that are ignored unless showSystem is true
func WithSystemSchemas(schemas []string) Option {
	return func(w *DefaultWriter) {
		w.systemSchemas = make(map[string]struct{}, len(schemas))
		for _, s := range schemas {
			w.systemSchemas[s] = struct{}{}
		}
	}
}

// WithListAllDbs that lists all catalogs
func WithListAllDbs(f func(string, bool) error) Option {
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
		return err
	}
	res, err := r.Functions("", sp, tp, types)
	if err != nil {
		return err
	}
	defer res.Close()

	if !showSystem {
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
				return err
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
	fmt.Fprintln(w.w, "List of functions")
	return tblfmt.EncodeAll(w.w, res, env.Pall())
}

func (w DefaultWriter) getFunctionColumns(c, s, f string) (string, error) {
	r := w.r.(FunctionColumnReader)
	cols, err := r.FunctionColumns(c, s, f)
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
		return err
	}

	found := 0

	tr, isTR := w.r.(TableReader)
	_, isCR := w.r.(ColumnReader)
	if isTR && isCR {
		res, err := tr.Tables("", sp, tp, []string{})
		if err != nil {
			return err
		}
		defer res.Close()
		if !showSystem {
			res.SetFilter(func(r Result) bool {
				_, ok := w.systemSchemas[r.(*Table).Schema]
				return !ok
			})
		}
		for res.Next() {
			t := res.Get()
			err = w.describeTableDetails(t.Type, t.Schema, t.Name, verbose, showSystem)
			if err != nil {
				return err
			}
			found++
		}
	}

	if _, ok := w.r.(SequenceReader); ok {
		foundSeq, err := w.describeSequences(sp, tp, verbose, showSystem)
		if err != nil {
			return err
		}
		found += foundSeq
	}

	ir, isIR := w.r.(IndexReader)
	_, isICR := w.r.(IndexColumnReader)
	if isIR && isICR {
		res, err := ir.Indexes("", sp, "", tp)
		if err != nil && err != ErrNotSupported {
			return err
		}
		if res != nil {
			defer res.Close()
			if !showSystem {
				res.SetFilter(func(r Result) bool {
					_, ok := w.systemSchemas[r.(*Index).Schema]
					return !ok
				})
			}
			for res.Next() {
				i := res.Get()
				err = w.describeIndexes(i.Schema, i.Table, i.Name)
				if err != nil {
					return err
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
	res, err := r.Columns("", sp, tp)
	if err != nil {
		return err
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
	fmt.Fprintf(w.w, "%s \"%s.%s\"\n", typ, sp, tp)
	err = tblfmt.EncodeAll(w.w, res, env.Pall())
	if err != nil {
		return err
	}

	err = w.describeTableIndexes(sp, tp)
	if err != nil {
		return err
	}

	// TODO also describe: FKs, references, triggers - using template encoder?
	return nil
}

func (w DefaultWriter) describeTableIndexes(sp, tp string) error {
	r, ok := w.r.(IndexReader)
	if !ok {
		return nil
	}
	res, err := r.Indexes("", sp, tp, "")
	if err != nil && err != ErrNotSupported {
		return err
	}
	if res == nil {
		return nil
	}
	defer res.Close()

	if res.Len() == 0 {
		return nil
	}
	fmt.Fprintln(w.w, "Indexes:")
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
			return err
		}
		fmt.Fprintf(w.w, "  \"%s\" %s%s%s (%s)\n", i.Name, primary, unique, i.Type, i.Columns)
	}
	fmt.Fprintln(w.w)
	return nil
}

func (w DefaultWriter) getIndexColumns(c, s, t, i string) (string, error) {
	r := w.r.(IndexColumnReader)
	cols, err := r.IndexColumns(c, s, t, i)
	if err != nil {
		return "", err
	}
	result := []string{}
	for cols.Next() {
		result = append(result, cols.Get().Name)
	}
	return strings.Join(result, ", "), nil
}

func (w DefaultWriter) describeSequences(sp, tp string, verbose, showSystem bool) (int, error) {
	r := w.r.(SequenceReader)
	res, err := r.Sequences("", sp, tp)
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
		fmt.Fprintf(w.w, "Sequence \"%s.%s\"\n", s.Schema, s.Name)
		err = tblfmt.EncodeAll(w.w, rows, params)
		if err != nil {
			return 0, err
		}
		// TODO footer should say which table this sequence belongs to
		found++
	}

	return found, nil
}

func (w DefaultWriter) describeIndexes(sp, tp, ip string) error {
	r := w.r.(IndexColumnReader)
	res, err := r.IndexColumns("", sp, tp, ip)
	if err != nil {
		return err
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

	// TODO footer should say if it's primary, index type and which table this index belongs to
	fmt.Fprintf(w.w, "Index \"%s.%s\"\n", sp, ip)
	err = tblfmt.EncodeAll(w.w, res, env.Pall())
	if err != nil {
		return err
	}

	return nil
}

// ListAllDbs matching pattern
func (w DefaultWriter) ListAllDbs(pattern string, verbose bool) error {
	if w.listAllDbs == nil {
		return fmt.Errorf(text.NotSupportedByDriver, `\l`)
	}
	return w.listAllDbs(pattern, verbose)
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
		return err
	}
	res, err := r.Tables("", sp, tp, types)
	if err != nil {
		return err
	}
	defer res.Close()
	if !showSystem {
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

	fmt.Fprintln(w.w, "List of relations")
	return tblfmt.EncodeAll(w.w, res, env.Pall())
}

// ListSchemas matching pattern
func (w DefaultWriter) ListSchemas(pattern string, verbose, showSystem bool) error {
	r, ok := w.r.(SchemaReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\d`)
	}
	res, err := r.Schemas("", pattern)
	if err != nil {
		return err
	}
	defer res.Close()

	if !showSystem {
		res.SetFilter(func(r Result) bool {
			_, ok := w.systemSchemas[r.(*Schema).Schema]
			return !ok
		})
	}
	fmt.Fprintln(w.w, "List of schemas")
	return tblfmt.EncodeAll(w.w, res, env.Pall())
}

// ListIndexes matching pattern
func (w DefaultWriter) ListIndexes(pattern string, verbose, showSystem bool) error {
	r, ok := w.r.(IndexReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\di`)
	}
	sp, tp, err := parsePattern(pattern)
	if err != nil {
		return err
	}
	res, err := r.Indexes("", sp, "", tp)
	if err != nil {
		return err
	}
	defer res.Close()

	if !showSystem {
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

	fmt.Fprintln(w.w, "List of indexes")
	return tblfmt.EncodeAll(w.w, res, env.Pall())
}

func parsePattern(pattern string) (string, string, error) {
	// TODO do proper escaping, quoting etc
	if strings.ContainsRune(pattern, '.') {
		parts := strings.SplitN(pattern, ".", 2)
		return strings.ReplaceAll(parts[0], "*", "%"), strings.ReplaceAll(parts[1], "*", "%"), nil
	}
	return "", strings.ReplaceAll(pattern, "*", "%"), nil
}
