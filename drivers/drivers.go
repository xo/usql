// Package drivers handles the registration, default implementation, and
// handles hooks for usql database drivers.
package drivers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/xo/dbmeta"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers/completer"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/stmt"
	"github.com/xo/usql/text"
)

// DB is the common interface for database operations, compatible with
// database/sql.DB and database/sql.Tx.
type DB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	Prepare(string) (*sql.Stmt, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

// Driver holds funcs for a driver.
type Driver struct {
	// Name is a name to override the driver name with.
	Name string
	// UseColumnTypes will cause database's ColumnTypes func to be used for
	// types.
	UseColumnTypes bool
	// ForceParams will be used to force parameters if defined.
	ForceParams func(*dburl.URL)
	// Open will be used by Open if defined.
	Open func(context.Context, *dburl.URL, func() io.Writer, func() io.Writer) (func(string, string) (*sql.DB, error), error)
	// Version will be used by Version if defined, for a database that dbmeta
	// has no model for. dbmeta reads the version of every other one.
	Version func(context.Context, DB) (string, error)
	// IsPasswordErr will be used by IsPasswordErr if defined.
	IsPasswordErr func(error) bool
	// AlwaysQuery runs every statement as a query, for a database whose
	// statements return rows that usql would otherwise run as exec, or that
	// sends no count of rows affected.
	AlwaysQuery bool
	// ColumnTypes is a callback that will be used if
	ColumnTypes func(*sql.ColumnType) (interface{}, error)
	// RowsAffected will be used by RowsAffected if defined.
	RowsAffected func(sql.Result) (int64, error)
	// Err will be used by Error.Error if defined.
	Err func(error) (string, string)
	// ConvertBytes will be used by ConvertBytes to convert a raw []byte
	// slice to a string if defined.
	ConvertBytes func([]byte, string) (string, error)
	// ConvertMap will be used by ConvertMap to convert a map[string]interface{}
	// to a string if defined.
	ConvertMap func(map[string]interface{}) (string, error)
	// ConvertSlice will be used by ConvertSlice to convert a []interface{} to
	// a string if defined.
	ConvertSlice func([]interface{}) (string, error)
	// ConvertDefault will be used by ConvertDefault to convert a interface{}
	// to a string if defined.
	ConvertDefault func(interface{}) (string, error)
	// Copy rows into the database table
	Copy func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error)
}

// drivers are registered drivers.
var drivers = make(map[string]Driver)

// Available returns the available drivers.
func Available() map[string]Driver {
	return drivers
}

// Register registers driver d with name and associated aliases.
func Register(name string, d Driver, aliases ...string) {
	if _, ok := drivers[name]; ok {
		panic(fmt.Sprintf("driver %s is already registered", name))
	}
	drivers[name] = d
	for _, alias := range aliases {
		if _, ok := drivers[alias]; ok {
			panic(fmt.Sprintf("alias %s is already registered", name))
		}
		drivers[alias] = d
	}
}

// Registered returns whether or not a driver is registered.
func Registered(name string) bool {
	_, ok := drivers[name]
	return ok
}

// LowerColumnNames reports whether column names are shown in lower case,
// which is so for a database that stores a name that is not quoted in upper
// case, such as Oracle.
func LowerColumnNames(u *dburl.URL) bool {
	info, ok := dbmeta.Dialect(u.Dialect).Info()
	return ok && info.Fold == dbmeta.FoldUpper
}

// UseColumnTypes reports whether the driver takes its column types from the
// database rather than from the raw bytes.
func UseColumnTypes(u *dburl.URL) bool {
	if d, ok := drivers[u.SchemeName]; ok {
		return d.UseColumnTypes
	}
	return false
}

// ForceParams forces parameters on the DSN for a driver.
func ForceParams(u *dburl.URL) {
	d, ok := drivers[u.SchemeName]
	if ok && d.ForceParams != nil {
		d.ForceParams(u)
	}
}

// Open opens a sql.DB connection for a driver.
func Open(ctx context.Context, u *dburl.URL, stdout, stderr func() io.Writer) (*sql.DB, error) {
	d, ok := drivers[u.SchemeName]
	if !ok {
		return nil, WrapErr(u.SchemeName, text.ErrDriverNotAvailable)
	}
	f := sql.Open
	if d.Open != nil {
		var err error
		if f, err = d.Open(ctx, u, stdout, stderr); err != nil {
			return nil, WrapErr(u.SchemeName, err)
		}
	}
	// the registry is keyed by the scheme, and database/sql by the driver
	db, err := f(u.Driver, u.DSN)
	if err != nil {
		return nil, WrapErr(u.SchemeName, err)
	}
	return db, nil
}

