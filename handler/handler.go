package handler

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/styles"
	"github.com/olekukonko/tablewriter"
	"github.com/xo/dburl"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
	"github.com/xo/usql/metacmd"
	"github.com/xo/usql/rline"
	"github.com/xo/usql/stmt"
	ustyles "github.com/xo/usql/styles"
	"github.com/xo/usql/text"
)

// Handler is a input process handler.
type Handler struct {
	l    rline.IO
	user *user.User
	wd   string
	nopw bool

	// slm is single line mode
	slm bool

	// query statement buffer
	buf        *stmt.Stmt
	lastPrefix string
	last       string

	// connection
	u  *dburl.URL
	db *sql.DB
	tx *sql.Tx
}

// New creates a new input handler.
func New(l rline.IO, user *user.User, wd string, nopw bool) *Handler {
	// set help intercept
	f, iactive := l.Next, l.Interactive()
	if iactive {
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
		nopw: nopw,
		buf:  stmt.New(f),
	}

	if iactive {
		l.SetOutput(h.outputHighlighter)
	}

	return h
}

// outputHighlighter returns s as a highlighted string, based on the current
// buffer and syntax highlighting settings.
func (h *Handler) outputHighlighter(s string) string {
	// bail when string is empty (ie, contains no printable, non-space
	// characters) or if syntax highlighting is not enabled
	if empty(s) || env.All()["SYNTAX_HL"] != "true" {
		return s
	}

	// count end lines
	var endl string
	for strings.HasSuffix(s, lineterm) {
		s = strings.TrimSuffix(s, lineterm)
		endl += lineterm
	}

	// capture current query statement buffer
	orig := h.buf.String()
	full := orig
	if full != "" {
		full += "\n"
	}
	full += s

	// setup statement parser
	st := drivers.NewStmt(h.u, func() func() ([]rune, error) {
		y := strings.Split(orig, "\n")
		if y[0] == "" {
			y[0] = s
		} else {
			y = append(y, s)
		}

		return func() ([]rune, error) {
			if len(y) > 0 {
				z := y[0]
				y = y[1:]
				return []rune(z), nil
			}
			return nil, io.EOF
		}
	}())

	// accumulate all "active" statements in buffer, breaking either at
	// EOF or when a \ cmd has been encountered
	var err error
	var cmd, final string
	for {
		cmd, _, err = st.Next()
		if err != nil && err != io.EOF {
			return s + endl
		} else if err == io.EOF {
			break
		}

		if st.Ready() || cmd != "" {
			final += st.String()
			st.Reset(nil)

			// grab remaining whitespace to add to final
			l := len(final)
			if i := strings.IndexFunc(full[l:], func(r rune) bool { return !unicode.IsSpace(r) }); i != -1 {
				final += full[l : l+i]
			}
		}
	}
	if !st.Ready() && cmd == "" {
		final += st.String()
	}

	// determine whatever is remaining after "active"
	var remaining string
	if len(final) < len(full) {
		remaining = full[len(final):]
	}

	// this happens when a read line is empty and/or has only
	// whitespace and a \ cmd
	if s == remaining {
		return s + endl
	}

	// highlight entire final accumulated buffer
	b := new(bytes.Buffer)
	if err := h.Highlight(b, final); err != nil {
		return s + endl
	}
	colored := b.String()

	// return only last line plus whatever remaining string (ie, after
	// a \ cmd) and the end line count
	ss := strings.Split(colored, "\n")
	return lastcolor(colored) + ss[len(ss)-1] + remaining + endl
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
		case h.slm && err == nil:
			execute = h.buf.Len != 0

		case err == rline.ErrInterrupt:
			h.buf.Reset(nil)
			continue

		case err != nil:
			return err
		}

		var res metacmd.Res
		if cmd != "" {
			cmd = strings.TrimPrefix(cmd, `\`)

			// decode
			var r metacmd.Runner
			r, err = metacmd.Decode(cmd, params)
			switch {
			case err == text.ErrUnknownCommand:
				fmt.Fprintf(stderr, text.InvalidCommand, cmd)
				fmt.Fprintln(stderr)
				continue

			case err == text.ErrMissingRequiredArgument:
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
				err = h.Execute(stdout, res, h.lastPrefix, h.last)
				if err != nil {
					fmt.Fprintf(stderr, "error: %v", err)
					fmt.Fprintln(stderr)
				}
			}
		}
	}
}

// Execute executes a query against the connected database.
func (h *Handler) Execute(w io.Writer, res metacmd.Res, prefix, qstr string) error {
	if h.db == nil {
		return text.ErrNotConnected
	}

	// determine type and pre process string
	prefix, qstr, qtyp, err := drivers.Process(h.u, prefix, qstr)
	if err != nil {
		return drivers.WrapErr(h.u.Driver, err)
	}

	f := h.execOnly
	switch res.Exec {
	case metacmd.ExecSet:
		f = h.execSet
	case metacmd.ExecExec:
		f = h.execExec
	}

	return drivers.WrapErr(h.u.Driver, f(w, prefix, qstr, qtyp, res.ExecParam))
}

// CommandRunner executes a set of commands.
func (h *Handler) CommandRunner(cmds []string) func() error {
	h.slm = true
	return func() error {
		for _, cmd := range cmds {
			h.Reset([]rune(cmd))
			err := h.Run()
			if err != nil && err != io.EOF {
				return err
			}
		}
		return nil
	}
}

// Reset resets the handler's query statement buffer.
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

	tx := ">"
	if h.tx != nil {
		tx = "~"
	}

	return s + h.buf.State() + tx + " "
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
func (h *Handler) DB() drivers.DB {
	if h.tx != nil {
		return h.tx
	}

	return h.db
}

// Last returns the last executed statement.
func (h *Handler) Last() string {
	return h.last
}

// Buf returns the current query statement buffer.
func (h *Handler) Buf() *stmt.Stmt {
	return h.buf
}

// Highlight highlights using the current environment settings.
func (h *Handler) Highlight(w io.Writer, buf string) error {
	vars := env.All()

	// create lexer, formatter, styler
	l := chroma.Coalesce(drivers.Lexer(h.u))
	f := formatters.Get(vars["SYNTAX_HL_FORMAT"])
	s := styles.Get(vars["SYNTAX_HL_STYLE"])

	// override background
	if vars["SYNTAX_HL_OVERRIDE_BG"] != "false" {
		s = ustyles.Get(vars["SYNTAX_HL_STYLE"])
	}

	// tokenize stream
	it, err := l.Tokenise(nil, buf)
	if err != nil {
		return err
	}

	// write formatted output
	return f.Format(w, s, it)
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

	if h.tx != nil {
		return text.ErrPreviousTransactionExists
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
		h.forceParams(h.u)
	} else {
		h.u = &dburl.URL{
			Driver: params[0],
			DSN:    strings.Join(params[1:], " "),
		}
	}

	// open connection
	h.db, err = drivers.Open(h.u)
	if err != nil && !drivers.IsPasswordErr(h.u, err) {
		defer h.Close()
		return err
	}

	// set buffer options
	drivers.ConfigStmt(h.u, h.buf)

	// force error/check connection
	if err == nil {
		if err = drivers.Ping(h.u, h.db); err == nil {
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

// forceParams forces connection parameters on a database URL, adding any
// driver specific required parameters, and the username/password when a
// matching entry exists in the PASS file.
func (h *Handler) forceParams(u *dburl.URL) {
	// force driver parameters
	drivers.ForceParams(u)

	// see if password entry is present
	user, err := env.PassFileEntry(h.user, u)
	if err != nil {
		errout := h.l.Stderr()
		fmt.Fprintf(errout, "error: %v", err)
		fmt.Fprintln(errout)
	} else if user != nil {
		u.User = user
	}

	// copy back to u
	z, _ := dburl.Parse(u.String())
	*u = *z
}

// Password collects a password from input, and returns a modified DSN
// including the collected password.
func (h *Handler) Password(dsn string) (string, error) {
	var err error

	if dsn == "" {
		return "", text.ErrMissingDSN
	}

	u, err := dburl.Parse(dsn)
	if err != nil {
		return "", err
	}

	user := h.user.Username
	if u.User != nil {
		user = u.User.Username()
	}
	pass, err := h.l.Password(text.EnterPassword)
	if err != nil {
		return "", err
	}

	u.User = url.UserPassword(user, pass)
	return u.String(), nil
}

// Close closes the database connection if it is open.
func (h *Handler) Close() error {
	if h.tx != nil {
		return text.ErrPreviousTransactionExists
	}

	if h.db != nil {
		err := h.db.Close()
		drv := h.u.Driver
		h.db, h.u = nil, nil
		return drivers.WrapErr(drv, err)
	}

	return nil
}

// ReadVar reads a variable from the interactive prompt, saving it to
// environment variables.
func (h *Handler) ReadVar(typ, prompt string) (string, error) {
	if !h.l.Interactive() {
		return "", text.ErrNotInteractive
	}

	var masked bool
	// check type
	switch typ {
	case "password":
		masked = true
	case "string", "int", "uint", "float", "bool":
	default:
		return "", text.ErrInvalidType
	}

	var v string
	var err error
	if masked {
		if prompt == "" {
			prompt = text.EnterPassword
		}
		v, err = h.l.Password(prompt)
	} else {
		h.l.Prompt(prompt)
		var r []rune
		r, err = h.l.Next()
		v = string(r)
	}

	var z interface{} = v
	switch typ {
	case "int":
		z, err = strconv.ParseInt(v, 10, 64)
	case "uint":
		z, err = strconv.ParseUint(v, 10, 64)
	case "float":
		z, err = strconv.ParseFloat(v, 64)
	case "bool":
		z, err = strconv.ParseBool(v)
	}
	if err != nil {
		return "", text.ErrInvalidValue
	}

	return fmt.Sprintf("%v", z), nil
}

// ChangePassword changes a password for the user.
func (h *Handler) ChangePassword(user string) (string, error) {
	if h.db == nil {
		return "", text.ErrNotConnected
	}

	if !h.l.Interactive() {
		return "", text.ErrNotInteractive
	}

	var err error

	err = drivers.CanChangePassword(h.u)
	if err != nil {
		return "", err
	}

	var newpw, newpw2, oldpw string

	// ask for previous password
	if user == "" && drivers.RequirePreviousPassword(h.u) {
		oldpw, err = h.l.Password(text.EnterPreviousPassword)
		if err != nil {
			return "", err
		}
	}

	// attempt to get passwords
	for i := 0; i < 3; i++ {
		newpw, err = h.l.Password(text.NewPassword)
		if err != nil {
			return "", err
		}
		newpw2, err = h.l.Password(text.ConfirmPassword)
		if err != nil {
			return "", err
		}

		if newpw == newpw2 {
			break
		}
		fmt.Fprintln(h.l.Stderr(), text.PasswordsDoNotMatch)
	}
	if newpw != newpw2 {
		return "", text.ErrPasswordAttemptsExhausted
	}

	return drivers.ChangePassword(h.u, h.DB(), user, newpw, oldpw)
}

// Version prints the database version information after a successful connection.
func (h *Handler) Version() error {
	if env.All()["SHOW_HOST_INFORMATION"] != "true" {
		return nil
	}

	if h.db == nil {
		return text.ErrNotConnected
	}

	ver, err := drivers.Version(h.u, h.DB())
	if err != nil {
		ver = fmt.Sprintf("<unknown, error: %v>", err)
	}

	if ver != "" {
		out := h.l.Stdout()
		fmt.Fprintf(out, text.ConnInfo, h.u.Driver, ver)
		fmt.Fprintln(out)
	}

	return nil
}

// timefmt returns the current time format setting.
func (h *Handler) timefmt() string {
	s, ok := env.All()["TIME_FORMAT"]
	if !ok || s == "" {
		s = time.RFC3339Nano
	}
	return s
}

// execOnly executes a query against the database.
func (h *Handler) execOnly(w io.Writer, prefix, qstr string, qtyp bool, _ string) error {
	// exec or query
	f := h.exec
	if qtyp {
		f = h.query
	}

	// exec
	return f(w, prefix, qstr)
}

// execSet executes a SQL query, setting all returned columns as variables.
func (h *Handler) execSet(w io.Writer, prefix, qstr string, _ bool, namePrefix string) error {
	// query
	q, err := h.DB().Query(qstr)
	if err != nil {
		return err
	}

	// get cols
	cols, err := drivers.Columns(h.u, q)
	if err != nil {
		return err
	}

	// process row(s)
	var i int
	var row []string
	clen, tfmt := len(cols), h.timefmt()
	for q.Next() {
		if i == 0 {
			row, err = h.scan(q, clen, tfmt)
			if err != nil {
				return err
			}
		}
		i++
	}
	if i > 1 {
		return text.ErrTooManyRows
	}

	// set vars
	for i, c := range cols {
		n := namePrefix + c
		if err = env.ValidIdentifier(n); err != nil {
			return fmt.Errorf(text.CouldNotSetVariable, n)
		}
		env.Set(n, row[i])
	}

	return nil
}

// execExec executes a query and re-executes all columns of all rows as if they
// were their own queries.
func (h *Handler) execExec(w io.Writer, prefix, qstr string, qtyp bool, _ string) error {
	// query
	q, err := h.DB().Query(qstr)
	if err != nil {
		return err
	}

	// execRows
	err = h.execRows(w, q)
	if err != nil {
		return err
	}

	// check for additional result sets ...
	for drivers.NextResultSet(q) {
		err = h.execRows(w, q)
		if err != nil {
			return err
		}
	}

	return nil
}

// query executes a query against the database.
func (h *Handler) query(w io.Writer, _, qstr string) error {
	var err error

	// run query
	q, err := h.DB().Query(qstr)
	if err != nil {
		return err
	}
	defer q.Close()

	// output rows
	err = h.outputRows(w, q)
	if err != nil {
		return err
	}

	// check for additional result sets ...
	for drivers.NextResultSet(q) {
		err = h.outputRows(w, q)
		if err != nil {
			return err
		}
	}

	return nil
}

// execRows executes all the columns in the row.
func (h *Handler) execRows(w io.Writer, q *sql.Rows) error {
	var err error

	// get columns
	cols, err := drivers.Columns(h.u, q)
	if err != nil {
		return err
	}

	// process rows
	res := metacmd.Res{Exec: metacmd.ExecOnly}
	clen, tfmt := len(cols), h.timefmt()
	for q.Next() {
		if clen != 0 {
			row, err := h.scan(q, clen, tfmt)
			if err != nil {
				return err
			}

			// execute
			for _, qstr := range row {
				err = h.Execute(w, res, stmt.FindPrefix(qstr, 4), qstr)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// outputRows outputs the supplied SQL rows to the supplied writer.
func (h *Handler) outputRows(w io.Writer, q *sql.Rows) error {
	// get columns
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

	// process rows
	var rows int
	clen, tfmt := len(cols), h.timefmt()
	for q.Next() {
		if clen != 0 {
			row, err := h.scan(q, clen, tfmt)
			if err != nil {
				return err
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

// scan scans a row.
func (h *Handler) scan(q *sql.Rows, clen int, tfmt string) ([]string, error) {
	r := make([]interface{}, clen)
	for i := range r {
		r[i] = new(interface{})
	}

	err := q.Scan(r...)
	if err != nil {
		return nil, err
	}

	row := make([]string, clen)
	for n, z := range r {
		j := z.(*interface{})
		//log.Printf(">>> %s: %s", cols[n], reflect.TypeOf(*j))
		switch x := (*j).(type) {
		case []byte:
			row[n] = drivers.ConvertBytes(h.u, tfmt, x)

		case string:
			row[n] = x

		case time.Time:
			row[n] = x.Format(tfmt)

		case fmt.Stringer:
			row[n] = x.String()

		default:
			row[n] = fmt.Sprintf("%v", *j)
		}
	}
	return row, err
}

// exec does a database exec.
func (h *Handler) exec(w io.Writer, typ, qstr string) error {
	var err error

	res, err := h.DB().Exec(qstr)
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

// Begin begins a transaction.
func (h *Handler) Begin() error {
	if h.db == nil {
		return text.ErrNotConnected
	}

	if h.tx != nil {
		return text.ErrPreviousTransactionExists
	}

	var err error
	h.tx, err = h.db.Begin()
	if err != nil {
		return drivers.WrapErr(h.u.Driver, err)
	}

	return nil
}

// Commit commits a transaction.
func (h *Handler) Commit() error {
	if h.db == nil {
		return text.ErrNotConnected
	}

	if h.tx == nil {
		return text.ErrNoPreviousTransactionExists
	}

	tx := h.tx
	h.tx = nil

	err := tx.Commit()
	if err != nil {
		return drivers.WrapErr(h.u.Driver, err)
	}

	return nil
}

// Rollback rollbacks a transaction.
func (h *Handler) Rollback() error {
	if h.db == nil {
		return text.ErrNotConnected
	}

	if h.tx == nil {
		return text.ErrNoPreviousTransactionExists
	}

	tx := h.tx
	h.tx = nil

	err := tx.Rollback()
	if err != nil {
		return drivers.WrapErr(h.u.Driver, err)
	}

	return nil
}

// Include includes the specified path.
func (h *Handler) Include(path string, relative bool) error {
	var err error

	if relative && !filepath.IsAbs(path) {
		path = filepath.Join(h.wd, path)
	}

	// read file
	path, f, err := env.OpenFile(h.user, path, relative)
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
		Pw:  h.l.Password,
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
