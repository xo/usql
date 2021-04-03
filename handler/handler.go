// Package handler provides a input process handler implementation for usql.
package handler

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/styles"
	"github.com/xo/dburl"
	"github.com/xo/tblfmt"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/completer"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/env"
	"github.com/xo/usql/metacmd"
	"github.com/xo/usql/rline"
	"github.com/xo/usql/stmt"
	ustyles "github.com/xo/usql/styles"
	"github.com/xo/usql/text"
)

// Handler is a input process handler.
//
// Glues together usql's components to provide a "read-eval-print loop" (REPL)
// for usql's interactive command-line and manages most of the core/high-level logic.
//
// Manages the active statement buffer, application IO, executing/querying SQL
// statements, and handles backslash (\) commands encountered in the input
// stream.
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
	// out file or pipe
	out io.WriteCloser
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
		var opt metacmd.Option
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
			opt, err = r.Run(h)
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
		if opt.Quit {
			if h.out != nil {
				h.out.Close()
			}
			return nil
		}
		// execute buf
		if execute || h.buf.Ready() || opt.Exec != metacmd.ExecNone {
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
				case h.batch && typ != h.batchEnd && opt.Exec != metacmd.ExecNone:
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
				out := stdout
				if h.out != nil {
					out = h.out
				}
				ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
				if err = h.Execute(ctx, out, opt, h.lastPrefix, h.last, forceBatch); err != nil {
					lastErr = WrapErr(h.last, err)
					fmt.Fprintf(stderr, "error: %v", err)
					fmt.Fprintln(stderr)
				}
				stop()
			}
		}
	}
}

