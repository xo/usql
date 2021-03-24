package postgres

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

var _ metadata.CatalogReader = &metaReader{}
var _ metadata.IndexReader = &metaReader{}

func (r *metaReader) SetLimit(l int) {
	r.limit = l
}

func (r metaReader) Catalogs(metadata.Filter) (*metadata.CatalogSet, error) {
	qstr := `
SELECT d.datname as "Name"
FROM pg_catalog.pg_database d`
	rows, closeRows, err := r.query(qstr, []string{}, "1")
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

func (r metaReader) Indexes(f metadata.Filter) (*metadata.IndexSet, error) {
	qstr := `
SELECT
  'postgres' as "Catalog",
  n.nspname as "Schema",
  c2.relname as "Table",
  c.relname as "Name",
  CASE i.indisprimary WHEN TRUE THEN 'YES' ELSE 'NO' END,
  CASE i.indisunique WHEN TRUE THEN 'YES' ELSE 'NO' END,
  CASE c.relkind WHEN 'r' THEN 'table' WHEN 'v' THEN 'view' WHEN 'm' THEN 'materialized view' WHEN 'i' THEN 'index' WHEN 'S' THEN 'sequence' WHEN 's' THEN 'special' WHEN 'f' THEN 'foreign table' WHEN 'p' THEN 'partitioned table' WHEN 'I' THEN 'partitioned index' END as "Type"
FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
     LEFT JOIN pg_catalog.pg_index i ON i.indexrelid = c.oid
     LEFT JOIN pg_catalog.pg_class c2 ON i.indrelid = c2.oid`
	conds := []string{"c.relkind IN ('i','I','')",
		"n.nspname <> 'pg_catalog'",
		"n.nspname <> 'information_schema'",
		"n.nspname !~ '^pg_toast'",
		"pg_catalog.pg_table_is_visible(c.oid)"}
	vals := []interface{}{}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, fmt.Sprintf("n.nspname LIKE $%d", len(vals)))
	}
	if f.Parent != "" {
		vals = append(vals, f.Parent)
		conds = append(conds, fmt.Sprintf("c2.relname LIKE $%d", len(vals)))
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, fmt.Sprintf("c.relname LIKE $%d", len(vals)))
	}
	rows, closeRows, err := r.query(qstr, conds, "1, 2", vals...)
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

func (r metaReader) query(qstr string, conds []string, order string, vals ...interface{}) (*sql.Rows, func(), error) {
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
