// Package informationschema provides metadata readers that query tables from
// the information_schema schema. It tries to be database agnostic,
// but there is a set of options to configure what tables and columns to expect.
package informationschema

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

// InformationSchema metadata reader
type InformationSchema struct {
	metadata.LoggingReader
	pf            func(int) string
	hasFunctions  bool
	hasSequences  bool
	hasIndexes    bool
	colExpr       map[ColumnName]string
	limit         int
	systemSchemas []string
	currentSchema string
}

var _ metadata.BasicReader = &InformationSchema{}

type Logger interface {
	Println(...interface{})
}

type ColumnName string

const (
	ColumnsColumnSize       = ColumnName("columns.column_size")
	ColumnsNumericScale     = ColumnName("columns.numeric_scale")
	ColumnsNumericPrecRadix = ColumnName("columns.numeric_precision_radix")
	ColumnsCharOctetLength  = ColumnName("columns.character_octet_length")

	FunctionColumnsColumnSize       = ColumnName("function_columns.column_size")
	FunctionColumnsNumericScale     = ColumnName("function_columns.numeric_scale")
	FunctionColumnsNumericPrecRadix = ColumnName("function_columns.numeric_precision_radix")
	FunctionColumnsCharOctetLength  = ColumnName("function_columns.character_octet_length")

	FunctionsSecurityType = ColumnName("functions.security_type")

	SequenceColumnsIncrement = ColumnName("sequence_columns.increment")
)

// New InformationSchema reader
func New(opts ...metadata.ReaderOption) func(drivers.DB, ...metadata.ReaderOption) metadata.Reader {
	s := &InformationSchema{
		pf:           func(n int) string { return fmt.Sprintf("$%d", n) },
		hasFunctions: true,
		hasSequences: true,
		hasIndexes:   true,
		colExpr: map[ColumnName]string{
			ColumnsColumnSize:               "COALESCE(character_maximum_length, numeric_precision, datetime_precision, 0)",
			ColumnsNumericScale:             "COALESCE(numeric_scale, 0)",
			ColumnsNumericPrecRadix:         "COALESCE(numeric_precision_radix, 10)",
			ColumnsCharOctetLength:          "COALESCE(character_octet_length, 0)",
			FunctionColumnsColumnSize:       "COALESCE(character_maximum_length, numeric_precision, datetime_precision, 0)",
			FunctionColumnsNumericScale:     "COALESCE(numeric_scale, 0)",
			FunctionColumnsNumericPrecRadix: "COALESCE(numeric_precision_radix, 10)",
			FunctionColumnsCharOctetLength:  "COALESCE(character_octet_length, 0)",
			FunctionsSecurityType:           "security_type",
			SequenceColumnsIncrement:        "increment",
		},
		systemSchemas: []string{"information_schema"},
	}
	// aply InformationSchema specific options
	for _, o := range opts {
		o(s)
	}

	return func(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
		s.LoggingReader = metadata.NewLoggingReader(db, opts...)
		return s
	}
}

// WithPlaceholder generator function, that usually returns either `?` or `$n`,
// where `n` is the argument.
func WithPlaceholder(pf func(int) string) metadata.ReaderOption {
	return func(r metadata.Reader) {
		r.(*InformationSchema).pf = pf
	}
}

// WithCustomColumns to use different expressions for some columns
func WithCustomColumns(cols map[ColumnName]string) metadata.ReaderOption {
	return func(r metadata.Reader) {
		for k, v := range cols {
			r.(*InformationSchema).colExpr[k] = v
		}
	}
}

// WithFunctions when the `routines` and `parameters` tables exists
func WithFunctions(fun bool) metadata.ReaderOption {
	return func(r metadata.Reader) {
		r.(*InformationSchema).hasFunctions = fun
	}
}

// WithIndexes when the `statistics` table exists
func WithIndexes(ind bool) metadata.ReaderOption {
	return func(r metadata.Reader) {
		r.(*InformationSchema).hasIndexes = ind
	}
}

// WithSequences when the `sequences` table exists
func WithSequences(seq bool) metadata.ReaderOption {
	return func(r metadata.Reader) {
		r.(*InformationSchema).hasSequences = seq
	}
}

