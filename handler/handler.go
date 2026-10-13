// Package handler provides a input process handler implementation for usql.
package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/go-git/go-billy/v5"
	"github.com/xo/dbmeta"
	"github.com/xo/dburl"
	"github.com/xo/dburl/passfile"
	"github.com/xo/rline"
	"github.com/xo/tblfmt"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
	"github.com/xo/usql/feed"
	"github.com/xo/usql/highlight"
	"github.com/xo/usql/metacmd"
	"github.com/xo/usql/metacmd/charts"
	"github.com/xo/usql/metacmd/complete"
	"github.com/xo/usql/metacmd/describe"
	"github.com/xo/usql/text"
)

// Handler is a input process handler.
//
// Glues together usql's components to provide a "read-eval-print loop" (REPL)
// for usql's interactive command-line and manages most of the core/high-level logic.
//
// Manages the active statement buffer, application IO, executing/querying SQL
// statements, and handles backslash (\) meta commands encountered in the input
// stream.
type Handler struct {
	// l is the session of the line editor. The handler writes through it, and
	// reads passwords and answers to \prompt with it.
	l *rline.Prompt
	// src is the input of the handler: the line editor at a terminal, a
	// script, a pipe, or an included file.
	src *feed.Source
	// hl draws the input in color, and is nil unless the handler is interactive
	hl *highlight.Highlighter
	// iactive is set when src is the line editor at a terminal.
	iactive bool
	// reading is set while \prompt reads an answer from the line editor.
	reading bool
	// user is the user.
	user *user.User
	// wd is the working directory.
	wd string
	// charts is the charts filesystem.
	charts billy.Filesystem
	// nopw indicates not asking for password.
	nopw bool
	// timing of every command executed.
	timing bool
	// singleLineMode is single line mode.
	singleLineMode bool
	// buf is psql's query buffer.
	buf feed.Buffer
	// lastExec is the last executed query statement.
	lastExec string
	// lastExecPrefix is the last executed query statement prefix.
	lastExecPrefix string
	// lastRaw is the last executed raw query statement.
	lastRaw string
	// bind are bound values for that will be used for statement execution.
	bind []interface{}
	// u is the active connection information.
	u *dburl.URL
	// db is the active database connection.
	db *sql.DB
	// meta reads the structure of the database, and is built the first time
	// it is needed.
	meta *dbmeta.Meta
	// tx is the active transaction, if any.
	tx *sql.Tx
	// out file or pipe
	out io.WriteCloser
}

// New returns a handler that reads its input from r, and talks to the user
// through l. When r is l and l is a terminal, the handler is interactive: it
// shows a prompt, completes, and reads whole entries.
func New(l *rline.Prompt, r feed.LineReader, user *user.User, wd string, charts billy.Filesystem, nopw bool) (*Handler, error) {
	h := &Handler{
		l:       l,
		iactive: l.Interactive() && r == feed.LineReader(l),
		user:    user,
		wd:      wd,
		charts:  charts,
		nopw:    nopw,
	}
	var prompt func() string
	if h.iactive {
		prompt = h.prompt
		l.SetCompleter(complete.New(complete.WithConnStrings(h.connStrings())))
		var err error
		if h.hl, err = highlight.New(l); err != nil {
			return nil, err
		}
		l.SetHighlighter(h.hl)
	}
	var err error
	if h.src, err = feed.NewSource(r, drivers.Language(nil), prompt); err != nil {
		return nil, err
	}
	return h, nil
}

// setLanguage sets the grammar of the dialect of the connection on the source of
// the input, and on the highlighter, which reads the same statements.
func (h *Handler) setLanguage() error {
	lang := drivers.Language(h.u)
	if err := h.src.SetLanguage(lang); err != nil {
		return err
	}
	if h.hl == nil {
		return nil
	}
	var dialect string
	if h.u != nil {
		dialect = h.u.Dialect
	}
	return h.hl.SetDialect(dialect, lang)
}

// prompt returns the prompt for the next entry, and sets the marker that the
// line editor draws on the rows after the first.
func (h *Handler) prompt() string {
	h.l.SetPrompt("", h.Prompt(env.Setting("prompt2")))
	return h.Prompt(env.Setting("prompt1"))
}

// Continue reports whether an entry typed at the terminal needs more input. It
// is the function that the line editor's WithContinue calls, with the whole
// entry typed so far.
func (h *Handler) Continue(text string) bool {
	if h.reading {
		return false
	}
	if h.buf.Empty() && isHelpQuitExit(text) {
		return false
	}
	items, err := h.src.Parser().Parse(text)
	if err != nil {
		return false
	}
	return !feed.Finished(items)
}

// isHelpQuitExit reports whether s is only help, quit or exit.
func isHelpQuitExit(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case text.HelpPrefix, text.QuitPrefix, text.ExitPrefix:
		return true
	}
	return false
}

// GetTiming gets the timing toggle.
func (h *Handler) GetTiming() bool {
	return h.timing
}

// SetTiming sets the timing toggle.
func (h *Handler) SetTiming(timing bool) {
	h.timing = timing
}