// Execute executes a query against the connected database.
func (h *Handler) Execute(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, qstr string, forceTrans bool) error {
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
		if err = h.BeginTx(ctx); err != nil {
			return err
		}
	}
	f := h.execOnly
	switch opt.Exec {
	case metacmd.ExecExec:
		f = h.execExec
	case metacmd.ExecSet:
		f = h.execSet
	case metacmd.ExecWatch:
		f = h.execWatch
	}
	if err = drivers.WrapErr(h.u.Driver, f(ctx, w, opt, prefix, qstr, qtyp)); err != nil {
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
	// build a list of all possible connStrings for the completer
	connStrings := h.connStrings()

	if len(params) == 0 || params[0] == "" {
		h.l.Completer(completer.NewDefaultCompleter(completer.WithConnStrings(connStrings)))
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
			h.l.Completer(drivers.NewCompleter(h.u, h.db, completer.WithConnStrings(connStrings)))
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

func (h *Handler) connStrings() []string {
	available := drivers.Available()
	entries, err := env.PassFileEntries(h.user)
	if err != nil {
		// ignore the error as this is only used for completer
		// and it'll be reported again when trying to force params before opening a conn
		entries = nil
	}
	names := make([]string, 0, len(available)+len(entries))

	for schema := range available {
		_, aliases := dburl.SchemeDriverAndAliases(schema)
		// TODO should we create all combinations of space, :, :// and +transport ?
		names = append(names, schema)
		names = append(names, aliases...)
	}

	for _, e := range entries {
		if e.Protocol == "*" {
			continue
		}
		user := ""
		host := ""
		port := ""
		dbname := ""
		if e.Username != "*" {
			user = e.Username + "@"
			if e.Host != "*" {
				host = e.Host
				if e.Port != "*" {
					port = ":" + e.Port
				}
				if e.DBName != "*" {
					dbname = "/" + e.DBName
				}
			}
		}
		names = append(names, fmt.Sprintf("%s://%s%s%s%s", e.Protocol, user, host, port, dbname))
	}

	sort.Strings(names)
	return names
}

// forceParams forces connection parameters on a database URL, adding any
// driver specific required parameters, and the username/password when a
// matching entry exists in the PASS file.
func (h *Handler) forceParams(u *dburl.URL) {
	// force driver parameters
	drivers.ForceParams(u)
	// see if password entry is present
	user, err := env.PassFileMatch(h.user, u)
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

// execWatch repeatedly executes a query against the database.
func (h *Handler) execWatch(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, qstr string, qtyp bool) error {
	for {
		// this is the actual output that psql has: "Mon Jan 2006 3:04:05 PM MST"
		// fmt.Fprintf(w, "%s (every %fs)\n\n", time.Now().Format("Mon Jan 2006 3:04:05 PM MST"), float64(opt.Watch)/float64(time.Second))
		fmt.Fprintf(w, "%s (every %v)\n\n", time.Now().Format(time.RFC1123), opt.Watch)
		if err := h.execOnly(ctx, w, opt, prefix, qstr, qtyp); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		case <-time.After(opt.Watch):
		}
	}
}

// execOnly executes a query against the database.
func (h *Handler) execOnly(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, qstr string, qtyp bool) error {
	// exec or query
	f := h.exec
	if qtyp {
		f = h.query
	}
	// exec
	return f(ctx, w, opt, prefix, qstr)
}

// execSet executes a SQL query, setting all returned columns as variables.
func (h *Handler) execSet(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, qstr string, _ bool) error {
	// query
	q, err := h.DB().QueryContext(ctx, qstr)
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
		n := opt.Params["prefix"] + c
		if err = env.ValidIdentifier(n); err != nil {
			return fmt.Errorf(text.CouldNotSetVariable, n)
		}
		_ = env.Set(n, row[i])
	}
	return nil
}

// execExec executes a query and re-executes all columns of all rows as if they
// were their own queries.
func (h *Handler) execExec(ctx context.Context, w io.Writer, _ metacmd.Option, prefix, qstr string, qtyp bool) error {
	// query
	q, err := h.DB().QueryContext(ctx, qstr)
	if err != nil {
		return err
	}
	// execRows
	err = h.execRows(ctx, w, q)
	if err != nil {
		return err
	}
	// check for additional result sets ...
	for drivers.NextResultSet(q) {
		err = h.execRows(ctx, w, q)
		if err != nil {
			return err
		}
	}
	return nil
}

// query executes a query against the database.
func (h *Handler) query(ctx context.Context, w io.Writer, opt metacmd.Option, _, qstr string) error {
	start := time.Now()
	// run query
	q, err := h.DB().QueryContext(ctx, qstr)
	if err != nil {
		return err
	}
	defer q.Close()
	params := env.Pall()
	for k, v := range opt.Params {
		params[k] = v
	}
	var pipe io.WriteCloser
	if pipeName := params["pipe"]; pipeName != "" || h.out != nil {
		if params["expanded"] == "auto" && params["columns"] == "" {
			// don't rely on terminal size when piping output to a file or cmd
			params["expanded"] = "off"
		}
		if pipeName != "" {
			if pipeName[0] == '|' {
				pipe, err = env.Pipe(pipeName[1:])
			} else {
				pipe, err = os.OpenFile(pipeName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
			}
			if err != nil {
				return err
			}
			w = pipe
		}
	} else {
		params["pager_cmd"] = env.All()["PAGER"]
	}
	// wrap query with crosstab
	resultSet := tblfmt.ResultSet(q)
	if opt.Exec == metacmd.ExecCrosstab {
		var err error
		resultSet, err = tblfmt.NewCrosstabView(q, tblfmt.WithParams(opt.Crosstab...))
		if err != nil {
			return err
		}
	}
	if err = tblfmt.EncodeAll(w, resultSet, params); err != nil {
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
	if pipe != nil {
		return pipe.Close()
	}
	return err
}

// execRows executes all the columns in the row.
func (h *Handler) execRows(ctx context.Context, w io.Writer, q *sql.Rows) error {
	// get columns
	cols, err := drivers.Columns(h.u, q)
	if err != nil {
		return err
	}
	// process rows
	res := metacmd.Option{Exec: metacmd.ExecOnly}
	clen, tfmt := len(cols), h.timefmt()
	for q.Next() {
		if clen != 0 {
			row, err := h.scan(q, clen, tfmt)
			if err != nil {
				return err
			}
			// execute
			for _, qstr := range row {
				if err = h.Execute(ctx, w, res, stmt.FindPrefix(qstr), qstr, false); err != nil {
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
func (h *Handler) exec(ctx context.Context, w io.Writer, _ metacmd.Option, typ, qstr string) error {
	res, err := h.DB().ExecContext(ctx, qstr)
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
	return h.BeginTx(context.Background())
}

// Begin begins a transaction in a context.
func (h *Handler) BeginTx(ctx context.Context) error {
	if h.db == nil {
		return text.ErrNotConnected
	}
	if h.tx != nil {
		return text.ErrPreviousTransactionExists
	}
	var err error
	h.tx, err = h.db.BeginTx(ctx, nil)
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

// ReaderOptions returns default reader options.
func (h *Handler) ReaderOptions() []metadata.ReaderOption {
	opts := []metadata.ReaderOption{}
	envs := env.All()
	if envs["ECHO_HIDDEN"] == "on" || envs["ECHO_HIDDEN"] == "noexec" {
		if envs["ECHO_HIDDEN"] == "noexec" {
			opts = append(opts, metadata.WithDryRun(true))
		}
		opts = append(
			opts,
			metadata.WithLogger(log.New(os.Stdout, "DEBUG: ", log.LstdFlags)),
			metadata.WithTimeout(30*time.Second),
		)
	}
	return opts
}

// GetOutput gets the output writer.
func (h *Handler) GetOutput() io.WriteCloser {
	return h.out
}

// SetOutput sets the output writer.
func (h *Handler) SetOutput(o io.WriteCloser) {
	h.out = o
}
