package metadata

import (
	"errors"
)

var (
	ErrNotSupported  = errors.New("error: not supported")
	ErrScanArgsCount = errors.New("error: wrong number of arguments for Scan()")
)

// ExtendedReader of all database metadata in a structured format.
type ExtendedReader interface {
	CatalogReader
	SchemaReader
	TableReader
	ColumnReader
	IndexReader
	IndexColumnReader
	TriggerReader
	ConstraintReader
	ConstraintColumnReader
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

// CatalogReader lists database schemas.
type CatalogReader interface {
	Reader
	Catalogs(Filter) (*CatalogSet, error)
}

// SchemaReader lists database schemas.
type SchemaReader interface {
	Reader
	Schemas(Filter) (*SchemaSet, error)
}

// TableReader lists database tables.
type TableReader interface {
	Reader
	Tables(Filter) (*TableSet, error)
}

// ColumnReader lists table columns.
type ColumnReader interface {
	Reader
	Columns(Filter) (*ColumnSet, error)
}

// IndexReader lists table indexes.
type IndexReader interface {
	Reader
	Indexes(Filter) (*IndexSet, error)
}

// IndexColumnReader lists index columns.
type IndexColumnReader interface {
	Reader
	IndexColumns(Filter) (*IndexColumnSet, error)
}

// TriggerReader lists table triggers.
type TriggerReader interface {
	Reader
	Triggers(Filter) (*TriggerSet, error)
}

// ConstraintReader lists table constraints.
type ConstraintReader interface {
	Reader
	Constraints(Filter) (*ConstraintSet, error)
}

// ConstraintColumnReader lists constraint columns.
type ConstraintColumnReader interface {
	Reader
	ConstraintColumns(Filter) (*ConstraintColumnSet, error)
}

// FunctionReader lists database functions.
type FunctionReader interface {
	Reader
	Functions(Filter) (*FunctionSet, error)
}

// FunctionColumnReader lists function parameters.
type FunctionColumnReader interface {
	Reader
	FunctionColumns(Filter) (*FunctionColumnSet, error)
}

// SequenceReader lists sequences.
type SequenceReader interface {
	Reader
	Sequences(Filter) (*SequenceSet, error)
}

// Reader of any database metadata in a structured format.
type Reader interface{}

// Filter objects returned by Readers
type Filter struct {
	// Catalog name pattern that objects must belong to;
	// use Name to filter catalogs by name
	Catalog string
	// Schema name pattern that objects must belong to;
	// use Name to filter schemas by name
	Schema string
	// Parent name pattern that objects must belong to;
	// does not apply to schema and catalog containing matching objects
	Parent string
	// Reference name pattern of other objects referencing this one,
	Reference string
	// Name pattern that object name must match
	Name string
	// Types of the object
	Types []string
	// WithSystem objects
	WithSystem bool
	// OnlyVisible objects
	OnlyVisible bool
}

// Writer of database metadata in a human readable format.
type Writer interface {
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

type CatalogSet struct {
	resultSet
}

func NewCatalogSet(v []Catalog) *CatalogSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &CatalogSet{
		resultSet: resultSet{
			results: r,
			columns: []string{"Catalog"},
		},
	}
}

func (s CatalogSet) Get() *Catalog {
	return s.results[s.current-1].(*Catalog)
}

type Catalog struct {
	Catalog string
}

func (s Catalog) values() []interface{} {
	return []interface{}{s.Catalog}
}

type SchemaSet struct {
	resultSet
}

func NewSchemaSet(v []Schema) *SchemaSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &SchemaSet{
		resultSet: resultSet{
			results: r,
			columns: []string{"Schema", "Catalog"},
		},
	}
}

func (s SchemaSet) Get() *Schema {
	return s.results[s.current-1].(*Schema)
}

type Schema struct {
	Schema  string
	Catalog string
}

func (s Schema) values() []interface{} {
	return []interface{}{s.Schema, s.Catalog}
}

type TableSet struct {
	resultSet
}

