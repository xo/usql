package impala

import (
	"context"
	"database/sql"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"

	driver "github.com/sclgo/impala-go"
)

type MetaReader struct {
	meta *driver.Metadata
}

func (r MetaReader) Columns(filter metadata.Filter) (*metadata.ColumnSet, error) {
	columnIds, err := r.meta.GetColumns(context.Background(), filter.Schema, filter.Parent, filter.Name)
	if err != nil {
		return nil, err
	}
	columns := make([]metadata.Column, len(columnIds))
	for i, columnId := range columnIds {
		columns[i] = metadata.Column{
			Schema: columnId.Schema,
			Table:  columnId.TableName,
			Name:   columnId.ColumnName,
		}
	}
	return metadata.NewColumnSet(columns), nil
}

func (r MetaReader) Schemas(filter metadata.Filter) (*metadata.SchemaSet, error) {
	schemaNames, err := r.meta.GetSchemas(context.Background(), filter.Name)
	if err != nil {
		return nil, err
	}
	schemas := make([]metadata.Schema, len(schemaNames))
	for i, name := range schemaNames {
		schemas[i] = metadata.Schema{
			Schema: name,
		}
	}
	return metadata.NewSchemaSet(schemas), nil
}

func (r MetaReader) Tables(filter metadata.Filter) (*metadata.TableSet, error) {
	tableIds, err := r.meta.GetTables(context.Background(), filter.Schema, filter.Name)
	if err != nil {
		return nil, err
	}
	tables := make([]metadata.Table, len(tableIds))
	for i, table := range tableIds {
		tables[i] = metadata.Table{
			Schema: table.Schema,
			Name:   table.Name,
			Type:   table.Type,
		}
	}
	return metadata.NewTableSet(tables), nil
}

var (
	_ metadata.SchemaReader = (*MetaReader)(nil)
	_ metadata.TableReader  = (*MetaReader)(nil)
	_ metadata.ColumnReader = (*MetaReader)(nil)
)

func New(db drivers.DB, _ ...metadata.ReaderOption) metadata.Reader {
	if sqlDb, ok := db.(*sql.DB); ok {
		return &MetaReader{
			meta: driver.NewMetadata(sqlDb),
		}
	} else {
		return struct{}{} // reader with no capabilities
	}
}
