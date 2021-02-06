package mssql

import (
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

type metaReader struct {
	db drivers.DB
}

func (r metaReader) Catalogs() (*metadata.CatalogSet, error) {
	qstr := `
SELECT name
FROM sys.databases
ORDER BY name
`

	rows, err := r.db.Query(qstr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (r metaReader) Indexes(catalog, schemaPattern, tablePattern, namePattern string) (*metadata.IndexSet, error) {
	qstr := `
SELECT
  db_name(),
  s.name,
  t.name,
  COALESCE(i.name, ''),
  CASE WHEN i.is_primary_key = 1 THEN 'YES' ELSE 'NO' END,
  CASE WHEN i.is_unique = 1 THEN 'YES' ELSE 'NO' END,
  i.type_desc
FROM sys.schemas s
JOIN sys.tables t on t.schema_id = s.schema_id
JOIN sys.indexes i ON i.object_id = t.object_id
`
	conds := []string{}
	vals := []interface{}{}
	if schemaPattern != "" {
		vals = append(vals, schemaPattern)
		conds = append(conds, "s.name LIKE ?")
	}
	if tablePattern != "" {
		vals = append(vals, tablePattern)
		conds = append(conds, "t.name LIKE ?")
	}
	if namePattern != "" {
		vals = append(vals, namePattern)
		conds = append(conds, "i.name LIKE ?")
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY s.name, t.name, i.name`

	rows, err := r.db.Query(qstr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []metadata.Index{}
	for rows.Next() {
		rec := metadata.Index{}
		err = rows.Scan(&rec.Catalog, &rec.Schema, &rec.Table, &rec.Name, &rec.IsUnique, &rec.IsPrimary, &rec.Type)
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
	qstr := `
SELECT
  db_name(),
  s.name,
  t.name,
  COALESCE(i.name, ''),
  c.name,
  t.name,
  ic.key_ordinal
FROM sys.schemas s
JOIN sys.tables t on t.schema_id = s.schema_id
JOIN sys.indexes i ON i.object_id = t.object_id
JOIN sys.index_columns ic ON i.object_id = ic.object_id and i.index_id = ic.index_id
JOIN sys.columns c ON ic.object_id = c.object_id and ic.column_id = c.column_id
JOIN sys.types ty ON ty.user_type_id = c.user_type_id
`
	conds := []string{}
	vals := []interface{}{}
	if schemaPattern != "" {
		vals = append(vals, schemaPattern)
		conds = append(conds, "s.name LIKE ?")
	}
	if tablePattern != "" {
		vals = append(vals, tablePattern)
		conds = append(conds, "t.name LIKE ?")
	}
	if indexPattern != "" {
		vals = append(vals, indexPattern)
		conds = append(conds, "i.name LIKE ?")
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY s.name, t.name, i.name, ic.index_column_id`
	rows, err := r.db.Query(qstr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []metadata.IndexColumn{}
	for rows.Next() {
		rec := metadata.IndexColumn{}
		err = rows.Scan(&rec.Catalog, &rec.Schema, &rec.Table, &rec.IndexName, &rec.Name, &rec.DataType, &rec.OrdinalPosition)
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
