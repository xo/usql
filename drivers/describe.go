package drivers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/xo/dbmeta"
	"github.com/xo/dburl"
)

// ErrDescribeNotSupported is the error when a database cannot describe the
// result of a query without running it.
var ErrDescribeNotSupported = errors.New("describing a query without running it is not supported")

// ResultType is the type of one result column as a driver reports it, by the
// number of the type, as PostgreSQL's protocol gives it.
type ResultType struct {
	// Name is the name of the column.
	Name string
	// OID is the number of the type.
	OID uint32
	// Modifier is the modifier of the type, or -1.
	Modifier int32
}

// ResultColumn is one column of the result of a query: its name, and the SQL
// name of its type, such as character varying(10).
type ResultColumn struct {
	Name string
	Type string
}

// Describe returns the columns of the result of query without running it, for
// \gdesc. db is where it runs, and conn is the connection of the session, for a
// driver that describes the query through its own connection.
//
// A database whose dialect has a statement that describes a query, which is
// SQL Server, runs that statement. A driver that describes a query itself,
// which is pgx, gives the numbers of the types, and dbmeta names them. Any
// other database returns ErrDescribeNotSupported, because it can describe a
// query only by running it.
func Describe(ctx context.Context, u *dburl.URL, db, conn DB, query string) ([]ResultColumn, error) {
	dialect := dbmeta.Dialect(u.Dialect)
	if desc, err := dialect.DescribeResult(query); err == nil {
		return describeByStatement(ctx, db, desc)
	}
	d, ok := drivers[u.SchemeName]
	if !ok || d.Describe == nil || conn == nil {
		return nil, ErrDescribeNotSupported
	}
	types, err := d.Describe(ctx, conn, query)
	if err != nil {
		return nil, err
	}
	cols := make([]ResultColumn, len(types))
	refs := make([]dbmeta.TypeRef, len(types))
	for i, t := range types {
		cols[i] = ResultColumn{Name: t.Name, Type: strconv.FormatUint(uint64(t.OID), 10)}
		refs[i] = dbmeta.TypeRef{OID: t.OID, Modifier: t.Modifier}
	}
	if len(refs) == 0 {
		return cols, nil
	}
	// the numbers of the types are what is left when the dialect cannot
	// name them
	stmt, err := dialect.TypeNames(refs)
	if err != nil {
		return cols, nil
	}
	rows, err := db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ord int
		var name string
		if err := rows.Scan(&ord, &name); err != nil {
			return nil, err
		}
		if 0 < ord && ord <= len(cols) {
			cols[ord-1].Type = name
		}
	}
	return cols, rows.Err()
}

// describeByStatement runs the statement of a dialect that describes a query,
// and reads the name and the type of each column from its rows.
func describeByStatement(ctx context.Context, db DB, desc dbmeta.ResultDescription) ([]ResultColumn, error) {
	rows, err := db.QueryContext(ctx, desc.Stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	names, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	nameAt, typeAt := -1, -1
	for i, n := range names {
		switch n {
		case desc.Name:
			nameAt = i
		case desc.Type:
			typeAt = i
		}
	}
	if nameAt == -1 || typeAt == -1 {
		return nil, fmt.Errorf("the statement that describes the query has no columns %s and %s", desc.Name, desc.Type)
	}
	var cols []ResultColumn
	for rows.Next() {
		vals := make([]any, len(names))
		for i := range vals {
			vals[i] = new(sql.NullString)
		}
		if err := rows.Scan(vals...); err != nil {
			return nil, err
		}
		cols = append(cols, ResultColumn{
			Name: vals[nameAt].(*sql.NullString).String,
			Type: vals[typeAt].(*sql.NullString).String,
		})
	}
	return cols, rows.Err()
}
