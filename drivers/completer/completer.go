// completer package provides a generic SQL command line completer
package completer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/gohxs/readline"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

const (
	WORD_BREAKS = "\t\n$><=;|&{() "
)

type caseType bool

var (
	IGNORE_CASE            = caseType(true)
	MATCH_CASE             = caseType(false)
	CommonSqlStartCommands = []string{
		"ABORT",
		"ALTER",
		"ANALYZE",
		"BEGIN",
		"CALL",
		"CHECKPOINT",
		"CLOSE",
		"CLUSTER",
		"COMMENT",
		"COMMIT",
		"COPY",
		"CREATE",
		"DEALLOCATE",
		"DECLARE",
		"DELETE FROM",
		"DESC",
		"DESCRIBE",
		"DISCARD",
		"DO",
		"DROP",
		"END",
		"EXEC",
		"EXECUTE",
		"EXPLAIN",
		"FETCH",
		"GRANT",
		"IMPORT",
		"INSERT",
		"LIST",
		"LISTEN",
		"LOAD",
		"LOCK",
		"MOVE",
		"NOTIFY",
		"PRAGMA",
		"PREPARE",
		"REASSIGN",
		"REFRESH MATERIALIZED VIEW",
		"REINDEX",
		"RELEASE",
		"RESET",
		"REVOKE",
		"ROLLBACK",
		"SAVEPOINT",
		"SECURITY LABEL",
		"SELECT",
		"SET",
		"SHOW",
		"START",
		"TABLE",
		"TRUNCATE",
		"UNLISTEN",
		"UPDATE",
		"VACUUM",
		"VALUES",
		"WITH",
	}
	CommonSqlCommands = []string{
		"AND",
		"CASE",
		"CROSS JOIN",
		"ELSE",
		"END",
		"FETCH",
		"FROM",
		"FULL OUTER JOIN",
		"GROUP BY",
		"HAVING",
		"IN",
		"INNER JOIN",
		"IS NOT NULL",
		"IS NULL",
		"JOIN",
		"LEFT JOIN",
		"LIMIT",
		"NOT",
		"ON",
		"OR",
		"ORDER BY",
		"THEN",
		"WHEN",
		"WHERE",
	}
)

