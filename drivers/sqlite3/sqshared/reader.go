package sqshared

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

type MetadataReader struct {
	metadata.LoggingReader
	limit int
}

// NewMetadataReader creates the metadata reader for sqlite3 databases.
func NewMetadataReader(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
	return &MetadataReader{
		LoggingReader: metadata.NewLoggingReader(db, opts...),
	}
}

var (
	_ metadata.BasicReader          = &MetadataReader{}
	_ metadata.FunctionReader       = &MetadataReader{}
	_ metadata.FunctionColumnReader = &MetadataReader{}
	_ metadata.IndexReader          = &MetadataReader{}
	_ metadata.IndexColumnReader    = &MetadataReader{}
)

func (r *MetadataReader) SetLimit(l int) {
	r.limit = l
}

// Columns from selected catalog (or all, if empty), matching schemas and tables
func (r MetadataReader) Columns(f metadata.Filter) (*metadata.ColumnSet, error) {
	tables, err := r.Tables(metadata.Filter{Catalog: f.Catalog, Schema: f.Schema, Name: f.Parent})
	if err != nil {
		return nil, err
	}
	results := []metadata.Column{}
	for tables.Next() {
		table := tables.Get()
		qstr := `SELECT
  cid,
  name,
  type,
  CASE WHEN "notnull" = 1 THEN 'NO' ELSE 'YES' END,
  COALESCE(dflt_value, '')
FROM pragma_table_info(?)`
		rows, closeRows, err := r.query(qstr, []string{}, "name", table.Name)
		if err != nil {
			return nil, err
		}
		defer closeRows()

		rec := metadata.Column{
			Catalog: table.Catalog,
			Schema:  table.Schema,
			Table:   table.Name,
		}
		for rows.Next() {
			err = rows.Scan(
				&rec.OrdinalPosition,
				&rec.Name,
				&rec.DataType,
				&rec.IsNullable,
				&rec.Default,
			)
			if err != nil {
				return nil, err
			}
			results = append(results, rec)
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}
	}

	return metadata.NewColumnSet(results), nil
}

func (r MetadataReader) Tables(f metadata.Filter) (*metadata.TableSet, error) {
	qstr := `SELECT
  '' AS table_catalog,
  '' AS table_schem,
  table_name,
  table_type
FROM (
    SELECT
      name AS table_name,
      UPPER(type) AS table_type
    FROM sqlite_master
    WHERE name NOT LIKE 'sqlite\_%' ESCAPE '\' AND UPPER(type) IN ('TABLE', 'VIEW')
    UNION ALL
    SELECT
      name AS table_name,
      'GLOBAL TEMPORARY' AS table_type
    FROM sqlite_temp_master
    UNION ALL
    SELECT
      name AS table_name,
      'SYSTEM TABLE' AS table_type
    FROM sqlite_master
    WHERE name LIKE 'sqlite\_%' ESCAPE '\' AND UPPER(type) IN ('TABLE', 'VIEW')
    UNION ALL
    SELECT
      name AS table_name,
      'SYSTEM TABLE' AS table_type
    FROM pragma_module_list
)`
	conds := []string{}
	vals := []interface{}{}
	if f.Catalog != "" {
		vals = append(vals, f.Catalog)
		conds = append(conds, "table_catalog = ?")
	}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, "table_schema LIKE ?")
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "table_name LIKE ?")
	}
	if len(f.Types) != 0 {
		pholders := []string{}
		for _, t := range f.Types {
			vals = append(vals, t)
			pholders = append(pholders, "?")
		}
		if len(pholders) != 0 {
			conds = append(conds, "table_type IN ("+strings.Join(pholders, ", ")+")")
		}
	}
	rows, closeRows, err := r.query(qstr, conds, "table_type, table_name", vals...)
	if err != nil {
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

func (r MetadataReader) Schemas(f metadata.Filter) (*metadata.SchemaSet, error) {
	qstr := `SELECT
  name AS schema_name,
  '' AS catalog_name
FROM pragma_database_list`
	conds := []string{}
	vals := []interface{}{}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "schema_name LIKE ?")
	}
	rows, closeRows, err := r.query(qstr, conds, "seq", vals...)
	if err != nil {
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

func (r MetadataReader) Functions(f metadata.Filter) (*metadata.FunctionSet, error) {
	qstr := `SELECT
  name AS specific_name,
  name AS routine_name,
  type AS routine_type
FROM pragma_function_list`
	conds := []string{}
	vals := []interface{}{}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "name LIKE ?")
	}
	if len(f.Types) != 0 {
		pholders := []string{}
		for _, t := range f.Types {
			vals = append(vals, t)
			pholders = append(pholders, "?")
		}
		if len(pholders) != 0 {
			conds = append(conds, "type IN ("+strings.Join(pholders, ", ")+")")
		}
	}
	rows, closeRows, err := r.query(qstr, conds, "name, type", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()

	results := []metadata.Function{}
	for rows.Next() {
		rec := metadata.Function{}
		err = rows.Scan(
			&rec.SpecificName,
			&rec.Name,
			&rec.Type,
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

func (r MetadataReader) FunctionColumns(metadata.Filter) (*metadata.FunctionColumnSet, error) {
	return &metadata.FunctionColumnSet{}, nil
}

func (r MetadataReader) Indexes(f metadata.Filter) (*metadata.IndexSet, error) {
	qstr := `SELECT
  m.name,
  i.name,
  CASE WHEN i."unique" = 1 THEN 'YES' ELSE 'NO' END,
  CASE WHEN i.origin = 'pk' THEN 'YES' ELSE 'NO' END
FROM sqlite_master m
JOIN pragma_index_list(m.name) i`
	conds := []string{"m.type = 'table'"}
	vals := []interface{}{}
	if f.Parent != "" {
		vals = append(vals, f.Parent)
		conds = append(conds, "m.name LIKE ?")
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "i.name LIKE ?")
	}
	rows, closeRows, err := r.query(qstr, conds, "m.name, i.seq", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()

	results := []metadata.Index{}
	for rows.Next() {
		rec := metadata.Index{}
		err = rows.Scan(&rec.Table, &rec.Name, &rec.IsUnique, &rec.IsPrimary)
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

func (r MetadataReader) IndexColumns(f metadata.Filter) (*metadata.IndexColumnSet, error) {
	qstr := `SELECT
  m.name,
  i.name,
  ic.name,
  ic.seqno
FROM sqlite_master m
JOIN pragma_index_list(m.name) i
JOIN pragma_index_xinfo(i.name) ic`
	conds := []string{"m.type = 'table' AND ic.cid >= 0"}
	vals := []interface{}{}
	if f.Parent != "" {
		vals = append(vals, f.Parent)
		conds = append(conds, "m.name LIKE ?")
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "i.name LIKE ?")
	}
	rows, closeRows, err := r.query(qstr, conds, "m.name, i.seq, ic.seqno", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()

	results := []metadata.IndexColumn{}
	for rows.Next() {
		rec := metadata.IndexColumn{}
		err = rows.Scan(&rec.Table, &rec.IndexName, &rec.Name, &rec.OrdinalPosition)
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

func (r MetadataReader) query(qstr string, conds []string, order string, vals ...interface{}) (*sql.Rows, func(), error) {
	if len(conds) != 0 {
		qstr += "\nWHERE " + strings.Join(conds, " AND ")
	}
	if order != "" {
		qstr += "\nORDER BY " + order
	}
	if r.limit != 0 {
		qstr += fmt.Sprintf("\nLIMIT %d", r.limit)
	}
	return r.Query(qstr, vals...)
}
