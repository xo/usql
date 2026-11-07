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
	"github.com/kenshaw/rasterm"
	"github.com/xo/dbfwd"
	"github.com/xo/dbmeta"
	"github.com/xo/dburl"
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
	"github.com/xo/usql/pager"
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
	// historyWarned is set once the handler has said that the history file
	// cannot be read or written.
	historyWarned bool
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
	// cond is the stack of \if blocks of src. A branch that does not run
	// sends no statement, and runs no meta command other than \if, \elif,
	// \else and \endif.
	cond conds
	// lastExec is the last executed query statement.
	lastExec string
	// lastRaw is the last executed raw query statement.
	lastRaw string
	// database is the current database that dbmeta reports, which the
	// prompt's %/ shows, and which USE changes.
	database sql.Null[string]
	// bind are bound values for that will be used for statement execution.
	bind []interface{}
	// u is the active connection information. For a forward, it is the URL
	// of the local end, and it holds the password.
	u *dburl.URL
	// fwd is the forward that dbfwd runs for the connection, or nil.
	fwd *dbfwd.Forward
	// fwdURL is the URL that fwd resolved, such as k+cnpg://pg.prod/app. It
	// holds no password.
	fwdURL *dburl.URL
	// fwdDSN is the text of fwdURL, which \c with no argument opens again.
	fwdDSN string
	// db is the pool of the active database connection.
	db *sql.DB
	// conn is the one connection of db that the statements of the session
	// run on, taken when the session starts and again after it is lost.
	conn *drivers.Session
	// meta reads the structure of the database, and is built the first time
	// it is needed.
	meta *dbmeta.Meta
	// tx is the active transaction, if any.
	tx *sql.Tx
	// out file or pipe
	out io.WriteCloser
	// page is where the entry that runs now writes its output. Stdout makes it
	// the first time an entry writes, and endPage closes it when the entry
	// ends.
	page *pager.Writer
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
		l.SetCompleter(complete.New(complete.WithCommands(metacmd.Completions()), complete.WithConnStrings(h.connStrings()), complete.WithKeywords(h.keywords)))
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
	h.l.SetPrompt("", h.render(env.Setting("prompt2"), true))
	return h.render(env.Setting("prompt1"), false)
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
	stderr := h.l.Stderr()
	// display welcome info
	if h.iactive && env.Setting("quiet") == "off" {
		// logo
		if typ := env.TermGraphics(); typ.Available() {
			if err := typ.Encode(h.l, text.Logo); err != nil {
				return err
			}
		}
		// welcome text
		fmt.Fprintln(h.l, text.WelcomeDesc)
		fmt.Fprintln(h.l)
	}
	// the output of the last entry goes to the pager or the terminal before
	// the handler stops
	defer h.endPage()
	var lastErr error
	for {
		// the pager of the last entry must exit before the next line is read
		h.endPage()
		h.checkHistory(stderr)
		// the keys typed while rasterm asked the terminal whether it draws
		// images go to the line editor, before the terminal
		if b := rasterm.PendingInput(); h.iactive && len(b) != 0 {
			h.l.PushInput(b)
		}
		it, err := h.src.Next()
		switch {
		case errors.Is(err, rline.ErrInterrupted):
			h.buf.Reset()
			h.src.Reset()
			// Ctrl-C also leaves the innermost \if block, as in psql
			if len(h.cond) != 0 {
				h.cond = h.cond[:len(h.cond)-1]
				fmt.Fprintln(stderr, "error:", text.ErrIfEscaped)
			}
			continue
		case errors.Is(err, io.EOF):
			// input that is not typed, from -c, -f, a pipe or \i, sends what
			// is left in the buffer at its end, as psql does, unless the
			// input ends in a branch that does not run
			if (h.singleLineMode || !h.iactive) && !h.buf.Empty() && h.cond.active() {
				stop, err := h.send(h.Stdout(), stderr, metacmd.Option{})
				if err != nil {
					lastErr = err
				}
				if stop {
					return err
				}
			}
			// each input closes its own \if blocks. psql reports the error
			// and keeps its exit status, unless on_error_stop is on.
			if len(h.cond) != 0 {
				fmt.Fprintln(stderr, "error:", text.ErrUnclosedIf)
				if env.Setting("on_error_stop") == "on" && !h.iactive {
					return WrapErr("", text.ErrUnclosedIf)
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
					fmt.Fprintln(h.Stdout(), text.HelpDesc)
					continue
				}
				return nil
			}
			// a branch that does not run drops its statements, so they do
			// not join the query buffer, and their variables are not
			// expanded
			if !h.cond.active() {
				if h.iactive && it.Terminated {
					fmt.Fprintln(stderr, text.QueryIgnoredUseEndIf)
				}
				continue
			}
			h.buf.Add(it)
			if !it.Terminated {
				continue
			}
		case feed.Command:
			var cont bool
			opt, cont, lastErr = h.apply(h.Stdout(), stderr, it)
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
		stop, err := h.send(h.Stdout(), stderr, opt)
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
	exp, raw := h.buf.Expand(h.Lookup), h.buf.Raw()
	if !h.buf.Empty() {
		h.lastExec, h.lastRaw = exp, raw
		h.buf.Reset()
	}
	if h.lastExec == "" || h.lastExec == ";" {
		return false, nil
	}
	out := stdout
	switch {
	case h.out != nil:
		out = h.out
	case opt.Exec == metacmd.ExecWatch:
		// \watch writes until it is interrupted, so a pager would never see
		// the end of its output
		out = h.l
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err := h.Execute(ctx, out, opt, h.lastExec, h.unbind()...)
	if err == nil {
		return false, nil
	}
	// the error goes after the output, not into the middle of it
	h.endPage()
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
		case errors.Is(err, text.ErrUnknownCommand):
			fmt.Fprintln(stderr, fmt.Sprintf(text.InvalidCommand, cmd))
		case errors.Is(err, text.ErrMissingRequiredArgument):
			fmt.Fprintln(stderr, fmt.Sprintf(text.MissingRequiredArg, cmd))
		default:
			fmt.Fprintln(stderr, "error:", err)
		}
		return metacmd.Option{}, true, err
	}
	// a branch that does not run runs no meta command but \if, \elif, \else
	// and \endif. psql reports an unknown command there too.
	if !h.cond.active() && !isConditional(cmd) {
		if h.iactive {
			fmt.Fprintf(stderr, text.CommandIgnoredUseEndIf+"\n", cmd)
		}
		return metacmd.Option{}, true, nil
	}
	opt, unread, err := f(h)
	if err != nil && !errors.Is(err, rline.ErrInterrupted) {
		// the error goes after the output, not into the middle of it
		h.endPage()
		fmt.Fprintln(stderr, "error:", err)
		return metacmd.Option{}, true, WrapErr(cmd, err)
	}
	for _, arg := range unread {
		fmt.Fprintln(stdout, fmt.Sprintf(text.ExtraArgumentIgnored, cmd, arg))
	}
	return opt, false, nil
}

