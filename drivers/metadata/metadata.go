package metadata

import (
	"errors"
)

// ExtendedReader of all database metadata in a structured format.
type ExtendedReader interface {
	SchemaReader
	TableReader
	ColumnReader
	IndexReader
	IndexColumnReader
	FunctionReader
	FunctionColumnReader
	SequenceReader
}

// BasicReader of common database metadata like schemas, tables and columns.
type BasicReader interface {
	SchemaReader
	TableReader
	ColumnReader
}

// SchemaReader lists database schemas.
type SchemaReader interface {
	Reader
	Schemas(catalog, schemaPattern string) (*SchemaSet, error)
}

// TableReader lists database tables.
type TableReader interface {
	Reader
	Tables(catalog, schemaPattern, namePattern string, types []string) (*TableSet, error)
}

// ColumnReader lists table columns.
type ColumnReader interface {
	Reader
	Columns(catalog, schemaPattern, tablePattern string) (*ColumnSet, error)
}

// IndexReader lists database indexes.
type IndexReader interface {
	Reader
	Indexes(catalog, schemaPattern, namePattern string) (*IndexSet, error)
}

// IndexColumnReader lists database indexes.
type IndexColumnReader interface {
	Reader
	IndexColumns(catalog, schemaPattern, indexPattern string) (*IndexColumnSet, error)
}

// FunctionReader lists database functions.
type FunctionReader interface {
	Reader
	Functions(catalog, schemaPattern, namePattern string, types []string) (*FunctionSet, error)
}

// FunctionColumnReader lists function parameters.
type FunctionColumnReader interface {
	Reader
	FunctionColumns(catalog, schemaPattern, functionPattern string) (*FunctionColumnSet, error)
}

// SequenceReader lists sequences.
type SequenceReader interface {
	Reader
	Sequences(catalog, schemaPattern, namePattern string) (*SequenceSet, error)
}

// Reader of any database metadata in a structured format.
type Reader interface{}

// Writer of database metadata in a human readable format.
type Writer interface {
	// DescribeAggregates \da
	DescribeAggregates(string, bool, bool) error
	// DescribeFunctions \df, \dfa, \dfn, \dft, \dfw, etc.
	DescribeFunctions(string, string, bool, bool) error
	// DescribeTableDetails \d foo
	DescribeTableDetails(string, bool, bool) error
	// ListAllDbs \l
	ListAllDbs(string, bool) error
	// ListTables \dt, \dv, \dm, etc.
	ListTables(string, string, bool, bool) error
	// ListSchemas \dn
	ListSchemas(string, bool, bool) error
	// ListIndexes \di
	ListIndexes(string, bool, bool) error
}

type SchemaSet struct {
	resultSet
}

func NewSchemaSet(v []Schema) *SchemaSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
		r[i].setVerbose(true)
	}
	return &SchemaSet{
		resultSet: resultSet{
			results: r,
			verbose: true,
		},
	}
}

func (s SchemaSet) Columns() ([]string, error) {
	return []string{"Schema", "Catalog"}, nil
}

func (s SchemaSet) Get() *Schema {
	return s.results[s.current-1].(*Schema)
}

type Schema struct {
	verbose bool

	Schema  string
	Catalog string
}

func (s Schema) values() []interface{} {
	return []interface{}{s.Schema, s.Catalog}
}

func (s *Schema) setVerbose(v bool) {
	s.verbose = v
}

type TableSet struct {
	resultSet
}

func NewTableSet(v []Table) *TableSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
		r[i].setVerbose(true)
	}
	return &TableSet{
		resultSet: resultSet{
			results: r,
			verbose: true,
		},
	}
}

func (t TableSet) Columns() ([]string, error) {
	if !t.verbose {
		return []string{"Schema", "Name", "Type"}, nil
	}
	return []string{"Catalog", "Schema", "Name", "Type", "Size", "Comment"}, nil
}

func (t TableSet) Get() *Table {
	return t.results[t.current-1].(*Table)
}

type Table struct {
	verbose bool

	Catalog string
	Schema  string
	Name    string
	Type    string
	Size    string
	Comment string
}

func (t Table) values() []interface{} {
	if !t.verbose {
		return []interface{}{t.Schema, t.Name, t.Type}
	}
	return []interface{}{t.Catalog, t.Schema, t.Name, t.Type, t.Size, t.Comment}
}

func (t *Table) setVerbose(v bool) {
	t.verbose = v
}

