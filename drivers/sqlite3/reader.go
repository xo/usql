package sqlite3

import (
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

type metaReader struct {
	db drivers.DB
}

// Columns from selected catalog (or all, if empty), matching schemas and tables
func (r metaReader) Columns(catalog, schemaPattern, tablePattern string) (*metadata.ColumnSet, error) {
	tables, err := r.Tables(catalog, schemaPattern, tablePattern, []string{})
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
		rows, err := r.db.Query(qstr, table.Name)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

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

func (r metaReader) Tables(catalog, schemaPattern, namePattern string, types []string) (*metadata.TableSet, error) {
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
)
`
	conds := []string{}
	vals := []interface{}{}
	if catalog != "" {
		vals = append(vals, catalog)
		conds = append(conds, "table_catalog = ?")
	}
	if schemaPattern != "" {
		vals = append(vals, schemaPattern)
		conds = append(conds, "table_schema LIKE ?")
	}
	if namePattern != "" {
		vals = append(vals, namePattern)
		conds = append(conds, "table_name LIKE ?")
	}
	if len(types) != 0 {
		pholders := []string{}
		for _, t := range types {
			vals = append(vals, t)
			pholders = append(pholders, "?")
		}
		if len(pholders) != 0 {
			conds = append(conds, "table_type IN ("+strings.Join(pholders, ", ")+")")
		}
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY table_type, table_name`
	rows, err := r.db.Query(qstr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (r metaReader) Schemas(catalog, namePattern string) (*metadata.SchemaSet, error) {
	qstr := `SELECT
  name AS schema_name,
  '' AS catalog_name
FROM pragma_database_list
`
	conds := []string{}
	vals := []interface{}{}
	if namePattern != "" {
		vals = append(vals, namePattern)
		conds = append(conds, "schema_name LIKE ?")
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY seq`
	rows, err := r.db.Query(qstr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (r metaReader) Functions(catalog, schemaPattern, namePattern string, types []string) (*metadata.FunctionSet, error) {
	qstr := `SELECT
  name AS specific_name,
  name AS routine_name,
  type AS routine_type
FROM pragma_function_list
`
	conds := []string{}
	vals := []interface{}{}
	if namePattern != "" {
		vals = append(vals, namePattern)
		conds = append(conds, "name LIKE ?")
	}
	if len(types) != 0 {
		pholders := []string{}
		for _, t := range types {
			vals = append(vals, t)
			pholders = append(pholders, "?")
		}
		if len(pholders) != 0 {
			conds = append(conds, "type IN ("+strings.Join(pholders, ", ")+")")
		}
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY name, type`
	rows, err := r.db.Query(qstr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (r metaReader) FunctionColumns(catalog, schemaPattern, functionPattern string) (*metadata.FunctionColumnSet, error) {
	return &metadata.FunctionColumnSet{}, nil
}

func (r metaReader) Indexes(catalog, schemaPattern, tablePattern, namePattern string) (*metadata.IndexSet, error) {
	qstr := `SELECT
  m.name,
  i.name,
  CASE WHEN i."unique" = 1 THEN 'YES' ELSE 'NO' END,
  CASE WHEN i.origin = 'pk' THEN 'YES' ELSE 'NO' END
FROM sqlite_master m
JOIN pragma_index_list(m.name) i
`
	conds := []string{"m.type = 'table'"}
	vals := []interface{}{}
	if tablePattern != "" {
		vals = append(vals, tablePattern)
		conds = append(conds, "m.name LIKE ?")
	}
	if namePattern != "" {
		vals = append(vals, namePattern)
		conds = append(conds, "i.name LIKE ?")
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY m.name, i.seq`

	rows, err := r.db.Query(qstr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (r metaReader) IndexColumns(catalog, schemaPattern, tablePattern, indexPattern string) (*metadata.IndexColumnSet, error) {
	qstr := `SELECT
  m.name,
  i.name,
  ic.name,
  ic.seqno
FROM sqlite_master m
JOIN pragma_index_list(m.name) i
JOIN pragma_index_xinfo(i.name) ic
`
	conds := []string{"m.type = 'table' AND ic.cid >= 0"}
	vals := []interface{}{}
	if tablePattern != "" {
		vals = append(vals, tablePattern)
		conds = append(conds, "m.name LIKE ?")
	}
	if indexPattern != "" {
		vals = append(vals, indexPattern)
		conds = append(conds, "i.name LIKE ?")
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY m.name, i.seq, ic.seqno`

	rows, err := r.db.Query(qstr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
