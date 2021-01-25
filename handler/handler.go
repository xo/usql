package handler

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/styles"
	"github.com/xo/dburl"
	"github.com/xo/tblfmt"
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
	// timing of every command executed
	timing bool
	// singleLineMode is single line mode
	singleLineMode bool
	// query statement buffer
	buf *stmt.Stmt
	// last statement
	last       string
	lastPrefix string
	lastRaw    string
	// batch
	batch    bool
	batchEnd string
	// connection
	u  *dburl.URL
	db *sql.DB
	tx *sql.Tx
}

// New creates a new input handler.
func New(l rline.IO, user *user.User, wd string, nopw bool) *Handler {
	f, iactive := l.Next, l.Interactive()
	if iactive {
		f = func() ([]rune, error) {
			// next line
			r, err := l.Next()
			if err != nil {
				return nil, err
			}
			// save history
			_ = l.Save(string(r))
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

// SetSingleLineMode sets the single line mode toggle.
func (h *Handler) SetSingleLineMode(singleLineMode bool) {
	h.singleLineMode = singleLineMode
}

// GetTiming gets the timing toggle.
func (h *Handler) GetTiming() bool {
	return h.timing
}

// SetTiming sets the timing toggle.
func (h *Handler) SetTiming(timing bool) {
	h.timing = timing
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
	// leading whitespace
	var leading string
	// capture current query statement buffer
	orig := h.buf.RawString()
	full := orig
	if full != "" {
		full += "\n"
	} else {
		// get leading whitespace
		if i := strings.IndexFunc(s, func(r rune) bool {
			return !stmt.IsSpaceOrControl(r)
		}); i != -1 {
			leading = s[:i]
		}
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
loop:
	for {
		cmd, _, err = st.Next(env.Unquote(h.user, false, env.All()))
		switch {
		case err != nil && err != io.EOF:
			return s + endl
		case err == io.EOF:
			break loop
		}
		if st.Ready() || cmd != "" {
			final += st.RawString()
			st.Reset(nil)
			// grab remaining whitespace to add to final
			l := len(final)
			// find first non empty character
			if i := strings.IndexFunc(full[l:], func(r rune) bool {
				return !stmt.IsSpaceOrControl(r)
			}); i != -1 {
				final += full[l : l+i]
			}
		}
	}
	if !st.Ready() && cmd == "" {
		final += st.RawString()
	}
	final = leading + final
	// determine whatever is remaining after "active"
	var remaining string
	if fnl := len(final); fnl < len(full) {
		remaining = full[fnl:]
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

// helpQuitExitRE is a regexp to use to match help, quit, or exit messages.
var helpQuitExitRE = regexp.MustCompile(fmt.Sprintf(`(?im)^(%s|%s|%s)\s*$`, text.HelpPrefix, text.QuitPrefix, text.ExitPrefix))

// Run executes queries and commands.
func (h *Handler) Run() error {
	stdout, stderr, iactive := h.l.Stdout(), h.l.Stderr(), h.l.Interactive()
	// display welcome info
	if iactive {
		fmt.Fprintln(h.l.Stdout(), text.WelcomeDesc)
		fmt.Fprintln(h.l.Stdout())
	}
	var lastErr error
	for {
		var execute bool
		// set prompt
		if iactive {
			h.l.Prompt(h.Prompt())
		}
		// read next statement/command
		cmd, paramstr, err := h.buf.Next(env.Unquote(h.user, false, env.All()))
		switch {
		case h.singleLineMode && err == nil:
			execute = h.buf.Len != 0
		case err == rline.ErrInterrupt:
			h.buf.Reset(nil)
			continue
		case err != nil:
			if err == io.EOF {
				return lastErr
			}
			return err
		}
		var res metacmd.Result
		if cmd != "" {
			cmd = strings.TrimPrefix(cmd, `\`)
			params := stmt.DecodeParams(paramstr)
			// decode
			r, err := metacmd.Decode(cmd, params)
			if err != nil {
				lastErr = WrapErr(cmd, err)
				switch {
				case err == text.ErrUnknownCommand:
					fmt.Fprintf(stderr, text.InvalidCommand, cmd)
				case err == text.ErrMissingRequiredArgument:
					fmt.Fprintf(stderr, text.MissingRequiredArg, cmd)
				default:
					fmt.Fprintf(stderr, "error: %v", err)
				}
				fmt.Fprintln(stderr)
				continue
			}
			// run
			res, err = r.Run(h)
			if err != nil && err != rline.ErrInterrupt {
				lastErr = WrapErr(cmd, err)
				fmt.Fprintf(stderr, "error: %v", err)
				fmt.Fprintln(stderr)
				continue
			}
			// print unused command parameters
			for {
				ok, arg, err := params.Get(func(s string, isvar bool) (bool, string, error) {
					return true, s, nil
				})
				if err != nil {
					fmt.Fprintf(stderr, "error: %v", err)
					fmt.Fprintln(stderr)
				}
				if !ok {
					break
				}
				fmt.Fprintf(stdout, text.ExtraArgumentIgnored, cmd, arg)
				fmt.Fprintln(stdout)
			}
		}
		// help, exit, quit intercept
		if iactive && len(h.buf.Buf) >= 4 {
			i, first := stmt.RunesLastIndex(h.buf.Buf, '\n'), false
			if i == -1 {
				i, first = 0, true
			}
			if s := strings.ToLower(helpQuitExitRE.FindString(string(h.buf.Buf[i:]))); s != "" {
				switch s {
				case "help":
					s = text.HelpDescShort
					if first {
						s = text.HelpDesc
						h.buf.Reset(nil)
					}
				case "quit", "exit":
					s = text.QuitDesc
					if first {
						return nil
					}
				}
				fmt.Fprintln(stdout, s)
			}
		}
		// quit
		if res.Quit {
			return nil
		}
		// execute buf
		if execute || h.buf.Ready() || res.Exec != metacmd.ExecNone {
			// intercept batch query
			if h.u != nil {
				typ, end, batch := drivers.IsBatchQueryPrefix(h.u, h.buf.Prefix)
				switch {
				case h.batch && batch:
					err = fmt.Errorf("cannot perform %s in existing batch", typ)
					lastErr = WrapErr(h.buf.String(), err)
					fmt.Fprintf(stderr, "error: %v", err)
					fmt.Fprintln(stderr)
					continue
				// cannot use \g* while accumulating statements for batch queries
				case h.batch && typ != h.batchEnd && res.Exec != metacmd.ExecNone:
					err = errors.New("cannot force batch execution")
					lastErr = WrapErr(h.buf.String(), err)
					fmt.Fprintf(stderr, "error: %v", err)
					fmt.Fprintln(stderr)
					continue
				case batch:
					h.batch, h.batchEnd = true, end
				case h.batch:
					var lend string
					if len(h.last) != 0 {
						lend = "\n"
					}
					// append to last
					h.last += lend + h.buf.String()
					h.lastPrefix = h.buf.Prefix
					h.lastRaw += lend + h.buf.RawString()
					h.buf.Reset(nil)
					// break
					if h.batchEnd != typ {
						continue
					}
					h.lastPrefix = h.batchEnd
					h.batch, h.batchEnd = false, ""
				}
			}
			if h.buf.Len != 0 {
				h.last, h.lastPrefix, h.lastRaw = h.buf.String(), h.buf.Prefix, h.buf.RawString()
				h.buf.Reset(nil)
			}
			// log.Printf(">> PROCESS EXECUTE: (%s) `%s`", h.lastPrefix, h.last)
			if !h.batch && h.last != "" && h.last != ";" {
				// force a transaction for batched queries for certain drivers
				var forceBatch bool
				if h.u != nil {
					_, _, forceBatch = drivers.IsBatchQueryPrefix(h.u, stmt.FindPrefix(h.last))
					forceBatch = forceBatch && drivers.BatchAsTransaction(h.u)
				}
				// execute
				if err = h.Execute(stdout, res, h.lastPrefix, h.last, forceBatch); err != nil {
					lastErr = WrapErr(h.last, err)
					fmt.Fprintf(stderr, "error: %v", err)
					fmt.Fprintln(stderr)
				}
			}
		}
	}
}

// Execute executes a query against the connected database.
func (h *Handler) Execute(w io.Writer, res metacmd.Result, prefix, qstr string, forceTrans bool) error {
	if h.db == nil {
		return text.ErrNotConnected
	}
	// determine type and pre process string
	prefix, qstr, qtyp, err := drivers.Process(h.u, prefix, qstr)
	if err != nil {
		return drivers.WrapErr(h.u.Driver, err)
	}
	// start a transaction if forced
	if forceTrans {
		if err = h.Begin(); err != nil {
			return err
		}
	}
	f := h.execOnly
	switch res.Exec {
	case metacmd.ExecSet:
		f = h.execSet
	case metacmd.ExecExec:
		f = h.execExec
	}
	if err = drivers.WrapErr(h.u.Driver, f(w, prefix, qstr, qtyp, res.ExecParam, res.Expanded)); err != nil {
		if forceTrans {
			defer h.tx.Rollback()
			h.tx = nil
		}
		return err
	}
	if forceTrans {
		return h.Commit()
	}
	return nil
}

// Reset resets the handler's query statement buffer.
func (h *Handler) Reset(r []rune) {
	h.buf.Reset(r)
	h.last, h.lastPrefix, h.lastRaw, h.batch, h.batchEnd = "", "", "", false, ""
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
	if h.tx != nil || h.batch {
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

// LastRaw returns the last raw (non-interpolated) executed statement.
func (h *Handler) LastRaw() string {
	return h.lastRaw
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
			if fi, err = os.Stat(urlstr); err != nil {
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
	switch {
	case err != nil:
		errout := h.l.Stderr()
		fmt.Fprintf(errout, "error: %v", err)
		fmt.Fprintln(errout)
	case user != nil:
		u.User = user
	}
	// copy back to u
	z, _ := dburl.Parse(u.String())
	*u = *z
}

// Password collects a password from input, and returns a modified DSN
// including the collected password.
func (h *Handler) Password(dsn string) (string, error) {
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
	switch typ {
	case "int":
		_, err = strconv.ParseInt(v, 10, 64)
	case "uint":
		_, err = strconv.ParseUint(v, 10, 64)
	case "float":
		_, err = strconv.ParseFloat(v, 64)
	case "bool":
		var b bool
		b, err = strconv.ParseBool(v)
		if err == nil {
			v = fmt.Sprintf("%v", b)
		}
	}
	if err != nil {
		errstr := err.Error()
		if i := strings.LastIndex(errstr, ":"); i != -1 {
			errstr = strings.TrimSpace(errstr[i+1:])
		}
		return "", fmt.Errorf(text.InvalidValue, typ, v, errstr)
	}
	return v, nil
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
	if err = drivers.CanChangePassword(h.u); err != nil {
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
		if newpw, err = h.l.Password(text.NewPassword); err != nil {
			return "", err
		}
		if newpw2, err = h.l.Password(text.ConfirmPassword); err != nil {
			return "", err
		}
		if newpw == newpw2 {
			break
		}
		fmt.Fprintln(h.l.Stderr(), text.PasswordsDoNotMatch)
	}
	// verify passwords match
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
	s := env.Timefmt()
	if s == "" {
		s = time.RFC3339Nano
	}
	return s
}

// execOnly executes a query against the database.
func (h *Handler) execOnly(w io.Writer, prefix, qstr string, qtyp bool, _ string, expanded bool) error {
	// exec or query
	f := h.exec
	if qtyp {
		f = h.query
	}
	// exec
	return f(w, prefix, qstr, expanded)
}

// execSet executes a SQL query, setting all returned columns as variables.
func (h *Handler) execSet(w io.Writer, prefix, qstr string, _ bool, namePrefix string, _ bool) error {
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
		_ = env.Set(n, row[i])
	}
	return nil
}

// execExec executes a query and re-executes all columns of all rows as if they
// were their own queries.
func (h *Handler) execExec(w io.Writer, prefix, qstr string, qtyp bool, _ string, _ bool) error {
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
func (h *Handler) query(w io.Writer, _, qstr string, expanded bool) error {
	start := time.Now()
	// run query
	q, err := h.DB().Query(qstr)
	if err != nil {
		return err
	}
	defer q.Close()
	params := env.Pall()
	if expanded {
		params["expanded"] = "on"
	}
	if err = tblfmt.EncodeAll(w, q, params); err != nil {
		return err
	}
	if h.timing {
		d := time.Since(start)
		fmt.Fprintf(h.l.Stdout(), text.TimingDesc, float64(d.Microseconds())/1000)
		if d > 1*time.Second {
			fmt.Fprintf(h.l.Stdout(), " (%v)", d.Round(1*time.Millisecond))
		}
		fmt.Fprintln(h.l.Stdout())
	}
	return err
}

// execRows executes all the columns in the row.
func (h *Handler) execRows(w io.Writer, q *sql.Rows) error {
	// get columns
	cols, err := drivers.Columns(h.u, q)
	if err != nil {
		return err
	}
	// process rows
	res := metacmd.Result{Exec: metacmd.ExecOnly}
	clen, tfmt := len(cols), h.timefmt()
	for q.Next() {
		if clen != 0 {
			row, err := h.scan(q, clen, tfmt)
			if err != nil {
				return err
			}
			// execute
			for _, qstr := range row {
				if err = h.Execute(w, res, stmt.FindPrefix(qstr), qstr, false); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// scan scans a row.
func (h *Handler) scan(q *sql.Rows, clen int, tfmt string) ([]string, error) {
	var err error
	// scan to []interface{}
	r := make([]interface{}, clen)
	for i := range r {
		r[i] = new(interface{})
	}
	if err = q.Scan(r...); err != nil {
		return nil, err
	}
	// get conversion funcs
	cb, cm, cs, cd := drivers.ConvertBytes(h.u), drivers.ConvertMap(h.u), drivers.ConvertSlice(h.u), drivers.ConvertDefault(h.u)
	row := make([]string, clen)
	for n, z := range r {
		j := z.(*interface{})
		switch x := (*j).(type) {
		case []byte:
			if x != nil {
				row[n], err = cb(x, tfmt)
				if err != nil {
					return nil, err
				}
			}
		case string:
			row[n] = x
		case time.Time:
			row[n] = x.Format(tfmt)
		case fmt.Stringer:
			row[n] = x.String()
		case map[string]interface{}:
			if x != nil {
				row[n], err = cm(x)
				if err != nil {
					return nil, err
				}
			}
		case []interface{}:
			if x != nil {
				row[n], err = cs(x)
				if err != nil {
					return nil, err
				}
			}
		default:
			if x != nil {
				row[n], err = cd(x)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return row, err
}

// exec does a database exec.
func (h *Handler) exec(w io.Writer, typ, qstr string, _ bool) error {
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
	drivers.ConfigStmt(p.u, p.buf)
	err = p.Run()
	h.db, h.u = p.db, p.u
	return err
}