// WithSystemSchemas that are ignored unless WithSystem filter is true
func WithSystemSchemas(schemas []string) metadata.ReaderOption {
	return func(r metadata.Reader) {
		r.(*InformationSchema).systemSchemas = schemas
	}
}

// WithCurrentSchema expression to filter by when OnlyVisible filter is true
func WithCurrentSchema(expr string) metadata.ReaderOption {
	return func(r metadata.Reader) {
		r.(*InformationSchema).currentSchema = expr
	}
}

func (s *InformationSchema) SetLimit(l int) {
	s.limit = l
}

// Columns from selected catalog (or all, if empty), matching schemas and tables
func (s InformationSchema) Columns(f metadata.Filter) (*metadata.ColumnSet, error) {
	columns := []string{
		"table_catalog",
		"table_schema",
		"table_name",
		"column_name",
		"ordinal_position",
		"data_type",
		"COALESCE(column_default, '')",
		"COALESCE(is_nullable, '') AS is_nullable",
		s.colExpr[ColumnsColumnSize],
		s.colExpr[ColumnsNumericScale],
		s.colExpr[ColumnsNumericPrecRadix],
		s.colExpr[ColumnsCharOctetLength],
	}

	qstr := "SELECT\n  " + strings.Join(columns, ",\n  ") + " FROM information_schema.columns\n"
	conds, vals := s.conditions(1, f, formats{
		catalog:    "table_catalog LIKE %s",
		schema:     "table_schema LIKE %s",
		notSchemas: "table_schema NOT IN (%s)",
		parent:     "table_name LIKE %s",
	})
	rows, closeRows, err := s.query(qstr, conds, "table_catalog, table_schema, table_name, ordinal_position", vals...)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewColumnSet([]metadata.Column{}), nil
		}
		return nil, err
	}
	defer closeRows()

	results := []metadata.Column{}
	for rows.Next() {
		rec := metadata.Column{}
		err = rows.Scan(
			&rec.Catalog,
			&rec.Schema,
			&rec.Table,
			&rec.Name,
			&rec.OrdinalPosition,
			&rec.DataType,
			&rec.Default,
			&rec.IsNullable,
			&rec.ColumnSize,
			&rec.DecimalDigits,
			&rec.NumPrecRadix,
			&rec.CharOctetLength,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewColumnSet(results), nil
}

// Tables from selected catalog (or all, if empty), matching schemas, names and types
func (s InformationSchema) Tables(f metadata.Filter) (*metadata.TableSet, error) {
	qstr := `SELECT
  table_catalog,
  table_schema,
  table_name,
  table_type
FROM information_schema.tables
`
	conds, vals := s.conditions(1, f, formats{
		catalog:    "table_catalog LIKE %s",
		schema:     "table_schema LIKE %s",
		notSchemas: "table_schema NOT IN (%s)",
		name:       "table_name LIKE %s",
		types:      "table_type IN (%s)",
	})
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	addSequences := false
	for _, t := range f.Types {
		if t == "SEQUENCE" && s.hasSequences {
			addSequences = true
		}
	}
	if addSequences {
		qstr += `
UNION ALL
SELECT
  sequence_catalog AS table_catalog,
  sequence_schema AS table_schema,
  sequence_name AS table_name,
  'SEQUENCE' AS table_type
FROM information_schema.sequences
`
		conds, seqVals := s.conditions(len(vals)+1, f, formats{
			catalog:    "sequence_catalog LIKE %s",
			schema:     "sequence_schema LIKE %s",
			notSchemas: "sequence_schema NOT IN (%s)",
			name:       "sequence_name LIKE %s",
		})
		vals = append(vals, seqVals...)
		if len(conds) != 0 {
			qstr += " WHERE " + strings.Join(conds, " AND ")
		}
	}
	rows, closeRows, err := s.query(qstr, []string{}, "table_catalog, table_schema, table_type, table_name", vals...)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewTableSet([]metadata.Table{}), nil
		}
		return nil, err
	}
	defer closeRows()

	results := []metadata.Table{}
	for rows.Next() {
		rec := metadata.Table{}
		err = rows.Scan(&rec.Catalog, &rec.Schema, &rec.Name, &rec.Type)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewTableSet(results), nil
}

