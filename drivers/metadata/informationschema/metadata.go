package informationschema

import (
	"fmt"
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

type InformationSchema struct {
	db drivers.DB
}

var _ metadata.Reader = &InformationSchema{}

func New(db drivers.DB) metadata.Reader {
	// TODO add options to make it work with dbs other than PostgreSQL, like MySQL (? as placeholders), Presto/Trino, MSSQL, ClickHouse
	return &InformationSchema{
		db: db,
	}
}

func (s InformationSchema) Columns(catalog, schema, table string) (*metadata.ColumnSet, error) {
	qstr := `SELECT
  table_catalog,
  table_schema,
  table_name,
  column_name,
  ordinal_position,
  data_type,
  COALESCE(column_default, ''),
  COALESCE(character_maximum_length, numeric_precision, datetime_precision, interval_precision, 0) AS column_size,
  COALESCE(numeric_scale, 0),
  COALESCE(numeric_precision_radix, 0),
  COALESCE(character_octet_length, 0),
  COALESCE(is_nullable, '') AS is_nullable,
  COALESCE(is_generated, '') AS is_generated,
  COALESCE(is_identity, '') AS is_identity
FROM information_schema.columns
`
	conds := []string{}
	vals := []interface{}{}
	if catalog != "" {
		vals = append(vals, catalog)
		conds = append(conds, fmt.Sprintf("table_catalog = $%d", len(vals)))
	}
	if schema != "" {
		vals = append(vals, schema)
		conds = append(conds, fmt.Sprintf("table_schema LIKE $%d", len(vals)))
	}
	if table != "" {
		vals = append(vals, table)
		conds = append(conds, fmt.Sprintf("table_name LIKE $%d", len(vals)))
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	qstr += `
ORDER BY table_catalog, table_schema, table_name, ordinal_position`
	rows, err := s.db.Query(qstr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []metadata.Column{}
	for rows.Next() {
		rec := metadata.Column{}
		err = rows.Scan(
			&rec.Catalog,
			&rec.Schema,
			&rec.Table,
			&rec.Name,
			&rec.OrdinalPosition,
			&rec.DataType,
			&rec.ColumnDefault,
			&rec.ColumnSize,
			&rec.DecimalDigits,
			&rec.NumPrecRadix,
			&rec.CharOctetLength,
			&rec.IsNullable,
			&rec.IsGenerated,
			&rec.IsIdentity,
		)
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

func (s InformationSchema) Tables(catalog, schemaPattern, tableNamePattern string, types []string) (*metadata.TableSet, error) {
	qstr := `SELECT
  table_catalog,
  table_schema,
  table_name,
  table_type
FROM information_schema.tables
`
	conds := []string{}
	vals := []interface{}{}
	if catalog != "" {
		vals = append(vals, catalog)
		conds = append(conds, fmt.Sprintf("table_catalog = $%d", len(vals)))
	}
	if schemaPattern != "" {
		vals = append(vals, schemaPattern)
		conds = append(conds, fmt.Sprintf("table_schema LIKE $%d", len(vals)))
	}
	if tableNamePattern != "" {
		vals = append(vals, tableNamePattern)
		conds = append(conds, fmt.Sprintf("table_name LIKE $%d", len(vals)))
	}
	addSequences := false
	if len(types) != 0 {
		pholders := []string{}
		for _, t := range types {
			if t == "SEQUENCE" {
				addSequences = true
				continue
			}
			vals = append(vals, t)
			pholders = append(pholders, fmt.Sprintf("$%d", len(vals)))
		}
		if len(pholders) != 0 {
			conds = append(conds, "table_type IN ("+strings.Join(pholders, ", ")+")")
		}
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
	if addSequences {
		qstr += `
UNION ALL
SELECT
  sequence_catalog AS table_catalog,
  sequence_schema AS table_schema,
  sequence_name AS table_name,
  'SEQUENCE' AS table_type
FROM information_schema.sequences
`
		conds = []string{}
		if catalog != "" {
			vals = append(vals, catalog)
			conds = append(conds, fmt.Sprintf("sequence_catalog = $%d", len(vals)))
		}
		if schemaPattern != "" {
			vals = append(vals, schemaPattern)
			conds = append(conds, fmt.Sprintf("sequence_schema LIKE $%d", len(vals)))
		}
		if tableNamePattern != "" {
			vals = append(vals, tableNamePattern)
			conds = append(conds, fmt.Sprintf("sequence_name LIKE $%d", len(vals)))
		}
		if len(conds) != 0 {
			qstr += " WHERE " + strings.Join(conds, " AND ")
		}
	}
	qstr += `
ORDER BY table_catalog, table_schema, table_type, table_name`
	rows, err := s.db.Query(qstr, vals...)
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

func (s InformationSchema) Schemas() (*metadata.SchemaSet, error) {
	qstr := `SELECT
  schema_name,
  catalog_name
FROM information_schema.schemata
ORDER BY catalog_name, schema_name`
	rows, err := s.db.Query(qstr)
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