// Execute executes a query against the connected database.
func (h *Handler) Execute(ctx context.Context, w io.Writer, opt metacmd.Option, sqlstr string, bind ...interface{}) error {
	if h.db == nil {
		return text.ErrNotConnected
	}
	// no statement reaches a database whose forward is lost
	if err := h.forwardLost(); err != nil {
		return err
	}
	// the command tag, the statement to send, and whether it returns rows,
	// from dbmeta
	prefix, sqlstr, qtyp := drivers.Process(h.u, sqlstr)
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
	err := f(ctx, w, opt, prefix, sqlstr, qtyp, bind)
	// a driver says that the connection is bad only before the statement
	// ran, so it runs once more on a new connection, as database/sql runs a
	// statement of its pool
	if h.lost(err) {
		fmt.Fprintln(h.l.Stderr(), text.ConnectionReset)
		err = f(ctx, w, opt, prefix, sqlstr, qtyp, bind)
		h.lost(err)
	}
	// USE chooses another database, which the prompt shows
	if err == nil && strings.EqualFold(prefix, "USE") {
		h.readDatabase(ctx)
	}
	// a forward that is lost explains the error of the driver
	if lost := h.forwardLost(); err != nil && lost != nil {
		return lost
	}
	return drivers.WrapErr(h.u.SchemeName, err)
}

