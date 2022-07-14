package sqlserver

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

func NewReader(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
	ir := infos.New(
		infos.WithPlaceholder(placeholder),
		infos.WithIndexes(false),
		infos.WithSequences(false),
		infos.WithConstraints(false),
		infos.WithCustomClauses(map[infos.ClauseName]string{
			infos.FunctionsSecurityType: "''",
		}),
		infos.WithSystemSchemas([]string{
			"db_accessadmin",
			"db_backupoperator",
			"db_datareader",
			"db_datawriter",
			"db_ddladmin",
			"db_denydatareader",
			"db_denydatawriter",
			"db_owner",
			"db_securityadmin",
			"INFORMATION_SCHEMA",
			"sys",
		}),
		infos.WithCurrentSchema("schema_name()"),
		infos.WithDataTypeFormatter(dataTypeFormatter),
	)(db, opts...)
	mr := &metaReader{
		LoggingReader: metadata.NewLoggingReader(db, opts...),
	}
	return metadata.NewPluginReader(ir, mr)
}

func dataTypeFormatter(col metadata.Column) string {
	switch col.DataType {
	case "numeric", "decimal":
		if col.ColumnSize == 18 && col.DecimalDigits == 0 {
			return col.DataType
		} else {
			return fmt.Sprintf("%s(%d,%d)", col.DataType, col.ColumnSize, col.DecimalDigits)
		}
	case "datetimeoffset", "datetime2", "time":
		if col.ColumnSize == 7 {
			return col.DataType
		} else {
			return fmt.Sprintf("%s(%d)", col.DataType, col.ColumnSize)
		}
	case "char", "nchar", "binary":
		if col.ColumnSize == 1 {
			return col.DataType
		} else {
			return fmt.Sprintf("%s(%d)", col.DataType, col.ColumnSize)
		}
	case "varchar", "nvarchar", "varbinary":
		if col.ColumnSize == -1 {
			return col.DataType + "(max)"
		} else if col.ColumnSize == 1 {
			return col.DataType
		} else {
			return fmt.Sprintf("%s(%d)", col.DataType, col.ColumnSize)
		}
	default:
		return col.DataType
	}
}

func (r *metaReader) SetLimit(l int) {
	r.limit = l
}

func (r metaReader) Catalogs(metadata.Filter) (*metadata.CatalogSet, error) {
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

func (r metaReader) Indexes(f metadata.Filter) (*metadata.IndexSet, error) {
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
	if f.OnlyVisible {
		conds = append(conds, "s.name = schema_name()")
	}
	if !f.WithSystem {
		conds = append(conds, "s.name NOT IN ('db_accessadmin', 'db_backupoperator', 'db_datareader', 'db_datawriter', 'db_ddladmin', 'db_denydatareader', 'db_denydatawriter', 'db_owner', 'db_securityadmin', 'INFORMATION_SCHEMA', 'sys')")
	}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, fmt.Sprintf("s.name LIKE @p%d", len(vals)))
	}
	if f.Parent != "" {
		vals = append(vals, f.Parent)
		conds = append(conds, fmt.Sprintf("t.name LIKE @p%d", len(vals)))
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, fmt.Sprintf("i.name LIKE @p%d", len(vals)))
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

func (r metaReader) IndexColumns(f metadata.Filter) (*metadata.IndexColumnSet, error) {
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
	if f.OnlyVisible {
		conds = append(conds, "s.name = schema_name()")
	}
	if !f.WithSystem {
		conds = append(conds, "s.name NOT IN ('db_accessadmin', 'db_backupoperator', 'db_datareader', 'db_datawriter', 'db_ddladmin', 'db_denydatareader', 'db_denydatawriter', 'db_owner', 'db_securityadmin', 'INFORMATION_SCHEMA', 'sys')")
	}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, fmt.Sprintf("s.name LIKE @p%d", len(vals)))
	}
	if f.Parent != "" {
		vals = append(vals, f.Parent)
		conds = append(conds, fmt.Sprintf("t.name LIKE @p%d", len(vals)))
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, fmt.Sprintf("i.name LIKE @p%d", len(vals)))
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
