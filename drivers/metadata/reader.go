package metadata

import (
	"context"
	"database/sql"
	"time"
)

// PluginReader allows to be easily composed from other readers
type PluginReader struct {
	catalogs        func(Filter) (*CatalogSet, error)
	schemas         func(Filter) (*SchemaSet, error)
	tables          func(Filter) (*TableSet, error)
	columns         func(Filter) (*ColumnSet, error)
	indexes         func(Filter) (*IndexSet, error)
	indexColumns    func(Filter) (*IndexColumnSet, error)
	functions       func(Filter) (*FunctionSet, error)
	functionColumns func(Filter) (*FunctionColumnSet, error)
	sequences       func(Filter) (*SequenceSet, error)
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

func (p PluginReader) Catalogs(f Filter) (*CatalogSet, error) {
	if p.catalogs == nil {
		return nil, ErrNotSupported
	}
	return p.catalogs(f)
}

func (p PluginReader) Schemas(f Filter) (*SchemaSet, error) {
	if p.schemas == nil {
		return nil, ErrNotSupported
	}
	return p.schemas(f)
}

func (p PluginReader) Tables(f Filter) (*TableSet, error) {
	if p.tables == nil {
		return nil, ErrNotSupported
	}
	return p.tables(f)
}

func (p PluginReader) Columns(f Filter) (*ColumnSet, error) {
	if p.columns == nil {
		return nil, ErrNotSupported
	}
	return p.columns(f)
}

func (p PluginReader) Indexes(f Filter) (*IndexSet, error) {
	if p.indexes == nil {
		return nil, ErrNotSupported
	}
	return p.indexes(f)
}

func (p PluginReader) IndexColumns(f Filter) (*IndexColumnSet, error) {
	if p.indexColumns == nil {
		return nil, ErrNotSupported
	}
	return p.indexColumns(f)
}

func (p PluginReader) Functions(f Filter) (*FunctionSet, error) {
	if p.functions == nil {
		return nil, ErrNotSupported
	}
	return p.functions(f)
}

func (p PluginReader) FunctionColumns(f Filter) (*FunctionColumnSet, error) {
	if p.functionColumns == nil {
		return nil, ErrNotSupported
	}
	return p.functionColumns(f)
}

func (p PluginReader) Sequences(f Filter) (*SequenceSet, error) {
	if p.sequences == nil {
		return nil, ErrNotSupported
	}
	return p.sequences(f)
}

type LoggingReader struct {
	db      DB
	logger  logger
	dryRun  bool
	timeout time.Duration
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

// WithTimeout for a single query
func WithTimeout(t time.Duration) ReaderOption {
	return func(r Reader) {
		r.(loggerSetter).setTimeout(t)
	}
}

// WithLimit for a single query, if the reader supports it
func WithLimit(l int) ReaderOption {
	return func(r Reader) {
		if rl, ok := r.(limiter); ok {
			rl.SetLimit(l)
		}
	}
}

type loggerSetter interface {
	setLogger(logger)
	setDryRun(bool)
	setTimeout(t time.Duration)
}

type limiter interface {
	SetLimit(l int)
}

func (r *LoggingReader) setLogger(l logger) {
	r.logger = l
}

func (r *LoggingReader) setDryRun(d bool) {
	r.dryRun = d
}

func (r *LoggingReader) setTimeout(t time.Duration) {
	r.timeout = t
}

func (r LoggingReader) Query(q string, v ...interface{}) (*sql.Rows, CloseFunc, error) {
	if r.logger != nil {
		r.logger.Println(q)
		r.logger.Println(v)
	}
	if r.dryRun {
		return nil, nil, sql.ErrNoRows
	}
	if r.timeout != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		rows, err := r.db.QueryContext(ctx, q, v...)
		return rows, func() { cancel(); rows.Close() }, err
	}
	rows, err := r.db.Query(q, v...)
	return rows, func() { rows.Close() }, err
}

// CloseFunc should be called when result wont be processed anymore
type CloseFunc func()