// checkHistory says once, at a terminal, that the history file could not be
// read or written, as the line editor goes on without it and says nothing.
func (h *Handler) checkHistory(w io.Writer) {
	if !h.iactive || h.historyWarned {
		return
	}
	if err := h.l.HistoryErr(); err != nil {
		h.historyWarned = true
		fmt.Fprintf(w, text.HistoryNotSaved+"\n", err)
	}
}

// readDatabase reads the current database from dbmeta, for the prompt. A
// database that dbmeta cannot ask leaves it unknown, and the prompt shows the
// database of the URL.
func (h *Handler) readDatabase(ctx context.Context) {
	h.database = sql.Null[string]{}
	if !h.iactive {
		return
	}
	m, err := h.metadata(ctx)
	if err != nil {
		return
	}
	if d, ok, err := dbmeta.First(dbmeta.CurrentDatabase.All(ctx, m, h.DB(), nil)); err == nil && ok {
		h.database = sql.Null[string]{V: d.Name.V, Valid: true}
	}
}

// Reset empties the query buffer and forgets the last query.
func (h *Handler) Reset() {
	h.buf.Reset()
	h.lastExec, h.lastRaw = "", ""
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
	// the text has \if blocks of its own, as a file has
	src, cond := h.src, h.cond
	defer func() { h.src, h.cond = src, cond }()
	h.cond = nil
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
	return h.render(prompt, false)
}

