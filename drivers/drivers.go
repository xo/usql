package drivers

import (
	"database/sql"
	"fmt"
	"strings"

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
	// N is a name to override the driver name with.
	N string

	// AD will be passed to query buffers to enable dollar ($$) style strings.
	AD bool

	// AMC will be passed to query buffers to enable multiline (/**/) style
	// comments.
	AMC bool

	// ACC will be passed to query buffers to enable C (//) style comments.
	ACC bool

	// AHC will be passed to query buffers to enable hash (#) style comments.
	AHC bool

	// ReqPP will be used by RequirePreviousPassword.
	ReqPP bool

	// O will be used by Open if defined.
	O func(*dburl.URL) (func(string, string) (*sql.DB, error), error)

	// V will be used by Version if defined.
	V func(DB) (string, error)

	// U will be used by User if defined.
	U func(DB) (string, error)

	// ChPw will be used by ChangePassword if defined.
	ChPw func(DB, string, string, string) error

	// PwErr will be used by IsPasswordErr if defined.
	PwErr func(error) bool

	// P will be used by Process if defined.
	P func(string, string) (string, string, bool, error)

	// Cols will be used to retrieve the columns for the rows if defined.
	Cols func(*sql.Rows) ([]string, error)

	// Cb will be used by ConvertBytes to convert a raw []byte slice to a
	// string if defined.
	Cb func([]byte) string

	// E will be used by Error.Error if defined.
	E func(error) (string, string)

	// EV will be used by Error.Verbose if defined.
	EV func(error) *ErrVerbose

	// A will be used by RowsAffected if defined.
	A func(sql.Result) (int64, error)
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

// Open opens a sql.DB connection for the registered driver.
func Open(u *dburl.URL, buf *stmt.Stmt) (*sql.DB, error) {
	var err error

	d, ok := drivers[u.Driver]
	if !ok {
		return nil, WrapErr(u.Driver, text.ErrDriverNotAvailable)
	}

	// force query buffer settings
	stmt.AllowDollar(d.AD)(buf)
	stmt.AllowMultilineComments(d.AMC)(buf)
	stmt.AllowCComments(d.ACC)(buf)
	stmt.AllowHashComments(d.AHC)(buf)

	f := sql.Open
	if d.O != nil {
		f, err = d.O(u)
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

// Version returns information about the database connection for the specified
// URL's driver.
func Version(u *dburl.URL, db DB) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.V != nil {
		ver, err := d.V(db)
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
	if d, ok := drivers[u.Driver]; ok && d.U != nil {
		user, err := d.U(db)
		return user, WrapErr(u.Driver, err)
	}

	var user string
	db.QueryRow(`select current_user`).Scan(&user)
	return user, nil
}

// Process processes the supplied SQL query for the specified URL's driver.
func Process(u *dburl.URL, prefix, sqlstr string) (string, string, bool, error) {
	if d, ok := drivers[u.Driver]; ok && d.P != nil {
		a, b, c, err := d.P(prefix, sqlstr)
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

	if d, ok := drivers[drv]; ok && d.PwErr != nil {
		return d.PwErr(err)
	}
	return false
}

// RequirePreviousPassword returns true if the specified URL's driver requires
// a previous password when changing a user's password.
func RequirePreviousPassword(u *dburl.URL) bool {
	if d, ok := drivers[u.Driver]; ok {
		return d.ReqPP
	}
	return false
}

// CanChangePassword returns whether or not the specified driver's URL supports
// changing passwords.
func CanChangePassword(u *dburl.URL) error {
	if d, ok := drivers[u.Driver]; ok && d.ChPw != nil {
		return nil
	}
	return text.ErrPasswordNotSupportedByDriver
}

// ChangePassword initiates a user password change for the specified URL's
// driver. If user is not supplied, then the current user will be retrieved
// from User.
func ChangePassword(u *dburl.URL, db DB, user, new, old string) (string, error) {
	if d, ok := drivers[u.Driver]; ok && d.ChPw != nil {
		var err error
		if user == "" {
			user, err = User(u, db)
			if err != nil {
				return "", err
			}
		}

		return user, d.ChPw(db, user, new, old)
	}
	return "", text.ErrPasswordNotSupportedByDriver
}

// Columns returns the columns for SQL result for the specified URL's driver.
func Columns(u *dburl.URL, rows *sql.Rows) ([]string, error) {
	var cols []string
	var err error

	if d, ok := drivers[u.Driver]; ok && d.Cols != nil {
		cols, err = d.Cols(rows)
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

// ConvertBytes converts a raw byte slice for a specified URL's driver.
func ConvertBytes(u *dburl.URL, buf []byte) string {
	if d, ok := drivers[u.Driver]; ok && d.Cb != nil {
		return d.Cb(buf)
	}
	return string(buf)
}

// RowsAffected returns the rows affected for the SQL result for a specified
// URL's driver.
func RowsAffected(u *dburl.URL, res sql.Result) (int64, error) {
	var count int64
	var err error
	if d, ok := drivers[u.Driver]; ok && d.A != nil {
		count, err = d.A(res)
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