type ColumnSet struct {
	resultSet
}

func NewColumnSet(v []Column) *ColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
		r[i].setVerbose(true)
	}
	return &ColumnSet{
		resultSet: resultSet{
			results: r,
			verbose: true,
		},
	}
}

func (c ColumnSet) Columns() ([]string, error) {
	if !c.verbose {
		return []string{"Name", "Type", "Nullable", "Default"}, nil
	}
	return []string{"Catalog", "Schema", "Table", "Name", "Type", "Nullable", "Default", "Size", "Decimal Digits", "Precision Radix", "Octet Length"}, nil
}

func (c ColumnSet) Get() *Column {
	return c.results[c.current-1].(*Column)
}

type Column struct {
	verbose bool

	Catalog         string
	Schema          string
	Table           string
	Name            string
	OrdinalPosition int
	DataType        string
	// ScanType        reflect.Type
	Default         string
	ColumnSize      int
	DecimalDigits   int
	NumPrecRadix    int
	CharOctetLength int
	IsNullable      Bool
}

type Bool string

var (
	UNKNOWN Bool = ""
	YES     Bool = "YES"
	NO      Bool = "NO"
)

func (c Column) values() []interface{} {
	if !c.verbose {
		return []interface{}{
			c.Name,
			c.DataType,
			c.IsNullable,
			c.Default,
		}
	}
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.Table,
		c.Name,
		c.DataType,
		c.IsNullable,
		c.Default,
		c.ColumnSize,
		c.DecimalDigits,
		c.NumPrecRadix,
		c.CharOctetLength,
	}
}

func (c *Column) setVerbose(v bool) {
	c.verbose = v
}

type IndexSet struct {
	resultSet
}

func NewIndexSet(v []Index) *IndexSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
		r[i].setVerbose(true)
	}
	return &IndexSet{
		resultSet: resultSet{
			results: r,
			verbose: true,
		},
	}
}

func (i IndexSet) Columns() ([]string, error) {
	if !i.verbose {
		return []string{"Schema", "Name", "Table"}, nil
	}
	return []string{"Catalog", "Schema", "Name", "Table", "Is primary", "Is unique", "Type"}, nil
}

func (i IndexSet) Get() *Index {
	return i.results[i.current-1].(*Index)
}

type Index struct {
	verbose bool

	Catalog   string
	Schema    string
	Table     string
	Name      string
	IsPrimary Bool
	IsUnique  Bool
	Type      string
	Columns   string
}

func (i Index) values() []interface{} {
	if !i.verbose {
		return []interface{}{i.Schema, i.Name, i.Table}
	}
	return []interface{}{i.Catalog, i.Schema, i.Name, i.Table, i.IsPrimary, i.IsUnique, i.Type}
}

func (i *Index) setVerbose(v bool) {
	i.verbose = v
}

type IndexColumnSet struct {
	resultSet
}

func NewIndexColumnSet(v []IndexColumn) *IndexColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
		r[i].setVerbose(true)
	}
	return &IndexColumnSet{
		resultSet: resultSet{
			results: r,
			verbose: true,
		},
	}
}

func (c IndexColumnSet) Columns() ([]string, error) {
	if !c.verbose {
		return []string{"Name", "Data type"}, nil
	}
	return []string{"Catalog", "Schema", "Table", "Index name", "Name", "Data type"}, nil
}

func (c IndexColumnSet) Get() *IndexColumn {
	return c.results[c.current-1].(*IndexColumn)
}

type IndexColumn struct {
	verbose bool

	Catalog         string
	Schema          string
	Table           string
	IndexName       string
	Name            string
	DataType        string
	OrdinalPosition int
}

func (c IndexColumn) values() []interface{} {
	if !c.verbose {
		return []interface{}{c.Name, c.DataType}
	}
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.Table,
		c.IndexName,
		c.Name,
	}
}

func (c *IndexColumn) setVerbose(v bool) {
	c.verbose = v
}

type FunctionSet struct {
	resultSet
}

func NewFunctionSet(v []Function) *FunctionSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
		r[i].setVerbose(true)
	}
	return &FunctionSet{
		resultSet: resultSet{
			results: r,
			verbose: true,
		},
	}
}

func (f FunctionSet) Columns() ([]string, error) {
	if !f.verbose {
		return []string{"Schema", "Name", "Result data type", "Argument data types", "Type"}, nil
	}
	return []string{"Catalog", "Schema", "Name", "Result data type", "Argument data types", "Type", "Volatility", "Security", "Language", "Source code"}, nil
}

