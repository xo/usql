package metadata

import (
	"errors"
)

// Reader of database metadata.
type Reader interface {
	Columns(catalog, schema, table string) (*ColumnSet, error)
	Tables(catalog, schemaPattern, tableNamePattern string, types []string) (*TableSet, error)
	Schemas() (*SchemaSet, error)
}

// Writer of database metadata.
type Writer interface {
	// DescribeAggregates \da
	DescribeAggregates(string, bool, bool) error
	// DescribeFunctions \df, \dfa, \dfn, \dft, \dfw, etc.
	DescribeFunctions(string, string, bool, bool) error
	// DescribeTableDetails \d foo
	DescribeTableDetails(string, bool, bool) error
	// ListAllDbs \l
	ListAllDbs(string, bool) error
	// ListTables \dt, \di, \ds, \dS, etc.
	ListTables(string, string, bool, bool) error
	// ListSchemas \dn
	ListSchemas(string, bool, bool) error
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
			verbose: true,
		},
	}
}

func (c ColumnSet) Columns() ([]string, error) {
	if !c.verbose {
		return []string{"Name", "Type", "Nullable", "Default"}, nil
	}
	return []string{"Catalog", "Schema", "Table", "Name", "Type", "Nullable", "Default", "Size", "Decimal Digits", "Precision Radix", "Octet Length", "Generated", "Identity"}, nil
}

func (c *ColumnSet) SetVerbose(v bool) {
	c.verbose = v
	for _, r := range c.results {
		r.setVerbose(v)
	}
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
	ColumnDefault   string
	ColumnSize      int
	DecimalDigits   int
	NumPrecRadix    int
	CharOctetLength int
	IsNullable      Bool
	IsGenerated     Bool
	IsIdentity      Bool
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
			c.ColumnDefault,
		}
	}
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.Table,
		c.Name,
		c.DataType,
		c.IsNullable,
		c.ColumnDefault,
		c.ColumnSize,
		c.DecimalDigits,
		c.NumPrecRadix,
		c.CharOctetLength,
		c.IsGenerated,
		c.IsIdentity,
	}
}

func (c *Column) setVerbose(v bool) {
	c.verbose = v
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

func (t *TableSet) SetVerbose(v bool) {
	t.verbose = v
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
			verbose: true,
		},
	}
}

func (s SchemaSet) Columns() ([]string, error) {
	return []string{"Schema", "Catalog"}, nil
}

func (s *SchemaSet) SetVerbose(v bool) {
	s.verbose = v
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

type resultSet struct {
	results []Result
	current int
	verbose bool
}

type Result interface {
	values() []interface{}
	setVerbose(bool)
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
