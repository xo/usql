package handler

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/knq/dburl"
	"github.com/olekukonko/tablewriter"

	"github.com/knq/usql/drivers"
	"github.com/knq/usql/metacmd"
	"github.com/knq/usql/rline"
	"github.com/knq/usql/stmt"
	"github.com/knq/usql/text"
)

// Handler is a input process handler.
type Handler struct {
	l    rline.IO
	user *user.User
	wd   string

	// statement buffer
	buf        *stmt.Stmt
	lastPrefix string
	last       string

	// connection
	u  *dburl.URL
	db *sql.DB
}

// New creates a new input handler.
func New(l rline.IO, user *user.User, wd string) *Handler {
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

	h := &Handler{
		l:    l,
		user: user,
		wd:   wd,
		buf:  stmt.New(f),
	}

	return h
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
		var exitWithErr error

		// set prompt
		if iactive {
			h.l.Prompt(h.Prompt())
		}

		// read next statement/command
		cmd, params, err := h.buf.Next()
		switch {
		case !iactive && err == io.EOF:
			execute, exitWithErr = true, io.EOF

		case err == rline.ErrInterrupt:
			h.buf.Reset()
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
				h.buf.Reset()
			}

			//log.Printf(">> PROCESS EXECUTE: (%) `%s`", last)
			if h.last != "" && h.last != ";" {
				err = h.Execute(stdout, h.lastPrefix, h.last)
				if err != nil {
					fmt.Fprintf(stderr, "error: %v", err)
					fmt.Fprintln(stderr)
				}
			}

			execute = false
		}

		if exitWithErr != nil {
			return exitWithErr
		}
	}
}