// stmtOpts returns statement options for a driver.
//
// dbmeta gives the grammar of each database it models. With no connection,
// every form is allowed. A database that dbmeta has no model for allows only
// block comments, which is what nearly every such driver set before dbmeta
// held the grammar. Allowing more is not safe: a hash comment would swallow
// SQL Server's #temp table names over ODBC.
func stmtOpts(u *dburl.URL) []stmt.Option {
	s := dbmeta.Syntax{
		DollarQuotes:  true,
		BlockComments: true,
		SlashComments: true,
		HashComments:  true,
		Backticks:     true,
	}
	if u != nil {
		var ok bool
		if s, ok = dbmeta.Dialect(u.Dialect).Syntax(); !ok {
			s = dbmeta.Syntax{BlockComments: true}
		}
	}
	return []stmt.Option{
		stmt.WithAllowDollar(s.DollarQuotes),
		stmt.WithAllowMultilineComments(s.BlockComments),
		stmt.WithAllowCComments(s.SlashComments),
		stmt.WithAllowHashComments(s.HashComments),
		stmt.WithAllowBacktick(s.Backticks),
	}
}

// NewStmt wraps creating a new stmt.Stmt for a driver.
func NewStmt(u *dburl.URL, f func() ([]rune, error), opts ...stmt.Option) *stmt.Stmt {
	return stmt.New(f, append(opts, stmtOpts(u)...)...)
}

// ConfigStmt sets the stmt.Stmt options for a driver.
func ConfigStmt(u *dburl.URL, s *stmt.Stmt) {
	if u == nil {
		return
	}
	for _, o := range stmtOpts(u) {
		o(s)
	}
}

// Version returns information about the database connection for a driver.
// dbmeta reads it for every database it has a model for.
func Version(ctx context.Context, u *dburl.URL, db DB) (string, error) {
	if metadata.Has(u.Dialect) {
		m, err := metadata.New(ctx, u.Dialect, db)
		if err != nil {
			return "", WrapErr(u.SchemeName, err)
		}
		return m.String(), nil
	}
	if d, ok := drivers[u.SchemeName]; ok && d.Version != nil {
		ver, err := d.Version(ctx, db)
		return ver, WrapErr(u.SchemeName, err)
	}
	var ver string
	err := db.QueryRowContext(ctx, `SELECT version();`).Scan(&ver)
	if err != nil || ver == "" {
		ver = "<unknown>"
	}
	return ver, nil
}

// User returns the current database user for a driver.
func User(ctx context.Context, u *dburl.URL, db DB) (string, error) {
	if metadata.Has(u.Dialect) {
		m, err := metadata.New(ctx, u.Dialect, db)
		if err != nil {
			return "", WrapErr(u.SchemeName, err)
		}
		if user, err := m.CurrentUser(ctx); err == nil && user != "" {
			return user, nil
		}
	}
	var user string
	_ = db.QueryRowContext(ctx, `SELECT current_user`).Scan(&user)
	return user, nil
}

// Process returns the type of the statement sqlstr, whose first words are
// prefix, the statement to run, and whether to run it as a query. It removes
// a trailing semicolon where dbmeta says the database refuses one.
func Process(u *dburl.URL, prefix, sqlstr string) (string, string, bool) {
	if info, ok := dbmeta.Dialect(u.Dialect).Info(); ok {
		switch info.Terminator {
		case dbmeta.TerminatorStripped:
			sqlstr = endRE.ReplaceAllString(sqlstr, "")
		case dbmeta.TerminatorStrippedUnlessEnd:
			if !endAnchorRE.MatchString(sqlstr) {
				sqlstr = endRE.ReplaceAllString(sqlstr, "")
			}
		}
	}
	typ, q := QueryExecType(prefix, sqlstr)
	return typ, sqlstr, q || drivers[u.SchemeName].AlwaysQuery
}

// ColumnTypes returns the column types callback for a driver.
func ColumnTypes(u *dburl.URL) func(*sql.ColumnType) (interface{}, error) {
	return drivers[u.SchemeName].ColumnTypes
}

// IsPasswordErr returns true if an err is a password error for a driver.
func IsPasswordErr(u *dburl.URL, err error) bool {
	drv := u.SchemeName
	if e, ok := err.(*Error); ok {
		drv, err = e.Driver, e.Err
	}
	if d, ok := drivers[drv]; ok && d.IsPasswordErr != nil {
		return d.IsPasswordErr(err)
	}
	return false
}

// RequirePreviousPassword returns true if the database requires the previous
// password when a user changes their own password.
func RequirePreviousPassword(u *dburl.URL) bool {
	info, ok := dbmeta.Dialect(u.Dialect).Info()
	return ok && info.OldPassword
}

