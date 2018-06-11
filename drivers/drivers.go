package drivers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/xo/dburl"

	"github.com/xo/usql/stmt"
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

	// ForceParams will be used to force parameters if defined.
	ForceParams func(*dburl.URL)

	// Open will be used by Open if defined.
	Open func(*dburl.URL) (func(string, string) (*sql.DB, error), error)

	// Version will be used by Version if defined.
	Version func(DB) (string, error)

	// User will be used by User if defined.
	User func(DB) (string, error)

	// ChangePassword will be used by ChangePassword if defined.
	ChangePassword func(DB, string, string, string) error

	// IsPasswordErr will be used by IsPasswordErr if defined.
	IsPasswordErr func(error) bool

	// Process will be used by Process if defined.
	Process func(string, string) (string, string, bool, error)

	// Columns will be used to retrieve the columns for the rows if
	// defined.
	Columns func(*sql.Rows) ([]string, error)

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
}

// drivers is the map of drivers funcs.
var drivers map[string]Driver

func init() {
	drivers = make(map[string]Driver)
}

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

// Registered returns whether or not a specific driver has been registered.
func Registered(name string) bool {
	_, ok := drivers[name]
	return ok
}

// ForceParams forces parameters on the supplied DSN for the registered driver.
func ForceParams(u *dburl.URL) {
	d, ok := drivers[u.Driver]
	if ok && d.ForceParams != nil {
		d.ForceParams(u)
	}
}

// Open opens a sql.DB connection for the registered driver.
func Open(u *dburl.URL) (*sql.DB, error) {
	var err error

	d, ok := drivers[u.Driver]
	if !ok {
		return nil, WrapErr(u.Driver, text.ErrDriverNotAvailable)
	}

	f := sql.Open
	if d.Open != nil {
		f, err = d.Open(u)
		if err != nil {
			return nil, WrapErr(u.Driver, err)
		}
	}

	db, err := f(u.Driver, u.DSN)
	if err != nil {
		return nil, WrapErr(u.Driver, err)
	}

	return db, nil
}

// stmtOpts returns statement options for the specified driver.
func stmtOpts(u *dburl.URL) []stmt.Option {
	if u != nil {
		if d, ok := drivers[u.Driver]; ok {
			return []stmt.Option{
				stmt.AllowDollar(d.AllowDollar),
				stmt.AllowMultilineComments(d.AllowMultilineComments),
				stmt.AllowCComments(d.AllowCComments),
				stmt.AllowHashComments(d.AllowHashComments),
			}
		}
	}

	return []stmt.Option{
		stmt.AllowDollar(true),
		stmt.AllowMultilineComments(true),
		stmt.AllowCComments(true),
		stmt.AllowHashComments(true),
	}
}

// NewStmt wraps creating a new stmt.Stmt for the specified driver.
func NewStmt(u *dburl.URL, f func() ([]rune, error), opts ...stmt.Option) *stmt.Stmt {
	return stmt.New(f, append(opts, stmtOpts(u)...)...)
}

// ConfigStmt sets the stmt.Stmt options for the specified driver.
func ConfigStmt(u *dburl.URL, s *stmt.Stmt) {
	if u == nil {
		return
	}
	for _, o := range stmtOpts(u) {
		o(s)
	}
}

// Version returns information about the database connection for the specified
// URL's driver.
func Version(u *dburl.URL, db DB) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.Version != nil {
		ver, err := d.Version(db)
		return ver, WrapErr(u.Driver, err)
	}

	var ver string
	err := db.QueryRow(`select version();`).Scan(&ver)
	if err != nil || ver == "" {
		ver = "<unknown>"
	}
	return ver, nil
}

// User returns the current database user for the specified URL's driver.
func User(u *dburl.URL, db DB) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.User != nil {
		user, err := d.User(db)
		return user, WrapErr(u.Driver, err)
	}

	var user string
	db.QueryRow(`select current_user`).Scan(&user)
	return user, nil
}

