package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// CopyTarget splits the target of a native copy, which is a table such as
// "schema.table" or "table(col, col)", into its schema, its name and the names of
// the columns that the copy fills. A table with no list of columns gets all of its
// columns, and a list of expressions is evaluated against the table, because
// splitting it at the commas is not reliable. The schema is empty for a name that
// has no dot. The names have no quotes, and the case of a letter is kept, so the
// caller quotes them for its driver.
func CopyTarget(ctx context.Context, db *sql.DB, table string) (schema, name string, columns []string, err error) {
	query := "SELECT * FROM " + table + " WHERE 1=0"
	if i := strings.IndexRune(table, '('); i != -1 {
		query = "SELECT " + table[i+1:len(table)-1] + " FROM " + table[:i] + " WHERE 1=0"
		table = table[:i]
	}
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to execute query to determine target table columns: %w", err)
	}
	defer rows.Close()
	if columns, err = rows.Columns(); err != nil {
		return "", "", nil, fmt.Errorf("failed to fetch target table columns: %w", err)
	}
	parts := splitName(strings.TrimSpace(table))
	if len(parts) < 2 {
		return "", strings.Join(parts, ""), columns, nil
	}
	return parts[0], strings.Join(parts[1:], "."), columns, nil
}

// splitName splits a name at the dots that are outside double quotes, and
// takes the quotes off each part.
func splitName(s string) []string {
	var parts []string
	var b strings.Builder
	var quoted bool
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"' && quoted && i+1 < len(s) && s[i+1] == '"':
			b.WriteByte('"')
			i++
		case c == '"':
			quoted = !quoted
		case c == '.' && !quoted:
			parts = append(parts, b.String())
			b.Reset()
		default:
			b.WriteByte(c)
		}
	}
	return append(parts, b.String())
}

// CopyRows adapts the rows of a query to the source of a native copy: it scans
// each row into values that the copy writes. It satisfies pgx's CopyFromSource.
type CopyRows struct {
	rows   *sql.Rows
	values []any
}

// NewCopyRows returns a CopyRows for rows, which have n columns.
func NewCopyRows(rows *sql.Rows, n int) *CopyRows {
	r := &CopyRows{rows: rows, values: make([]any, n)}
	for i := range r.values {
		r.values[i] = new(any)
	}
	return r
}

// Next advances to the next row.
func (r *CopyRows) Next() bool {
	return r.rows.Next()
}

// Values returns the values of the current row.
func (r *CopyRows) Values() ([]any, error) {
	if err := r.rows.Scan(r.values...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}
	actuals := make([]any, len(r.values))
	for i, v := range r.values {
		actuals[i] = *(v.(*any))
	}
	return actuals, nil
}

// Err returns the error of the rows.
func (r *CopyRows) Err() error {
	return r.rows.Err()
}
