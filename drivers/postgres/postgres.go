// Package postgres defines and registers usql's PostgreSQL driver.
//
// See: https://github.com/lib/pq
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/lib/pq" // DRIVER: postgres
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	pgmeta "github.com/xo/usql/drivers/metadata/postgres"
)

func init() {
	drivers.Register("postgres", drivers.Driver{
		Name:                   "pq",
		AllowDollar:            true,
		AllowMultilineComments: true,
		LexerName:              "postgres",
		ForceParams: func(u *dburl.URL) {
			if u.Scheme == "cockroachdb" {
				drivers.ForceQueryParameters([]string{"sslmode", "disable"})(u)
			}
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			// numeric version
			// SHOW server_version_num;
			var ver string
			err := db.QueryRowContext(ctx, `SHOW server_version`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "PostgreSQL " + ver, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, _ string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` PASSWORD '` + newpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(*pq.Error); ok {
				return string(e.Code), e.Message
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*pq.Error); ok {
				return e.Code.Name() == "invalid_password"
			}
			return false
		},
		NewMetadataReader: pgmeta.NewReader(),
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(pgmeta.NewReader()(db, opts...))(db, w)
		},
		Copy: func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
			columns, err := rows.Columns()
			if err != nil {
				return 0, fmt.Errorf("failed to fetch source rows columns: %w", err)
			}
			clen := len(columns)

			query := table
			if !strings.HasPrefix(strings.ToLower(query), "insert into") {
				leftParen := strings.IndexRune(table, '(')
				colQuery := "SELECT * FROM " + table + " WHERE 1=0"
				if leftParen != -1 {
					colQuery = "SELECT " + table[leftParen+1:len(table)-1] + " FROM " + table[:leftParen] + " WHERE 1=0"
					table = table[:leftParen]
				}
				colStmt, err := db.PrepareContext(ctx, colQuery)
				if err != nil {
					return 0, fmt.Errorf("failed to prepare query to determine target table columns: %w", err)
				}
				colRows, err := colStmt.QueryContext(ctx)
				if err != nil {
					return 0, fmt.Errorf("failed to execute query to determine target table columns: %w", err)
				}
				columns, err := colRows.Columns()
				if err != nil {
					return 0, fmt.Errorf("failed to fetch target table columns: %w", err)
				}
				query = pq.CopyIn(table, columns...)
			}
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return 0, fmt.Errorf("failed to begin transaction: %w", err)
			}
			stmt, err := tx.PrepareContext(ctx, query)
			if err != nil {
				return 0, fmt.Errorf("failed to prepare insert query: %w", err)
			}

			values := make([]interface{}, clen)
			for i := 0; i < clen; i++ {
				values[i] = new(interface{})
			}

			var n int64
			for rows.Next() {
				err = rows.Scan(values...)
				if err != nil {
					return n, fmt.Errorf("failed to scan row: %w", err)
				}
				res, err := stmt.ExecContext(ctx, values...)
				if err != nil {
					return n, fmt.Errorf("failed to exec insert: %w", err)
				}
				rn, err := res.RowsAffected()
				if err != nil {
					return n, fmt.Errorf("failed to check rows affected: %w", err)
				}
				n += rn
			}

			err = tx.Commit()
			if err != nil {
				return n, fmt.Errorf("failed to commit transaction: %w", err)
			}

			return n, rows.Err()
		},
	}, "cockroachdb", "redshift")
}
