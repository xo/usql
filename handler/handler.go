package handler

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/knq/dburl"
	"github.com/olekukonko/tablewriter"

	"github.com/knq/usql/drivers"
	"github.com/knq/usql/env"
	"github.com/knq/usql/metacmd"
	"github.com/knq/usql/rline"
	"github.com/knq/usql/stmt"
	"github.com/knq/usql/text"
)

var (
	// ErrNotConnected is the not connected error.
	ErrNotConnected = errors.New("not connected")

	// ErrNoSuchFileOrDirectory is the no such file or directory error.
	ErrNoSuchFileOrDirectory = errors.New("no such file or directory")

	// ErrCannotIncludeDirectories is the cannot include directories error.
	ErrCannotIncludeDirectories = errors.New("cannot include directories")

	// ErrMissingDSN is the missing dsn error.
	ErrMissingDSN = errors.New("missing dsn")
)

// Handler is a input process handler.
type Handler struct {
	l    rline.IO
	user *user.User
	wd   string
	nopw bool

	// statement buffer
	buf        *stmt.Stmt
	lastPrefix string
	last       string

	// connection
	u  *dburl.URL
	db *sql.DB
}

// New creates a new input handler.
func New(l rline.IO, user *user.User, wd string, nopw bool) *Handler {
	// set help intercept
	f := l.Next
	if l.Interactive() {
		f = func() ([]rune, error) {
			// next line
			r, err := l.Next()
			if err != nil {
				return nil, err
			}

			// check if line starts with help
			rlen := len(r)
			if rlen >= 4 && stmt.StartsWith(r, 0, rlen, text.HelpPrefix) {
				fmt.Fprintln(l.Stdout(), text.HelpDesc)
				return nil, nil
			}

			// save history
			l.Save(string(r))

			return r, nil
		}
	}

	return &Handler{
		l:    l,
		user: user,
		wd:   wd,
		nopw: nopw,
		buf:  stmt.New(f),
	}
}

// Run executes queries and commands.
func (h *Handler) Run() error {
	stdout, stderr, iactive := h.l.Stdout(), h.l.Stderr(), h.l.Interactive()

	// display welcome info
	if iactive {
		fmt.Fprintln(h.l.Stdout(), text.WelcomeDesc)
		fmt.Fprintln(h.l.Stdout())
	}

	for {
		var err error
		var execute bool

		// set prompt
		if iactive {
			h.l.Prompt(h.Prompt())
		}

		// read next statement/command
		cmd, params, err := h.buf.Next()
		switch {
		case !iactive && err == nil:
			execute = h.buf.Len != 0

		case err == rline.ErrInterrupt:
			h.buf.Reset(nil)
			continue

		case err != nil:
			return err
		}

		var res metacmd.Res
		if cmd != "" {
			// decode
			var r metacmd.Runner
			r, err = metacmd.Decode(cmd, params)
			switch {
			case err == metacmd.ErrUnknownCommand:
				fmt.Fprintf(stderr, text.InvalidCommand, cmd)
				fmt.Fprintln(stderr)
				continue
			case err == metacmd.ErrMissingRequiredArgument:
				fmt.Fprintf(stderr, text.MissingRequiredArg, cmd)
				fmt.Fprintln(stderr)
				continue
			case err != nil:
				fmt.Fprintf(stderr, "error: %v", err)
				fmt.Fprintln(stderr)
				continue
			}

			// run
			res, err = r.Run(h)
			if err != nil && err != rline.ErrInterrupt {
				fmt.Fprintf(stderr, "error: %v", err)
				fmt.Fprintln(stderr)
				continue
			}

			// print unused command parameters
			for i := res.Processed; i < len(params); i++ {
				fmt.Fprintf(stdout, text.ExtraArgumentIgnored, cmd, params[i])
				fmt.Fprintln(stdout)
			}
		}

		// quit
		if res.Quit {
			return nil
		}

		// execute buf
		if execute || h.buf.Ready() || res.Exec != metacmd.ExecNone {
			if h.buf.Len != 0 {
				h.lastPrefix, h.last = h.buf.Prefix, h.buf.String()
				h.buf.Reset(nil)
			}

			// log.Printf(">> PROCESS EXECUTE: (%s) `%s`", h.lastPrefix, h.last)
			if h.last != "" && h.last != ";" {
				err = h.Execute(stdout, h.lastPrefix, h.last)
				if err != nil {
					fmt.Fprintf(stderr, "error: %v", err)
					fmt.Fprintln(stderr)
				}
			}
		}
	}
}

// Reset resets the handler's statement buffer.
func (h *Handler) Reset(r []rune) {
	h.buf.Reset(r)
	h.last, h.lastPrefix = "", ""
}

// Prompt creates the prompt text.
func (h *Handler) Prompt() string {
	s := text.NotConnected

	if h.db != nil {
		s = h.u.Short()
		if s == "" {
			s = "(" + h.u.Driver + ")"
		}
	}

	return s + h.buf.State() + "> "
}