// Schemas from selected catalog (or all, if empty), matching schemas and tables
func (s InformationSchema) Schemas(f metadata.Filter) (*metadata.SchemaSet, error) {
	qstr := `SELECT
  schema_name,
  catalog_name
FROM information_schema.schemata
`
	conds, vals := s.conditions(1, f, formats{
		catalog:    "catalog_name LIKE %s",
		name:       "schema_name LIKE %s",
		notSchemas: "schema_name NOT IN (%s)",
	})
	rows, closeRows, err := s.query(qstr, conds, "catalog_name, schema_name", vals...)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewSchemaSet([]metadata.Schema{}), nil
		}
		return nil, err
	}
	defer closeRows()

	results := []metadata.Schema{}
	for rows.Next() {
		rec := metadata.Schema{}
		err = rows.Scan(&rec.Schema, &rec.Catalog)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewSchemaSet(results), nil
}

// Functions from selected catalog (or all, if empty), matching schemas, names and types
func (s InformationSchema) Functions(f metadata.Filter) (*metadata.FunctionSet, error) {
	if !s.hasFunctions {
		return nil, metadata.ErrNotSupported
	}

	columns := []string{
		"specific_name",
		"routine_catalog",
		"routine_schema",
		"routine_name",
		"COALESCE(routine_type, '')",
		"data_type",
		"routine_definition",
		"COALESCE(external_language, routine_body) AS language",
		"is_deterministic",
		s.colExpr[FunctionsSecurityType],
	}

	qstr := "SELECT\n  " + strings.Join(columns, ",\n  ") + " FROM information_schema.routines\n"
	conds, vals := s.conditions(1, f, formats{
		catalog:    "routine_catalog LIKE %s",
		schema:     "routine_schema LIKE %s",
		notSchemas: "routine_schema NOT IN (%s)",
		name:       "routine_name LIKE %s",
		types:      "routine_type IN (%s)",
	})
	rows, closeRows, err := s.query(qstr, conds, "routine_catalog, routine_schema, routine_name, COALESCE(routine_type, '')", vals...)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewFunctionSet([]metadata.Function{}), nil
		}
		return nil, err
	}
	defer closeRows()

	results := []metadata.Function{}
	for rows.Next() {
		rec := metadata.Function{}
		err = rows.Scan(
			&rec.SpecificName,
			&rec.Catalog,
			&rec.Schema,
			&rec.Name,
			&rec.Type,
			&rec.ResultType,
			&rec.Source,
			&rec.Language,
			&rec.Volatility,
			&rec.Security,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewFunctionSet(results), nil
}

// FunctionColumns (arguments) from selected catalog (or all, if empty), matching schemas and functions
func (s InformationSchema) FunctionColumns(f metadata.Filter) (*metadata.FunctionColumnSet, error) {
	if !s.hasFunctions {
		return nil, metadata.ErrNotSupported
	}

	columns := []string{
		"specific_catalog",
		"specific_schema",
		"specific_name",
		"COALESCE(parameter_name, '')",
		"ordinal_position",
		"COALESCE(parameter_mode, '')",
		"data_type",
		s.colExpr[FunctionColumnsColumnSize],
		s.colExpr[FunctionColumnsNumericScale],
		s.colExpr[FunctionColumnsNumericPrecRadix],
		s.colExpr[FunctionColumnsCharOctetLength],
	}

	qstr := "SELECT\n  " + strings.Join(columns, ",\n  ") + " FROM information_schema.parameters\n"

	conds, vals := s.conditions(1, f, formats{
		catalog:    "specific_catalog LIKE %s",
		schema:     "specific_schema LIKE %s",
		notSchemas: "specific_schema NOT IN (%s)",
		parent:     "specific_name LIKE %s",
	})
	rows, closeRows, err := s.query(qstr, conds, "specific_catalog, specific_schema, specific_name, ordinal_position, COALESCE(parameter_name, '')", vals...)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewFunctionColumnSet([]metadata.FunctionColumn{}), nil
		}
		return nil, err
	}
	defer closeRows()

	results := []metadata.FunctionColumn{}
	for rows.Next() {
		rec := metadata.FunctionColumn{}
		err = rows.Scan(
			&rec.Catalog,
			&rec.Schema,
			&rec.FunctionName,
			&rec.Name,
			&rec.OrdinalPosition,
			&rec.Type,
			&rec.DataType,
			&rec.ColumnSize,
			&rec.DecimalDigits,
			&rec.NumPrecRadix,
			&rec.CharOctetLength,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewFunctionColumnSet(results), nil
}

// Indexes from selected catalog (or all, if empty), matching schemas and names
func (s InformationSchema) Indexes(f metadata.Filter) (*metadata.IndexSet, error) {
	if !s.hasIndexes {
		return nil, metadata.ErrNotSupported
	}

	qstr := `SELECT
  table_catalog,
  index_schema,
  table_name,
  index_name,
  CASE WHEN non_unique = 0 THEN 'YES' ELSE 'NO' END AS is_unique,
  CASE WHEN index_name = 'PRIMARY' THEN 'YES' ELSE 'NO' END AS is_primary,
  index_type
FROM information_schema.statistics
`
	conds, vals := s.conditions(1, f, formats{
		catalog:    "table_catalog LIKE %s",
		schema:     "index_schema LIKE %s",
		notSchemas: "index_schema NOT IN (%s)",
		parent:     "table_name LIKE %s",
		name:       "index_name LIKE %s",
	})
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
GROUP BY table_catalog, index_schema, table_name, index_name,
  CASE WHEN non_unique = 0 THEN 'YES' ELSE 'NO' END,
  CASE WHEN index_name = 'PRIMARY' THEN 'YES' ELSE 'NO' END,
  index_type`
	rows, closeRows, err := s.query(qstr, []string{}, "table_catalog, index_schema, table_name, index_name", vals...)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewIndexSet([]metadata.Index{}), nil
		}
		return nil, err
	}
	defer closeRows()

	results := []metadata.Index{}
	for rows.Next() {
		rec := metadata.Index{}
		err = rows.Scan(&rec.Catalog, &rec.Schema, &rec.Table, &rec.Name, &rec.IsUnique, &rec.IsPrimary, &rec.Type)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewIndexSet(results), nil
}

// IndexColumns from selected catalog (or all, if empty), matching schemas and indexes
func (s InformationSchema) IndexColumns(f metadata.Filter) (*metadata.IndexColumnSet, error) {
	if !s.hasIndexes {
		return nil, metadata.ErrNotSupported
	}

	qstr := `SELECT
  i.table_catalog,
  i.table_schema,
  i.table_name,
  i.index_name,
  i.column_name,
  c.data_type,
  i.seq_in_index

FROM information_schema.statistics i
JOIN information_schema.columns c ON
  i.table_catalog = c.table_catalog AND
  i.table_schema = c.table_schema AND
  i.table_name = c.table_name AND
  i.column_name = c.column_name
`
	conds, vals := s.conditions(1, f, formats{
		catalog:    "i.table_catalog LIKE %s",
		schema:     "index_schema LIKE %s",
		notSchemas: "index_schema NOT IN (%s)",
		parent:     "i.table_name LIKE %s",
		name:       "index_name LIKE %s",
	})
	rows, closeRows, err := s.query(qstr, conds, "i.table_catalog, index_schema, table_name, index_name, seq_in_index", vals...)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewIndexColumnSet([]metadata.IndexColumn{}), nil
		}
		return nil, err
	}
	defer closeRows()

	results := []metadata.IndexColumn{}
	for rows.Next() {
		rec := metadata.IndexColumn{}
		err = rows.Scan(&rec.Catalog, &rec.Schema, &rec.Table, &rec.IndexName, &rec.Name, &rec.DataType, &rec.OrdinalPosition)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewIndexColumnSet(results), nil
}

// Sequences from selected catalog (or all, if empty), matching schemas and names
func (s InformationSchema) Sequences(f metadata.Filter) (*metadata.SequenceSet, error) {
	if !s.hasSequences {
		return nil, metadata.ErrNotSupported
	}

	columns := []string{
		"sequence_catalog",
		"sequence_schema",
		"sequence_name",
		"data_type",
		"start_value",
		"minimum_value",
		"maximum_value",
		s.colExpr[SequenceColumnsIncrement],
		"cycle_option",
	}

	qstr := "SELECT\n  " + strings.Join(columns, ",\n  ") + " FROM information_schema.sequences\n"

	conds, vals := s.conditions(1, f, formats{
		catalog:    "sequence_catalog LIKE %s",
		schema:     "sequence_schema LIKE %s",
		notSchemas: "sequence_schema NOT IN (%s)",
		name:       "sequence_name LIKE %s",
	})
	rows, closeRows, err := s.query(qstr, conds, "sequence_catalog, sequence_schema, sequence_name", vals...)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewSequenceSet([]metadata.Sequence{}), nil
		}
		return nil, err
	}
	defer closeRows()

	results := []metadata.Sequence{}
	for rows.Next() {
		rec := metadata.Sequence{}
		err = rows.Scan(&rec.Catalog, &rec.Schema, &rec.Name, &rec.DataType, &rec.Start, &rec.Min, &rec.Max, &rec.Increment, &rec.Cycles)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewSequenceSet(results), nil
}

func (s InformationSchema) conditions(baseParam int, filter metadata.Filter, formats formats) ([]string, []interface{}) {
	conds := []string{}
	vals := []interface{}{}
	if filter.Catalog != "" && formats.catalog != "" {
		vals = append(vals, filter.Catalog)
		conds = append(conds, fmt.Sprintf(formats.catalog, s.pf(baseParam)))
		baseParam++
	}
	if filter.Schema != "" && formats.schema != "" {
		vals = append(vals, filter.Schema)
		conds = append(conds, fmt.Sprintf(formats.schema, s.pf(baseParam)))
		baseParam++
	}

	if !filter.WithSystem && formats.notSchemas != "" && len(s.systemSchemas) != 0 {
		pholders := []string{}
		for _, v := range s.systemSchemas {
			if v == filter.Schema {
				continue
			}
			vals = append(vals, v)
			pholders = append(pholders, s.pf(baseParam))
			baseParam++
		}
		if len(pholders) != 0 {
			conds = append(conds, fmt.Sprintf(formats.notSchemas, strings.Join(pholders, ", ")))
		}
	}
	if filter.OnlyVisible && formats.schema != "" && s.currentSchema != "" {
		conds = append(conds, fmt.Sprintf(formats.schema, s.currentSchema))
	}
	if filter.Parent != "" && formats.parent != "" {
		vals = append(vals, filter.Parent)
		conds = append(conds, fmt.Sprintf(formats.parent, s.pf(baseParam)))
		baseParam++
	}
	if filter.Name != "" && formats.name != "" {
		vals = append(vals, filter.Name)
		conds = append(conds, fmt.Sprintf(formats.name, s.pf(baseParam)))
		baseParam++
	}
	if len(filter.Types) != 0 && formats.types != "" {
		pholders := []string{}
		for _, t := range filter.Types {
			vals = append(vals, t)
			pholders = append(pholders, s.pf(baseParam))
			baseParam++
		}
		if len(pholders) != 0 {
			conds = append(conds, fmt.Sprintf(formats.types, strings.Join(pholders, ", ")))
		}
	}

	return conds, vals
}

type formats struct {
	catalog    string
	schema     string
	notSchemas string
	parent     string
	name       string
	types      string
}

func (s InformationSchema) query(qstr string, conds []string, order string, vals ...interface{}) (*sql.Rows, func(), error) {
	if len(conds) != 0 {
		qstr += "\nWHERE " + strings.Join(conds, " AND ")
	}
	if order != "" {
		qstr += "\nORDER BY " + order
	}
	if s.limit != 0 {
		qstr += fmt.Sprintf("\nLIMIT %d", s.limit)
	}
	return s.Query(qstr, vals...)
}