func NewDefaultCompleter(opts ...Option) readline.AutoCompleter {
	c := completer{
		// an empty struct satisfies the metadata.Reader interface, because it is actually empty
		reader:           struct{}{},
		logger:           log.New(os.Stdout, "ERROR: ", log.LstdFlags),
		sqlStartCommands: CommonSqlStartCommands,
		// TODO do we need to add built-in functions like, COALESCE, CAST, NULLIF, CONCAT etc?
		sqlCommands: CommonSqlCommands,
		backslashCommands: []string{
			`\!`,
			`\?`,
			`\a`,
			`\begin`,
			`\c`,
			`\connect`,
			`\C`,
			`\cd`,
			`\commit`,
			`\conninfo`,
			`\copyright`,
			`\copy`,
			`\d+`,
			`\da+`,
			`\da`,
			`\daS+`,
			`\daS`,
			`\df+`,
			`\df`,
			`\dfS+`,
			`\dfS`,
			`\di+`,
			`\di`,
			`\diS+`,
			`\diS`,
			`\dm+`,
			`\dm`,
			`\dmS+`,
			`\dmS`,
			`\dn+`,
			`\dn`,
			`\dnS+`,
			`\dnS`,
			`\drivers`,
			`\ds+`,
			`\ds`,
			`\dS+`,
			`\dS`,
			`\dsS+`,
			`\dsS`,
			`\dt+`,
			`\dt`,
			`\dtS+`,
			`\dtS`,
			`\dv+`,
			`\dv`,
			`\dvS+`,
			`\dvS`,
			`\e`,
			`\echo`,
			`\f`,
			`\g`,
			`\gexec`,
			`\gset`,
			`\gx`,
			`\H`,
			`\i`,
			`\ir`,
			`\l+`,
			`\l`,
			`\p`,
			`\password`,
			`\prompt`,
			`\pset`,
			`\q`,
			`\r`,
			`\raw`,
			`\rollback`,
			`\set`,
			`\setenv`,
			`\t`,
			`\T`,
			`\timing`,
			`\unset`,
			`\w`,
			`\watch`,
			`\x`,
			`\Z`,
		},
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

// Option to configure the reader
type Option func(*completer)

// WithDB option
func WithDB(db metadata.DB) Option {
	return func(c *completer) {
		c.db = db
	}
}

// WithReader option
func WithReader(r metadata.Reader) Option {
	return func(c *completer) {
		c.reader = r
	}
}

// WithLogger option
func WithLogger(l logger) Option {
	return func(c *completer) {
		c.logger = l
	}
}

// WithSQLStartCommands that can begin a query
func WithSQLStartCommands(commands []string) Option {
	return func(c *completer) {
		c.sqlStartCommands = commands
	}
}

// WithSQLCommands that can be any part of a query
func WithSQLCommands(commands []string) Option {
	return func(c *completer) {
		c.sqlCommands = commands
	}
}

// WithConnStrings option
func WithConnStrings(connStrings []string) Option {
	return func(c *completer) {
		c.connStrings = connStrings
	}
}

// WithBeforeComplete option
func WithBeforeComplete(f CompleteFunc) Option {
	return func(c *completer) {
		c.beforeComplete = f
	}
}

// completer based on https://github.com/postgres/postgres/blob/9f3665fbfc34b963933e51778c7feaa8134ac885/src/bin/psql/tab-complete.c
type completer struct {
	db                metadata.DB
	reader            metadata.Reader
	logger            logger
	sqlStartCommands  []string
	sqlCommands       []string
	backslashCommands []string
	connStrings       []string
	beforeComplete    CompleteFunc
}

// CompleteFunc returns patterns completing current text, using previous words as context
type CompleteFunc func(previousWords []string, text []rune) [][]rune

type logger interface {
	Println(...interface{})
}

func (c completer) Do(line []rune, start int) (newLine [][]rune, length int) {
	var i int
	for i = start - 1; i > 0; i-- {
		if strings.ContainsRune(WORD_BREAKS, line[i]) {
			i++
			break
		}
	}
	if i == -1 {
		i = 0
	}
	previousWords := getPreviousWords(start, line)
	text := line[i:start]

	if c.beforeComplete != nil {
		result := c.beforeComplete(previousWords, text)
		if result != nil {
			return result, len(text)
		}
	}
	result := c.complete(previousWords, text)
	if result != nil {
		return result, len(text)
	}
	return nil, 0
}

func (c completer) complete(previousWords []string, text []rune) [][]rune {
	if len(text) > 0 {
		if len(previousWords) == 0 && text[0] == '\\' {
			/* If current word is a backslash command, offer completions for that */
			return CompleteFromListCase(MATCH_CASE, text, c.backslashCommands...)
		}
		if text[0] == ':' {
			if len(text) == 1 || text[1] == ':' {
				return nil
			}
			/* If current word is a variable interpolation, handle that case */
			if text[1] == '\'' {
				return completeFromVariables(text, ":'", "'", true)
			}
			if text[1] == '"' {
				return completeFromVariables(text, ":\"", "\"", true)
			}
			return completeFromVariables(text, ":", "", true)
		}
	}
	if len(previousWords) == 0 {
		/* If no previous word, suggest one of the basic sql commands */
		return CompleteFromList(text, c.sqlStartCommands...)
	}
	/* DELETE --- can be inside EXPLAIN, RULE, etc */
	/* ... despite which, only complete DELETE with FROM at start of line */
	if matches(IGNORE_CASE, previousWords, "DELETE") {
		return CompleteFromList(text, "FROM")
	}
	/* Complete DELETE FROM with a list of tables */
	if TailMatches(IGNORE_CASE, previousWords, "DELETE", "FROM") {
		return c.completeWithUpdatables(text)
	}
	/* Complete DELETE FROM <table> */
	if TailMatches(IGNORE_CASE, previousWords, "DELETE", "FROM", "*") {
		return CompleteFromList(text, "USING", "WHERE")
	}
	/* XXX: implement tab completion for DELETE ... USING */

	/* Complete CREATE */
	if TailMatches(IGNORE_CASE, previousWords, "CREATE") {
		return CompleteFromList(text, "DATABASE", "SEQUENCE", "TABLE", "VIEW", "TEMPORARY")
	}
	if TailMatches(IGNORE_CASE, previousWords, "CREATE", "TEMP|TEMPORARY") {
		return CompleteFromList(text, "TABLE", "VIEW")
	}
	if TailMatches(IGNORE_CASE, previousWords, "CREATE", "TABLE", "*") || TailMatches(IGNORE_CASE, previousWords, "CREATE", "TEMP|TEMPORARY", "TABLE", "*") {
		return CompleteFromList(text, "(")
	}
	/* INSERT --- can be inside EXPLAIN, RULE, etc */
	/* Complete INSERT with "INTO" */
	if TailMatches(IGNORE_CASE, previousWords, "INSERT") {
		return CompleteFromList(text, "INTO")
	}
	/* Complete INSERT INTO with table names */
	if TailMatches(IGNORE_CASE, previousWords, "INSERT", "INTO") {
		return c.completeWithUpdatables(text)
	}
	/* Complete "INSERT INTO <table> (" with attribute names */
	if TailMatches(IGNORE_CASE, previousWords, "INSERT", "INTO", "*", "(") {
		return c.completeWithAttributes(IGNORE_CASE, previousWords[1], text)
	}

	/*
	 * Complete INSERT INTO <table> with "(" or "VALUES" or "SELECT" or
	 * "TABLE" or "DEFAULT VALUES" or "OVERRIDING"
	 */
	if TailMatches(IGNORE_CASE, previousWords, "INSERT", "INTO", "*") {
		return CompleteFromList(text, "(", "DEFAULT VALUES", "SELECT", "TABLE", "VALUES", "OVERRIDING")
	}

	/*
	 * Complete INSERT INTO <table> (attribs) with "VALUES" or "SELECT" or
	 * "TABLE" or "OVERRIDING"
	 */
	if TailMatches(IGNORE_CASE, previousWords, "INSERT", "INTO", "*", "*") &&
		strings.HasSuffix(previousWords[0], ")") {
		return CompleteFromList(text, "SELECT", "TABLE", "VALUES", "OVERRIDING")
	}

	/* Complete OVERRIDING */
	if TailMatches(IGNORE_CASE, previousWords, "OVERRIDING") {
		return CompleteFromList(text, "SYSTEM VALUE", "USER VALUE")
	}

	/* Complete after OVERRIDING clause */
	if TailMatches(IGNORE_CASE, previousWords, "OVERRIDING", "*", "VALUE") {
		return CompleteFromList(text, "SELECT", "TABLE", "VALUES")
	}

	/* Insert an open parenthesis after "VALUES" */
	if TailMatches(IGNORE_CASE, previousWords, "VALUES") && !TailMatches(IGNORE_CASE, previousWords, "DEFAULT", "VALUES") {
		return CompleteFromList(text, "(")
	}
	/* UPDATE --- can be inside EXPLAIN, RULE, etc */
	/* If prev. word is UPDATE suggest a list of tables */
	if TailMatches(IGNORE_CASE, previousWords, "UPDATE") {
		return c.completeWithUpdatables(text)
	}
	/* Complete UPDATE <table> with "SET" */
	if TailMatches(IGNORE_CASE, previousWords, "UPDATE", "*") {
		return CompleteFromList(text, "SET")
	}
	/* Complete UPDATE <table> SET with list of attributes */
	if TailMatches(IGNORE_CASE, previousWords, "UPDATE", "*", "SET") {
		return c.completeWithAttributes(IGNORE_CASE, previousWords[1], text)
	}
	/* UPDATE <table> SET <attr> = */
	if TailMatches(IGNORE_CASE, previousWords, "UPDATE", "*", "SET", "!*=") {
		return CompleteFromList(text, "=")
	}
	/* WHERE */
	/* Simple case of the word before the where being the table name */
	if TailMatches(IGNORE_CASE, previousWords, "*", "WHERE") {
		// TODO would be great to _try_ to parse the (incomplete) query
		// and get a list of possible selectables to filter by
		return c.completeWithAttributes(IGNORE_CASE, previousWords[1], text,
			"AND",
			"OR",
			"CASE",
			"WHEN",
			"THEN",
			"ELSE",
			"END",
		)
	}

	/* ... FROM | JOIN ... */
	if TailMatches(IGNORE_CASE, previousWords, "FROM|JOIN") {
		return c.completeWithSelectables(text)
	}
	/* Backslash commands */
	if TailMatches(MATCH_CASE, previousWords, `\cd|\e|\edit|\g|\gx|\i|\include|\ir|\include_relative|\o|\out|\s|\w|\write`) {
		return completeFromFiles(text)
	}
	if TailMatches(MATCH_CASE, previousWords, `\c|\connect|\copy`) ||
		TailMatches(MATCH_CASE, previousWords, `\copy`, `*`) {
		return CompleteFromList(text, c.connStrings...)
	}
	if TailMatches(MATCH_CASE, previousWords, `\copy`, `*`, `*`) {
		return nil
	}
	if TailMatches(MATCH_CASE, previousWords, `\da*`) {
		return c.completeWithFunctions(text, []string{"AGGREGATE"})
	}
	if TailMatches(MATCH_CASE, previousWords, `\df*`) {
		return c.completeWithFunctions(text, []string{})
	}
	if TailMatches(MATCH_CASE, previousWords, `\di*`) {
		return c.completeWithIndexes(text)
	}
	if TailMatches(MATCH_CASE, previousWords, `\dn*`) {
		return c.completeWithSchemas(text)
	}
	if TailMatches(MATCH_CASE, previousWords, `\ds*`) {
		return c.completeWithSequences(text)
	}
	if TailMatches(MATCH_CASE, previousWords, `\dt*`) {
		return c.completeWithTables(text, []string{"TABLE", "BASE TABLE", "SYSTEM TABLE", "SYNONYM", "LOCAL TEMPORARY", "GLOBAL TEMPORARY"})
	}
	if TailMatches(MATCH_CASE, previousWords, `\dv*`) {
		return c.completeWithTables(text, []string{"VIEW", "SYSTEM VIEW"})
	}
	if TailMatches(MATCH_CASE, previousWords, `\dm*`) {
		return c.completeWithTables(text, []string{"MATERIALIZED VIEW"})
	}
	if TailMatches(MATCH_CASE, previousWords, `\d*`) {
		return c.completeWithSelectables(text)
	}
	if TailMatches(MATCH_CASE, previousWords, `\l*`) ||
		TailMatches(MATCH_CASE, previousWords, `\lo*`) {
		return c.completeWithCatalogs(text)
	}
	if TailMatches(MATCH_CASE, previousWords, `\pset`) {
		return CompleteFromList(text, `border`, `columns`, `expanded`, `fieldsep`, `fieldsep_zero`,
			`footer`, `format`, `linestyle`, `null`, `numericlocale`, `pager`, `pager_min_lines`,
			`recordsep`, `recordsep_zero`, `tableattr`, `title`, `title`, `tuples_only`,
			`unicode_border_linestyle`, `unicode_column_linestyle`, `unicode_header_linestyle`)
	}
	if TailMatches(MATCH_CASE, previousWords, `\pset`, `expanded`) {
		return CompleteFromList(text, "auto", "on", "off")
	}
	if TailMatches(MATCH_CASE, previousWords, `\pset`, `pager`) {
		return CompleteFromList(text, "always", "on", "off")
	}
	if TailMatches(MATCH_CASE, previousWords, `\pset`, `fieldsep_zero|footer|numericlocale|pager|recordsep_zero|tuples_only`) {
		return CompleteFromList(text, "on", "off")
	}
	if TailMatches(MATCH_CASE, previousWords, `\pset`, `format`) {
		return CompleteFromList(text, "unaligned", "aligned", "wrapped", "html", "asciidoc", "latex", "latex-longtable", "troff-ms", "csv", "json", "vertical")
	}
	if TailMatches(MATCH_CASE, previousWords, `\pset`, `linestyle`) {
		return CompleteFromList(text, "ascii", "old-ascii", "unicode")
	}
	if TailMatches(MATCH_CASE, previousWords, `\pset`, `unicode_border_linestyle|unicode_column_linestyle|unicode_header_linestyle`) {
		return CompleteFromList(text, "single", "double")
	}
	if TailMatches(MATCH_CASE, previousWords, `\pset`, `*`) ||
		TailMatches(MATCH_CASE, previousWords, `\pset`, `*`, `*`) {
		return nil
	}
	// is suggesting basic sql commands better than nothing?
	return CompleteFromList(text, c.sqlCommands...)
}

func getPreviousWords(point int, buf []rune) []string {
	var i int

	/*
	 * Allocate a slice of strings (rune slices). The worst case is that the line contains only
	 * non-whitespace WORD_BREAKS characters, making each one a separate word.
	 * This is usually much more space than we need, but it's cheaper than
	 * doing a separate malloc() for each word.
	 */
	previousWords := make([]string, 0, point*2)

	/*
	 * First we look for a non-word char before the current point.  (This is
	 * probably useless, if readline is on the same page as we are about what
	 * is a word, but if so it's cheap.)
	 */
	for i = point - 1; i >= 0; i-- {
		if strings.ContainsRune(WORD_BREAKS, buf[i]) {
			break
		}
	}
	point = i

	/*
	 * Now parse words, working backwards, until we hit start of line.  The
	 * backwards scan has some interesting but intentional properties
	 * concerning parenthesis handling.
	 */
	for point >= 0 {
		var start, end int
		inquotes := false
		parentheses := 0

		/* now find the first non-space which then constitutes the end */
		end = -1
		for i = point; i >= 0; i-- {
			if !unicode.IsSpace(buf[i]) {
				end = i
				break
			}
		}
		/* if no end found, we're done */
		if end < 0 {
			break
		}

		/*
		 * Otherwise we now look for the start.  The start is either the last
		 * character before any word-break character going backwards from the
		 * end, or it's simply character 0.  We also handle open quotes and
		 * parentheses.
		 */
		for start = end; start > 0; start-- {
			if buf[start] == '"' {
				inquotes = !inquotes
			}
			if inquotes {
				continue
			}
			if buf[start] == ')' {
				parentheses++
			} else if buf[start] == '(' {
				parentheses -= 1
				if parentheses <= 0 {
					break
				}
			} else if parentheses == 0 && strings.ContainsRune(WORD_BREAKS, buf[start-1]) {
				break
			}
		}

		/* Return the word located at start to end inclusive */
		i = end - start + 1
		previousWords = append(previousWords, string(buf[start:start+i]))

		/* Continue searching */
		point = start - 1
	}

	return previousWords
}

// TailMatches when last words match all patterns
func TailMatches(ct caseType, words []string, patterns ...string) bool {
	if len(words) < len(patterns) {
		return false
	}
	for i, p := range patterns {
		if !wordMatches(ct, p, words[len(patterns)-i-1]) {
			return false
		}
	}
	return true
}

func matches(ct caseType, words []string, patterns ...string) bool {
	if len(words) != len(patterns) {
		return false
	}
	for i, p := range patterns {
		if !wordMatches(ct, p, words[len(patterns)-i-1]) {
			return false
		}
	}
	return true
}

func wordMatches(ct caseType, pattern, word string) bool {
	if pattern == "*" {
		return true
	}

	if pattern[0] == '!' {
		return !wordMatches(ct, pattern[1:], word)
	}

	cmp := func(a, b string) bool { return a == b }
	if ct == IGNORE_CASE {
		cmp = strings.EqualFold
	}

	for _, p := range strings.Split(pattern, "|") {
		star := strings.IndexByte(p, '*')
		if star == -1 {
			if cmp(p, word) {
				return true
			}
		} else {
			if len(word) >= len(p)-1 && cmp(p[0:star], word[0:star]) && (star >= len(p) || cmp(p[star+1:], word[len(word)-len(p)+star+1:])) {
				return true
			}
		}
	}

	return false
}

// CompleteFromList where items starts with text, ignoring case
func CompleteFromList(text []rune, options ...string) [][]rune {
	return CompleteFromListCase(IGNORE_CASE, text, options...)
}

// CompleteFromList where items starts with text
func CompleteFromListCase(ct caseType, text []rune, options ...string) [][]rune {
	if len(options) == 0 {
		return nil
	}
	isLower := false
	if len(text) > 0 {
		isLower = unicode.IsLower(text[0])
	}
	prefix := string(text)
	if ct == IGNORE_CASE {
		prefix = strings.ToUpper(prefix)
	}
	result := make([][]rune, 0, len(options))
	for _, o := range options {
		if (ct == IGNORE_CASE && !strings.HasPrefix(strings.ToUpper(o), prefix)) ||
			(ct == MATCH_CASE && !strings.HasPrefix(o, prefix)) {
			continue
		}
		match := o[len(text):]
		if ct == IGNORE_CASE && isLower {
			match = strings.ToLower(match)
		}
		result = append(result, []rune(match))
	}
	return result
}

func completeFromVariables(text []rune, prefix, suffix string, needValue bool) [][]rune {
	vars := env.All()
	names := make([]string, 0, len(vars))
	for name, value := range vars {
		if needValue && value == "" {
			continue
		}
		names = append(names, fmt.Sprintf("%s%s%s", prefix, name, suffix))
	}
	return CompleteFromListCase(MATCH_CASE, text, names...)
}

func (c completer) completeWithSelectables(text []rune) [][]rune {
	filter := parseIdentifier(string(text))
	names := c.getNamespaces(filter)
	if r, ok := c.reader.(metadata.TableReader); ok {
		tables := c.getNames(
			func() (iterator, error) {
				return r.Tables(filter)
			},
			func(res interface{}) string {
				t := res.(*metadata.TableSet).Get()
				return qualifiedIdentifier(filter, t.Catalog, t.Schema, t.Name)
			},
		)
		names = append(names, tables...)
	}
	if r, ok := c.reader.(metadata.FunctionReader); ok {
		functions := c.getNames(
			func() (iterator, error) {
				return r.Functions(filter)
			},
			func(res interface{}) string {
				f := res.(*metadata.FunctionSet).Get()
				return qualifiedIdentifier(filter, f.Catalog, f.Schema, f.Name)
			},
		)
		names = append(names, functions...)
	}
	if r, ok := c.reader.(metadata.SequenceReader); ok {
		sequences := c.getNames(
			func() (iterator, error) {
				return r.Sequences(filter)
			},
			func(res interface{}) string {
				s := res.(*metadata.SequenceSet).Get()
				return qualifiedIdentifier(filter, s.Catalog, s.Schema, s.Name)
			},
		)
		names = append(names, sequences...)
	}
	sort.Strings(names)
	// TODO make sure CompleteFromList would properly handle quoted identifiers
	return CompleteFromList(text, names...)
}

func (c completer) completeWithTables(text []rune, types []string) [][]rune {
	r, ok := c.reader.(metadata.TableReader)
	if !ok {
		return [][]rune{}
	}

	filter := parseIdentifier(string(text))
	filter.Types = types
	names := c.getNamespaces(filter)
	tables := c.getNames(
		func() (iterator, error) {
			return r.Tables(filter)
		},
		func(res interface{}) string {
			t := res.(*metadata.TableSet).Get()
			return qualifiedIdentifier(filter, t.Catalog, t.Schema, t.Name)
		},
	)
	names = append(names, tables...)
	sort.Strings(names)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithFunctions(text []rune, types []string) [][]rune {
	r, ok := c.reader.(metadata.FunctionReader)
	if !ok {
		return [][]rune{}
	}
	filter := parseIdentifier(string(text))
	filter.Types = types
	names := c.getNamespaces(filter)
	functions := c.getNames(
		func() (iterator, error) {
			return r.Functions(filter)
		},
		func(res interface{}) string {
			f := res.(*metadata.FunctionSet).Get()
			return qualifiedIdentifier(filter, f.Catalog, f.Schema, f.Name)
		},
	)
	names = append(names, functions...)
	sort.Strings(names)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithIndexes(text []rune) [][]rune {
	r, ok := c.reader.(metadata.IndexReader)
	if !ok {
		return [][]rune{}
	}
	filter := parseIdentifier(string(text))
	names := c.getNamespaces(filter)
	indexes := c.getNames(
		func() (iterator, error) {
			return r.Indexes(filter)
		},
		func(res interface{}) string {
			f := res.(*metadata.IndexSet).Get()
			return qualifiedIdentifier(filter, f.Catalog, f.Schema, f.Name)
		},
	)
	names = append(names, indexes...)
	sort.Strings(names)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithSequences(text []rune) [][]rune {
	r, ok := c.reader.(metadata.SequenceReader)
	if !ok {
		return [][]rune{}
	}
	filter := parseIdentifier(string(text))
	names := c.getNamespaces(filter)
	sequences := c.getNames(
		func() (iterator, error) {
			return r.Sequences(filter)
		},
		func(res interface{}) string {
			s := res.(*metadata.SequenceSet).Get()
			return qualifiedIdentifier(filter, s.Catalog, s.Schema, s.Name)
		},
	)
	names = append(names, sequences...)
	sort.Strings(names)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithSchemas(text []rune) [][]rune {
	r, ok := c.reader.(metadata.SchemaReader)
	if !ok {
		return [][]rune{}
	}
	filter := parseIdentifier(string(text))
	names := c.getNames(
		func() (iterator, error) {
			if filter.Schema != "" {
				// name should already have a wildcard appended
				return r.Schemas(metadata.Filter{Catalog: filter.Schema, Name: filter.Name, WithSystem: true})
			}
			return r.Schemas(filter)
		},
		func(res interface{}) string {
			s := res.(*metadata.SchemaSet).Get()
			return qualifiedIdentifier(filter, "", s.Catalog, s.Schema)
		},
	)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithCatalogs(text []rune) [][]rune {
	r, ok := c.reader.(metadata.CatalogReader)
	if !ok {
		return [][]rune{}
	}
	filter := parseIdentifier(string(text))
	names := c.getNames(
		func() (iterator, error) {
			return r.Catalogs(filter)
		},
		func(res interface{}) string {
			s := res.(*metadata.CatalogSet).Get()
			return s.Catalog
		},
	)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithUpdatables(text []rune) [][]rune {
	filter := parseIdentifier(string(text))
	names := c.getNamespaces(filter)
	if r, ok := c.reader.(metadata.TableReader); ok {
		// exclude materialized views, sequences, system tables, synonyms
		filter.Types = []string{"TABLE", "BASE TABLE", "LOCAL TEMPORARY", "GLOBAL TEMPORARY", "VIEW"}
		tables := c.getNames(
			func() (iterator, error) {
				return r.Tables(filter)
			},
			func(res interface{}) string {
				t := res.(*metadata.TableSet).Get()
				return qualifiedIdentifier(filter, t.Catalog, t.Schema, t.Name)
			},
		)
		names = append(names, tables...)
	}
	sort.Strings(names)
	// TODO make sure CompleteFromList would properly handle quoted identifiers
	return CompleteFromList(text, names...)
}

func (c completer) getNamespaces(f metadata.Filter) []string {
	names := make([]string, 0, 10)
	if f.Catalog == "" && f.Schema == "" {
		if r, ok := c.reader.(metadata.CatalogReader); ok {
			catalogs := c.getNames(
				func() (iterator, error) { return r.Catalogs(metadata.Filter{}) },
				func(res interface{}) string {
					return res.(*metadata.CatalogSet).Get().Catalog
				},
			)
			names = append(names, catalogs...)
		}
	}
	if f.Catalog != "" {
		// filter is already fully qualified, so don't return any namespaces
		return names
	}
	if r, ok := c.reader.(metadata.SchemaReader); ok {
		schemas := c.getNames(
			func() (iterator, error) {
				if f.Schema != "" {
					// name should already have a wildcard appended
					return r.Schemas(metadata.Filter{Catalog: f.Schema, Name: f.Name, WithSystem: true})
				}
				return r.Schemas(f)
			},
			func(res interface{}) string {
				s := res.(*metadata.SchemaSet).Get()
				return qualifiedIdentifier(f, "", s.Catalog, s.Schema)
			},
		)
		names = append(names, schemas...)
	}
	return names
}

func (c completer) completeWithAttributes(ct caseType, selectable string, text []rune, options ...string) [][]rune {
	names := make([]string, 0, 10)
	if r, ok := c.reader.(metadata.ColumnReader); ok {
		parent := parseParentIdentifier(selectable)
		columns := c.getNames(
			func() (iterator, error) {
				return r.Columns(parent)
			},
			func(res interface{}) string {
				return res.(*metadata.ColumnSet).Get().Name
			},
		)
		names = append(names, columns...)
	}
	if r, ok := c.reader.(metadata.FunctionReader); ok {
		filter := parseIdentifier(string(text))
		// functions don't have to be fully qualified to be callable
		filter.OnlyVisible = false
		functions := c.getNames(
			func() (iterator, error) {
				return r.Functions(filter)
			},
			func(res interface{}) string {
				return res.(*metadata.FunctionSet).Get().Name
			},
		)
		names = append(names, functions...)
	}
	names = append(names, options...)
	return CompleteFromList(text, names...)
}

// parseIdentifier into catalog, schema and name
func parseIdentifier(name string) metadata.Filter {
	// TODO handle quoted identifiers
	result := metadata.Filter{}
	if !strings.ContainsRune(name, '.') {
		result.Name = name + "%"
		result.OnlyVisible = true
	} else {
		parts := strings.SplitN(name, ".", 3)
		if len(parts) == 2 {
			result.Schema = parts[0]
			result.Name = parts[1] + "%"
		} else {
			result.Catalog = parts[0]
			result.Schema = parts[1]
			result.Name = parts[2] + "%"
		}
	}

	if result.Schema != "" || len(result.Name) > 3 {
		result.WithSystem = true
	}
	return result
}

// parseParentIdentifier into catalog, schema and parent
func parseParentIdentifier(name string) metadata.Filter {
	// TODO handle quoted identifiers
	result := metadata.Filter{}
	if !strings.ContainsRune(name, '.') {
		result.Parent = name
		result.OnlyVisible = true
	} else {
		parts := strings.SplitN(name, ".", 3)
		if len(parts) == 2 {
			result.Schema = parts[0]
			result.Parent = parts[1]
		} else {
			result.Catalog = parts[0]
			result.Schema = parts[1]
			result.Parent = parts[2]
		}
	}

	if result.Schema != "" {
		result.WithSystem = true
	}
	return result
}

func qualifiedIdentifier(filter metadata.Filter, catalog, schema, name string) string {
	// TODO handle quoted identifiers
	if filter.Catalog != "" && filter.Schema != "" {
		return catalog + "." + schema + "." + name
	}
	if filter.Schema != "" {
		return schema + "." + name
	}
	return name
}

func (c completer) getNames(query func() (iterator, error), mapper func(interface{}) string) []string {
	res, err := query()
	if err != nil {
		if err != text.ErrNotSupported {
			c.logger.Println("Error getting selectables", err)
		}
		return nil
	}
	defer res.Close()

	// there can be duplicates if names are not qualified
	values := make(map[string]struct{}, 10)
	for res.Next() {
		values[mapper(res)] = struct{}{}
	}
	result := make([]string, 0, len(values))
	for v := range values {
		result = append(result, v)
	}
	return result
}

type iterator interface {
	Next() bool
	Close() error
}

func completeFromFiles(text []rune) [][]rune {
	// TODO handle quotes properly
	dir := filepath.Dir(string(text))
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	matches := make([]string, 0, len(dirs))
	switch dir {
	case ".":
		dir = ""
	case "/":
		// pass
	default:
		dir += "/"
	}
	for _, entry := range dirs {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		matches = append(matches, dir+name)
	}
	return CompleteFromList(text, matches...)
}