// IO returns the io for the handler.
func (h *Handler) IO() rline.IO {
	return h.l
}

// User returns the user for the handler.
func (h *Handler) User() *user.User {
	return h.user
}

// URL returns the URL for the handler.
func (h *Handler) URL() *dburl.URL {
	return h.u
}

// DB returns the sql.DB for the handler.
func (h *Handler) DB() *sql.DB {
	return h.db
}

// Last returns the last executed statement.
func (h *Handler) Last() string {
	return h.last
}

// Buf returns the current statement buffer.
func (h *Handler) Buf() *stmt.Stmt {
	return h.buf
}

// Open handles opening a specified database URL, passing either a single
// string in the form of a URL, or more than one string, in which case the
// first string is treated as a driver name, and the remaining strings are
// joined (with a space) and passed as a DSN to sql.Open.
//
// If there is only one parameter, and it is not a well formatted URL, but
// appears to be a file on disk, then an attempt will be made to open it with
// an appropriate driver (mysql, postgres, sqlite3) depending on the type (unix
// domain socket, directory, or regular file, respectively).
func (h *Handler) Open(params ...string) error {
	if len(params) == 0 || params[0] == "" {
		return nil
	}

	var err error
	if len(params) < 2 {
		urlstr := params[0]

		// parse dsn
		h.u, err = dburl.Parse(urlstr)
		switch {
		case err == dburl.ErrInvalidDatabaseScheme:
			var fi os.FileInfo
			fi, err = os.Stat(urlstr)
			if err != nil {
				return err
			}

			switch {
			case fi.IsDir():
				return h.Open("postgres+unix:" + urlstr)

			case fi.Mode()&os.ModeSocket != 0:
				return h.Open("mysql+unix:" + urlstr)
			}

			// it is a file, so reattempt to open it with sqlite3
			return h.Open("sqlite3:" + urlstr)

		case err != nil:
			return err
		}

		// force parameters
		h.ForceParams(h.u)
	} else {
		h.u = &dburl.URL{
			Driver: params[0],
			DSN:    strings.Join(params[1:], " "),
		}
	}

	// open connection
	h.db, err = drivers.Open(h.u, h.buf)
	if err != nil && !drivers.IsPasswordErr(h.u, err) {
		defer h.Close()
		return err
	} else if err == nil {
		// force error/check connection
		err = drivers.Ping(h.u, h.db)
		if err == nil {
			return h.Version()
		}
	}

	// bail without getting password
	if h.nopw || !drivers.IsPasswordErr(h.u, err) || len(params) > 1 || !h.l.Interactive() {
		defer h.Close()
		return err
	}

	// print the error
	fmt.Fprintf(h.l.Stderr(), "error: %v", err)
	fmt.Fprintln(h.l.Stderr())

	// otherwise, try to collect a password ...
	dsn, err := h.Password(params[0])
	if err != nil {
		// close connection
		defer h.Close()
		return err
	}

	// reconnect
	return h.Open(dsn)
}

// forceParamMap are the params to force for specific database drivers/schemes.
var forceParamMap = map[string][]string{
	"mysql": []string{
		"parseTime", "true",
		"loc", "Local",
		"sql_mode", "ansi",
	},
	"mymysql":     []string{"sql_mode", "ansi"},
	"sqlite3":     []string{"loc", "auto"},
	"cockroachdb": []string{"sslmode", "disable"},
}

// ForceParams forces connection parameters on a database URL.
//
// Note: also forces/sets the username/password when a matching entry exists in
// the PASS file.
func (h *Handler) ForceParams(u *dburl.URL) {
	var z *dburl.URL

	// force driver parameters
	fp, ok := forceParamMap[u.Driver]
	if !ok {
		fp = forceParamMap[u.Scheme]
	}
	if len(fp) != 0 {
		v := u.Query()
		for i := 0; i < len(fp); i += 2 {
			v.Set(fp[i], fp[i+1])
		}
		u.RawQuery = v.Encode()
	}

	// see if password entry is present
	user, err := env.PassFileEntry(u, h.user)
	if err != nil {
		errout := h.l.Stderr()
		fmt.Fprintf(errout, "error: %v", err)
		fmt.Fprintln(errout)
	} else if user != nil {
		u.User = user
	}

	// copy back to u
	z, _ = dburl.Parse(u.String())
	*u = *z
}

// Password collects a password from input, and returning a modified DSN
// including the collected password.
func (h *Handler) Password(dsn string) (string, error) {
	var err error

	if dsn == "" {
		return "", ErrMissingDSN
	}

	u, err := dburl.Parse(dsn)
	if err != nil {
		return "", err
	}

	user := h.user.Username
	if u.User != nil {
		user = u.User.Username()
	}
	pass, err := h.l.Password()
	if err != nil {
		return "", err
	}

	u.User = url.UserPassword(user, pass)
	return u.String(), nil
}