// Run reads the input, and runs each statement and meta command in turn, until
// the input ends.
func (h *Handler) Run() error {
	stdout, stderr := h.l, h.l.Stderr()
	// display welcome info
	if h.iactive && env.Setting("quiet") == "off" {
		// logo
		if typ := env.TermGraphics(); typ.Available() {
			if err := typ.Encode(stdout, text.Logo); err != nil {
				return err
			}
		}
		// welcome text
		fmt.Fprintln(stdout, text.WelcomeDesc)
		fmt.Fprintln(stdout)
	}
	var lastErr error
	for {
		it, err := h.src.Next()
		switch {
		case errors.Is(err, rline.ErrInterrupted):
			h.buf.Reset()
			h.src.Reset()
			continue
		case errors.Is(err, io.EOF):
			// -c sends what is left in the buffer at the end of its text
			if h.singleLineMode && !h.buf.Empty() {
				stop, err := h.send(stdout, stderr, metacmd.Option{})
				if err != nil {
					lastErr = err
				}
				if stop {
					return err
				}
			}
			return lastErr
		case err != nil:
			return err
		}
		var opt metacmd.Option
		switch it.Kind {
		case feed.Statement:
			if h.iactive && h.buf.Empty() && !it.Terminated && isHelpQuitExit(it.Text) {
				if strings.EqualFold(strings.TrimSpace(it.Text), text.HelpPrefix) {
					fmt.Fprintln(stdout, text.HelpDesc)
					continue
				}
				return nil
			}
			h.buf.Add(it)
			if !it.Terminated {
				continue
			}
		case feed.Command:
			var cont bool
			opt, cont, lastErr = h.apply(stdout, stderr, it)
			switch {
			case cont:
				continue
			case opt.Quit:
				if h.out != nil {
					h.out.Close()
				}
				return nil
			case opt.Exec == metacmd.ExecNone:
				continue
			}
		}
		stop, err := h.send(stdout, stderr, opt)
		if err != nil {
			lastErr = err
		}
		if stop {
			return err
		}
	}
}

