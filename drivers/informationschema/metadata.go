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
FROM information_schema.columns`
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
FROM information_schema.tables`
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
	if len(types) != 0 {
		pholders := []string{}
		for _, t := range types {
			vals = append(vals, t)
			pholders = append(pholders, fmt.Sprintf("$%d", len(vals)))
		}
		conds = append(conds, "table_type IN ("+strings.Join(pholders, ", ")+")")
	}
	if len(conds) != 0 {
		qstr += " WHERE " + strings.Join(conds, " AND ")
	}
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
	qstr := "SELECT catalog_name, schema_name FROM information_schema.schemata"
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
