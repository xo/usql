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
	"maps"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/go-git/go-billy/v5"
	"github.com/xo/dburl"
	"github.com/xo/dburl/passfile"
	"github.com/xo/echartsgoja"
	"github.com/xo/resvg"
	"github.com/xo/tblfmt"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/completer"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/env"
	"github.com/xo/usql/metacmd"
	"github.com/xo/usql/metacmd/charts"
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
// statements, and handles backslash (\) meta commands encountered in the input
// stream.
type Handler struct {
	// l is the readline handler.
	l rline.IO
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
	// buf is the query statement buffer.
	buf *stmt.Stmt
	// lastExec is the last executed query statement.
	lastExec string
	// lastExecPrefix is the last executed query statement prefix.
	lastExecPrefix string
	// lastPrint is the last executed printable query statement.
	lastPrint string
	// lastRaw is the last executed raw query statement.
	lastRaw string
	// batch indicates a batch has been started.
	batch bool
	// batchEnd is the batch end string.
	batchEnd string
	// bind are bound values for that will be used for statement execution.
	bind []interface{}
	// u is the active connection information.
	u *dburl.URL
	// db is the active database connection.
	db *sql.DB
	// tx is the active transaction, if any.
	tx *sql.Tx
	// out file or pipe
	out io.WriteCloser
}

// New creates a new input handler.
func New(l rline.IO, user *user.User, wd string, charts billy.Filesystem, nopw bool) *Handler {
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
		l:      l,
		user:   user,
		wd:     wd,
		charts: charts,
		nopw:   nopw,
		buf:    stmt.New(f),
	}
	return h
}

// GetTiming gets the timing toggle.
func (h *Handler) GetTiming() bool {
	return h.timing
}

// SetTiming sets the timing toggle.
func (h *Handler) SetTiming(timing bool) {
	h.timing = timing
}

// SetSingleLineMode sets the single line mode toggle.
func (h *Handler) SetSingleLineMode(singleLineMode bool) {
	h.singleLineMode = singleLineMode
}

