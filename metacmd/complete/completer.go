// Package complete completes SQL and meta commands at the cursor, for usql's
// line editor.
package complete

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/xo/rline"

	"github.com/xo/usql/env"
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

// Completer returns completion candidates for a line of input. It is an
// rline.Completer, so the line editor takes it as it is.
type Completer interface {
	rline.Completer
	// Do returns the candidates for line at pos, and the length of the prefix
	// they share. Each candidate is what follows that prefix.
	Do(line []rune, pos int) (newLine [][]rune, length int)
}

// Complete satisfies rline.Completer. rline hands it the whole entry and the
// cursor. Each candidate replaces the word before the cursor, which is why it
// is added with the word deleted before it, counted in bytes.
func (c completer) Complete(comp *rline.Completion, prefix string) {
	line := []rune(comp.Text())
	pos := utf8.RuneCountInString(prefix)
	suffixes, length := c.Do(line, pos)
	if length > pos {
		length = pos
	}
	word := string(line[pos-length : pos])
	for _, s := range suffixes {
		if !comp.AddCandidate(rline.Candidate{Replacement: word + string(s), DeleteBefore: len(word)}) {
			return
		}
	}
}

func New(opts ...Option) Completer {
	c := completer{
		sqlStartCommands: CommonSqlStartCommands,
		// TODO do we need to add built-in functions like, COALESCE, CAST, NULLIF, CONCAT etc?
		sqlCommands: CommonSqlCommands,
		backslashCommands: []string{
			`\!`,
			`\?`,
			`\C`,
			`\H`,
			`\T`,
			`\Z`,
			`\a`,
			`\begin`,
			`\bind`,
			`\c`,
			`\cd`,
			`\commit`,
			`\connect`,
			`\conninfo`,
			`\copy`,
			`\copyright`,
			`\cset`,
			`\d+`,
			`\dS+`,
			`\dS`,
			`\da+`,
			`\daS+`,
			`\daS`,
			`\da`,
			`\df+`,
			`\dfS+`,
			`\dfS`,
			`\df`,
			`\di+`,
			`\diS+`,
			`\diS`,
			`\di`,
			`\dm+`,
			`\dmS+`,
			`\dmS`,
			`\dm`,
			`\dn+`,
			`\dnS+`,
			`\dnS`,
			`\dn`,
			`\drivers`,
			`\ds+`,
			`\dsS+`,
			`\dsS`,
			`\ds`,
			`\dt+`,
			`\dtS+`,
			`\dtS`,
			`\dt`,
			`\dv+`,
			`\dvS+`,
			`\dvS`,
			`\dv`,
			`\e`,
			`\echo`,
			`\f`,
			`\g`,
			`\getenv`,
			`\gexec`,
			`\gset`,
			`\gx`,
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
			`\timing`,
			`\unset`,
			`\w`,
			`\watch`,
			`\x`,
		},
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

// Option to configure the reader
type Option func(*completer)

// WithSource sets the source of the names of the objects in the database.
func WithSource(s Source) Option {
	return func(c *completer) {
		c.source = s
	}
}

// WithConnStrings option
func WithConnStrings(connStrings []string) Option {
	return func(c *completer) {
		c.connStrings = connStrings
	}
}

// completer based on https://github.com/postgres/postgres/blob/9f3665fbfc34b963933e51778c7feaa8134ac885/src/bin/psql/tab-complete.c
type completer struct {
	source            Source
	sqlStartCommands  []string
	sqlCommands       []string
	backslashCommands []string
	connStrings       []string
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
		return CompleteFromList(text, "DATABASE", "SCHEMA", "SEQUENCE", "TABLE", "VIEW", "TEMPORARY")
	}
	if TailMatches(IGNORE_CASE, previousWords, "CREATE", "TEMP|TEMPORARY") {
		return CompleteFromList(text, "TABLE", "VIEW")
	}
	if TailMatches(IGNORE_CASE, previousWords, "CREATE", "TABLE", "*") || TailMatches(IGNORE_CASE, previousWords, "CREATE", "TEMP|TEMPORARY", "TABLE", "*") {
		return CompleteFromList(text, "(")
	}

	/* Complete ALTER */
	if TailMatches(IGNORE_CASE, previousWords, "ALTER") {
		return CompleteFromList(text, "DATABASE", "FUNCTION", "INDEX", "SCHEMA", "SEQUENCE", "TABLE", "TRIGGER", "USER", "VIEW")
	}

	/* Complete DROP */
	if TailMatches(IGNORE_CASE, previousWords, "DROP") {
		return CompleteFromList(text, "DATABASE", "FUNCTION", "INDEX", "SCHEMA", "SEQUENCE", "TABLE", "TRIGGER", "USER", "VIEW")
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
	/* TABLE, but not TABLE embedded in other commands */
	if matches(IGNORE_CASE, previousWords, "TABLE") {
		return c.completeWithUpdatables(text)
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
		return c.completeWithFunctions(text, true)
	}
	if TailMatches(MATCH_CASE, previousWords, `\df*`) {
		return c.completeWithFunctions(text, false)
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
		return c.completeWithTables(text, "t")
	}
	if TailMatches(MATCH_CASE, previousWords, `\dv*`) {
		return c.completeWithTables(text, "v")
	}
	if TailMatches(MATCH_CASE, previousWords, `\dm*`) {
		return c.completeWithTables(text, "m")
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
	if TailMatches(MATCH_CASE, previousWords, `\?`) {
		return CompleteFromList(text, "commands", "options", "variables")
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
	vars := env.Vars().Vars()
	names := make([]string, 0, len(vars))
	for name, value := range vars {
		if needValue && value == "" {
			continue
		}
		names = append(names, prefix+name+suffix)
	}
	return CompleteFromListCase(MATCH_CASE, text, names...)
}

func (c completer) completeWithSelectables(text []rune) [][]rune {
	filter := parseIdentifier(string(text))
	names := c.getNamespaces(filter)
	for _, kind := range []Kind{Tables, Functions, Sequences} {
		names = append(names, c.getNames(kind, filter)...)
	}
	sort.Strings(names)
	// TODO make sure CompleteFromList would properly handle quoted identifiers
	return CompleteFromList(text, names...)
}

func (c completer) completeWithTables(text []rune, types string) [][]rune {
	filter := parseIdentifier(string(text))
	filter.Types = types
	names := append(c.getNamespaces(filter), c.getNames(Tables, filter)...)
	sort.Strings(names)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithFunctions(text []rune, aggregates bool) [][]rune {
	filter := parseIdentifier(string(text))
	filter.Aggregates = aggregates
	names := append(c.getNamespaces(filter), c.getNames(Functions, filter)...)
	sort.Strings(names)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithIndexes(text []rune) [][]rune {
	filter := parseIdentifier(string(text))
	names := append(c.getNamespaces(filter), c.getNames(Indexes, filter)...)
	sort.Strings(names)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithSequences(text []rune) [][]rune {
	filter := parseIdentifier(string(text))
	names := append(c.getNamespaces(filter), c.getNames(Sequences, filter)...)
	sort.Strings(names)
	return CompleteFromList(text, names...)
}

func (c completer) completeWithSchemas(text []rune) [][]rune {
	filter := parseIdentifier(string(text))
	return CompleteFromList(text, c.getNames(Schemas, filter)...)
}

func (c completer) completeWithCatalogs(text []rune) [][]rune {
	filter := parseIdentifier(string(text))
	return CompleteFromList(text, c.getNames(Catalogs, filter)...)
}

func (c completer) completeWithUpdatables(text []rune) [][]rune {
	filter := parseIdentifier(string(text))
	// tables and views, and not materialized views or sequences
	filter.Types = "tv"
	names := append(c.getNamespaces(filter), c.getNames(Tables, filter)...)
	sort.Strings(names)
	// TODO make sure CompleteFromList would properly handle quoted identifiers
	return CompleteFromList(text, names...)
}

func (c completer) getNamespaces(f Filter) []string {
	names := make([]string, 0, 10)
	if f.Catalog == "" && f.Schema == "" {
		names = append(names, c.getNames(Catalogs, Filter{})...)
	}
	if f.Catalog != "" {
		// filter is already fully qualified, so don't return any namespaces
		return names
	}
	return append(names, c.getNames(Schemas, f)...)
}

func (c completer) completeWithAttributes(_ caseType, selectable string, text []rune, options ...string) [][]rune {
	names := make([]string, 0, 10)
	for _, o := range c.objects(Columns, parseParentIdentifier(selectable)) {
		names = append(names, o.Name)
	}
	// functions don't have to be fully qualified to be callable
	for _, o := range c.objects(Functions, parseIdentifier(string(text))) {
		names = append(names, o.Name)
	}
	names = append(names, options...)
	return CompleteFromList(text, names...)
}

// Kind is a kind of object that a completion names.
type Kind int

// The kinds of object.
const (
	Catalogs Kind = iota
	Schemas
	Tables
	Functions
	Sequences
	Indexes
	Columns
)

// Filter selects the objects of one kind. Catalog, Schema, Parent and Name
// are SQL LIKE patterns, and an empty one matches everything.
type Filter struct {
	Catalog string
	Schema  string
	// Parent is the table of the columns.
	Parent string
	Name   string
	// Types are the classes of relation, as the letters of \d: t tables, v
	// views, m materialized views and s sequences. None means every class.
	Types string
	// Aggregates selects only aggregate functions.
	Aggregates bool
	// WithSystem includes the objects the database keeps for itself.
	WithSystem bool
}

// Object is one object that a completion names.
type Object struct {
	Catalog string
	Schema  string
	Name    string
}

// Source answers the objects of a database. It returns text.ErrNotSupported
// for a kind that the database cannot list.
type Source interface {
	Objects(kind Kind, f Filter) ([]Object, error)
}

// parseIdentifier into catalog, schema and name
func parseIdentifier(name string) Filter {
	// TODO handle quoted identifiers
	result := Filter{}
	if !strings.ContainsRune(name, '.') {
		result.Name = name + "%"
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
func parseParentIdentifier(name string) Filter {
	// TODO handle quoted identifiers
	result := Filter{}
	if !strings.ContainsRune(name, '.') {
		result.Parent = name
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

func qualifiedIdentifier(filter Filter, catalog, schema, name string) string {
	// TODO handle quoted identifiers
	if filter.Catalog != "" && filter.Schema != "" {
		return catalog + "." + schema + "." + name
	}
	if filter.Schema != "" {
		return schema + "." + name
	}
	return name
}

// objects returns the objects of kind that f selects, and none when the
// database cannot list them.
func (c completer) objects(kind Kind, f Filter) []Object {
	if c.source == nil {
		return nil
	}
	objs, err := c.source.Objects(kind, f)
	if err != nil {
		return nil
	}
	return objs
}

// getNames returns the qualified names of the objects of kind that f
// selects, without duplicates.
func (c completer) getNames(kind Kind, f Filter) []string {
	// there can be duplicates if names are not qualified
	values := make(map[string]struct{}, 10)
	for _, o := range c.objects(kind, f) {
		switch kind {
		case Catalogs:
			values[o.Name] = struct{}{}
		case Schemas:
			values[qualifiedIdentifier(f, "", o.Catalog, o.Name)] = struct{}{}
		default:
			values[qualifiedIdentifier(f, o.Catalog, o.Schema, o.Name)] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for v := range values {
		result = append(result, v)
	}
	return result
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
