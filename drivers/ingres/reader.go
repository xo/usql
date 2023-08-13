package ingres

import (
	"database/sql"
	"fmt"
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
  table_name AS Name,
  (case
    when table_type = 'T' then 'Table'
    when table_type = 'P' then 'Partition'
    when table_type = 'V' then 'View'
    when table_type = 'I' then 'Index'
    else ''
  end) as TableType,
  (case
    when table_subtype = 'N' then 'Native'
    when table_subtype = 'I' then 'Gateway'
    when table_subtype = 'L' then 'Link (STAR)'
    else ''
  end) as TableSubtype,
  storage_structure as Storage,
  (table_pagesize * number_pages) AS Size,
  num_rows AS Rows,
  coalesce(c.long_remark, '') as Comment
FROM iitables t
LEFT JOIN iidbms_comment c
ON t.table_reltid = c.comtabbase and t.table_reltidx = c.comtabidx
`
	var conds []string
	var vals []interface{}

	vals = append(vals, "I")
	conds = append(conds, "table_type != ~V ")

	if f.Name != "" {
		vals = append(vals, f.Name, f.Name)
		conds = append(conds, "table_name = ~V or table_name LIKE ~V ")
	}
	if len(f.Types) != 0 {
		tableTypes := map[string][]rune{
			"TABLE":             {'T', 'P'},
			"VIEW":              {'V'},
			"MATERIALIZED VIEW": {'V'},
			"SEQUENCE":          {'S'},
		}
		pholders := []string{"''"}
		for _, t := range f.Types {
			for _, k := range tableTypes[t] {
				val := string(k)
				vals = append(vals, val)
				pholders = append(pholders, " ~V ")
			}
		}
		conds = append(conds, fmt.Sprintf("table_type IN (%s)", strings.Join(pholders, ", ")))
	}
	rows, closeRows, err := r.query(qstr, conds, "Name", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()
	var results []metadata.Table
	for rows.Next() {
		var tableType, subType, storage string

		var rec metadata.Table
		if err := rows.Scan(&rec.Name,
			&tableType,
			&subType,
			&storage,
			&rec.Size,
			&rec.Rows,
			&rec.Comment); err != nil {
			return nil, err
		}

		var parts []string
		if subType != "Native" {
			parts = append(parts, subType)
		}
		parts = append(parts, storage, tableType)
		rec.Type = strings.Join(parts, " ")

		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	set := metadata.NewTableSet(results)

	return set, nil
}

func (r MetadataReader) Columns(f metadata.Filter) (*metadata.ColumnSet, error) {
	qstr := `SELECT
  column_sequence as Position,
  column_name as Name,
  column_datatype as Type,
  int(column_length) as ColumnSize,
  coalesce(column_default_val, ''),
  column_nulls as Nullable,
  column_internal_length as OctetLength,
  column_scale as Scale
FROM
  iicolumns`
	vals := []interface{}{f.Parent}
	conds := []string{"table_name = ~V "}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "column_name = ~V ")
	}
	if len(f.Types) != 0 {
		var pholders []string
		for _, t := range f.Types {
			vals = append(vals, t)
			pholders = append(pholders, " ~V ")
		}
		if len(pholders) != 0 {
			conds = append(conds, "column_datatype IN ("+strings.Join(pholders, ", ")+")")
		}
	}
	rows, closeRows, err := r.query(qstr, conds, "column_sequence", vals...)
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
			&rec.Name,
			&rec.DataType,
			&rec.ColumnSize,
			&rec.Default,
			&rec.IsNullable,
			&rec.CharOctetLength,
			&rec.DecimalDigits,
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
	qstr := `SELECT schema_name FROM iischema`

	var conds []string
	var vals []interface{}
	if f.Name != "" {
		vals = append(vals, f.Name, f.Name)
		conds = append(conds, "schema_name = ~V OR schema_name LIKE ~V ")
	}
	rows, closeRows, err := r.query(qstr, conds, "schema_name", vals...)
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
  procedure_name AS name
FROM
  iiprocedures`
	var conds []string
	var vals []interface{}
	if f.Name != "" {
		vals = append(vals, f.Name, f.Name)
		conds = append(conds, "procedure_name = ~V or procedure_name LIKE ~V ")
	}
	if len(f.Types) != 0 {
		var pholders []string
		for _, t := range f.Types {
			vals = append(vals, t)
			pholders = append(pholders, " ~V ")
		}
		if len(pholders) != 0 {
			conds = append(conds, "proc_subtype IN ("+strings.Join(pholders, ", ")+")")
		}
	}
	rows, closeRows, err := r.query(qstr, conds, "procedure_name", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()
	var results []metadata.Function
	for rows.Next() {
		var rec metadata.Function
		if err := rows.Scan(
			&rec.Name,
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

func (r MetadataReader) Indexes(f metadata.Filter) (*metadata.IndexSet, error) {
	qstr := `SELECT
    t.table_name as Name,
    storage_structure as Type,
    r.relid as Table,
    case when c.constraint_type = 'P' then 'Y' else 'N' end as IsPrimary,
    case when t.unique_rule = 'U' then 'Y' else 'N' end as IsUnique
FROM iitables t
JOIN iirelation r ON reltid=t.table_reltid and reltidx = 0 and table_type = 'I'
LEFT JOIN iiconstraints c ON t.table_name = c.constraint_name
`

	var conds []string
	var vals []interface{}
	if f.Parent != "" {
		vals = append(vals, f.Parent)
		conds = append(conds, "r.relid = ~V ")
	}
	if f.Name != "" {
		vals = append(vals, f.Name, f.Name)
		conds = append(conds, "t.table_name = ~V OR t.table_name LIKE ~V ")
	}
	rows, closeRows, err := r.query(qstr, conds, "t.table_name", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()
	var results []metadata.Index
	for rows.Next() {
		var rec metadata.Index
		if err := rows.Scan(&rec.Name, &rec.Type,
			&rec.Table, &rec.IsPrimary, &rec.IsUnique); err != nil {
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
  rm.relid as Table,
  ri.relid as IndexName,
  attname as Name,
  uppercase(iitypename(ii_ext_type(a.attfrmt, a.attfrml))) as DataType
FROM iiattribute a
LEFT JOIN iirelation rm ON rm.reltid = attrelid and rm.reltidx = 0
LEFT JOIN iirelation ri ON ri.reltid = attrelid and ri.reltidx = attrelidx`
	vals := []interface{}{f.Parent}
	conds := []string{"varchar(rm.relid) = ~V "}

	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, "ri.relid = ~V ")
	}

	rows, closeRows, err := r.query(qstr, conds, "attid", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()
	var results []metadata.IndexColumn
	for rows.Next() {
		rec := metadata.IndexColumn{}
		if err := rows.Scan(
			&rec.Table,
			&rec.IndexName,
			&rec.Name,
			&rec.DataType,
		); err != nil {
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
	return r.Query(qstr, vals...)
}