// CanChangePassword returns whether or not the a driver supports changing
// passwords.
func CanChangePassword(u *dburl.URL) error {
	if metadata.CanChangePassword(u.Dialect) {
		return nil
	}
	return text.ErrPasswordNotSupportedByDriver
}

// ChangePassword initiates a user password change for the a driver. If user is
// not supplied, then the current user will be retrieved from User.
//
// dbmeta builds the statement, with the quoting rules of the database, and
// usql runs it.
func ChangePassword(u *dburl.URL, db DB, user, new, old string) (string, error) {
	if err := CanChangePassword(u); err != nil {
		return "", err
	}
	ctx := context.Background()
	if user == "" {
		var err error
		if user, err = User(ctx, u, db); err != nil {
			return "", err
		}
	}
	stmt, err := metadata.ChangePassword(ctx, u.Dialect, db, user, new, old)
	if err != nil {
		return "", WrapErr(u.SchemeName, err)
	}
	_, err = db.ExecContext(ctx, stmt)
	return user, WrapErr(u.SchemeName, err)
}

// Columns returns the column names for the SQL row result for a driver.
func Columns(u *dburl.URL, rows *sql.Rows) ([]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, WrapErr(u.SchemeName, err)
	}
	if LowerColumnNames(u) {
		for i, s := range cols {
			if j := strings.IndexFunc(s, func(r rune) bool {
				return unicode.IsLetter(r) && unicode.IsLower(r)
			}); j == -1 {
				cols[i] = strings.ToLower(s)
			}
		}
	}
	for i, c := range cols {
		if strings.TrimSpace(c) == "" {
			cols[i] = fmt.Sprintf("col%d", i)
		}
	}
	return cols, nil
}

// ConvertBytes returns a func to handle converting bytes for a driver.
func ConvertBytes(u *dburl.URL) func([]byte, string) (string, error) {
	if d, ok := drivers[u.SchemeName]; ok && d.ConvertBytes != nil {
		return d.ConvertBytes
	}
	return func(buf []byte, _ string) (string, error) {
		return string(buf), nil
	}
}