func NewTableSet(v []Table) *TableSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &TableSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",

				"Name",
				"Type",

				"Size",
				"Comment",
			},
		},
	}
}

func (t TableSet) Get() *Table {
	return t.results[t.current-1].(*Table)
}

type Table struct {
	Catalog string
	Schema  string
	Name    string
	Type    string
	Size    string
	Comment string
}

func (t Table) values() []interface{} {
	return []interface{}{
		t.Catalog,
		t.Schema,
		t.Name,
		t.Type,
		t.Size,
		t.Comment,
	}
}

type ColumnSet struct {
	resultSet
}

func NewColumnSet(v []Column) *ColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &ColumnSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",

				"Name",
				"Type",
				"Nullable",
				"Default",

				"Size",
				"Decimal Digits",
				"Precision Radix",
				"Octet Length",
			},
		},
	}
}

func (c ColumnSet) Get() *Column {
	return c.results[c.current-1].(*Column)
}

type Column struct {
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

type IndexSet struct {
	resultSet
}

func NewIndexSet(v []Index) *IndexSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &IndexSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",

				"Name",
				"Table",

				"Is primary",
				"Is unique",
				"Type",
			},
		},
	}
}

func (i IndexSet) Get() *Index {
	return i.results[i.current-1].(*Index)
}

type Index struct {
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
	return []interface{}{
		i.Catalog,
		i.Schema,
		i.Name,
		i.Table,
		i.IsPrimary,
		i.IsUnique,
		i.Type,
	}
}

type IndexColumnSet struct {
	resultSet
}

func NewIndexColumnSet(v []IndexColumn) *IndexColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &IndexColumnSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",
				"Index name",

				"Name",
				"Data type",
			},
		},
	}
}

func (c IndexColumnSet) Get() *IndexColumn {
	return c.results[c.current-1].(*IndexColumn)
}

type IndexColumn struct {
	Catalog         string
	Schema          string
	Table           string
	IndexName       string
	Name            string
	DataType        string
	OrdinalPosition int
}

func (c IndexColumn) values() []interface{} {
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.Table,
		c.IndexName,
		c.Name,
		c.DataType,
	}
}

type ConstraintSet struct {
	resultSet
}

func NewConstraintSet(v []Constraint) *ConstraintSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &ConstraintSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",
				"Name",

				"Type",
				"Is deferrable",
				"Initially deferred",

				"Foreign catalog",
				"Foreign schema",
				"Foreign table",
				"Foreign name",
				"Match type",
				"Update rule",
				"Delete rule",

				"Check Clause",
			},
		},
	}
}

func (i ConstraintSet) Get() *Constraint {
	return i.results[i.current-1].(*Constraint)
}

type Constraint struct {
	Catalog string
	Schema  string
	Table   string
	Name    string
	Type    string

	IsDeferrable        Bool
	IsInitiallyDeferred Bool

	ForeignCatalog string
	ForeignSchema  string
	ForeignTable   string
	ForeignName    string
	MatchType      string
	UpdateRule     string
	DeleteRule     string

	CheckClause string
}

func (i Constraint) values() []interface{} {
	return []interface{}{
		i.Catalog,
		i.Schema,
		i.Table,
		i.Name,
		i.Type,
		i.IsDeferrable,
		i.IsInitiallyDeferred,
		i.ForeignCatalog,
		i.ForeignSchema,
		i.ForeignTable,
		i.ForeignName,
		i.MatchType,
		i.UpdateRule,
		i.DeleteRule,
	}
}

type ConstraintColumnSet struct {
	resultSet
}

func NewConstraintColumnSet(v []ConstraintColumn) *ConstraintColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &ConstraintColumnSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",
				"Constraint",
				"Name",
				"Foreign Catalog",
				"Foreign Schema",
				"Foreign Table",
				"Foreign Constraint",
				"Foreign Name",
			},
		},
	}
}

func (c ConstraintColumnSet) Get() *ConstraintColumn {
	return c.results[c.current-1].(*ConstraintColumn)
}

