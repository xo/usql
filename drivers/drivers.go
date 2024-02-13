// Package drivers handles the registration, default implementation, and
// handles hooks for usql database drivers.
package drivers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/gohxs/readline"
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
	// AllowDollar will be passed to query buffers to enable dollar ($$) style
	// strings.
	AllowDollar bool
	// AllowMultilineComments will be passed to query buffers to enable
	// multiline (/**/) style comments.
	AllowMultilineComments bool
	// AllowCComments will be passed to query buffers to enable C (//) style
	// comments.
	AllowCComments bool
	// AllowHashComments will be passed to query buffers to enable hash (#)
	// style comments.
	AllowHashComments bool
	// RequirePreviousPassword will be used by RequirePreviousPassword.
	RequirePreviousPassword bool
	// LexerName is the name of the syntax lexer to use.
	LexerName string
	// LowerColumnNames will cause column names to be lowered cased.
	LowerColumnNames bool
	// UseColumnTypes will cause database's ColumnTypes func to be used for
	// types.
	UseColumnTypes bool
	// ForceParams will be used to force parameters if defined.
	ForceParams func(*dburl.URL)
	// Open will be used by Open if defined.
	Open func(context.Context, *dburl.URL, func() io.Writer, func() io.Writer) (func(string, string) (*sql.DB, error), error)
	// Version will be used by Version if defined.
	Version func(context.Context, DB) (string, error)
	// User will be used by User if defined.
	User func(context.Context, DB) (string, error)
	// ChangePassword will be used by ChangePassword if defined.
	ChangePassword func(DB, string, string, string) error
	// IsPasswordErr will be used by IsPasswordErr if defined.
	IsPasswordErr func(error) bool
	// Process will be used by Process if defined.
	Process func(*dburl.URL, string, string) (string, string, bool, error)
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
	// BatchAsTransaction will cause batched queries to be done in a
	// transaction block.
	BatchAsTransaction bool
	// BatchQueryPrefixes will be used by BatchQueryPrefixes if defined.
	BatchQueryPrefixes map[string]string
	// NewMetadataReader returns a db metadata introspector.
	NewMetadataReader func(db DB, opts ...metadata.ReaderOption) metadata.Reader
	// NewMetadataWriter returns a db metadata printer.
	NewMetadataWriter func(db DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer
	// NewCompleter returns a db auto-completer.
	NewCompleter func(db DB, opts ...completer.Option) readline.AutoCompleter
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

// LowerColumnNames returns whether or not column names should be converted to
// lower case for a driver.
func LowerColumnNames(u *dburl.URL) bool {
	if d, ok := drivers[u.Driver]; ok {
		return d.LowerColumnNames
	}
	return false
}

// UseColumnTypes returns whether or not a driver should uses column types.
func UseColumnTypes(u *dburl.URL) bool {
	if d, ok := drivers[u.Driver]; ok {
		return d.UseColumnTypes
	}
	return false
}

// ForceParams forces parameters on the DSN for a driver.
func ForceParams(u *dburl.URL) {
	d, ok := drivers[u.Driver]
	if ok && d.ForceParams != nil {
		d.ForceParams(u)
	}
}

// Open opens a sql.DB connection for a driver.
func Open(ctx context.Context, u *dburl.URL, stdout, stderr func() io.Writer) (*sql.DB, error) {
	d, ok := drivers[u.Driver]
	if !ok {
		return nil, WrapErr(u.Driver, text.ErrDriverNotAvailable)
	}
	f := sql.Open
	if d.Open != nil {
		var err error
		if f, err = d.Open(ctx, u, stdout, stderr); err != nil {
			return nil, WrapErr(u.Driver, err)
		}
	}
	driver := u.Driver
	if u.GoDriver != "" {
		driver = u.GoDriver
	}
	db, err := f(driver, u.DSN)
	if err != nil {
		return nil, WrapErr(u.Driver, err)
	}
	return db, nil
}

// stmtOpts returns statement options for a driver.
func stmtOpts(u *dburl.URL) []stmt.Option {
	if u != nil {
		if d, ok := drivers[u.Driver]; ok {
			return []stmt.Option{
				stmt.WithAllowDollar(d.AllowDollar),
				stmt.WithAllowMultilineComments(d.AllowMultilineComments),
				stmt.WithAllowCComments(d.AllowCComments),
				stmt.WithAllowHashComments(d.AllowHashComments),
			}
		}
	}
	return []stmt.Option{
		stmt.WithAllowDollar(true),
		stmt.WithAllowMultilineComments(true),
		stmt.WithAllowCComments(true),
		stmt.WithAllowHashComments(true),
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
func Version(ctx context.Context, u *dburl.URL, db DB) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.Version != nil {
		ver, err := d.Version(ctx, db)
		return ver, WrapErr(u.Driver, err)
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
	if d, ok := drivers[u.Driver]; ok && d.User != nil {
		user, err := d.User(ctx, db)
		return user, WrapErr(u.Driver, err)
	}
	var user string
	_ = db.QueryRowContext(ctx, `SELECT current_user`).Scan(&user)
	return user, nil
}

// Process processes the sql query for a driver.
func Process(u *dburl.URL, prefix, sqlstr string) (string, string, bool, error) {
	if d, ok := drivers[u.Driver]; ok && d.Process != nil {
		a, b, c, err := d.Process(u, prefix, sqlstr)
		return a, b, c, WrapErr(u.Driver, err)
	}
	typ, q := QueryExecType(prefix, sqlstr)
	return typ, sqlstr, q, nil
}

// ColumnTypes returns the column types callback for a driver.
func ColumnTypes(u *dburl.URL) func(*sql.ColumnType) (interface{}, error) {
	return drivers[u.Driver].ColumnTypes
}

// IsPasswordErr returns true if an err is a password error for a driver.
func IsPasswordErr(u *dburl.URL, err error) bool {
	drv := u.Driver
	if e, ok := err.(*Error); ok {
		drv, err = e.Driver, e.Err
	}
	if d, ok := drivers[drv]; ok && d.IsPasswordErr != nil {
		return d.IsPasswordErr(err)
	}
	return false
}

// RequirePreviousPassword returns true if a driver requires a previous
// password when changing a user's password.
func RequirePreviousPassword(u *dburl.URL) bool {
	if d, ok := drivers[u.Driver]; ok {
		return d.RequirePreviousPassword
	}
	return false
}

// CanChangePassword returns whether or not the a driver supports changing
// passwords.
func CanChangePassword(u *dburl.URL) error {
	if d, ok := drivers[u.Driver]; ok && d.ChangePassword != nil {
		return nil
	}
	return text.ErrPasswordNotSupportedByDriver
}

// ChangePassword initiates a user password change for the a driver. If user is
// not supplied, then the current user will be retrieved from User.
func ChangePassword(u *dburl.URL, db DB, user, new, old string) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.ChangePassword != nil {
		if user == "" {
			var err error
			if user, err = User(context.Background(), u, db); err != nil {
				return "", err
			}
		}
		return user, d.ChangePassword(db, user, new, old)
	}
	return "", text.ErrPasswordNotSupportedByDriver
}

// Columns returns the column names for the SQL row result for a driver.
func Columns(u *dburl.URL, rows *sql.Rows) ([]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, WrapErr(u.Driver, err)
	}
	if drivers[u.Driver].LowerColumnNames {
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
	if d, ok := drivers[u.Driver]; ok && d.ConvertBytes != nil {
		return d.ConvertBytes
	}
	return func(buf []byte, _ string) (string, error) {
		return string(buf), nil
	}
}

// ConvertMap returns a func to handle converting a map[string]interface{} for
// a driver.
func ConvertMap(u *dburl.URL) func(map[string]interface{}) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.ConvertMap != nil {
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
	if d, ok := drivers[u.Driver]; ok && d.ConvertSlice != nil {
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
	if d, ok := drivers[u.Driver]; ok && d.ConvertDefault != nil {
		return d.ConvertDefault
	}
	return func(v interface{}) (string, error) {
		return fmt.Sprintf("%v", v), nil
	}
}

// BatchAsTransaction returns whether or not a driver requires batched queries
// to be done within a transaction block.
func BatchAsTransaction(u *dburl.URL) bool {
	if d, ok := drivers[u.Driver]; ok {
		return d.BatchAsTransaction
	}
	return false
}

// IsBatchQueryPrefix returns whether or not the supplied query prefix is a
// batch query prefix, and the closing prefix. Used to direct the handler to
// continue accumulating statements.
func IsBatchQueryPrefix(u *dburl.URL, prefix string) (string, string, bool) {
	// normalize
	typ, q := QueryExecType(prefix, "")
	d, ok := drivers[u.Driver]
	if q || !ok || d.BatchQueryPrefixes == nil {
		return typ, "", false
	}
	end, ok := d.BatchQueryPrefixes[typ]
	return typ, end, ok
}

// RowsAffected returns the rows affected for the SQL result for a driver.
func RowsAffected(u *dburl.URL, res sql.Result) (int64, error) {
	var count int64
	var err error
	if d, ok := drivers[u.Driver]; ok && d.RowsAffected != nil {
		count, err = d.RowsAffected(res)
	} else {
		count, err = res.RowsAffected()
	}
	if err != nil && err.Error() == "no RowsAffected available after DDL statement" {
		return 0, nil
	}
	if err != nil {
		return 0, WrapErr(u.Driver, err)
	}
	return count, nil
}

// Ping pings the database for a driver.
func Ping(ctx context.Context, u *dburl.URL, db *sql.DB) error {
	return WrapErr(u.Driver, db.PingContext(ctx))
}

// Lexer returns the syntax lexer for a driver.
func Lexer(u *dburl.URL) chroma.Lexer {
	var l chroma.Lexer
	if u != nil {
		if d, ok := drivers[u.Driver]; ok && d.LexerName != "" {
			l = lexers.Get(d.LexerName)
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

// NewMetadataReader wraps creating a new database introspector for a driver.
func NewMetadataReader(ctx context.Context, u *dburl.URL, db DB, w io.Writer, opts ...metadata.ReaderOption) (metadata.Reader, error) {
	d, ok := drivers[u.Driver]
	if !ok || d.NewMetadataReader == nil {
		return nil, fmt.Errorf(text.NotSupportedByDriver, `describe commands`, u.Driver)
	}
	return d.NewMetadataReader(db, opts...), nil
}

// NewMetadataWriter wraps creating a new database metadata printer for a driver.
func NewMetadataWriter(ctx context.Context, u *dburl.URL, db DB, w io.Writer, opts ...metadata.ReaderOption) (metadata.Writer, error) {
	d, ok := drivers[u.Driver]
	if !ok {
		return nil, fmt.Errorf(text.NotSupportedByDriver, `describe commands`, u.Driver)
	}
	if d.NewMetadataWriter != nil {
		return d.NewMetadataWriter(db, w, opts...), nil
	}
	if d.NewMetadataReader == nil {
		return nil, fmt.Errorf(text.NotSupportedByDriver, `describe commands`, u.Driver)
	}
	newMetadataWriter := metadata.NewDefaultWriter(d.NewMetadataReader(db, opts...))
	return newMetadataWriter(db, w), nil
}

// NewCompleter creates a metadata completer for a driver and database
// connection.
func NewCompleter(ctx context.Context, u *dburl.URL, db DB, readerOpts []metadata.ReaderOption, opts ...completer.Option) readline.AutoCompleter {
	d, ok := drivers[u.Driver]
	if !ok {
		return nil
	}
	if d.NewCompleter != nil {
		return d.NewCompleter(db, opts...)
	}
	if d.NewMetadataReader == nil {
		return nil
	}
	// prepend to allow to override default options
	readerOpts = append([]metadata.ReaderOption{
		// this needs to be relatively low, since autocomplete is very interactive
		metadata.WithTimeout(3 * time.Second),
		metadata.WithLimit(1000),
	}, readerOpts...)
	opts = append([]completer.Option{
		completer.WithReader(d.NewMetadataReader(db, readerOpts...)),
		completer.WithDB(db),
	}, opts...)
	return completer.NewDefaultCompleter(opts...)
}

// Copy copies the result set to the destination sql.DB.
func Copy(ctx context.Context, u *dburl.URL, stdout, stderr func() io.Writer, rows *sql.Rows, table string) (int64, error) {
	d, ok := drivers[u.Driver]
	if !ok {
		return 0, WrapErr(u.Driver, text.ErrDriverNotAvailable)
	}
	if d.Copy == nil {
		return 0, fmt.Errorf(text.NotSupportedByDriver, "copy", u.Driver)
	}
	db, err := Open(ctx, u, stdout, stderr)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	return d.Copy(ctx, db, rows, table)
}

// CopyWithInsert builds a copy handler based on insert.
func CopyWithInsert(placeholder func(int) string) func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
	if placeholder == nil {
		placeholder = func(n int) string { return fmt.Sprintf("$%d", n) }
	}
	return func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
		columns, err := rows.Columns()
		if err != nil {
			return 0, fmt.Errorf("failed to fetch source rows columns: %w", err)
		}
		clen := len(columns)
		query := table
		if !strings.HasPrefix(strings.ToLower(query), "insert into") {
			leftParen := strings.IndexRune(table, '(')
			if leftParen == -1 {
				colStmt, err := db.PrepareContext(ctx, "SELECT * FROM "+table+" WHERE 1=0")
				if err != nil {
					return 0, fmt.Errorf("failed to prepare query to determine target table columns: %w", err)
				}
				defer colStmt.Close()
				colRows, err := colStmt.QueryContext(ctx)
				if err != nil {
					return 0, fmt.Errorf("failed to execute query to determine target table columns: %w", err)
				}
				columns, err := colRows.Columns()
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
		for i := 0; i < len(columnTypes); i++ {
			values[i] = reflect.New(columnTypes[i].ScanType()).Interface()
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
		// TODO if using batches, flush the last batch,
		// TODO prepare another statement and count remaining rows
		err = tx.Commit()
		if err != nil {
			return n, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return n, rows.Err()
	}
}

func init() {
	dburl.OdbcIgnoreQueryPrefixes = []string{"usql_"}
}
