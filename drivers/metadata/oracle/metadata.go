// Package oracle provides a metadata reader
package oracle

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

type metaReader struct {
	metadata.LoggingReader
	systemSchemas string
}

var _ metadata.BasicReader = &metaReader{}
var _ metadata.IndexReader = &metaReader{}
var _ metadata.IndexColumnReader = &metaReader{}

func NewReader() func(drivers.DB, ...metadata.ReaderOption) metadata.Reader {
	return func(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
		r := &metaReader{
			LoggingReader: metadata.NewLoggingReader(db, opts...),
			systemSchemas: "'CTXSYS', 'FLOWS_FILES', 'MDSYS', 'OUTLN', 'SYS', 'SYSTEM', 'XDB', 'XS$NULL'",
		}
		return r
	}
}

func (r metaReader) Catalogs(metadata.Filter) (*metadata.CatalogSet, error) {
	qstr := `SELECT
  UPPER(Value) AS catalog
FROM v$parameter o
WHERE name = 'db_name'
UNION ALL
SELECT
  db_link AS catalog
FROM dba_db_links
ORDER BY catalog
`

	rows, closeRows, err := r.Query(qstr)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewCatalogSet([]metadata.Catalog{}), nil
		}
		return nil, err
	}
	defer closeRows()

	results := []metadata.Catalog{}
	for rows.Next() {
		rec := metadata.Catalog{}
		err = rows.Scan(&rec.Catalog)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewCatalogSet(results), nil
}

