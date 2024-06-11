// Package duckdb defines and registers usql's DuckDB driver. Requires CGO.
//
// See: https://github.com/marcboeker/go-duckdb
package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	_ "github.com/marcboeker/go-duckdb" // DRIVER
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
	mymeta "github.com/xo/usql/drivers/metadata/mysql"
)

type metaReader struct {
	metadata.LoggingReader
}

var (
	_ metadata.CatalogReader    = &metaReader{}
	_ metadata.ColumnStatReader = &metaReader{}
)

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

func init() {
	newReader := func(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
		ir := infos.New(
			infos.WithPlaceholder(func(int) string { return "?" }),
			infos.WithCustomClauses(map[infos.ClauseName]string{
				infos.ColumnsColumnSize:       "0",
				infos.ColumnsNumericScale:     "0",
				infos.ColumnsNumericPrecRadix: "0",
				infos.ColumnsCharOctetLength:  "0",
			}),
			infos.WithFunctions(false),
			infos.WithSequences(false),
			infos.WithIndexes(false),
			infos.WithConstraints(false),
			infos.WithColumnPrivileges(false),
			infos.WithUsagePrivileges(false),
		)(db, opts...)
		mr := &metaReader{
			LoggingReader: metadata.NewLoggingReader(db, opts...),
		}
		return metadata.NewPluginReader(ir, mr)
	}
	drivers.Register("duckdb", drivers.Driver{
		AllowMultilineComments: true,
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRowContext(ctx, `SELECT library_version FROM pragma_version()`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "DuckDB " + ver, nil
		},
		NewMetadataReader: newReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(newReader(db, opts...))(db, w)
		},
		Copy:         drivers.CopyWithInsert(func(int) string { return "?" }),
		NewCompleter: mymeta.NewCompleter,
	})
}
