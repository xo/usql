package clickhouse

import (
	"database/sql"
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

type MetadataReader struct {
	metadata.LoggingReader
}

// NewMetadataReader creates the metadata reader for clickhouse databases.
func NewMetadataReader(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
	return &MetadataReader{
		LoggingReader: metadata.NewLoggingReader(db, opts...),
	}
}

func (r MetadataReader) Tables(f metadata.Filter) (*metadata.TableSet, error) {
	qstr := `SELECT
  database AS Schema,
  name AS Name,
  COALESCE(
    IF(database LIKE 'system', 'SYSTEM TABLE', null),
    IF(is_temporary,'LOCAL TEMPORARY', null),
    IF(engine LIKE 'View', 'VIEW', null),
    'TABLE'
  ) AS Type,
  COALESCE(total_bytes, 0) AS Size,
  comment as Comment
FROM
  system.tables`
	var conds []string
	var vals []interface{}
	if !f.WithSystem {
		conds = append(conds, "database NOT LIKE 'system'")
	}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, "database LIKE ?")
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "name LIKE ?")
	}
	if len(f.Types) != 0 {
		var pholders []string
		for _, t := range f.Types {
			vals = append(vals, t)
			pholders = append(pholders, "?")
		}
		if len(pholders) != 0 {
			conds = append(conds, "Type IN ("+strings.Join(pholders, ", ")+")")
		}
	}
	rows, closeRows, err := r.query(qstr, conds, "Schema, Name", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()
	var results []metadata.Table
	for rows.Next() {
		var rec metadata.Table
		if err := rows.Scan(&rec.Schema, &rec.Name, &rec.Type, &rec.Size, &rec.Comment); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewTableSet(results), nil
}

func (r MetadataReader) Columns(f metadata.Filter) (*metadata.ColumnSet, error) {
	qstr := `SELECT
  position,
  database as schema,
  name,
  type,
  COALESCE(default_expression, '')
FROM
  system.columns`
	vals := []interface{}{f.Parent}
	conds := []string{"table LIKE ?"}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, "database LIKE ?")
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "name LIKE ?")
	}
	if len(f.Types) != 0 {
		var pholders []string
		for _, t := range f.Types {
			vals = append(vals, t)
			pholders = append(pholders, "?")
		}
		if len(pholders) != 0 {
			conds = append(conds, "Type IN ("+strings.Join(pholders, ", ")+")")
		}
	}
	rows, closeRows, err := r.query(qstr, conds, "name", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()
	var results []metadata.Column
	for rows.Next() {
		rec := metadata.Column{
			Catalog: f.Catalog,
			Table:   f.Parent,
		}
		if err := rows.Scan(
			&rec.OrdinalPosition,
			&rec.Schema,
			&rec.Name,
			&rec.DataType,
			&rec.Default,
		); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewColumnSet(results), nil
}

func (r MetadataReader) Schemas(f metadata.Filter) (*metadata.SchemaSet, error) {
	qstr := `SELECT
  name
FROM
  system.databases`
	var conds []string
	var vals []interface{}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "name LIKE ?")
	}
	rows, closeRows, err := r.query(qstr, conds, "name", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()
	var results []metadata.Schema
	for rows.Next() {
		var rec metadata.Schema
		if err := rows.Scan(&rec.Schema); err != nil {
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
  (IF(is_aggregate = 1,'AGGREGATE','FUNCTION')) AS type
FROM
  system.functions`
	var conds []string
	var vals []interface{}
	if f.Name != "" {
		conds = append(conds, "name LIKE ?")
		vals = append(vals, f.Name)
	}
	if len(f.Types) != 0 {
		var pholders []string
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
	var results []metadata.Function
	for rows.Next() {
		var rec metadata.Function
		if err := rows.Scan(
			&rec.SpecificName,
			&rec.Name,
			&rec.Type,
		); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewFunctionSet(results), nil
}

func (r MetadataReader) query(qstr string, conds []string, order string, vals ...interface{}) (*sql.Rows, func(), error) {
	if len(conds) != 0 {
		qstr += "\nWHERE " + strings.Join(conds, " AND ")
	}
	if order != "" {
		qstr += "\nORDER BY " + order
	}
	return r.Query(qstr, vals...)
}
