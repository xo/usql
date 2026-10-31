package describe

import (
	"io"

	"github.com/xo/tblfmt"
	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

// rows is a table of values that tblfmt renders. Each command builds one per
// section it prints.
type rows struct {
	columns []string
	values  [][]any
	current int
}

// add appends a row.
func (r *rows) add(v ...any) {
	r.values = append(r.values, v)
}

// Next satisfies tblfmt.ResultSet.
func (r *rows) Next() bool {
	r.current++
	return r.current <= len(r.values)
}

// Scan satisfies tblfmt.ResultSet.
func (r *rows) Scan(dest ...any) error {
	v := r.values[r.current-1]
	if len(v) != len(dest) {
		return text.ErrWrongNumberOfArguments
	}
	for i, d := range dest {
		*d.(*any) = v[i]
	}
	return nil
}

// Columns satisfies tblfmt.ResultSet.
func (r *rows) Columns() ([]string, error) {
	return r.columns, nil
}

// Close satisfies tblfmt.ResultSet.
func (r *rows) Close() error {
	return nil
}

// Err satisfies tblfmt.ResultSet.
func (r *rows) Err() error {
	return nil
}

// NextResultSet satisfies tblfmt.ResultSet.
func (r *rows) NextResultSet() bool {
	return false
}

// Table writes values under columns and title to w, in the output format of
// the session, as the lists of the describe commands are written.
func Table(w io.Writer, title string, columns []string, values [][]any) error {
	return encode(w, &rows{columns: columns, values: values}, title, nil, nil)
}

// noFooter prints nothing below a table, where psql prints no count of rows,
// as for each section of \d NAME.
func noFooter(io.Writer) error {
	return nil
}

// encode renders r to w under title, in the output format of the session. A
// non-nil footer replaces the count of rows with the lines that psql prints
// below the table, such as the indexes of a table.
func encode(w io.Writer, r *rows, title string, footer func(io.Writer) error, params map[string]string) error {
	p := env.Vars().Print()
	for k, v := range params {
		p[k] = v
	}
	p["title"] = title
	newEnc, opts := tblfmt.FromMap(p)
	if footer != nil {
		opts = append(opts, tblfmt.WithSummary(map[int]func(io.Writer, int) (int, error){
			-1: func(w io.Writer, _ int) (int, error) {
				return 0, footer(w)
			},
		}))
	}
	enc, err := newEnc(r, opts...)
	if err != nil {
		return err
	}
	return enc.EncodeAll(w)
}