// ConvertMap returns a func to handle converting a map[string]interface{} for
// a driver.
func ConvertMap(u *dburl.URL) func(map[string]interface{}) (string, error) {
	if d, ok := drivers[u.SchemeName]; ok && d.ConvertMap != nil {
		return d.ConvertMap
	}
	return func(v map[string]interface{}) (string, error) {
		buf, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
}

// ConvertSlice returns a func to handle converting a []interface{} for a
// driver.
func ConvertSlice(u *dburl.URL) func([]interface{}) (string, error) {
	if d, ok := drivers[u.SchemeName]; ok && d.ConvertSlice != nil {
		return d.ConvertSlice
	}
	return func(v []interface{}) (string, error) {
		buf, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
}

// ConvertDefault returns a func to handle converting a interface{} for a
// driver.
func ConvertDefault(u *dburl.URL) func(interface{}) (string, error) {
	if d, ok := drivers[u.SchemeName]; ok && d.ConvertDefault != nil {
		return d.ConvertDefault
	}
	return func(v interface{}) (string, error) {
		return fmt.Sprintf("%v", v), nil
	}
}

// IsBatchQueryPrefix returns whether or not the supplied query prefix is a
// batch query prefix, and the closing prefix. Used to direct the handler to
// continue accumulating statements.
func IsBatchQueryPrefix(u *dburl.URL, prefix string) (string, string, bool) {
	// normalize
	typ, q := QueryExecType(prefix, "")
	if q {
		return typ, "", false
	}
	if info, ok := dbmeta.Dialect(u.Dialect).Info(); ok {
		for _, b := range info.Batches {
			if b.Begin == typ {
				return typ, b.End, true
			}
		}
	}
	return typ, "", false
}

// RowsAffected returns the rows affected for the SQL result for a driver.
func RowsAffected(u *dburl.URL, res sql.Result) (int64, error) {
	var count int64
	var err error
	if d, ok := drivers[u.SchemeName]; ok && d.RowsAffected != nil {
		count, err = d.RowsAffected(res)
	} else {
		count, err = res.RowsAffected()
	}
	if err != nil && err.Error() == "no RowsAffected available after DDL statement" {
		return 0, nil
	}
	if err != nil {
		return 0, WrapErr(u.SchemeName, err)
	}
	return count, nil
}

// Ping pings the database for a driver.
func Ping(ctx context.Context, u *dburl.URL, db *sql.DB) error {
	return WrapErr(u.SchemeName, db.PingContext(ctx))
}

// lexerNames are the names of the Chroma lexers that highlight the SQL of a
// dialect. A dialect with none uses the generic SQL lexer.
var lexerNames = map[string]string{
	"cockroachdb": "postgres",
	"cql":         "cql",
	"cratedb":     "postgres",
	"memsql":      "mysql",
	"mysql":       "mysql",
	"odbc":        "tsql",
	"postgres":    "postgres",
	"questdb":     "postgres",
	"redshift":    "postgres",
	"sqlserver":   "tsql",
	"tidb":        "mysql",
	"vitess":      "mysql",
}

// Lexer returns the syntax lexer for the dialect of a URL.
func Lexer(u *dburl.URL) chroma.Lexer {
	var l chroma.Lexer
	if u != nil {
		if name, ok := lexerNames[u.Dialect]; ok {
			l = lexers.Get(name)
		}
	}
	if l == nil {
		l = lexers.Get("sql")
	}
	l.Config().EnsureNL = false
	return l
}

// ForceQueryParameters is a utility func that wraps forcing params of name,
// value pairs.
func ForceQueryParameters(params []string) func(*dburl.URL) {
	if len(params)%2 != 0 {
		panic("invalid query params")
	}
	return func(u *dburl.URL) {
		if len(params) != 0 {
			v := u.Query()
			for i := 0; i < len(params); i += 2 {
				v.Set(params[i], params[i+1])
			}
			u.RawQuery = v.Encode()
		}
	}
}

// NewMeta returns the metadata of the database, which the describe commands
// and the completer read. It fails for a database that dbmeta has no model
// for.
func NewMeta(ctx context.Context, u *dburl.URL, db DB, opts ...metadata.Option) (*metadata.Meta, error) {
	m, err := metadata.New(ctx, u.Dialect, db, opts...)
	if errors.Is(err, metadata.ErrNotSupported) {
		return nil, fmt.Errorf(text.NotSupportedByDriver, `describe commands`, u.SchemeName)
	}
	return m, err
}

// NewCompleter returns a completer that completes the names of the objects
// in the database that m reads.
func NewCompleter(m *metadata.Meta, opts ...completer.Option) completer.Completer {
	return completer.NewDefaultCompleter(append([]completer.Option{completer.WithSource(m.Source())}, opts...)...)
}

// Copy copies the result set to the destination sql.DB.
func Copy(ctx context.Context, u *dburl.URL, stdout, stderr func() io.Writer, rows *sql.Rows, table string) (int64, error) {
	d, ok := drivers[u.SchemeName]
	if !ok {
		return 0, WrapErr(u.SchemeName, text.ErrDriverNotAvailable)
	}
	if d.Copy == nil {
		return 0, fmt.Errorf(text.NotSupportedByDriver, "copy", u.SchemeName)
	}
	db, err := Open(ctx, u, stdout, stderr)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	return d.Copy(ctx, db, rows, table)
}

// CopyWithInsert builds a typical copy handler based on insert.
func CopyWithInsert(placeholder func(int) string) func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
	if placeholder == nil {
		placeholder = func(n int) string { return fmt.Sprintf("$%d", n) }
	}
	return func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
		return FlexibleCopyWithInsert(ctx, db, rows, table, placeholder, true)
	}
}

func FlexibleCopyWithInsert(ctx context.Context, db *sql.DB, rows *sql.Rows, table string, placeholder func(int) string, withTransaction bool) (int64, error) {
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
		// TODO if the db supports multiple rows per insert, create batches of 100 rows
		placeholders := make([]string, clen)
		for i := 0; i < clen; i++ {
			placeholders[i] = placeholder(i + 1)
		}
		query = "INSERT INTO " + table + " VALUES (" + strings.Join(placeholders, ", ") + ")"
	}
	var stmt *sql.Stmt
	var tx *sql.Tx
	if withTransaction {
		tx, err = db.BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to begin transaction: %w", err)
		}
		stmt, err = tx.PrepareContext(ctx, query)
	} else {
		stmt, err = db.PrepareContext(ctx, query)
	}
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
		//We can't use values... in Exec() below, because some drivers
		//don't accept pointer to an argument instead of the arg itself.
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
	// TODO if using batches, flush the last batch,
	// TODO prepare another statement and count remaining rows
	if tx != nil {
		err = tx.Commit()
		if err != nil {
			return n, fmt.Errorf("failed to commit transaction: %w", err)
		}
	}
	return n, rows.Err()
}

func init() {
	dburl.OdbcIgnoreQueryPrefixes = []string{"usql_"}
}

var (
	endRE       = regexp.MustCompile(`;?\s*$`)
	endAnchorRE = regexp.MustCompile(`(?i)\send\s*;\s*$`)
)