// Reset resets the handler's statement buffer.
func (h *Handler) Reset() {
	h.buf.Reset()
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
// joined (with a space) and passed as the DSN to sql.Open.
func (h *Handler) Open(params ...string) error {
	if len(params) == 0 {
		return nil
	}

	var err error
	if len(params) < 2 {
		urlstr := params[0]
		if urlstr == "" {
			return nil
		}

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
	} else {
		h.u = &dburl.URL{
			Driver: params[0],
			DSN:    strings.Join(params[1:], " "),
		}
	}

	// check driver
	if _, ok := drivers.Drivers[h.u.Driver]; !ok {
		return &Error{
			Driver: h.u.Driver,
			Err:    ErrDriverNotAvailable,
		}
	}

	// force connection parameters for drivers that are "url" style DSNs
	h.u.DSN = addQueryParam(h.u.Driver, "mysql", h.u.DSN, "parseTime", "true")
	h.u.DSN = addQueryParam(h.u.Driver, "mysql", h.u.DSN, "loc", "Local")
	h.u.DSN = addQueryParam(h.u.Driver, "mysql", h.u.DSN, "sql_mode", "ansi")
	h.u.DSN = addQueryParam(h.u.Driver, "sqlite3", h.u.DSN, "loc", "auto")
	h.u.DSN = addQueryParam(h.u.Scheme, "cockroachdb", h.u.DSN, "sslmode", "disable")

	// force connection parameter for mymysql
	if h.u.Driver == "mymysql" {
		q := h.u.Query()
		q.Set("sql_mode", "ansi")
		h.u.RawQuery = q.Encode()
		h.u.DSN, _ = dburl.GenMyMySQL(h.u)
	}

	// use special open func for pgx
	f := sql.Open
	if h.u.Driver == "pgx" {
		f = drivers.PgxOpen(h.u)
	}

	// force statement parse settings
	isPG := h.u.Driver == "postgres" || h.u.Driver == "pgx"
	stmt.AllowDollar(isPG)(h.buf)
	stmt.AllowMultilineComments(isPG)(h.buf)

	// connect
	h.db, err = f(h.u.Driver, h.u.DSN)
	if err != nil && !drivers.IsPasswordErr(h.u.Driver, err) {
		return err
	} else if err == nil {
		// do ping to force an error (if any)
		err = h.db.Ping()
		if err == nil {
			return nil
		}
	}

	// bail without getting password
	if !drivers.IsPasswordErr(h.u.Driver, err) || len(params) > 1 || !h.l.Interactive() {
		h.Close()
		return h.WrapError(err)
	}

	// print the error
	fmt.Fprintf(h.l.Stderr(), "error: %v", h.WrapError(err))
	fmt.Fprintln(h.l.Stderr())

	// otherwise, try to collect a password ...
	user := h.user.Username
	if h.u.User != nil {
		user = h.u.User.Username()
	}
	pass, err := h.l.Password()
	if err != nil {
		// close connection
		h.Close()
		return err
	}

	// reconnect using the user/pass ...
	h.u.User = url.UserPassword(user, pass)
	return h.Open(h.u.String())
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

//var beginTransactionRE = regexp.MustCompile(`(?i)^BEGIN\s*TRANSACTION;?$`)

// Execute executes a sql query against the connected database.
func (h *Handler) Execute(w io.Writer, prefix, sqlstr string) error {
	if h.db == nil {
		return ErrNotConnected
	}

	// determine if query or exec
	typ, q := h.ProcessPrefix(prefix, sqlstr)
	f := h.Exec
	if q {
		f = h.Query
	}

	switch h.u.Driver {
	case "ora":
		sqlstr = strings.TrimSuffix(sqlstr, ";")

		//	case "ql":
		//		if typ == "BEGIN" && beginTransactionRE.MatchString(sqlstr) {
		//			log.Printf("GOT BEGIN TRANSACTION")
		//			//tx, err := h.db.Begin()
		//			/*if err != nil {
		//				return err
		//			}*/
		//		}
		//		if typ == "COMMIT" {
		//
		//		}
	}

	//log.Printf(">>>> EXECUTE: %s", runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())

	// exec
	return h.WrapError(f(w, typ, sqlstr))
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
	for nextResultSet(q) {
		err = h.OutputRows(w, q)
		if err != nil {
			return err
		}
	}

	return nil
}

var allcapsRE = regexp.MustCompile(`^[A-Z_]+$`)

// OutputRows outputs the supplied SQL rows to the supplied writer.
func (h *Handler) OutputRows(w io.Writer, q *sql.Rows) error {
	// get column names
	cols, err := q.Columns()
	if err != nil {
		return err
	}

	// fix display column names
	for i, s := range cols {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			cols[i] = fmt.Sprintf("col%d", i)
		}

		// fix case on oracle column names
		if h.u.Driver == "ora" && allcapsRE.MatchString(cols[i]) {
			cols[i] = strings.ToLower(cols[i])
		}
	}

	// create output table
	t := tablewriter.NewWriter(w)
	t.SetAutoFormatHeaders(false)
	t.SetBorder(false)
	t.SetAutoWrapText(false)
	t.SetHeader(cols)

	clen := len(cols)
	var rows int
	if clen != 0 {
		for q.Next() {
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
					if h.u.Driver == "sqlite3" {
						row[n] = drivers.Sqlite3Parse(x)
					} else {
						row[n] = string(x)
					}

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
	fmt.Fprintln(w, "\n")

	return nil
}

// Exec does a database exec.
func (h *Handler) Exec(w io.Writer, typ, sqlstr string) error {
	res, err := h.db.Exec(sqlstr)
	if err != nil {
		return err
	}

	// get count
	var count int64
	if h.u.Driver != "adodb" {
		count, err = res.RowsAffected()
		if err != nil {
			return err
		}
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

// WrapError conditionally wraps an error if the error occurs while connected
// to a database.
func (h *Handler) WrapError(err error) error {
	if err == nil {
		return nil
	}

	if h.db != nil {
		// attempt to clean up and standardize errors
		driver := h.u.Driver
		if s, ok := drivers.Drivers[driver]; ok {
			driver = s
		}

		return &Error{driver, err}
	}

	return err
}

// ProcessPrefix processes a prefix.
func (h *Handler) ProcessPrefix(prefix, sqlstr string) (string, bool) {
	if prefix == "" {
		return "EXEC", false
	}

	s := strings.Split(prefix, " ")
	if len(s) > 0 {
		// check query map
		if _, ok := queryMap[s[0]]; ok {
			typ := s[0]
			switch {
			case typ == "SELECT" && len(s) >= 2 && s[1] == "INTO":
				return "SELECT INTO", false
			case typ == "PRAGMA":
				return typ, !strings.ContainsRune(sqlstr, '=')
			}
			return typ, true
		}

		// find longest match
		for i := len(s); i > 0; i-- {
			typ := strings.Join(s[:i], " ")
			if _, ok := execMap[typ]; ok {
				return typ, false
			}
		}
	}

	return "EXEC", false
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

	p := New(l, h.user, filepath.Dir(path))
	p.db, p.u = h.db, h.u

	err = p.Run()
	if err == io.EOF {
		err = nil
	}

	h.db, h.u = p.db, p.u
	return err
}
