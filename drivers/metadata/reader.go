package metadata

import "database/sql"

// PluginReader allows to be easily composed from other readers
type PluginReader struct {
	catalogs        func() (*CatalogSet, error)
	schemas         func(catalog, schemaPattern string) (*SchemaSet, error)
	tables          func(catalog, schemaPattern, namePattern string, types []string) (*TableSet, error)
	columns         func(catalog, schemaPattern, tablePattern string) (*ColumnSet, error)
	indexes         func(catalog, schemaPattern, tablePattern, namePattern string) (*IndexSet, error)
	indexColumns    func(catalog, schemaPattern, tablePattern, indexPattern string) (*IndexColumnSet, error)
	functions       func(catalog, schemaPattern, namePattern string, types []string) (*FunctionSet, error)
	functionColumns func(catalog, schemaPattern, functionPattern string) (*FunctionColumnSet, error)
	sequences       func(catalog, schemaPattern, namePattern string) (*SequenceSet, error)
}

var _ ExtendedReader = &PluginReader{}

// NewPluginReader allows to be easily composed from other readers
func NewPluginReader(readers ...Reader) Reader {
	p := PluginReader{}
	for _, i := range readers {
		if r, ok := i.(CatalogReader); ok {
			p.catalogs = r.Catalogs
		}
		if r, ok := i.(SchemaReader); ok {
			p.schemas = r.Schemas
		}
		if r, ok := i.(TableReader); ok {
			p.tables = r.Tables
		}
		if r, ok := i.(ColumnReader); ok {
			p.columns = r.Columns
		}
		if r, ok := i.(IndexReader); ok {
			p.indexes = r.Indexes
		}
		if r, ok := i.(IndexColumnReader); ok {
			p.indexColumns = r.IndexColumns
		}
		if r, ok := i.(FunctionReader); ok {
			p.functions = r.Functions
		}
		if r, ok := i.(FunctionColumnReader); ok {
			p.functionColumns = r.FunctionColumns
		}
		if r, ok := i.(SequenceReader); ok {
			p.sequences = r.Sequences
		}
	}
	return &p
}

func (p PluginReader) Catalogs() (*CatalogSet, error) {
	if p.catalogs == nil {
		return nil, ErrNotSupported
	}
	return p.catalogs()
}

func (p PluginReader) Schemas(catalog, schemaPattern string) (*SchemaSet, error) {
	if p.schemas == nil {
		return nil, ErrNotSupported
	}
	return p.schemas(catalog, schemaPattern)
}

func (p PluginReader) Tables(catalog, schemaPattern, namePattern string, types []string) (*TableSet, error) {
	if p.tables == nil {
		return nil, ErrNotSupported
	}
	return p.tables(catalog, schemaPattern, namePattern, types)
}

func (p PluginReader) Columns(catalog, schemaPattern, tablePattern string) (*ColumnSet, error) {
	if p.columns == nil {
		return nil, ErrNotSupported
	}
	return p.columns(catalog, schemaPattern, tablePattern)
}

func (p PluginReader) Indexes(catalog, schemaPattern, tablePattern, namePattern string) (*IndexSet, error) {
	if p.indexes == nil {
		return nil, ErrNotSupported
	}
	return p.indexes(catalog, schemaPattern, tablePattern, namePattern)
}

func (p PluginReader) IndexColumns(catalog, schemaPattern, tablePattern, indexPattern string) (*IndexColumnSet, error) {
	if p.indexColumns == nil {
		return nil, ErrNotSupported
	}
	return p.indexColumns(catalog, schemaPattern, tablePattern, indexPattern)
}

func (p PluginReader) Functions(catalog, schemaPattern, namePattern string, types []string) (*FunctionSet, error) {
	if p.functions == nil {
		return nil, ErrNotSupported
	}
	return p.functions(catalog, schemaPattern, namePattern, types)
}

func (p PluginReader) FunctionColumns(catalog, schemaPattern, functionPattern string) (*FunctionColumnSet, error) {
	if p.functionColumns == nil {
		return nil, ErrNotSupported
	}
	return p.functionColumns(catalog, schemaPattern, functionPattern)
}

func (p PluginReader) Sequences(catalog, schemaPattern, namePattern string) (*SequenceSet, error) {
	if p.sequences == nil {
		return nil, ErrNotSupported
	}
	return p.sequences(catalog, schemaPattern, namePattern)
}

type LoggingReader struct {
	db     DB
	logger logger
	dryRun bool
}

type logger interface {
	Println(...interface{})
}

func NewLoggingReader(db DB, opts ...ReaderOption) LoggingReader {
	r := LoggingReader{
		db: db,
	}
	for _, o := range opts {
		o(&r)
	}
	return r
}

// ReaderOption to configure the reader
type ReaderOption func(Reader)

// WithLogger used to log queries before executing them
func WithLogger(l logger) ReaderOption {
	return func(r Reader) {
		r.(loggerSetter).setLogger(l)
	}
}

// WithDryRun allows to avoid running any queries
func WithDryRun(d bool) ReaderOption {
	return func(r Reader) {
		r.(loggerSetter).setDryRun(d)
	}
}

type loggerSetter interface {
	setLogger(logger)
	setDryRun(bool)
}

func (r *LoggingReader) setLogger(l logger) {
	r.logger = l
}

func (r *LoggingReader) setDryRun(d bool) {
	r.dryRun = d
}

func (r LoggingReader) Query(q string, v ...interface{}) (*sql.Rows, error) {
	if r.logger != nil {
		r.logger.Println(q)
		r.logger.Println(v)
	}
	if r.dryRun {
		return nil, sql.ErrNoRows
	}
	return r.db.Query(q, v...)
}
