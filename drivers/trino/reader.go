package trino

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/xo/usql/drivers/metadata"
)

type metaReader struct {
	metadata.LoggingReader
}

var _ metadata.CatalogReader = &metaReader{}
var _ metadata.ColumnStatReader = &metaReader{}

func (r metaReader) Catalogs(metadata.Filter) (*metadata.CatalogSet, error) {
	qstr := `SHOW catalogs`
	rows, closeRows, err := r.Query(qstr)
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

func (r metaReader) ColumnStats(f metadata.Filter) (*metadata.ColumnStatSet, error) {
	names := []string{}
	if f.Catalog != "" {
		names = append(names, f.Catalog+".")
	}
	if f.Schema != "" {
		names = append(names, f.Schema+".")
	}
	names = append(names, f.Parent)
	rows, closeRows, err := r.Query(fmt.Sprintf("SHOW STATS FOR %s", strings.Join(names, "")))
	if err != nil {
		return nil, err
	}
	defer closeRows()

	results := []metadata.ColumnStat{}
	for rows.Next() {
		rec := metadata.ColumnStat{Catalog: f.Catalog, Schema: f.Schema, Table: f.Parent}
		name := sql.NullString{}
		avgWidth := sql.NullInt32{}
		numDistinct := sql.NullInt64{}
		nullFrac := sql.NullFloat64{}
		numRows := sql.NullInt64{}
		min := sql.NullString{}
		max := sql.NullString{}
		err = rows.Scan(
			&name,
			&avgWidth,
			&numDistinct,
			&nullFrac,
			&numRows,
			&min,
			&max,
		)
		if err != nil {
			return nil, err
		}
		if !name.Valid {
			continue
		}
		rec.Name = name.String
		if avgWidth.Valid {
			rec.AvgWidth = int(avgWidth.Int32)
		}
		if numDistinct.Valid {
			rec.NumDistinct = numDistinct.Int64
		}
		if nullFrac.Valid {
			rec.NullFrac = nullFrac.Float64
		}
		if min.Valid {
			rec.Min = min.String
		}
		if max.Valid {
			rec.Max = max.String
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return metadata.NewColumnStatSet(results), nil
}