func (r metaReader) Schemas(f metadata.Filter) (*metadata.SchemaSet, error) {
	qstr := `SELECT
  username
FROM all_users
`
	conds, vals := r.conditions(f, formats{
		name:       "username LIKE :%d",
		notSchemas: "username NOT IN (%s)",
	})
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY username`
	rows, closeRows, err := r.Query(qstr, vals...)
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
		err = rows.Scan(&rec.Schema)
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

// Tables from selected catalog (or all, if empty), matching schemas, names and types
func (r metaReader) Tables(f metadata.Filter) (*metadata.TableSet, error) {
	qstr := `SELECT
o.owner AS table_schem,
o.object_name AS table_name,
o.object_type AS table_type
FROM all_objects o
`
	conds, vals := r.conditions(f, formats{
		schema:     "o.owner LIKE :%d",
		notSchemas: "o.owner NOT IN (%s)",
		name:       "o.object_name LIKE :%d",
		types:      "o.object_type IN (%s)",
	})
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	addSynonyms := false
	for _, t := range f.Types {
		if t == "SYNONYM" {
			addSynonyms = true
		}
	}
	if addSynonyms {
		qstr += `
UNION ALL
SELECT
  s.owner AS table_schem,
  s.synonym_name AS table_name,
  'SYNONYM' AS table_type
FROM all_synonyms s
`
		conds, seqVals := r.conditions(f, formats{
			schema:     "s.owner LIKE :%d",
			notSchemas: "s.owner NOT IN (%s)",
			name:       "s.synonym_name LIKE :%d",
		})
		vals = append(vals, seqVals...)
		if len(conds) != 0 {
			qstr += " WHERE " + strings.Join(conds, " AND ")
		}
	}
	qstr += `
ORDER BY table_schem, table_name, table_type`
	rows, closeRows, err := r.Query(qstr, vals...)
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
		err = rows.Scan(&rec.Schema, &rec.Name, &rec.Type)
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

func (r metaReader) Columns(f metadata.Filter) (*metadata.ColumnSet, error) {
	qstr := `SELECT
  c.owner,
  c.table_name,
  c.column_name,
  c.column_id AS ordinal_position,
  c.data_type,
  CASE c.nullable
    WHEN 'Y' THEN 'YES'
    ELSE  'NO'  END AS nullable,
  COALESCE(c.data_length, c.data_precision, 0),
  COALESCE(c.data_scale, 0),
  CASE c.data_type
           WHEN 'FLOAT'  THEN  2
           WHEN 'NUMBER' THEN 10
  ELSE  0  END AS num_prec_radix,
  COALESCE(c.char_col_decl_length, 0) as char_octet_length
FROM all_tab_columns c
`
	conds, vals := r.conditions(f, formats{
		schema:     "c.owner LIKE :%d",
		notSchemas: "c.owner NOT IN (%s)",
		parent:     "c.table_name LIKE :%d",
	})
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY c.owner, c.table_name, c.column_id`
	rows, closeRows, err := r.Query(qstr, vals...)
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
		targets := []interface{}{
			&rec.Schema,
			&rec.Table,
			&rec.Name,
			&rec.OrdinalPosition,
			&rec.DataType,
			&rec.IsNullable,
			&rec.ColumnSize,
			&rec.DecimalDigits,
			&rec.NumPrecRadix,
			&rec.CharOctetLength,
		}
		err = rows.Scan(targets...)
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

func (r metaReader) Functions(f metadata.Filter) (*metadata.FunctionSet, error) {
	qstr := `SELECT
  decode (b.object_type,'PACKAGE',CONCAT(CONCAT(b.object_name,'.'), a.object_name)
         ,b.object_name) as specific_name,
  b.owner   as procedure_schem,
  decode (b.object_type,'PACKAGE',CONCAT(CONCAT(b.object_name,'.'), a.object_name)
         ,b.object_name) as procedure_name,
  decode (b.object_type,'PACKAGE',decode(a.position,0,2,1,1,0),
          decode(b.object_type,'PROCEDURE',1,'FUNCTION',2,0)) as procedure_type
FROM all_arguments a
JOIN all_objects b ON b.object_id = a.object_id AND a.sequence  = 1
`
	conds, vals := r.conditions(f, formats{
		schema:     "b.owner LIKE :%d",
		notSchemas: "b.owner NOT IN (%s)",
		name:       "b.object_name LIKE :%d",
		types:      "b.object_type IN (%s)",
	})
	conds = append(conds, "(b.object_type = 'PROCEDURE' OR b.object_type = 'FUNCTION' OR b.object_type = 'PACKAGE')")
	qstr += " WHERE " + strings.Join(conds, " AND ")
	qstr += `
ORDER BY procedure_schem, procedure_name, procedure_type`
	rows, closeRows, err := r.Query(qstr, vals...)
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
			&rec.Schema,
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

func (r metaReader) FunctionColumns(f metadata.Filter) (*metadata.FunctionColumnSet, error) {
	qstr := `SELECT
     a.owner   as procedure_schem,
     decode (b.object_type,'PACKAGE',CONCAT(CONCAT(b.object_name,'.'),a.object_name),
             b.object_name) as procedure_name,
     decode(a.position,0,'RETURN_VALUE',a.argument_name) as column_name,
     a.position       as ordinal_position,
     decode(a.position,0,5,decode(a.in_out,'IN',1,'IN/OUT',2,'OUT',4)) as column_type,
     a.data_type      as type_name,
     COALESCE(a.data_length, a.data_precision, 0) as column_size,
     COALESCE(a.data_scale, 0) as decimal_digits,
     COALESCE(a.radix, 0) as num_prec_radix
FROM all_objects b
JOIN all_arguments a ON b.object_id = a.object_id AND a.data_level = 0
`
	conds, vals := r.conditions(f, formats{
		schema:     "a.owner LIKE :%d",
		notSchemas: "a.owner NOT IN (%s)",
		parent:     "b.object_name LIKE :%d",
	})
	conds = append(conds, "b.object_type = 'PROCEDURE' OR b.object_type = 'FUNCTION'")
	qstr += " WHERE " + strings.Join(conds, " AND ")
	qstr += `
ORDER BY procedure_schem, procedure_name, ordinal_position`
	rows, closeRows, err := r.Query(qstr, vals...)
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
			&rec.Schema,
			&rec.FunctionName,
			&rec.Name,
			&rec.OrdinalPosition,
			&rec.Type,
			&rec.DataType,
			&rec.ColumnSize,
			&rec.DecimalDigits,
			&rec.NumPrecRadix,
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

func (r metaReader) Indexes(f metadata.Filter) (*metadata.IndexSet, error) {
	qstr := `SELECT
  o.owner,
  o.table_name,
  o.index_name,
  decode(o.uniqueness,'UNIQUE','NO','YES')
FROM all_indexes o
`
	conds, vals := r.conditions(f, formats{
		schema:     "o.owner LIKE :%d",
		notSchemas: "o.owner NOT IN (%s)",
		parent:     "o.table_name LIKE :%d",
		name:       "o.index_name LIKE :%d",
	})
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY o.owner, o.table_name, o.index_name`

	rows, closeRows, err := r.Query(qstr, vals...)
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
		err = rows.Scan(&rec.Schema, &rec.Table, &rec.Name, &rec.IsUnique)
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

func (r metaReader) IndexColumns(f metadata.Filter) (*metadata.IndexColumnSet, error) {
	qstr := `SELECT
  o.owner,
  o.table_name,
  o.index_name,
  b.column_name,
  b.column_position
FROM all_indexes o
JOIN all_ind_columns b ON o.owner = b.index_owner AND o.index_name = b.index_name
`
	conds, vals := r.conditions(f, formats{
		schema:     "o.owner LIKE :%d",
		notSchemas: "o.owner NOT IN (%s)",
		parent:     "o.table_name LIKE :%d",
		name:       "o.index_name LIKE :%d",
	})
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY o.owner, o.table_name, o.index_name, b.column_position`
	rows, closeRows, err := r.Query(qstr, vals...)
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
		err = rows.Scan(&rec.Schema, &rec.Table, &rec.IndexName, &rec.Name, &rec.OrdinalPosition)
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

func (r metaReader) conditions(filter metadata.Filter, formats formats) ([]string, []interface{}) {
	baseParam := 1
	conds := []string{}
	vals := []interface{}{}
	if filter.Schema != "" && formats.schema != "" {
		vals = append(vals, filter.Schema)
		conds = append(conds, fmt.Sprintf(formats.schema, baseParam))
		baseParam++
	}

	if !filter.WithSystem && formats.notSchemas != "" {
		conds = append(conds, fmt.Sprintf(formats.notSchemas, r.systemSchemas))
	}
	if filter.Parent != "" && formats.parent != "" {
		vals = append(vals, filter.Parent)
		conds = append(conds, fmt.Sprintf(formats.parent, baseParam))
		baseParam++
	}
	if filter.Name != "" && formats.name != "" {
		vals = append(vals, filter.Name)
		conds = append(conds, fmt.Sprintf(formats.name, baseParam))
		baseParam++
	}
	if len(filter.Types) != 0 && formats.types != "" {
		pholders := []string{}
		for _, t := range filter.Types {
			vals = append(vals, t)
			pholders = append(pholders, fmt.Sprintf(":%d", baseParam))
			baseParam++
		}
		if len(pholders) != 0 {
			conds = append(conds, fmt.Sprintf(formats.types, strings.Join(pholders, ", ")))
		}
	}

	return conds, vals
}

type formats struct {
	schema     string
	notSchemas string
	parent     string
	name       string
	types      string
}