type ConstraintColumn struct {
	Catalog         string
	Schema          string
	Table           string
	Constraint      string
	Name            string
	OrdinalPosition int

	ForeignCatalog    string
	ForeignSchema     string
	ForeignTable      string
	ForeignConstraint string
	ForeignName       string
}

func (c ConstraintColumn) values() []interface{} {
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.Table,
		c.Constraint,
		c.Name,
		c.ForeignCatalog,
		c.ForeignSchema,
		c.ForeignTable,
		c.ForeignConstraint,
		c.ForeignName,
	}
}

type FunctionSet struct {
	resultSet
}

func NewFunctionSet(v []Function) *FunctionSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &FunctionSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",

				"Name",
				"Result data type",
				"Argument data types",
				"Type",

				"Volatility",
				"Security",
				"Language",
				"Source code",
			},
		},
	}
}

func (f FunctionSet) Get() *Function {
	return f.results[f.current-1].(*Function)
}

type Function struct {
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

type FunctionColumnSet struct {
	resultSet
}

func NewFunctionColumnSet(v []FunctionColumn) *FunctionColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &FunctionColumnSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Function name",

				"Name",
				"Type",
				"Data type",

				"Size",
				"Decimal Digits",
				"Precision Radix",
				"Octet Length",
			},
		},
	}
}

func (c FunctionColumnSet) Get() *FunctionColumn {
	return c.results[c.current-1].(*FunctionColumn)
}

type FunctionColumn struct {
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

type SequenceSet struct {
	resultSet
}

func NewSequenceSet(v []Sequence) *SequenceSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &SequenceSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Type",
				"Start",
				"Min",
				"Max",
				"Increment",
				"Cycles?",
			},
		},
	}
}

func (s SequenceSet) Get() *Sequence {
	return s.results[s.current-1].(*Sequence)
}

type Sequence struct {
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
	return []interface{}{
		s.DataType,
		s.Start,
		s.Min,
		s.Max,
		s.Increment,
		s.Cycles,
	}
}

type resultSet struct {
	results    []Result
	columns    []string
	current    int
	filter     func(Result) bool
	scanValues func(Result) []interface{}
}

type Result interface {
	values() []interface{}
}

func (r *resultSet) SetFilter(f func(Result) bool) {
	r.filter = f
}

func (r *resultSet) SetColumns(c []string) {
	r.columns = c
}

func (r *resultSet) SetScanValues(s func(Result) []interface{}) {
	r.scanValues = s
}

func (r *resultSet) Len() int {
	if r.filter == nil {
		return len(r.results)
	}
	len := 0
	for _, rec := range r.results {
		if r.filter(rec) {
			len++
		}
	}
	return len
}

func (r *resultSet) Reset() {
	r.current = 0
}

func (r *resultSet) Next() bool {
	r.current++
	if r.filter != nil {
		for r.current <= len(r.results) && !r.filter(r.results[r.current-1]) {
			r.current++
		}
	}
	return r.current <= len(r.results)
}

func (r resultSet) Columns() ([]string, error) {
	return r.columns, nil
}

func (r resultSet) Scan(dest ...interface{}) error {
	var v []interface{}
	if r.scanValues == nil {
		v = r.results[r.current-1].values()
	} else {
		v = r.scanValues(r.results[r.current-1])
	}
	if len(v) != len(dest) {
		return ErrScanArgsCount
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

type Trigger struct {
	Catalog    string
	Schema     string
	Table      string
	Name       string
	Definition string
}

func (t Trigger) values() []interface{} {
	return []interface{}{
		t.Catalog,
		t.Schema,
		t.Table,
		t.Name,
		t.Definition,
	}
}

type TriggerSet struct {
	resultSet
}

func NewTriggerSet(t []Trigger) *TriggerSet {
	r := make([]Result, len(t))
	for i := range t {
		r[i] = &t[i]
	}
	return &TriggerSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",
				"Name",
				"Definition",
			},
		},
	}
}

func (t TriggerSet) Get() *Trigger {
	return t.results[t.current-1].(*Trigger)
}