// send sends the query buffer, or the last query when the buffer is empty, as
// \g does. It returns the error of the statement, which it has already
// written out, and reports whether the handler must stop, as on_error_stop
// asks outside a terminal.
func (h *Handler) send(stdout, stderr io.Writer, opt metacmd.Option) (bool, error) {
	prefix, exp, raw := h.buf.Prefix(), h.buf.Expand(h.Lookup), h.buf.Raw()
	if !h.buf.Empty() {
		h.lastExec, h.lastExecPrefix, h.lastRaw = exp, prefix, raw
		h.buf.Reset()
	}
	if h.lastExec == "" || h.lastExec == ";" {
		return false, nil
	}
	out := stdout
	if h.out != nil {
		out = h.out
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err := h.Execute(ctx, out, opt, h.lastExecPrefix, h.lastExec, h.unbind()...)
	if err == nil {
		return false, nil
	}
	fmt.Fprintln(stderr, "error:", err)
	err = WrapErr(h.lastExec, err)
	if env.Setting("on_error_stop") == "on" {
		if h.iactive {
			// empty the input, so that no other statement runs
			h.buf.Reset()
			h.src.Reset()
			return false, err
		}
		return true, err
	}
	return false, err
}

// apply runs the meta command it. It reports whether the command failed, so
// that the handler goes on to the next item.
func (h *Handler) apply(stdout, stderr io.Writer, it feed.Item) (metacmd.Option, bool, error) {
	cmd := it.Name
	f, err := metacmd.Decode(it)
	if err != nil {
		switch err = WrapErr(cmd, err); {
		case err == text.ErrUnknownCommand:
			fmt.Fprintln(stderr, fmt.Sprintf(text.InvalidCommand, cmd))
		case err == text.ErrMissingRequiredArgument:
			fmt.Fprintln(stderr, fmt.Sprintf(text.MissingRequiredArg, cmd))
		default:
			fmt.Fprintln(stderr, "error:", err)
		}
		return metacmd.Option{}, true, err
	}
	opt, unread, err := f(h)
	if err != nil && !errors.Is(err, rline.ErrInterrupted) {
		fmt.Fprintln(stderr, "error:", err)
		return metacmd.Option{}, true, WrapErr(cmd, err)
	}
	for _, arg := range unread {
		fmt.Fprintln(stdout, fmt.Sprintf(text.ExtraArgumentIgnored, cmd, arg))
	}
	return opt, false, nil
}

// Execute executes a query against the connected database.
func (h *Handler) Execute(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, sqlstr string, bind ...interface{}) error {
	if h.db == nil {
		return text.ErrNotConnected
	}
	// determine type and pre process string
	prefix, sqlstr, qtyp := drivers.Process(h.u, prefix, sqlstr)
	f := h.doExecSingle
	switch opt.Exec {
	case metacmd.ExecExec:
		f = h.doExecExec
	case metacmd.ExecSet:
		f = h.doExecSet
	case metacmd.ExecWatch:
		f = h.doExecWatch
	case metacmd.ExecChart:
		f = h.doExecChart
	}
	return drivers.WrapErr(h.u.SchemeName, f(ctx, w, opt, prefix, sqlstr, qtyp, bind))
}

// prefix returns the prefix of the first statement of sqlstr, as \gexec needs
// for each statement that a query returns.
func (h *Handler) prefix(sqlstr string) string {
	items, err := h.src.Parser().Parse(sqlstr)
	if err != nil || len(items) == 0 {
		return ""
	}
	return items[0].Prefix()
}

// Reset empties the query buffer and forgets the last query.
func (h *Handler) Reset() {
	h.buf.Reset()
	h.lastExec, h.lastExecPrefix, h.lastRaw = "", "", ""
}

// Push puts text in front of the input, as \e does with the edited query.
func (h *Handler) Push(text string) {
	h.src.Push(text)
}

// Buffer returns the query buffer.
func (h *Handler) Buffer() *feed.Buffer {
	return &h.buf
}

// Lookup returns the value of a variable, for expanding a statement.
func (h *Handler) Lookup(name string) (string, bool) {
	return env.Vars().Get(name)
}

// RunCommand runs the text of -c, as its own input. A statement left in the
// buffer at the end of the text is sent.
func (h *Handler) RunCommand(text string) error {
	src := h.src
	defer func() { h.src = src }()
	var err error
	if h.src, err = feed.NewSource(feed.NewLines(strings.NewReader(text)), drivers.Language(h.u), nil); err != nil {
		return err
	}
	h.singleLineMode = true
	defer func() { h.singleLineMode = false }()
	return h.Run()
}

// Bind sets the bind parameters for the next query execution.
func (h *Handler) Bind(bind []interface{}) {
	h.bind = bind
}

// unbind returns the bind parameters.
func (h *Handler) unbind() []interface{} {
	v := h.bind
	h.bind = nil
	return v
}

// Prompt parses a prompt.
//
// NOTE: the documentation below is INCORRECT, as it is just copied from
// https://www.postgresql.org/docs/current/app-psql.html#APP-PSQL-PROMPTING
//
// TODO/FIXME: complete this functionality (from psql documentation):
//
//	%M - The full host name (with domain name) of the database server, or
//	[local] if the connection is over a Unix domain socket, or
//	[local:/dir/name], if the Unix domain socket is not at the compiled in
//	default location.
//
//	%m - The host name of the database server, truncated at the first dot, or
//	[local] if the connection is over a Unix domain socket.
//
//	%> - The port number at which the database server is listening.
//
//	%n - The database session user name. (The expansion of this value might
//	change during a database session as the result of the command SET SESSION
//	AUTHORIZATION.)
//
//	%/ - The name of the current database.
//
//	%~ - Like %/, but the output is ~ (tilde) if the database is your default
//	database.
//
//	%# - If the session user is a database superuser, then a #, otherwise a >.
//	(The expansion of this value might change during a database session as the
//	result of the command SET SESSION AUTHORIZATION.)
//
//	%p - The process ID of the backend currently connected to.
//
//	%R - In prompt 1 normally =, but @ if the session is in an inactive branch
//	of a conditional block, or ^ if in single-line mode, or ! if the session is
//	disconnected from the database (which can happen if \connect fails). In
//	prompt 2 %R is replaced by a character that depends on why psql expects
//	more input: - if the command simply wasn't terminated yet, but * if there
//	is an unfinished /* ... */ comment, a single quote if there is an
//	unfinished quoted string, a double quote if there is an unfinished quoted
//	identifier, a dollar sign if there is an unfinished dollar-quoted string,
//	or ( if there is an unmatched left parenthesis. In prompt 3 %R doesn't
//	produce anything.
//
//	%x - Transaction status: an empty string when not in a transaction block,
//	or * when in a transaction block, or ! when in a failed transaction block,
//	or ? when the transaction state is indeterminate (for example, because
//	there is no connection).
//
//	%l - The line number inside the current statement, starting from 1.
//
//	%digits - The character with the indicated octal code is substituted.
//
//	%:name: - The value of the psql variable name. See Variables, above, for
//	details.
//
//	%`command` - The output of command, similar to ordinary “back-tick”
//	substitution.
//
//	%[ ... %] - Prompts can contain terminal control characters which, for
//	example, change the color, background, or style of the prompt text, or
//	change the title of the terminal window. In order for the line editing
//	features of Readline to work properly, these non-printing control
//	characters must be designated as invisible by surrounding them with %[ and
//	%]. Multiple pairs of these can occur within the prompt. For example:
//
//	testdb=> \uset prompt1 '%[%033[1;33;40m%]%n@%/%R%[%033[0m%]%# '
//
//	results in a boldfaced (1;) yellow-on-black (33;40) prompt on
//	VT100-compatible, color-capable terminals.
//
//	%w - Whitespace of the same width as the most recent output of prompt1.
//	This can be used as a prompt2 setting, so that multi-line statements are
//	aligned with the first line, but there is no visible secondary prompt.
//
// To insert a percent sign into your prompt, write %%. The default prompts are
// '%/%R%x%# ' for prompts 1 and 2, and '>> ' for prompt 3.
func (h *Handler) Prompt(prompt string) string {
	r, connected := []rune(prompt), h.db != nil
	end := len(r)
	var buf []byte
	for i := 0; i < end; i++ {
		if r[i] != '%' {
			buf = append(buf, string(r[i])...)
			continue
		}
		switch grab(r, i+1, end) {
		case '%': // literal
			buf = append(buf, '%')
		case 'S': // short driver name
			if connected {
				s := dburl.ShortAlias(h.u.Scheme)
				if s == "" {
					s = dburl.ShortAlias(h.u.SchemeName)
				}
				if s == "" {
					s = text.UnknownShortAlias
				}
				buf = append(buf, s+":"...)
			} else {
				buf = append(buf, text.NotConnected...)
			}
		case 'u': // dburl short
			if connected {
				buf = append(buf, h.u.Short()...)
			} else {
				buf = append(buf, text.NotConnected...)
			}
		case 'M': // full host name with domain
			if connected {
				buf = append(buf, h.u.Hostname()...)
			}
		case 'm': // host name truncated at first dot, or [local] if it's a domain socket
			if connected {
				s := h.u.Hostname()
				if i := strings.Index(s, "."); i != -1 {
					s = s[:i]
				}
				buf = append(buf, s...)
			}
		case '>': // the port number
			if connected {
				s := h.u.Port()
				if s != "" {
					s = ":" + s
				}
				buf = append(buf, s...)
			}
		case 'N': // database user
			if connected && h.u.User != nil {
				s := h.u.User.Username()
				if s != "" {
					buf = append(buf, s+"@"...)
				}
			}
		case 'n': // database user
			if connected && h.u.User != nil {
				buf = append(buf, h.u.User.Username()...)
			}
		case '/': // database name
			switch {
			case connected && h.u.Opaque != "":
				buf = append(buf, h.u.Opaque...)
			case connected && h.u.Path != "" && h.u.Path != "/":
				buf = append(buf, h.u.Path...)
			}
		case 'O':
			if connected {
				buf = append(buf, h.u.Opaque...)
			}
		case 'o':
			if connected {
				buf = append(buf, filepath.Base(h.u.Opaque)...)
			}
		case 'P':
			if connected {
				buf = append(buf, h.u.Path...)
			}
		case 'p':
			if connected {
				buf = append(buf, path.Base(h.u.Path)...)
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			j := i + 1
			base := 10
			if grab(r, j, end) == '0' {
				j++
				base = 8
			}
			if grab(r, j, end) == 'x' {
				j++
				base = 16
			}
			i = j
			for unicode.IsDigit(grab(r, i+1, end)) {
				i++
			}

			n, err := strconv.ParseInt(string(r[j:i+1]), base, 16)
			if err == nil {
				buf = append(buf, byte(n))
			}
			i--
		case '~': // like %/ but ~ when default database
		case '#': // when superuser, a #, otherwise >
			if h.tx != nil {
				buf = append(buf, '~')
			} else {
				buf = append(buf, '>')
			}
		// case 'p': // the process id of the connected backend -- never going to be supported
		case 'R': // statement state
			// = for a new statement and - while the buffer holds one. The
			// line editor draws one prompt for an entry, so the state inside
			// a string or a comment is not shown.
			if h.buf.Empty() {
				buf = append(buf, '=')
			} else {
				buf = append(buf, '-')
			}
		case 'x': // empty when not in a transaction block, * in transaction block, ! in failed transaction block, or ? when indeterminate
		case 'l': // line number
		case ':': // variable value
		case '`': // value of the evaluated command
		case '[', ']':
		case 'w':
		}
		i++
	}
	return string(buf)
}

// Interactive reports whether the handler reads entries from the line editor
// at a terminal.
func (h *Handler) Interactive() bool {
	return h.iactive
}

// IO returns the session of the line editor.
func (h *Handler) IO() *rline.Prompt {
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

// LastExec returns the last executed statement.
func (h *Handler) LastExec() string {
	return h.lastExec
}

// LastRaw returns the last raw (non-interpolated) executed statement.
func (h *Handler) LastRaw() string {
	return h.lastRaw
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
func (h *Handler) Open(ctx context.Context, params ...string) error {
	if len(params) == 0 || params[0] == "" {
		return nil
	}
	if h.tx != nil {
		return text.ErrPreviousTransactionExists
	}
	if len(params) == 1 {
		if v, ok := env.Vars().GetConn(params[0]); ok {
			params = v
		}
	}
	if len(params) < 2 {
		dsn := params[0]
		// parse dsn
		u, err := dburl.Parse(dsn)
		if err != nil {
			return err
		}
		h.u = u
		// force parameters
		h.forceParams(h.u)
	} else {
		h.u = &dburl.URL{
			Driver: params[0],
			DSN:    strings.Join(params[1:], " "),
		}
	}
	// open connection
	var err error
	h.db, err = drivers.Open(ctx, h.u, h.GetOutput, h.l.Stderr)
	if err != nil && !drivers.IsPasswordErr(h.u, err) {
		defer h.Close()
		return err
	}
	// read what follows in the grammar of the dialect
	if err := h.setLanguage(); err != nil {
		return err
	}
	// force error/check connection
	if err == nil {
		if err = drivers.Ping(ctx, h.u, h.db); err == nil {
			if h.l.Interactive() {
				c := complete.New(complete.WithConnStrings(h.connStrings()))
				if m, err := h.metadata(ctx); err == nil {
					c = complete.New(
						complete.WithSource(describe.Source(m, h.db)),
						complete.WithConnStrings(h.connStrings()),
					)
				}
				h.l.SetCompleter(c)
			}
			return h.Version(ctx)
		}
	}
	// bail without getting password
	if h.nopw || !drivers.IsPasswordErr(h.u, err) || len(params) > 1 || !h.l.Interactive() {
		defer h.Close()
		return err
	}
	// print the error
	fmt.Fprintln(h.l.Stderr(), "error:", err)
	// otherwise, try to collect a password ...
	dsn, err := h.Password(params[0])
	if err != nil {
		// close connection
		defer h.Close()
		return err
	}
	// reconnect
	return h.Open(ctx, dsn)
}

func (h *Handler) connStrings() []string {
	entries, err := passfile.Entries(h.user.HomeDir, text.PassfileName)
	if err != nil {
		// ignore the error as this is only used for completer
		// and it'll be reported again when trying to force params before opening a conn
		entries = nil
	}
	available := drivers.Available()
	names := make([]string, 0, len(available)+len(entries))
	for schema := range available {
		_, aliases := dburl.SchemeNameAndAliases(schema)
		// TODO should we create all combinations of space, :, :// and +transport ?
		names = append(names, schema)
		names = append(names, aliases...)
	}
	for _, entry := range entries {
		if entry.Protocol == "*" {
			continue
		}
		user, host, port, dbname := "", "", "", ""
		if entry.Username != "*" {
			user = entry.Username + "@"
			if entry.Host != "*" {
				host = entry.Host
				if entry.Port != "*" {
					port = ":" + entry.Port
				}
				if entry.DBName != "*" {
					dbname = "/" + entry.DBName
				}
			}
		}
		names = append(names, entry.Protocol+"://"+user+host+port+dbname)
	}
	return append(names, slices.Sorted(maps.Keys(env.Vars().Conn()))...)
}

// forceParams forces connection parameters on a database URL, adding any
// driver specific required parameters, and the username/password when a
// matching entry exists in the PASS file.
func (h *Handler) forceParams(u *dburl.URL) {
	// force driver parameters
	drivers.ForceParams(u)
	// see if password entry is present
	user, err := passfile.Match(u, h.user.HomeDir, text.PassfileName)
	switch {
	case err != nil:
		fmt.Fprintln(h.l.Stderr(), "error:", err)
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
	switch conn, ok := env.Vars().GetConn(dsn); {
	case dsn == "":
		return "", text.ErrMissingDSN
	case ok && len(conn) < 2:
		return "", text.ErrNamedConnectionIsNotAURL
	case ok:
		dsn = conn[0]
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
		drv := h.u.SchemeName
		h.db, h.u, h.meta = nil, nil, nil
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
	} else if h.iactive {
		// the answer is not an entry, so it is always finished, and it stays
		// out of the history
		h.reading = true
		v, err = h.l.ReadLine(prompt)
		h.reading = false
		if err == nil {
			h.l.RemoveLastHistory()
			_ = h.l.SaveHistory()
		}
	} else {
		v, err = h.src.ReadLine(prompt)
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
		if b, err = strconv.ParseBool(v); err == nil {
			v = fmt.Sprintf("%t", b)
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
func (h *Handler) Version(ctx context.Context) error {
	if env.Setting("show_host_information") != "true" || !h.l.Interactive() {
		return nil
	}
	if h.db == nil {
		return text.ErrNotConnected
	}
	var ver string
	var err error
	if _, ok := dbmeta.Dialect(h.u.Dialect).Info(); ok && !drivers.OwnVersion(h.u) {
		var m *dbmeta.Meta
		if m, err = h.metadata(ctx); err == nil {
			ver = m.String()
		}
	} else {
		ver, err = drivers.Version(ctx, h.u, h.DB())
	}
	switch {
	case err != nil:
		ver = fmt.Sprintf("<unknown, error: %v>", err)
	case ver == "":
		ver = "<unknown>"
	}
	h.Print(text.ConnInfo, h.u.SchemeName, ver)
	return nil
}

// Print formats according to a format specifier and writes to handler's standard output.
func (h *Handler) Print(s string, v ...interface{}) {
	if env.Setting("quiet") == "on" {
		return
	}
	fmt.Fprintln(h.l, fmt.Sprintf(s, v...))
}

// doExecWatch repeatedly executes a query against the database.
func (h *Handler) doExecWatch(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, sqlstr string, qtyp bool, bind []interface{}) error {
	for {
		// the actual output that psql has: "Mon Jan 2006 3:04:05 PM MST" -- which is _slightly_ different than RFC1123
		// fmt.Fprintf(w, "%s (every %fs)\n\n", time.Now().Format("Mon Jan 2006 3:04:05 PM MST"), float64(opt.Watch)/float64(time.Second))
		fmt.Fprintf(w, "%s (every %v)\n", time.Now().Format(time.RFC1123), opt.Watch)
		fmt.Fprintln(w)
		if err := h.doExecSingle(ctx, w, opt, prefix, sqlstr, qtyp, bind); err != nil {
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

// doExecChart executes a single query against the database, displaying its output as a chart.
func (h *Handler) doExecChart(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, sqlstr string, qtyp bool, bind []interface{}) error {
	stdout := h.l
	// Report a missing renderer before anything else. Without one no terminal
	// and no argument can make \chart work, and reporting the terminal first
	// would name the wrong cause.
	if !charts.Available() {
		return text.ErrChartsNotSupported
	}
	typ := env.TermGraphics()
	if !typ.Available() {
		return text.ErrGraphicsNotSupported
	}
	if _, ok := opt.Params["help"]; ok {
		fmt.Fprintln(stdout, text.ChartUsage)
		return nil
	}
	cfg, err := charts.ParseArgs(opt.Params)
	if err != nil {
		return err
	}
	start := time.Now()
	// query
	rows, err := h.DB().QueryContext(ctx, sqlstr, bind...)
	if err != nil {
		return err
	}
	defer rows.Close()
	// get cols
	cols, err := drivers.Columns(h.u, rows)
	if err != nil {
		return err
	}
	// process row(s)
	transposed := make([][]string, len(cols))
	clen, tfmt := len(cols), env.Vars().PrintTimeFormat()
	for rows.Next() {
		row, err := h.scan(rows, clen, tfmt)
		if err != nil {
			return err
		}
		for i := range row {
			transposed[i] = append(transposed[i], row[i])
		}
	}
	// display
	c, err := charts.MakeChart(cfg, cols, transposed)
	if err != nil {
		return err
	}
	data, err := c.ToEcharts()
	if err != nil {
		return err
	}
	res, err := charts.RenderSVG(ctx, data, cfg.W, cfg.H)
	if err != nil {
		return err
	}
	if cfg.File != "" {
		fmt.Println("writing to", cfg.File)
		return os.WriteFile(cfg.File, res, 0o644)
	}
	img, err := charts.Rasterize(res, cfg.Background)
	if err != nil {
		return err
	}
	if err := typ.Encode(stdout, img); err != nil {
		return err
	}
	if h.timing {
		d := time.Since(start)
		s := text.TimingDesc
		v := []interface{}{float64(d.Microseconds()) / 1000}
		if d > 1*time.Second {
			s += " (%v)"
			v = append(v, d.Round(1*time.Millisecond))
		}
		fmt.Fprintln(h.l, fmt.Sprintf(s, v...))
	}
	return nil
}

// doExecSingle executes a single query against the database based on its query type.
func (h *Handler) doExecSingle(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, sqlstr string, qtyp bool, bind []interface{}) error {
	// exec or query
	f := h.doExec
	if qtyp {
		f = h.doQuery
	}
	// exec
	start := time.Now()
	if err := f(ctx, w, opt, prefix, sqlstr, bind); err != nil {
		return err
	}
	if h.timing {
		d := time.Since(start)
		s := text.TimingDesc
		v := []interface{}{float64(d.Microseconds()) / 1000}
		if d > 1*time.Second {
			s += " (%v)"
			v = append(v, d.Round(1*time.Millisecond))
		}
		fmt.Fprintln(h.l, fmt.Sprintf(s, v...))
	}
	return nil
}

// doExecSet executes a SQL query, setting all returned columns as variables.
func (h *Handler) doExecSet(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, sqlstr string, _ bool, bind []interface{}) error {
	// query
	rows, err := h.DB().QueryContext(ctx, sqlstr, bind...)
	if err != nil {
		return err
	}
	defer rows.Close()
	// get cols
	cols, err := drivers.Columns(h.u, rows)
	if err != nil {
		return err
	}
	// process row(s)
	var i int
	var row []string
	clen, tfmt := len(cols), env.Vars().PrintTimeFormat()
	for rows.Next() {
		if i == 0 {
			row, err = h.scan(rows, clen, tfmt)
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
		_ = env.Vars().Set(n, row[i])
	}
	return nil
}

// doExecExec executes a query and re-executes all columns of all rows as if they
// were their own queries.
func (h *Handler) doExecExec(ctx context.Context, w io.Writer, _ metacmd.Option, prefix, sqlstr string, qtyp bool, bind []interface{}) error {
	// query
	rows, err := h.DB().QueryContext(ctx, sqlstr, bind...)
	if err != nil {
		return err
	}
	defer rows.Close()
	// exec resulting rows
	if err := h.doExecRows(ctx, w, rows); err != nil {
		return err
	}
	// check for additional result sets ...
	for rows.NextResultSet() {
		if err := h.doExecRows(ctx, w, rows); err != nil {
			return err
		}
	}
	return nil
}

// doQuery executes a doQuery against the database.
func (h *Handler) doQuery(ctx context.Context, w io.Writer, opt metacmd.Option, typ, sqlstr string, bind []interface{}) error {
	// run query
	rows, err := h.DB().QueryContext(ctx, sqlstr, bind...)
	if err != nil {
		return err
	}
	defer rows.Close()
	params := env.Vars().Print()
	params["time"] = env.Vars().PrintTimeFormat()
	for k, v := range opt.Params {
		params[k] = v
	}
	var pipe io.WriteCloser
	var cmd *exec.Cmd
	if pipeName := params["pipe"]; pipeName != "" || h.out != nil {
		if params["expanded"] == "auto" && params["columns"] == "" {
			// don't rely on terminal size when piping output to a file or cmd
			params["expanded"] = "off"
		}
		if pipeName != "" {
			if pipeName[0] == '|' {
				pipe, cmd, err = env.Pipe(h.l, h.l.Stderr(), pipeName[1:])
			} else {
				pipe, err = os.OpenFile(pipeName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
			}
			if err != nil {
				return err
			}
			w = pipe
		}
	} else if opt.Exec != metacmd.ExecWatch {
		params["pager_cmd"] = env.Setting("pager")
	}
	// set up column type config
	var extra []tblfmt.Option
	switch f := drivers.ColumnTypes(h.u); {
	case f != nil:
		extra = append(extra, tblfmt.WithColumnTypesFunc(f))
	case drivers.UseColumnTypes(h.u):
		// Not tblfmt.WithUseColumnTypes, which scans straight into the type
		// the driver names and so fails on a NULL in a column the driver
		// described as not nullable. See drivers.NullSafeColumnType.
		extra = append(extra, tblfmt.WithColumnTypesFunc(drivers.NullSafeColumnType))
	}
	resultSet := tblfmt.ResultSet(rows)
	// wrap query with crosstab
	if opt.Exec == metacmd.ExecCrosstab {
		var err error
		if resultSet, err = tblfmt.NewCrosstabView(rows, append(extra, tblfmt.WithParams(opt.Crosstab...))...); err != nil {
			return err
		}
		extra = nil
	}
	if drivers.LowerColumnNames(h.u) {
		params["lower_column_names"] = "true"
	}
	// a sqlserver EXEC can return a result set with no columns and nothing
	// after it, so report it as an exec rather than as an empty result set
	if h.u.SchemeName == "sqlserver" && strings.HasPrefix(typ, "EXEC") && opt.Exec != metacmd.ExecCrosstab {
		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		if len(cols) == 0 && !rows.NextResultSet() {
			fmt.Fprintln(w, typ)
			return nil
		}
	}
	// encode and handle error conditions
	switch err := tblfmt.EncodeAll(w, resultSet, params, extra...); {
	case err != nil && cmd != nil && errors.Is(err, syscall.EPIPE):
		// broken pipe means pager quit before consuming all data, which might be expected
		return nil
	case err != nil:
		return err
	case params["format"] == "aligned":
		fmt.Fprintln(w)
	}
	if pipe != nil {
		pipe.Close()
		if cmd != nil {
			cmd.Wait()
		}
	}
	return err
}

// doExecRows executes all the columns in the row.
func (h *Handler) doExecRows(ctx context.Context, w io.Writer, rows *sql.Rows) error {
	// get columns
	cols, err := drivers.Columns(h.u, rows)
	if err != nil {
		return err
	}
	// process rows
	res := metacmd.Option{
		Exec: metacmd.ExecOnly,
	}
	clen, tfmt := len(cols), env.Vars().PrintTimeFormat()
	for rows.Next() {
		if clen != 0 {
			row, err := h.scan(rows, clen, tfmt)
			if err != nil {
				return err
			}
			// execute
			for _, sqlstr := range row {
				if err = h.Execute(ctx, w, res, h.prefix(sqlstr), sqlstr); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// scan scans a row.
func (h *Handler) scan(rows *sql.Rows, clen int, tfmt string) ([]string, error) {
	// scan to []interface{}
	r := make([]interface{}, clen)
	for i := range r {
		r[i] = new(interface{})
	}
	if err := rows.Scan(r...); err != nil {
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
				var err error
				if row[n], err = cb(x, tfmt); err != nil {
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
				var err error
				if row[n], err = cm(x); err != nil {
					return nil, err
				}
			}
		case []interface{}:
			if x != nil {
				var err error
				if row[n], err = cs(x); err != nil {
					return nil, err
				}
			}
		default:
			if x != nil {
				var err error
				if row[n], err = cd(x); err != nil {
					return nil, err
				}
			}
		}
	}
	return row, nil
}

// doExec does a database exec.
func (h *Handler) doExec(ctx context.Context, w io.Writer, _ metacmd.Option, typ, sqlstr string, bind []interface{}) error {
	res, err := h.DB().ExecContext(ctx, sqlstr, bind...)
	if err != nil {
		_ = env.Vars().Set("ROW_COUNT", "0")
		return err
	}
	// get affected
	count, err := drivers.RowsAffected(h.u, res)
	if err != nil {
		_ = env.Vars().Set("ROW_COUNT", "0")
		return err
	}
	// print name
	if env.Setting("quiet") == "off" {
		fmt.Fprint(w, typ)
		// print count
		if count > 0 {
			fmt.Fprint(w, " ", count)
		}
		fmt.Fprintln(w)
	}
	return env.Vars().Set("ROW_COUNT", strconv.FormatInt(count, 10))
}

// Begin begins a transaction.
func (h *Handler) Begin(txOpts *sql.TxOptions) error {
	return h.BeginTx(context.Background(), txOpts)
}

// Begin begins a transaction in a context.
func (h *Handler) BeginTx(ctx context.Context, txOpts *sql.TxOptions) error {
	if h.db == nil {
		return text.ErrNotConnected
	}
	if h.tx != nil {
		return text.ErrPreviousTransactionExists
	}
	var err error
	h.tx, err = h.db.BeginTx(ctx, txOpts)
	if err != nil {
		return drivers.WrapErr(h.u.SchemeName, err)
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
	if err := tx.Commit(); err != nil {
		return drivers.WrapErr(h.u.SchemeName, err)
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
	if err := tx.Rollback(); err != nil {
		return drivers.WrapErr(h.u.SchemeName, err)
	}
	return nil
}

// If starts an if block.
func (h *Handler) If(ok bool) error {
	return nil
}

// ElseIf starts an else if block.
func (h *Handler) ElseIf(ok bool) error {
	return nil
}

// Else starts an else block.
func (h *Handler) Else(bool) error {
	return nil
}

// EndIf closes an if block.
func (h *Handler) EndIf(bool) error {
	return nil
}

// IncludeReader includes the content of rdr, as \i does. A handler of its
// own reads rdr, on the same connection, and the connection it ends with is
// kept, so that a \c in the file applies after it.
func (h *Handler) IncludeReader(rdr io.Reader, path string) error {
	p, err := New(h.l, feed.NewLines(rdr), h.user, filepath.Dir(path), h.charts, h.nopw)
	if err != nil {
		return err
	}
	p.db, p.u = h.db, h.u
	if err := p.src.SetLanguage(drivers.Language(p.u)); err != nil {
		return err
	}
	err = p.Run()
	h.db, h.u = p.db, p.u
	if lerr := h.setLanguage(); err == nil {
		err = lerr
	}
	return err
}

// Include includes the specified path.
func (h *Handler) Include(path string, relative bool) error {
	if relative && !filepath.IsAbs(path) {
		path = filepath.Join(h.wd, path)
	}
	// fmt.Fprintf(os.Stderr, "include: %s relative: %t\n", path, relative)
	// open
	path, f, err := env.OpenFile(h.user, path)
	if err != nil {
		return err
	}
	defer f.Close()
	return h.IncludeReader(f, path)
}

// Formatter returns the formatter of the describe commands, such as \d, with
// the options of the session, such as echo_hidden, applied. It writes to the
// output of the session, so \o redirects it, as it does in psql.
func (h *Handler) Formatter(ctx context.Context) (*describe.Formatter, error) {
	m, err := h.metadata(ctx)
	if err != nil {
		return nil, err
	}
	return describe.NewFormatter(m, h.db, h.GetOutput(), h.u.SchemeName, formatterOpts()...), nil
}

// metadata returns dbmeta's model of the database. It is read once for each
// connection.
func (h *Handler) metadata(ctx context.Context) (*dbmeta.Meta, error) {
	if h.db == nil {
		return nil, text.ErrNotConnected
	}
	if h.meta == nil {
		m, err := drivers.Meta(ctx, h.u, h.db)
		if err != nil {
			return nil, err
		}
		h.meta = m
	}
	return h.meta, nil
}

// GetOutput gets the output writer.
func (h *Handler) GetOutput() io.Writer {
	if h.out == nil {
		return h.l
	}
	return h.out
}

// SetOutput sets the output writer.
func (h *Handler) SetOutput(o io.WriteCloser) {
	if h.out != nil {
		h.out.Close()
	}
	h.out = o
}

// FS is the filesystem interface.
type FS interface{}

// Error wraps handler errors.
type Error struct {
	Buf string
	Err error
}

// WrapErr wraps an [error] using the specified driver when err is not nil.
func WrapErr(buf string, err error) error {
	if err == nil {
		return nil
	}
	// avoid double wrapping error
	if _, ok := err.(*Error); ok {
		return err
	}
	return &Error{buf, err}
}

// Error satisfies the [error] interface, returning the original error message.
func (e *Error) Error() string {
	return e.Err.Error()
}

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.Err
}

// formatterOpts returns the options of the describe commands that
// echo_hidden sets.
func formatterOpts() []describe.Option {
	var opts []describe.Option
	if env.Setting("echo_hidden") == "on" || env.Setting("echo_hidden") == "noexec" {
		if env.Setting("echo_hidden") == "noexec" {
			opts = append(opts, describe.WithDryRun(true))
		}
		opts = append(
			opts,
			describe.WithLogger(log.New(os.Stdout, "DEBUG: ", log.LstdFlags)),
			describe.WithTimeout(30*time.Second),
		)
	}
	return opts
}

// grab returns the i'th rune from r when i < end, otherwise 0.
func grab(r []rune, i, end int) rune {
	if i < end {
		return r[i]
	}
	return 0
}