// Run executes queries and commands.
func (h *Handler) Run() error {
	stdout, stderr, iactive := h.l.Stdout(), h.l.Stderr(), h.l.Interactive()
	// display welcome info
	if iactive && env.Get("QUIET") == "off" {
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
	var cmd string
	var paramstr string
	var err error
	var opt metacmd.Option
	var cont bool
	var lastErr error
	var execute bool
	for {
		execute = false
		// set prompt
		if iactive {
			h.l.Prompt(h.Prompt(env.Get("PROMPT1")))
		}
		// read next statement/command
		switch cmd, paramstr, err = h.buf.Next(env.Untick(h.user, env.Vars(), false)); {
		case h.singleLineMode && err == nil:
			execute = h.buf.Len != 0
		case err == rline.ErrInterrupt:
			h.buf.Reset(nil)
			continue
		case err == io.EOF:
			return lastErr
		case err != nil:
			return err
		case cmd != "":
			opt, cont, lastErr = h.apply(stdout, stderr, strings.TrimPrefix(cmd, `\`), paramstr)
		}
		if cont {
			continue
		}
		// help, exit, quit intercept
		if iactive && len(h.buf.Buf) >= 4 {
			i, first := lastIndex(h.buf.Buf, '\n'), false
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
					fmt.Fprintln(stderr, "error:", err)
					continue
				// cannot use \g* while accumulating statements for batch queries
				case h.batch && typ != h.batchEnd && opt.Exec != metacmd.ExecNone:
					err = errors.New("cannot force batch execution")
					lastErr = WrapErr(h.buf.String(), err)
					fmt.Fprintln(stderr, "error:", err)
					continue
				case batch:
					h.batch, h.batchEnd = true, end
				case h.batch:
					var lend string
					if len(h.lastExec) != 0 {
						lend = "\n"
					}
					// append to last
					h.lastExec += lend + h.buf.String()
					h.lastExecPrefix = h.buf.Prefix
					h.lastPrint += lend + h.buf.PrintString()
					h.lastRaw += lend + h.buf.RawString()
					h.buf.Reset(nil)
					// break
					if h.batchEnd != typ {
						continue
					}
					h.lastExecPrefix = h.batchEnd
					h.batch, h.batchEnd = false, ""
				}
			}
			if h.buf.Len != 0 {
				h.lastExec, h.lastExecPrefix, h.lastPrint, h.lastRaw = h.buf.String(), h.buf.Prefix, h.buf.PrintString(), h.buf.RawString()
				h.buf.Reset(nil)
			}
			// log.Printf(">> PROCESS EXECUTE: (%s) `%s`", h.lastPrefix, h.last)
			if !h.batch && h.lastExec != "" && h.lastExec != ";" {
				// force a transaction for batched queries for certain drivers
				var forceBatch bool
				if h.u != nil {
					_, _, forceBatch = drivers.IsBatchQueryPrefix(h.u, stmt.FindPrefix(h.lastExec, true, true, true))
					forceBatch = forceBatch && drivers.BatchAsTransaction(h.u)
				}
				// execute
				out := stdout
				if h.out != nil {
					out = h.out
				}
				ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
				if err = h.Execute(ctx, out, opt, h.lastExecPrefix, h.lastExec, forceBatch, h.unbind()...); err != nil {
					lastErr = WrapErr(h.lastExec, err)
					if env.Get("ON_ERROR_STOP") == "on" {
						if iactive {
							fmt.Fprintln(stderr, "error:", err)
							h.buf.Reset([]rune{}) // empty the buffer so no other statements are run
							continue
						} else {
							stop()
							return err
						}
					} else {
						fmt.Fprintln(stderr, "error:", err)
					}
				}
				stop()
			}
		}
	}
}

// apply applies the command against the handler.
func (h *Handler) apply(stdout, stderr io.Writer, cmd, paramstr string) (metacmd.Option, bool, error) {
	// cmd = strings.TrimPrefix(cmd, `\`)
	params := stmt.NewParams(paramstr)
	// decode
	f, err := metacmd.Decode(cmd, params)
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
	// run
	opt, err := f(h)
	if err != nil && err != rline.ErrInterrupt {
		fmt.Fprintln(stderr, "error:", err)
		return metacmd.Option{}, true, WrapErr(cmd, err)
	}
loop:
	// print unused command parameters
	for {
		switch arg, ok, err := params.Arg(); {
		case err != nil:
			fmt.Fprintln(stderr, "error:", err)
		case !ok:
			break loop
		default:
			fmt.Fprintln(stdout, fmt.Sprintf(text.ExtraArgumentIgnored, cmd, arg))
		}
	}
	return opt, false, nil
}

// outputHighlighter returns s as a highlighted string, based on the current
// buffer and syntax highlighting settings.
func (h *Handler) outputHighlighter(s string) string {
	// bail when string is empty (ie, contains no printable, non-space
	// characters) or if syntax highlighting is not enabled
	if empty(s) || env.Get("SYNTAX_HL") != "true" {
		return s
	}
	// count end lines
	var endl string
	if m := lineendRE.FindStringSubmatch(s); m != nil {
		s = strings.TrimSuffix(s, m[0])
		endl += m[0]
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
			return !isSpaceOrControl(r)
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
		cmd, _, err = st.Next(env.Untick(h.user, env.Vars(), false))
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
				return !isSpaceOrControl(r)
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

// Execute executes a query against the connected database.
func (h *Handler) Execute(ctx context.Context, w io.Writer, opt metacmd.Option, prefix, sqlstr string, forceTrans bool, bind ...interface{}) error {
	if h.db == nil {
		return text.ErrNotConnected
	}
	// determine type and pre process string
	prefix, sqlstr, qtyp, err := drivers.Process(h.u, prefix, sqlstr)
	if err != nil {
		return drivers.WrapErr(h.u.Driver, err)
	}
	// start a transaction if forced
	if forceTrans {
		if err = h.BeginTx(ctx, nil); err != nil {
			return err
		}
	}
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
	if err = drivers.WrapErr(h.u.Driver, f(ctx, w, opt, prefix, sqlstr, qtyp, bind)); err != nil {
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
	h.lastExec, h.lastExecPrefix, h.lastPrint, h.lastRaw, h.batch, h.batchEnd = "", "", "", "", false, ""
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
//	testdb=> \set PROMPT1 '%[%033[1;33;40m%]%n@%/%R%[%033[0m%]%# '
//
//	results in a boldfaced (1;) yellow-on-black (33;40) prompt on
//	VT100-compatible, color-capable terminals.
//
//	%w - Whitespace of the same width as the most recent output of PROMPT1.
//	This can be used as a PROMPT2 setting, so that multi-line statements are
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
					s = dburl.ShortAlias(h.u.Driver)
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
			if h.tx != nil || h.batch {
				buf = append(buf, '~')
			} else {
				buf = append(buf, '>')
			}
		// case 'p': // the process id of the connected backend -- never going to be supported
		case 'R': // statement state
			buf = append(buf, h.buf.State()...)
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

// LastExec returns the last executed statement.
func (h *Handler) LastExec() string {
	return h.lastExec
}

// LastPrint returns the last printable statement.
func (h *Handler) LastPrint() string {
	return h.lastPrint
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
	// create lexer, formatter, styler
	l := chroma.Coalesce(drivers.Lexer(h.u))
	f := formatters.Get(env.Get("SYNTAX_HL_FORMAT"))
	s := styles.Get(env.Get("SYNTAX_HL_STYLE"))
	// override background
	if env.Get("SYNTAX_HL_OVERRIDE_BG") != "false" {
		s = ustyles.Get(env.Get("SYNTAX_HL_STYLE"))
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
	// set buffer options
	drivers.ConfigStmt(h.u, h.buf)
	// force error/check connection
	if err == nil {
		if err = drivers.Ping(ctx, h.u, h.db); err == nil {
			if h.l.Interactive() {
				h.l.SetOutput(h.outputHighlighter)
				h.l.Completer(completer.NewDefaultCompleter(completer.WithConnStrings(h.connStrings())))
				h.l.Completer(drivers.NewCompleter(ctx, h.u, h.db, readerOpts(), completer.WithConnStrings(h.connStrings())))
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
		_, aliases := dburl.SchemeDriverAndAliases(schema)
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
	if env.Get("SHOW_HOST_INFORMATION") != "true" || !h.l.Interactive() {
		return nil
	}
	if h.db == nil {
		return text.ErrNotConnected
	}
	ver, err := drivers.Version(ctx, h.u, h.DB())
	switch {
	case err != nil:
		ver = fmt.Sprintf("<unknown, error: %v>", err)
	case ver == "":
		ver = "<unknown>"
	}
	h.Print(text.ConnInfo, h.u.Driver, ver)
	return nil
}

// Print formats according to a format specifier and writes to handler's standard output.
func (h *Handler) Print(s string, v ...interface{}) {
	if env.Get("QUIET") == "on" {
		return
	}
	fmt.Fprintln(h.l.Stdout(), fmt.Sprintf(s, v...))
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
	stdout, _, _ := h.l.Stdout(), h.l.Stderr(), h.l.Interactive()
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
	echarts := echartsgoja.New(echartsgoja.WithWidthHeight(cfg.W, cfg.H))
	res, err := echarts.RenderOptions(ctx, data)
	if err != nil {
		return err
	}
	if cfg.File != "" {
		fmt.Println("writing to", cfg.File)
		return os.WriteFile(cfg.File, []byte(res), 0o644)
	}
	img, err := resvg.Render([]byte(res), resvg.WithBackground(cfg.Background))
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
		fmt.Fprintln(h.l.Stdout(), fmt.Sprintf(s, v...))
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
		fmt.Fprintln(h.l.Stdout(), fmt.Sprintf(s, v...))
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
				pipe, cmd, err = env.Pipe(h.l.Stdout(), h.l.Stderr(), pipeName[1:])
			} else {
				pipe, err = os.OpenFile(pipeName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
			}
			if err != nil {
				return err
			}
			w = pipe
		}
	} else if opt.Exec != metacmd.ExecWatch {
		params["pager_cmd"] = env.Get("PAGER")
	}
	// set up column type config
	var extra []tblfmt.Option
	switch f := drivers.ColumnTypes(h.u); {
	case f != nil:
		extra = append(extra, tblfmt.WithColumnTypesFunc(f))
	case drivers.UseColumnTypes(h.u):
		extra = append(extra, tblfmt.WithUseColumnTypes(true))
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
	// encode and handle error conditions
	switch err := tblfmt.EncodeAll(w, resultSet, params, extra...); {
	case err != nil && cmd != nil && errors.Is(err, syscall.EPIPE):
		// broken pipe means pager quit before consuming all data, which might be expected
		return nil
	case err != nil && h.u.Driver == "sqlserver" && err == tblfmt.ErrResultSetHasNoColumns && strings.HasPrefix(typ, "EXEC"):
		// sqlserver EXEC statements sometimes do not have results, fake that
		// it was executed as a exec and not a query
		fmt.Fprintln(w, typ)
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
				if err = h.Execute(ctx, w, res, stmt.FindPrefix(sqlstr, true, true, true), sqlstr, false); err != nil {
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
	if env.Get("QUIET") == "off" {
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
	if err := tx.Commit(); err != nil {
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
	if err := tx.Rollback(); err != nil {
		return drivers.WrapErr(h.u.Driver, err)
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

// IncludeReader includes the content of rdr.
func (h *Handler) IncludeReader(rdr io.Reader, path string) error {
	r := bufio.NewReader(rdr)
	// setup rline
	l := &rline.Rline{
		N: func() ([]rune, error) {
			buf := new(bytes.Buffer)
			var b []byte
			var isPrefix bool
			var err error
			for {
				// read
				b, isPrefix, err = r.ReadLine()
				// when not EOF
				if err != nil && err != io.EOF {
					return nil, err
				}
				// append
				if _, werr := buf.Write(b); werr != nil {
					return nil, werr
				}
				// end of line
				if !isPrefix || err != nil {
					break
				}
			}
			// peek and read possible line ending \n or \r\n
			if err != io.EOF {
				if err := peekEnding(buf, r); err != nil {
					return nil, err
				}
			}
			return []rune(buf.String()), err
		},
		Out: h.l.Stdout(),
		Err: h.l.Stderr(),
		Pw:  h.l.Password,
	}
	p := New(l, h.user, filepath.Dir(path), h.charts, h.nopw)
	p.db, p.u = h.db, h.u
	drivers.ConfigStmt(p.u, p.buf)
	err := p.Run()
	h.db, h.u = p.db, p.u
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

// MetadataWriter loads the metadata writer for the
func (h *Handler) MetadataWriter(ctx context.Context) (metadata.Writer, error) {
	if h.db == nil {
		return nil, text.ErrNotConnected
	}
	return drivers.NewMetadataWriter(ctx, h.u, h.db, h.l.Stdout(), readerOpts()...)
}

// GetOutput gets the output writer.
func (h *Handler) GetOutput() io.Writer {
	if h.out == nil {
		return h.l.Stdout()
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

func readerOpts() []metadata.ReaderOption {
	var opts []metadata.ReaderOption
	if env.Get("ECHO_HIDDEN") == "on" || env.Get("ECHO_HIDDEN") == "noexec" {
		if env.Get("ECHO_HIDDEN") == "noexec" {
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

// peekEnding peeks to see if the next successive bytes in r is \n or \r\n,
// writing to w if it is. Does not advance r if the next bytes are not \n or
// \r\n.
func peekEnding(w io.Writer, r *bufio.Reader) error {
	// peek first byte
	buf, err := r.Peek(1)
	switch {
	case err != nil && err != io.EOF:
		return err
	case err == nil && buf[0] == '\n':
		if _, rerr := r.ReadByte(); err != nil && err != io.EOF {
			return rerr
		}
		_, werr := w.Write([]byte{'\n'})
		return werr
	case err == nil && buf[0] != '\r':
		return nil
	}
	// peek second byte
	buf, err = r.Peek(1)
	switch {
	case err != nil && err != io.EOF:
		return err
	case err == nil && buf[0] != '\n':
		return nil
	}
	if _, rerr := r.ReadByte(); err != nil && err != io.EOF {
		return rerr
	}
	_, werr := w.Write([]byte{'\n'})
	return werr
}

// grab returns the i'th rune from r when i < end, otherwise 0.
func grab(r []rune, i, end int) rune {
	if i < end {
		return r[i]
	}
	return 0
}

// empty reports whether s contains at least one printable, non-space
// character.
func empty(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return unicode.IsPrint(r) && !unicode.IsSpace(r)
	}) == -1
}

// lastcolor returns the last defined color in s, if any.
func lastcolor(s string) string {
	if i := strings.LastIndex(s, "\n"); i != -1 {
		s = s[:i]
	}
	if i := strings.LastIndex(s, "\x1b[0m"); i != -1 {
		s = s[i+4:]
	}
	return strings.Join(ansiRE.FindAllString(s, -1), "")
}

// isSpaceOrControl returns true when r is a space or control character.
func isSpaceOrControl(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r)
}

// lastIndex returns the last index in r of needle, or -1 if not found.
func lastIndex(r []rune, needle rune) int {
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] == needle {
			return i
		}
	}
	return -1
}

// ansiRE matches ansi escape (color) codes.
var ansiRE = regexp.MustCompile(`\x1b[[0-9]+([:;][0-9]+)*m`)

// lineendRE is the end of line terminal.
var lineendRE = regexp.MustCompile(`(?:\r?\n)+$`)

// helpQuitExitRE is a regexp to use to match help, quit, or exit messages.
var helpQuitExitRE = regexp.MustCompile(`(?im)^+(` + strings.Join([]string{text.HelpPrefix, text.QuitPrefix, text.ExitPrefix}, "|") + `)\s*$`)
