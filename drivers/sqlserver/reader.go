package sqlserver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/xo/usql/drivers/metadata"
)

type metaReader struct {
	metadata.LoggingReader
	limit int
}

func (r *metaReader) SetLimit(l int) {
	r.limit = l
}

func (r metaReader) Catalogs() (*metadata.CatalogSet, error) {
	qstr := `SELECT name
FROM sys.databases`
	rows, closeRows, err := r.query(qstr, []string{}, "name")
	if err != nil {
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
	rows, closeRows, err := r.query(qstr, conds, "s.name, t.name, i.name", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()

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
	rows, closeRows, err := r.query(qstr, conds, "s.name, t.name, i.name, ic.index_column_id", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()

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

func (r metaReader) query(qstr string, conds []string, order string, vals ...interface{}) (*sql.Rows, func(), error) {
	if len(conds) != 0 {
		qstr += "\nWHERE " + strings.Join(conds, " AND ")
	}
	if order != "" {
		qstr += "\nORDER BY " + order
	}
	if r.limit != 0 {
		qstr += fmt.Sprintf("\nFETCH FIRST %d ROWS ONLY", r.limit)
	}
	return r.Query(qstr, vals...)
}
