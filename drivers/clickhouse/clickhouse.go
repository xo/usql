// Package clickhouse defines and registers usql's ClickHouse driver.
//
// Group: base
// See: https://github.com/ClickHouse/clickhouse-go
package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("clickhouse", drivers.Driver{
		AllowMultilineComments: true,
		RowsAffected: func(sql.Result) (int64, error) {
			return 0, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, oldpw string) error {
			_, err := db.Exec(`ALTER USER ` + user + ` IDENTIFIED BY '` + newpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(*clickhouse.Exception); ok {
				return strconv.Itoa(int(e.Code)), strings.TrimPrefix(e.Message, "clickhouse: ")
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*clickhouse.Exception); ok {
				return e.Code == 516
			}
			return false
		},
		Copy:              CopyWithInsert,
		NewMetadataReader: NewMetadataReader,
	})
}

// CopyWithInsert builds a copy handler based on insert.
func CopyWithInsert(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
	columns, err := rows.Columns()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch source rows columns: %w", err)
	}
	clen := len(columns)
	query := table
	if !strings.HasPrefix(strings.ToLower(query), "insert into") {
		leftParen := strings.IndexRune(table, '(')
		if leftParen == -1 {
			colRows, err := db.QueryContext(ctx, "SELECT * FROM "+table+" WHERE 1=0")
			if err != nil {
				return 0, fmt.Errorf("failed to execute query to determine target table columns: %w", err)
			}
			columns, err := colRows.Columns()
			_ = colRows.Close()
			if err != nil {
				return 0, fmt.Errorf("failed to fetch target table columns: %w", err)
			}
			table += "(" + strings.Join(columns, ", ") + ")"
		}
		query = "INSERT INTO " + table + " VALUES (" + strings.Repeat("?, ", clen-1) + "?)"
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare insert query: %w", err)
	}
	defer stmt.Close()
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch source column types: %w", err)
	}
	values := make([]interface{}, clen)
	valueRefs := make([]reflect.Value, clen)
	actuals := make([]interface{}, clen)
	for i := 0; i < len(columnTypes); i++ {
		valueRefs[i] = reflect.New(columnTypes[i].ScanType())
		values[i] = valueRefs[i].Interface()
	}
	var n int64
	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return n, fmt.Errorf("failed to scan row: %w", err)
		}
		//We can't use values... in Exec() below, because, in some cases, clickhouse
		//driver doesn't accept pointer to an argument instead of the arg itself.
		for i := range values {
			actuals[i] = valueRefs[i].Elem().Interface()
		}
		res, err := stmt.ExecContext(ctx, actuals...)
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
}
