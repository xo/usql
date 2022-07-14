// Package postgres provides a metadata reader
package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
)

type metaReader struct {
	metadata.LoggingReader
	limit int
}

var _ metadata.CatalogReader = &metaReader{}
var _ metadata.TableReader = &metaReader{}
var _ metadata.ColumnStatReader = &metaReader{}
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
			infos.WithDataTypeFormatter(dataTypeFormatter))
		return metadata.NewPluginReader(
			newIS(db, opts...),
			&metaReader{
				LoggingReader: metadata.NewLoggingReader(db, opts...),
			},
		)
	}
}

func dataTypeFormatter(col metadata.Column) string {
	switch col.DataType {
	case "bit", "character":
		return fmt.Sprintf("%s(%d)", col.DataType, col.ColumnSize)
	case "bit varying", "character varying":
		if col.ColumnSize != 0 {
			return fmt.Sprintf("%s(%d)", col.DataType, col.ColumnSize)
		} else {
			return col.DataType
		}
	case "numeric":
		if col.ColumnSize != 0 {
			return fmt.Sprintf("numeric(%d,%d)", col.ColumnSize, col.DecimalDigits)
		} else {
			return col.DataType
		}
	case "time without time zone":
		return fmt.Sprintf("time(%d) without time zone", col.ColumnSize)
	case "time with time zone":
		return fmt.Sprintf("time(%d) with time zone", col.ColumnSize)
	case "timestamp without time zone":
		return fmt.Sprintf("timestamp(%d) without time zone", col.ColumnSize)
	case "timestamp with time zone":
		return fmt.Sprintf("timestamp(%d) with time zone", col.ColumnSize)
	default:
		return col.DataType
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

func (r metaReader) Tables(f metadata.Filter) (*metadata.TableSet, error) {
	qstr := `SELECT n.nspname as "Schema",
  c.relname as "Name",
  CASE c.relkind WHEN 'r' THEN 'table' WHEN 'v' THEN 'view' WHEN 'm' THEN 'materialized view' WHEN 'i' THEN 'index' WHEN 'S' THEN 'sequence' WHEN 's' THEN 'special' WHEN 'f' THEN 'foreign table' WHEN 'p' THEN 'partitioned table' WHEN 'I' THEN 'partitioned index' END as "Type",
  COALESCE((c.reltuples / NULLIF(c.relpages, 0)) * (pg_catalog.pg_relation_size(c.oid) / current_setting('block_size')::int), 0)::bigint as "Rows",
  pg_catalog.pg_size_pretty(pg_catalog.pg_table_size(c.oid)) as "Size",
  COALESCE(pg_catalog.obj_description(c.oid, 'pg_class'), '') as "Description"
FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
`
	conds := []string{"n.nspname !~ '^pg_toast'"}
	vals := []interface{}{}
	if f.OnlyVisible {
		conds = append(conds, "pg_catalog.pg_table_is_visible(c.oid)")
	}
	if !f.WithSystem {
		conds = append(conds, "n.nspname NOT IN ('pg_catalog', 'information_schema')")
	}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, fmt.Sprintf("n.nspname LIKE $%d", len(vals)))
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, fmt.Sprintf("c.relname LIKE $%d", len(vals)))
	}
	if len(f.Types) != 0 {
		tableTypes := map[string][]rune{
			"TABLE":             {'r', 'p', 's', 'f'},
			"VIEW":              {'v'},
			"MATERIALIZED VIEW": {'m'},
			"SEQUENCE":          {'S'},
		}
		pholders := []string{"''"}
		for _, t := range f.Types {
			for _, k := range tableTypes[t] {
				vals = append(vals, string(k))
				pholders = append(pholders, fmt.Sprintf("$%d", len(vals)))
			}
		}
		conds = append(conds, fmt.Sprintf("c.relkind IN (%s)", strings.Join(pholders, ", ")))
	}
	rows, closeRows, err := r.query(qstr, conds, "1, 2", vals...)
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
		err = rows.Scan(&rec.Schema, &rec.Name, &rec.Type, &rec.Rows, &rec.Size, &rec.Comment)
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

func (r metaReader) ColumnStats(f metadata.Filter) (*metadata.ColumnStatSet, error) {
	tables, err := r.Tables(metadata.Filter{Schema: f.Schema, Name: f.Parent, WithSystem: true})
	if err != nil {
		return nil, err
	}
	rowNum := int64(0)
	if tables.Next() {
		rowNum = tables.Get().Rows
	}

	qstr := `
SELECT
  s.schemaname,
  s.tablename,
  s.attname,
  s.avg_width,
  s.null_frac,
  CASE WHEN n_distinct >= 0 THEN n_distinct ELSE (-n_distinct * $1) END::bigint AS n_distinct,
  COALESCE((histogram_bounds::text::text[])[1], ''),
  COALESCE((histogram_bounds::text::text[])[array_length(histogram_bounds::text::text[], 1)], ''),
  most_common_vals::text::text[],
  most_common_freqs::text::text[]
FROM pg_catalog.pg_stats s
JOIN pg_catalog.pg_namespace n ON n.nspname = s.schemaname
JOIN pg_catalog.pg_class c ON c.relnamespace = n.oid AND c.relname = s.tablename
JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid AND a.attname = s.attname`
	conds := []string{}
	vals := []interface{}{rowNum}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, fmt.Sprintf("s.schemaname LIKE $%d", len(vals)))
	}
	if f.Parent != "" {
		vals = append(vals, f.Parent)
		conds = append(conds, fmt.Sprintf("s.tablename LIKE $%d", len(vals)))
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, fmt.Sprintf("s.attname LIKE $%d", len(vals)))
	}
	rows, closeRows, err := r.query(qstr, conds, "a.attnum", vals...)
	if err != nil {
		return nil, err
	}
	defer closeRows()

	results := []metadata.ColumnStat{}
	for rows.Next() {
		rec := metadata.ColumnStat{}
		err = rows.Scan(
			&rec.Schema,
			&rec.Table,
			&rec.Name,
			&rec.AvgWidth,
			&rec.NullFrac,
			&rec.NumDistinct,
			&rec.Min,
			&rec.Max,
			pq.Array(&rec.TopN),
			pq.Array(&rec.TopNFreqs),
		)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewColumnStatSet(results), nil
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
		"n.nspname !~ '^pg_toast'",
	}
	if f.OnlyVisible {
		conds = append(conds, "pg_catalog.pg_table_is_visible(c.oid)")
	}
	vals := []interface{}{}
	if !f.WithSystem {
		conds = append(conds, "n.nspname NOT IN ('pg_catalog', 'information_schema')")
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