func (f FunctionSet) Get() *Function {
	return f.results[f.current-1].(*Function)
}

type Function struct {
	verbose bool

	Catalog    string
	Schema     string
	Name       string
	ResultType string
	ArgTypes   string
	Type       string
	Volatility string
	Security   string
	Language   string
	Source     string

	SpecificName string
}

func (f Function) values() []interface{} {
	if !f.verbose {
		return []interface{}{f.Schema, f.Name, f.ResultType, f.ArgTypes, f.Type}
	}
	return []interface{}{
		f.Catalog,
		f.Schema,
		f.Name,
		f.ResultType,
		f.ArgTypes,
		f.Type,
		f.Volatility,
		f.Security,
		f.Language,
		f.Source,
	}
}

func (f *Function) setVerbose(v bool) {
	f.verbose = v
}

type FunctionColumnSet struct {
	resultSet
}

func NewFunctionColumnSet(v []FunctionColumn) *FunctionColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
		r[i].setVerbose(true)
	}
	return &FunctionColumnSet{
		resultSet: resultSet{
			results: r,
			verbose: true,
		},
	}
}

func (c FunctionColumnSet) Columns() ([]string, error) {
	if !c.verbose {
		return []string{"Name", "Type", "Data type"}, nil
	}
	return []string{"Catalog", "Schema", "Function name", "Name", "Type", "Data type", "Size", "Decimal Digits", "Precision Radix", "Octet Length"}, nil
}

func (c FunctionColumnSet) Get() *FunctionColumn {
	return c.results[c.current-1].(*FunctionColumn)
}

type FunctionColumn struct {
	verbose bool

	Catalog         string
	Schema          string
	Table           string
	Name            string
	FunctionName    string
	OrdinalPosition int
	Type            string
	DataType        string
	// ScanType        reflect.Type
	ColumnSize      int
	DecimalDigits   int
	NumPrecRadix    int
	CharOctetLength int
}

func (c FunctionColumn) values() []interface{} {
	if !c.verbose {
		return []interface{}{
			c.Name,
			c.Type,
			c.DataType,
		}
	}
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.FunctionName,
		c.Name,
		c.Type,
		c.DataType,
		c.ColumnSize,
		c.DecimalDigits,
		c.NumPrecRadix,
		c.CharOctetLength,
	}
}

func (c *FunctionColumn) setVerbose(v bool) {
	c.verbose = v
}

type SequenceSet struct {
	resultSet
}

func NewSequenceSet(v []Sequence) *SequenceSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
		r[i].setVerbose(true)
	}
	return &SequenceSet{
		resultSet: resultSet{
			results: r,
			verbose: true,
		},
	}
}

func (s SequenceSet) Columns() ([]string, error) {
	return []string{"Type", "Start", "Min", "Max", "Increment", "Cycles?"}, nil
}

func (s SequenceSet) Get() *Sequence {
	return s.results[s.current-1].(*Sequence)
}

type Sequence struct {
	verbose bool

	Catalog   string
	Schema    string
	Name      string
	DataType  string
	Start     string
	Min       string
	Max       string
	Increment string
	Cycles    Bool
}

func (s Sequence) values() []interface{} {
	return []interface{}{s.DataType, s.Start, s.Min, s.Max, s.Increment, s.Cycles}
}

func (s *Sequence) setVerbose(v bool) {
	s.verbose = v
}

type resultSet struct {
	results []Result
	current int
	verbose bool
}

type Result interface {
	values() []interface{}
	setVerbose(bool)
}

func (r *resultSet) SetVerbose(v bool) {
	r.verbose = v
	for _, rec := range r.results {
		rec.setVerbose(v)
	}
}

func (r *resultSet) Len() int {
	return len(r.results)
}

func (r *resultSet) Reset() {
	r.current = 0
}

func (r *resultSet) Next() bool {
	r.current++
	return r.current <= len(r.results)
}

func (r resultSet) Scan(dest ...interface{}) error {
	v := r.results[r.current-1].values()
	if len(v) != len(dest) {
		return errors.New("error: wrong number of arguments for Scan()")
	}
	for i, d := range dest {
		p := d.(*interface{})
		*p = v[i]
	}
	return nil
}

func (r resultSet) Close() error {
	return nil
}

func (r resultSet) Err() error {
	return nil
}

func (r resultSet) NextResultSet() bool {
	return false
}
