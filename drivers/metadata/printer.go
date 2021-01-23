package metadata

import (
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/xo/tblfmt"
	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

// DB is the common interface for database operations, compatible with
// database/sql.DB and database/sql.Tx.
type DB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	Prepare(string) (*sql.Stmt, error)
}

// DefaultWriter using an existing db introspector
type DefaultWriter struct {
	r        Reader
	db       DB
	w        io.Writer
	typesMap map[rune][]string
}

var _ Writer = &DefaultWriter{}

func NewDefaultWriter(r Reader) func(db DB, w io.Writer) Writer {
	return func(db DB, w io.Writer) Writer {
		return &DefaultWriter{
			r:  r,
			db: db,
			w:  w,
			typesMap: map[rune][]string{
				't': {"TABLE", "BASE TABLE"},
				'v': {"VIEW"},
			},
		}
	}
}

// DescribeAggregates matching pattern
func (w DefaultWriter) DescribeAggregates(pattern string, verbose, showSystem bool) error {
	return fmt.Errorf(text.NotSupportedByDriver, `\da`)
}

// DescribeFunctions matching pattern
func (w DefaultWriter) DescribeFunctions(funcTypes, pattern string, verbose, showSystem bool) error {
	return fmt.Errorf(text.NotSupportedByDriver, `\df`)
}

// DescribeTableDetails matching pattern
func (w DefaultWriter) DescribeTableDetails(pattern string, verbose, showSystem bool) error {
	sp, tp, err := parsePattern(pattern)
	if err != nil {
		return err
	}
	// TODO also describe: views, indexes, sequences
	r, err := w.r.Columns("", sp, tp)
	if err != nil {
		return err
	}
	defer r.Close()

	r.SetVerbose(verbose)
	params := env.All()
	params["format"] = "aligned"
	return tblfmt.EncodeAll(w.w, r, params)
}

// ListAllDbs matching pattern
func (w DefaultWriter) ListAllDbs(pattern string, verbose bool) error {
	return fmt.Errorf(text.NotSupportedByDriver, `\l`)
}

// ListTables matching pattern
func (w DefaultWriter) ListTables(tableTypes, pattern string, verbose, showSystem bool) error {
	types := []string{}
	for k, v := range w.typesMap {
		if strings.ContainsRune(tableTypes, k) {
			types = append(types, v...)
		}
	}
	sp, tp, err := parsePattern(pattern)
	if err != nil {
		return err
	}
	r, err := w.r.Tables("", sp, tp, types)
	if err != nil {
		return err
	}
	defer r.Close()

	r.SetVerbose(verbose)
	params := env.All()
	params["format"] = "aligned"
	return tblfmt.EncodeAll(w.w, r, params)
}

// ListSchemas matching pattern
func (w DefaultWriter) ListSchemas(pattern string, verbose, showSystem bool) error {
	r, err := w.r.Schemas()
	if err != nil {
		return err
	}
	// TODO do pattern matching locally
	defer r.Close()

	r.SetVerbose(verbose)
	params := env.All()
	params["format"] = "aligned"
	return tblfmt.EncodeAll(w.w, r, params)
}

func parsePattern(pattern string) (string, string, error) {
	// TODO do proper escaping, quoting etc
	if strings.ContainsRune(pattern, '.') {
		parts := strings.SplitN(pattern, ".", 2)
		return parts[0], parts[1], nil
	}
	return "", pattern, nil
}
