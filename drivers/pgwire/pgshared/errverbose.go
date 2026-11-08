package pgshared

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// ErrorFields are the fields of an error that a PostgreSQL server sends. The
// pgx and lib/pq drivers each have their own type, and fill this one in.
type ErrorFields struct {
	Severity         string
	Code             string
	Message          string
	Detail           string
	Hint             string
	Position         int
	InternalPosition int
	InternalQuery    string
	Where            string
	SchemaName       string
	TableName        string
	ColumnName       string
	DataTypeName     string
	ConstraintName   string
	File             string
	Line             int
	Routine          string
}

// Verbose returns the error as psql's \errverbose writes it, which is libpq's
// verbose form: the severity, the SQLSTATE and the message, the line of query
// that the position points at, and each other field that the server sent.
func Verbose(e ErrorFields, query string) string {
	var b strings.Builder
	b.WriteString(e.Severity + ":  ")
	if e.Code != "" {
		b.WriteString(e.Code + ": ")
	}
	b.WriteString(e.Message)
	if e.Position > 0 {
		b.WriteString(errorPosition(query, e.Position))
	}
	b.WriteByte('\n')
	field := func(name, value string) {
		if value != "" {
			b.WriteString(name + ":  " + value + "\n")
		}
	}
	field("DETAIL", e.Detail)
	field("HINT", e.Hint)
	if e.InternalQuery != "" {
		b.WriteString("QUERY:  " + e.InternalQuery)
		if e.InternalPosition > 0 {
			b.WriteString(errorPosition(e.InternalQuery, e.InternalPosition))
		}
		b.WriteByte('\n')
	}
	field("CONTEXT", e.Where)
	field("SCHEMA NAME", e.SchemaName)
	field("TABLE NAME", e.TableName)
	field("COLUMN NAME", e.ColumnName)
	field("DATATYPE NAME", e.DataTypeName)
	field("CONSTRAINT NAME", e.ConstraintName)
	var location string
	if e.File != "" {
		location = e.File
		if e.Line > 0 {
			location += ":" + strconv.Itoa(e.Line)
		}
		if e.Routine != "" {
			location = e.Routine + ", " + location
		}
	}
	field("LOCATION", location)
	return strings.TrimSuffix(b.String(), "\n")
}

// errorPosition returns the line of query that holds the character at
// position, which counts from 1, and a caret under that character, each on
// its own line, as libpq writes them after the message.
func errorPosition(query string, position int) string {
	lineNo, start := 1, 0
	n := 0
	for i := range query {
		n++
		if n == position {
			end := strings.IndexByte(query[i:], '\n')
			if end == -1 {
				end = len(query)
			} else {
				end += i
			}
			line := strings.TrimSuffix(query[start:end], "\r")
			prefix := "LINE " + strconv.Itoa(lineNo) + ": "
			col := utf8.RuneCountInString(query[start:i])
			return "\n" + prefix + line + "\n" + strings.Repeat(" ", len(prefix)+col) + "^"
		}
		if query[i] == '\n' {
			lineNo, start = lineNo+1, i+1
		}
	}
	return ""
}