// render expands the escapes of prompt. cont is set for the prompt of the rows
// after the first, where %R is always -.
func (h *Handler) render(prompt string, cont bool) string {
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
				buf = append(buf, h.host().Hostname()...)
			}
		case 'm': // host name truncated at first dot, or [local] if it's a domain socket
			if connected {
				s := h.host().Hostname()
				if i := strings.Index(s, "."); i != -1 {
					s = s[:i]
				}
				buf = append(buf, s...)
			}
		case '>': // the port number
			if connected {
				s := h.host().Port()
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
			case connected && h.database.Valid:
				// the current database, which follows USE, after the / that
				// the path of a URL has
				if h.database.V != "" {
					buf = append(buf, "/"+h.database.V...)
				}
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
			// = for a new statement, @ for a new statement in a branch of
			// \if that does not run, and - while the buffer holds a
			// statement or the entry continues on another row. The line
			// editor draws one prompt for the rows after the first, so the
			// state inside a string or a comment is not shown.
			switch {
			case cont || !h.buf.Empty():
				buf = append(buf, '-')
			case !h.cond.active():
				buf = append(buf, '@')
			default:
				buf = append(buf, '=')
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

// host returns the URL whose host and port the prompt shows. For a forward,
// it is the URL that the forward resolved, such as pg.prod in
// k+cnpg://pg.prod/app, because the local port changes on each connection.
func (h *Handler) host() *dburl.URL {
	if h.fwd != nil {
		return h.fwdURL
	}
	return h.u
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

// DB returns where the statements of the session run: the transaction when
// one is open, and otherwise the connection of the session. When the session
// has no connection, and one cannot be taken, it returns the pool.
func (h *Handler) DB() drivers.DB {
	switch {
	case h.tx != nil:
		return h.tx
	case h.db == nil:
		return nil
	case h.conn == nil:
		conn, err := drivers.NewSession(context.Background(), h.db)
		if err != nil {
			return h.db
		}
		h.conn = conn
	}
	return h.conn
}

// lost drops the connection of the session when err says that it is gone, so
// that the next statement takes a new one, and reports whether it did. The
// state of the old session, such as a SET, is gone with it.
func (h *Handler) lost(err error) bool {
	if h.conn == nil || h.tx != nil || !drivers.IsConnLost(err) {
		return false
	}
	_ = h.conn.Close()
	h.conn = nil
	return true
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
		// with no argument, \c opens a forward again, as after it was lost
		if h.fwdDSN == "" {
			return nil
		}
		params = []string{h.fwdDSN}
	}
	if h.tx != nil {
		return text.ErrPreviousTransactionExists
	}
	if len(params) == 1 {
		if v, ok := env.Vars().GetConn(params[0]); ok {
			params = v
		}
	}
	var u, fwdURL *dburl.URL
	var fwd *dbfwd.Forward
	if len(params) < 2 {
		dsn := params[0]
		// parse dsn
		var err error
		if u, err = dburl.Parse(dsn); err != nil {
			return err
		}
		// a URL such as k+cnpg://pg.prod/app opens through a forward. The
		// current connection stays when the forward cannot start.
		if dbfwd.Handles(u) {
			if fwd, err = h.forward(ctx, u); err != nil {
				return err
			}
			fwdURL = u
			if u, err = fwd.URL(ctx); err != nil {
				_ = fwd.Close()
				return err
			}
		}
		// force parameters
		h.forceParams(u)
	} else {
		u = &dburl.URL{
			Driver: params[0],
			DSN:    strings.Join(params[1:], " "),
		}
	}
	// the new connection replaces the current one
	h.drop()
	h.u, h.fwd, h.fwdURL, h.fwdDSN = u, fwd, fwdURL, ""
	if fwd != nil {
		h.fwdDSN = params[0]
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
				c := complete.New(complete.WithCommands(metacmd.Completions()), complete.WithConnStrings(h.connStrings()), complete.WithKeywords(h.keywords))
				if m, err := h.metadata(ctx); err == nil {
					c = complete.New(
						complete.WithCommands(metacmd.Completions()),
						complete.WithSource(describe.Source(m, h.db)),
						complete.WithConnStrings(h.connStrings()),
						complete.WithKeywords(h.keywords),
					)
				}
				h.l.SetCompleter(c)
			}
			h.readDatabase(ctx)
			return h.Version(ctx)
		}
	}
	// bail without getting password. A forward reads its password where the
	// database runs, so a prompt cannot help.
	if h.nopw || h.fwd != nil || !drivers.IsPasswordErr(h.u, err) || len(params) > 1 || !h.l.Interactive() {
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
	available := drivers.Available()
	names := make([]string, 0, len(available))
	for schema := range available {
		_, aliases := dburl.SchemeNameAndAliases(schema)
		// TODO should we create all combinations of space, :, :// and +transport ?
		names = append(names, schema)
		names = append(names, aliases...)
	}
	// the schemes that dbfwd opens through a forward
	if name, aliases := dburl.SchemeNameAndAliases("k8s"); name != "" && dbfwd.Handles(&dburl.URL{SchemeName: name}) {
		names = append(names, name)
		names = append(names, aliases...)
	}
	return append(names, slices.Sorted(maps.Keys(env.Vars().Conn()))...)
}

// forceParams forces connection parameters on a database URL, adding any
// driver specific required parameters.
func (h *Handler) forceParams(u *dburl.URL) {
	drivers.ForceParams(u)
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
	if dbfwd.Handles(u) {
		return "", text.ErrPasswordForward
	}
	user := h.user.Username
	if u.User != nil {
		user = u.User.Username()
	}
	// the output so far goes out before the question
	h.endPage()
	pass, err := h.l.Password(text.EnterPassword)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword(user, pass)
	return u.String(), nil
}

// Close closes the database connection if it is open, and stops its forward.
func (h *Handler) Close() error {
	if h.tx != nil {
		return text.ErrPreviousTransactionExists
	}
	var drv string
	if h.u != nil {
		drv = h.u.SchemeName
	}
	err := h.drop()
	h.u, h.fwdURL, h.fwdDSN = nil, nil, ""
	return drivers.WrapErr(drv, err)
}

// drop closes the connection of the session, the pool and the forward, and
// returns the error of closing the pool.
func (h *Handler) drop() error {
	var err error
	if h.conn != nil {
		_ = h.conn.Close()
	}
	if h.db != nil {
		err = h.db.Close()
	}
	if h.fwd != nil {
		_ = h.fwd.Close()
	}
	h.db, h.conn, h.fwd, h.meta, h.database = nil, nil, nil, nil, sql.Null[string]{}
	return err
}

// forward starts the forward for u. At a terminal, kubectl and the other
// tools that dbfwd runs get the terminal, so that a login can prompt.
func (h *Handler) forward(ctx context.Context, u *dburl.URL) (*dbfwd.Forward, error) {
	var opts []dbfwd.Option
	if h.l.Interactive() {
		opts = append(opts, dbfwd.WithInteractive(h.l.Release))
	}
	return dbfwd.Resolve(ctx, u, opts...)
}

// Forward returns the URL that the forward of the connection resolved, and
// the forward. Both are nil when the connection has no forward.
func (h *Handler) Forward() (*dburl.URL, *dbfwd.Forward) {
	if h.fwd == nil {
		return nil, nil
	}
	return h.fwdURL, h.fwd
}

// forwardLost returns why the forward of the connection was lost, and nil
// while it runs or when there is none.
func (h *Handler) forwardLost() error {
	if h.fwd == nil {
		return nil
	}
	select {
	case <-h.fwd.Done():
	default:
		return nil
	}
	if err := h.fwd.Err(); err != nil {
		return fmt.Errorf(text.ForwardLost, err)
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
	// the output so far goes out before the question
	h.endPage()
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
	// the output so far goes out before the question
	h.endPage()
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
	// the message goes to the terminal and not to the pager, as in psql, so
	// that a long version does not open the pager
	if env.Setting("quiet") != "on" {
		fmt.Fprintln(h.l, fmt.Sprintf(text.ConnInfo, h.u.SchemeName, ver))
		if h.fwd != nil {
			fmt.Fprintln(h.l, fmt.Sprintf(text.ForwardInfo, h.fwd))
		}
	}
	return nil
}

// Print formats according to a format specifier and writes to handler's standard output.
func (h *Handler) Print(s string, v ...interface{}) {
	if env.Setting("quiet") == "on" {
		return
	}
	fmt.Fprintln(h.Stdout(), fmt.Sprintf(s, v...))
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
		// as psql does, the time goes to the terminal after the pager exits
		h.endPage()
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
		// as psql does, the time goes to the terminal after the pager exits
		h.endPage()
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
	// align each column by its type, as the model of the dialect records it
	extra = append(extra, tblfmt.WithColumnAlign(columnAlign(dbmeta.Dialect(h.u.Dialect))))
	// show the rows in groups, so that a large result does not have to fit
	// in memory, as psql's FETCH_COUNT does
	if n, _ := strconv.Atoi(env.Setting("fetch_count")); n > 0 {
		extra = append(extra, tblfmt.WithCount(n))
	}
	// print a binary value as psql prints a bytea, on every database
	extra = append(extra, tblfmt.WithBytesFormat(dbmeta.Dialect(h.u.Dialect).FormatBinary))
	// encode and handle error conditions
	if indent := env.JSONIndent(params); indent != "" {
		extra = append(extra, tblfmt.WithIndent(indent))
	}
	switch err = tblfmt.EncodeAll(w, resultSet, params, extra...); {
	case err != nil && cmd != nil && errors.Is(err, syscall.EPIPE):
		// the command of \g |command quit before it read all the output, as
		// head does
		err = nil
	case err == nil && env.BlankLineAfter(params):
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
				if err = h.Execute(ctx, w, res, sqlstr); err != nil {
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
	// the transaction runs on the connection of the session
	db, ok := h.DB().(interface {
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	})
	if !ok {
		return text.ErrNotConnected
	}
	var err error
	h.tx, err = db.BeginTx(ctx, txOpts)
	if err != nil {
		h.lost(err)
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

// If starts a conditional block, as \if does. It calls eval for the value of
// the expression only when the block starts in a branch that runs.
func (h *Handler) If(eval func() (bool, error)) error {
	return h.cond.begin(eval)
}

// ElseIf starts the next branch of the conditional block, as \elif does. It
// calls eval only when no branch of the block ran yet.
func (h *Handler) ElseIf(eval func() (bool, error)) error {
	return h.cond.elif(eval)
}

// Else starts the last branch of the conditional block, as \else does.
func (h *Handler) Else() error {
	return h.cond.els()
}

// EndIf ends the conditional block, as \endif does.
func (h *Handler) EndIf() error {
	return h.cond.end()
}

// IncludeReader includes the content of rdr, as \i does. A handler of its
// own reads rdr, on the same connection, and the connection it ends with is
// kept, so that a \c in the file applies after it.
//
// rdr is decoded from the encoding setting, or from the byte order mark at
// its start (see feed.Decode).
func (h *Handler) IncludeReader(rdr io.Reader, path string) error {
	rdr, err := feed.Decode(rdr, env.Setting("encoding"))
	if err != nil {
		return err
	}
	p, err := New(h.l, feed.NewLines(rdr), h.user, filepath.Dir(path), h.charts, h.nopw)
	if err != nil {
		return err
	}
	p.db, p.conn, p.tx, p.u, p.meta = h.db, h.conn, h.tx, h.u, h.meta
	p.fwd, p.fwdURL, p.fwdDSN = h.fwd, h.fwdURL, h.fwdDSN
	if err := p.src.SetLanguage(drivers.Language(p.u)); err != nil {
		return err
	}
	err = p.Run()
	if errors.Is(err, feed.ErrInvalidUTF8) {
		err = fmt.Errorf("%s: %w", path, err)
	}
	h.db, h.conn, h.tx, h.u, h.meta = p.db, p.conn, p.tx, p.u, p.meta
	h.fwd, h.fwdURL, h.fwdDSN = p.fwd, p.fwdURL, p.fwdDSN
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
	opts := append(formatterOpts(), describe.WithHighlight(h.Highlight))
	return describe.NewFormatter(m, h.DB(), h.GetOutput(), h.u.SchemeName, opts...), nil
}

// Highlight returns s in the colors of syntax highlighting, as the escape
// sequences of a terminal, when the input is highlighted and the output goes
// to the terminal. Otherwise, such as with \o or when the output is not a
// terminal, it returns s as it is.
func (h *Handler) Highlight(s string) string {
	if h.hl == nil || h.out != nil {
		return s
	}
	if width, _ := h.l.Size(); width <= 0 {
		return s
	}
	return h.hl.Text(s)
}

// metadata returns dbmeta's model of the database. It is read once for each
// connection.
func (h *Handler) metadata(ctx context.Context) (*dbmeta.Meta, error) {
	if h.db == nil {
		return nil, text.ErrNotConnected
	}
	if h.meta == nil {
		m, err := drivers.Meta(ctx, h.u, h.DB())
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
		return h.Stdout()
	}
	return h.out
}

// Stdout returns where the entry that runs now writes its output. When the
// output is taller or wider than the terminal, it goes to the pager, as \pset
// pager and \pset pager_min_lines say.
func (h *Handler) Stdout() io.Writer {
	if h.page == nil {
		h.page = pager.New(h.l, h.pagerConfig())
	}
	return h.page
}

// endPage writes the output of the entry that ran to the terminal, or waits for
// its pager to exit.
func (h *Handler) endPage() {
	if h.page == nil {
		return
	}
	page := h.page
	h.page = nil
	if err := page.Close(); err != nil {
		fmt.Fprintln(h.l.Stderr(), "error:", err)
	}
}

// pagerConfig returns the configuration of the pager from the settings. The
// pager runs only when the line editor writes to a terminal, which has a size.
func (h *Handler) pagerConfig() pager.Config {
	params := env.Vars().Print()
	mode, command := params["pager"], env.Setting("pager")
	width, height := h.l.Size()
	if mode == "off" || command == "" || width <= 0 || height <= 0 {
		return pager.Config{}
	}
	if c, _ := strconv.Atoi(params["columns"]); c > 0 {
		width = c
	}
	minLines, _ := strconv.Atoi(params["pager_min_lines"])
	shell, param := env.Getshell()
	if shell == "" {
		return pager.Config{}
	}
	return pager.Config{
		Command:  command,
		Shell:    []string{shell, param},
		Always:   mode == "always",
		Width:    width,
		Height:   height,
		MinLines: minLines,
		Terminal: h.l.Release,
	}
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

// keywords returns the keywords that the grammar of the database accepts
// after stmt, for the completer. Before a connection, the grammar is the one
// of PostgreSQL, as in psql.
func (h *Handler) keywords(stmt string) []string {
	var dialect string
	if h.u != nil {
		dialect = h.u.Dialect
	}
	words, _ := highlight.Keywords(context.Background(), dialect, []byte(stmt))
	return words
}

// columnAlign returns the alignment of a column of d, from its type. It
// returns false when the model of d records no alignment for the type, and the
// encoder then aligns the column by its values.
func columnAlign(d dbmeta.Dialect) func(*sql.ColumnType) (tblfmt.Align, bool) {
	return func(col *sql.ColumnType) (tblfmt.Align, bool) {
		switch align, ok := d.ColumnAlign(col); {
		case !ok:
			return 0, false
		case align == dbmeta.AlignRight:
			return tblfmt.AlignRight, true
		}
		return tblfmt.AlignLeft, true
	}
}