// Process processes the supplied SQL query for the specified URL's driver.
func Process(u *dburl.URL, prefix, sqlstr string) (string, string, bool, error) {
	if d, ok := drivers[u.Driver]; ok && d.Process != nil {
		a, b, c, err := d.Process(prefix, sqlstr)
		return a, b, c, WrapErr(u.Driver, err)
	}

	typ, q := QueryExecType(prefix, sqlstr)
	return typ, sqlstr, q, nil
}

// IsPasswordErr returns true if the specified err is a password error for the
// specified URL's driver.
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

// RequirePreviousPassword returns true if the specified URL's driver requires
// a previous password when changing a user's password.
func RequirePreviousPassword(u *dburl.URL) bool {
	if d, ok := drivers[u.Driver]; ok {
		return d.RequirePreviousPassword
	}
	return false
}

// CanChangePassword returns whether or not the specified driver's URL supports
// changing passwords.
func CanChangePassword(u *dburl.URL) error {
	if d, ok := drivers[u.Driver]; ok && d.ChangePassword != nil {
		return nil
	}
	return text.ErrPasswordNotSupportedByDriver
}

// ChangePassword initiates a user password change for the specified URL's
// driver. If user is not supplied, then the current user will be retrieved
// from User.
func ChangePassword(u *dburl.URL, db DB, user, new, old string) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.ChangePassword != nil {
		var err error
		if user == "" {
			user, err = User(u, db)
			if err != nil {
				return "", err
			}
		}

		return user, d.ChangePassword(db, user, new, old)
	}
	return "", text.ErrPasswordNotSupportedByDriver
}

// Columns returns the columns for SQL result for the specified URL's driver.
func Columns(u *dburl.URL, rows *sql.Rows) ([]string, error) {
	var cols []string
	var err error

	if d, ok := drivers[u.Driver]; ok && d.Columns != nil {
		cols, err = d.Columns(rows)
	} else {
		cols, err = rows.Columns()
	}

	if err != nil {
		return nil, WrapErr(u.Driver, err)
	}

	for i, c := range cols {
		if strings.TrimSpace(c) == "" {
			cols[i] = fmt.Sprintf("col%d", i)
		}
	}

	return cols, nil
}

// ConvertBytes returns a func to handle converting bytes for the specified
// URL's driver.
func ConvertBytes(u *dburl.URL) func([]byte, string) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.ConvertBytes != nil {
		return d.ConvertBytes
	}
	return func(buf []byte, _ string) (string, error) {
		return string(buf), nil
	}
}

// ConvertMap returns a func to handle converting a map[string]interface{} for
// the specified URL's driver.
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

// ConvertSlice returns a func to handle converting a []interface{} for
// the specified URL's driver.
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

// ConvertDefault returns a func to handle converting a interface{} for
// the specified URL's driver.
func ConvertDefault(u *dburl.URL) func(interface{}) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.ConvertDefault != nil {
		return d.ConvertDefault
	}
	return func(v interface{}) (string, error) {
		return fmt.Sprintf("%v", v), nil
	}
}

// BatchAsTransaction returns whether or not the the specified URL's driver requires
// batched queries to be done within a transaction block.
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

// RowsAffected returns the rows affected for the SQL result for a specified
// URL's driver.
func RowsAffected(u *dburl.URL, res sql.Result) (int64, error) {
	var count int64
	var err error
	if d, ok := drivers[u.Driver]; ok && d.RowsAffected != nil {
		count, err = d.RowsAffected(res)
	} else {
		count, err = res.RowsAffected()
	}
	if err != nil {
		return 0, WrapErr(u.Driver, err)
	}

	return count, nil
}

// Ping pings the database for a specified URL's driver.
func Ping(u *dburl.URL, db *sql.DB) error {
	return WrapErr(u.Driver, db.Ping())
}

// Lexer returns the syntax lexer for a specified URL's driver.
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
