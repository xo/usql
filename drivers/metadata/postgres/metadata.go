// Package postgres provides a metadata reader
package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
)

type metaReader struct {
	metadata.LoggingReader
	limit int
}

var _ metadata.CatalogReader = &metaReader{}
var _ metadata.IndexReader = &metaReader{}
var _ metadata.IndexColumnReader = &metaReader{}
var _ metadata.TriggerReader = &metaReader{}

func NewReader() func(drivers.DB, ...metadata.ReaderOption) metadata.Reader {
	return func(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
		newIS := infos.New(
			infos.WithIndexes(false),
			infos.WithCustomClauses(map[infos.ClauseName]string{
				infos.ColumnsColumnSize:         "COALESCE(character_maximum_length, numeric_precision, datetime_precision, interval_precision, 0)",
				infos.FunctionColumnsColumnSize: "COALESCE(character_maximum_length, numeric_precision, datetime_precision, interval_precision, 0)",
			}),
			infos.WithSystemSchemas([]string{"pg_catalog", "pg_toast", "information_schema"}),
			infos.WithCurrentSchema("CURRENT_SCHEMA"),
		)
		return metadata.NewPluginReader(
			newIS(db, opts...),
			&metaReader{
				LoggingReader: metadata.NewLoggingReader(db, opts...),
			},
		)
	}
}

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
	conds := []string{
		"c.relkind IN ('i','I','')",
		"n.nspname <> 'pg_catalog'",
		"n.nspname <> 'information_schema'",
		"n.nspname !~ '^pg_toast'",
	}
	if f.OnlyVisible {
		conds = append(conds, "pg_catalog.pg_table_is_visible(c.oid)")
	}
	vals := []interface{}{}
	if !f.WithSystem {
		conds = append(conds, "n.nspname NOT IN ('pg_catalog', 'pg_toast', 'information_schema')")
	}
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

func (r metaReader) IndexColumns(f metadata.Filter) (*metadata.IndexColumnSet, error) {
	qstr := `
SELECT
  'postgres' as "Catalog",
  n.nspname as "Schema",
  c2.relname as "Table",
  c.relname as "IndexName",
  a.attname AS "Name",
  pg_catalog.format_type(a.atttypid, a.atttypmod) AS "DataType",
  a.attnum AS "OrdinalPosition"
FROM pg_catalog.pg_class c
     JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
     JOIN pg_catalog.pg_index i ON i.indexrelid = c.oid
     JOIN pg_catalog.pg_class c2 ON i.indrelid = c2.oid
     JOIN pg_catalog.pg_attribute a ON c.oid = a.attrelid
`
	conds := []string{
		"c.relkind IN ('i','I','')",
		"n.nspname <> 'pg_catalog'",
		"n.nspname <> 'information_schema'",
		"n.nspname !~ '^pg_toast'",
		"a.attnum > 0",
		"NOT a.attisdropped",
	}
	if f.OnlyVisible {
		conds = append(conds, "pg_catalog.pg_table_is_visible(c.oid)")
	}
	vals := []interface{}{}
	if !f.WithSystem {
		conds = append(conds, "n.nspname NOT IN ('pg_catalog', 'pg_toast', 'information_schema')")
	}
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
	rows, closeRows, err := r.query(qstr, conds, "1, 2, 3, 4, 7", vals...)
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

func (r metaReader) Triggers(f metadata.Filter) (*metadata.TriggerSet, error) {
	qstr := `SELECT
	n.nspname,
	c.relname,
    t.tgname, 
    pg_catalog.pg_get_triggerdef(t.oid, true)
FROM 
    pg_catalog.pg_trigger t 
    JOIN pg_catalog.pg_class c ON c.oid = t.tgrelid
	LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace`
	conds := []string{`(
	NOT t.tgisinternal OR (t.tgisinternal AND t.tgenabled = 'D') 
			OR 
				EXISTS (SELECT 1 FROM pg_catalog.pg_depend WHERE objid = t.oid 
			AND 
				refclassid = 'pg_catalog.pg_trigger'::pg_catalog.regclass)
	)`}
	vals := []interface{}{}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, fmt.Sprintf("n.nspname LIKE $%d", len(vals)))
	}
	if f.Parent != "" {
		vals = append(vals, f.Parent)
		conds = append(conds, fmt.Sprintf("c.relname LIKE $%d", len(vals)))
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, fmt.Sprintf("t.tgname LIKE $%d", len(vals)))
	}
	rows, closeRows, err := r.query(qstr, conds, "t.tgname", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()

	results := []metadata.Trigger{}
	for rows.Next() {
		rec := metadata.Trigger{}
		err = rows.Scan(
			&rec.Schema,
			&rec.Table,
			&rec.Name,
			&rec.Definition,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewTriggerSet(results), nil
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
