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
	r          Reader
	db         DB
	w          io.Writer
	tableTypes map[rune][]string
	funcTypes  map[rune][]string
}

var _ Writer = &DefaultWriter{}

func NewDefaultWriter(r Reader) func(db DB, w io.Writer) Writer {
	return func(db DB, w io.Writer) Writer {
		return &DefaultWriter{
			r:  r,
			db: db,
			w:  w,
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
		}
	}
}

// DescribeAggregates matching pattern
func (w DefaultWriter) DescribeAggregates(pattern string, verbose, showSystem bool) error {
	r, ok := w.r.(FunctionReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\da`)
	}
	sp, tp, err := parsePattern(pattern)
	if err != nil {
		return err
	}
	res, err := r.Functions("", sp, tp, []string{"AGGREGATE"})
	if err != nil {
		return err
	}
	defer res.Close()

	res.SetVerbose(verbose)
	params := env.All()
	params["format"] = "aligned"
	return tblfmt.EncodeAll(w.w, res, params)
}

// DescribeFunctions matching pattern
// TODO does it make sense to implement showSystem for DefaultWriter? it would require
// pushing it down to readers, which definitely would require configuration
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

	// this is inefficient but multiple databases supporting information_schema
	// aggregate strings in different ways (GROUP_CONCAT() vs string_agg/array_to_string(array_agg))
	// TODO work around by making such expression generator an option, same as placeholder
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

	res.SetVerbose(verbose)
	params := env.All()
	params["format"] = "aligned"
	return tblfmt.EncodeAll(w.w, res, params)
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
		if c.Type != "IN" {
			typ = c.Type + " "
		}
		args = append(args, fmt.Sprintf("%s%s %s", typ, c.Name, c.DataType))
	}
	return strings.Join(args, ", "), nil
}

// DescribeTableDetails matching pattern
func (w DefaultWriter) DescribeTableDetails(pattern string, verbose, showSystem bool) error {
	sp, tp, err := parsePattern(pattern)
	if err != nil {
		return err
	}

	if _, ok := w.r.(ColumnReader); ok {
		err = w.describeTableDetails(sp, tp, verbose, showSystem)
		if err != nil {
			return err
		}
	}

	if _, ok := w.r.(SequenceReader); ok {
		err = w.describeSequences(sp, tp, verbose, showSystem)
		if err != nil {
			return err
		}
	}

	if _, ok := w.r.(IndexColumnReader); ok {
		err = w.describeIndexes(sp, tp, verbose, showSystem)
		if err != nil {
			return err
		}
	}
	// TODO if no table, seq or index were found, should return this:
	// fmt.Fprintf(w.w, text.RelationNotFound, pattern)
	return nil
}

func (w DefaultWriter) describeTableDetails(sp, tp string, verbose, showSystem bool) error {
	// TODO first use TableReader to match tables, then describe each one separately
	r := w.r.(ColumnReader)
	res, err := r.Columns("", sp, tp)
	if err != nil {
		return err
	}
	defer res.Close()

	res.SetVerbose(verbose)
	params := env.All()
	params["format"] = "aligned"
	err = tblfmt.EncodeAll(w.w, res, params)
	if err != nil {
		return err
	}

	if r, ok := w.r.(IndexReader); ok {
		res, err := r.Indexes("", sp, tp)
		if err != nil {
			return err
		}
		defer res.Close()
		indexes := []*Index{}
		for res.Next() {
			indexes = append(indexes, res.Get())
		}
		if len(indexes) != 0 {
			fmt.Fprintln(w.w, "Indexes:")
			for _, i := range indexes {
				primary := ""
				unique := ""
				if i.IsPrimary == YES {
					primary = "PRIMARY_KEY, "
				}
				if i.IsUnique == YES {
					unique = "UNIQUE, "
				}
				i.Columns, err = w.getIndexColumns(i.Catalog, i.Schema, i.Name)
				if err != nil {
					return err
				}
				fmt.Fprintf(w.w, "\"%s\" %s%s%s (%s)\n", i.Name, primary, unique, i.Type, i.Columns)
			}
		}
	}
	// TODO also describe: FKs, references, triggers - using template encoder?
	return nil
}

func (w DefaultWriter) getIndexColumns(c, s, i string) (string, error) {
	r := w.r.(IndexColumnReader)
	cols, err := r.IndexColumns(c, s, i)
	if err != nil {
		return "", err
	}
	result := []string{}
	for cols.Next() {
		result = append(result, cols.Get().Name)
	}
	return strings.Join(result, ", "), nil
}

func (w DefaultWriter) describeSequences(sp, tp string, verbose, showSystem bool) error {
	r := w.r.(SequenceReader)
	res, err := r.Sequences("", sp, tp)
	if err != nil {
		return err
	}
	defer res.Close()
	if res.Len() == 0 {
		return nil
	}

	params := env.All()
	params["format"] = "aligned"
	// TODO footer should say which table this sequence belongs to
	err = tblfmt.EncodeAll(w.w, res, params)
	if err != nil {
		return err
	}

	return nil
}

func (w DefaultWriter) describeIndexes(sp, tp string, verbose, showSystem bool) error {
	// TODO first use IndexReader to match indexes, then describe each one separately
	r := w.r.(IndexColumnReader)
	res, err := r.IndexColumns("", sp, tp)
	if err != nil {
		return err
	}
	defer res.Close()
	if res.Len() == 0 {
		return nil
	}

	params := env.All()
	params["format"] = "aligned"
	// TODO footer should say if it's primary, index type and which table this index belongs to
	err = tblfmt.EncodeAll(w.w, res, params)
	if err != nil {
		return err
	}

	return nil
}

// ListAllDbs matching pattern
func (w DefaultWriter) ListAllDbs(pattern string, verbose bool) error {
	return fmt.Errorf(text.NotSupportedByDriver, `\l`)
}

// ListTables matching pattern
func (w DefaultWriter) ListTables(tableTypes, pattern string, verbose, showSystem bool) error {
	r, ok := w.r.(TableReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\d`)
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
	if res.Len() == 0 {
		fmt.Fprintf(w.w, text.RelationNotFound, pattern)
		return nil
	}

	res.SetVerbose(verbose)
	params := env.All()
	params["format"] = "aligned"
	return tblfmt.EncodeAll(w.w, res, params)
}

// ListSchemas matching pattern
func (w DefaultWriter) ListSchemas(pattern string, verbose, showSystem bool) error {
	r, ok := w.r.(SchemaReader)
	if !ok {
		return fmt.Errorf(text.NotSupportedByDriver, `\d`)
	}
	res, err := r.Schemas()
	if err != nil {
		return err
	}
	// TODO do pattern matching locally or add paterns to Schemas()
	defer res.Close()

	res.SetVerbose(verbose)
	params := env.All()
	params["format"] = "aligned"
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
		return err
	}
	res, err := r.Indexes("", sp, tp)
	if err != nil {
		return err
	}
	defer res.Close()
	if res.Len() == 0 {
		fmt.Fprintf(w.w, text.RelationNotFound, pattern)
		return nil
	}

	res.SetVerbose(verbose)
	params := env.All()
	params["format"] = "aligned"
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