// Close closes the database connection if it is open.
func (h *Handler) Close() error {
	if h.db != nil {
		err := h.db.Close()
		h.db, h.u = nil, nil
		return err
	}

	return nil
}

// Version prints the database version information after a successful connection.
func (h *Handler) Version() error {
	ver, err := drivers.Version(h.u, h.db)
	if err != nil {
		return err
	}

	if ver != "" {
		out := h.IO().Stdout()
		fmt.Fprintf(out, text.ConnInfo, h.u.Driver, ver)
		fmt.Fprintln(out)
	}

	return nil
}

// Execute executes a sql query against the connected database.
func (h *Handler) Execute(w io.Writer, prefix, sqlstr string) error {
	if h.db == nil {
		return ErrNotConnected
	}

	// determine type and pre process string
	typ, s, q, err := drivers.Process(h.u, prefix, sqlstr)
	if err != nil {
		return err
	}

	// exec or query
	f := h.Exec
	if q {
		f = h.Query
	}

	// exec
	return drivers.WrapErr(h.u.Driver, f(w, typ, s))
}

// Query executes a query against the database.
func (h *Handler) Query(w io.Writer, _, sqlstr string) error {
	var err error

	// run query
	q, err := h.db.Query(sqlstr)
	if err != nil {
		return err
	}
	defer q.Close()

	// output rows
	err = h.OutputRows(w, q)
	if err != nil {
		return err
	}

	// check for additional result sets ...
	for drivers.NextResultSet(q) {
		err = h.OutputRows(w, q)
		if err != nil {
			return err
		}
	}

	return nil
}

// OutputRows outputs the supplied SQL rows to the supplied writer.
func (h *Handler) OutputRows(w io.Writer, q *sql.Rows) error {
	// get column names
	cols, err := drivers.Columns(h.u, q)
	if err != nil {
		return err
	}

	// create output table
	t := tablewriter.NewWriter(w)
	t.SetAutoFormatHeaders(false)
	t.SetBorder(false)
	t.SetAutoWrapText(false)
	t.SetHeader(cols)

	clen := len(cols)
	var rows int
	for q.Next() {
		if clen != 0 {
			r := make([]interface{}, clen)
			for i := range r {
				r[i] = new(interface{})
			}

			err = q.Scan(r...)
			if err != nil {
				return err
			}

			row := make([]string, clen)
			for n, z := range r {
				j := z.(*interface{})
				//log.Printf(">>> %s: %s", cols[n], reflect.TypeOf(*j))
				switch x := (*j).(type) {
				case []byte:
					row[n] = drivers.ConvertBytes(h.u, x)

				case string:
					row[n] = x

				case time.Time:
					row[n] = x.Format(time.RFC3339Nano)

				case fmt.Stringer:
					row[n] = x.String()

				default:
					row[n] = fmt.Sprintf("%v", *j)
				}

			}
			t.Append(row)
			rows++
		}
	}

	t.Render()

	// row count
	fmt.Fprintf(w, text.RowCount, rows)
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	return nil
}

// Exec does a database exec.
func (h *Handler) Exec(w io.Writer, typ, sqlstr string) error {
	var err error

	res, err := h.db.Exec(sqlstr)
	if err != nil {
		return err
	}

	// get affected
	count, err := drivers.RowsAffected(h.u, res)
	if err != nil {
		return err
	}

	// print name
	fmt.Fprint(w, typ)

	// print count
	if count > 0 {
		fmt.Fprint(w, " ", count)
	}

	fmt.Fprintln(w)

	return nil
}

// Include includes the specified path.
func (h *Handler) Include(path string, relative bool) error {
	var err error

	if relative && !filepath.IsAbs(path) {
		path = filepath.Join(h.wd, path)
	}

	path, err = filepath.EvalSymlinks(path)
	switch {
	case err != nil && os.IsNotExist(err):
		return ErrNoSuchFileOrDirectory
	case err != nil:
		return err
	}

	fi, err := os.Stat(path)
	switch {
	case err != nil && os.IsNotExist(err):
		return ErrNoSuchFileOrDirectory
	case err != nil:
		return err
	case fi.IsDir():
		return ErrCannotIncludeDirectories
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)

	l := &rline.Rline{
		N: func() ([]rune, error) {
			if !s.Scan() {
				err := s.Err()
				if err == nil {
					return nil, io.EOF
				}
				return nil, err
			}
			return []rune(s.Text()), nil
		},
		Out: h.l.Stdout(),
		Err: h.l.Stderr(),
		Pw: func() (string, error) {
			return h.l.Password()
		},
	}

	p := New(l, h.user, filepath.Dir(path), h.nopw)
	p.db, p.u = h.db, h.u

	err = p.Run()
	if err == io.EOF {
		err = nil
	}

	h.db, h.u = p.db, p.u
	return err
}
