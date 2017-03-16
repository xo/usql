package handler

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/knq/dburl"
	"github.com/olekukonko/tablewriter"

	"github.com/knq/usql/handler/buf"
)

// Handler is a input process handler.
type Handler struct {
	histfile            string
	root, wd            string
	interactive, cygwin bool

	u  *dburl.URL
	db *sql.DB

	// parse settings
	allowdollar, allowmc bool

	// accumulated buffer
	buf *buf.Buf

	// quoted string state
	q       bool
	qdbl    bool
	qdollar bool
	qid     string

	// multicomment state
	mc bool

	// balanced paren count
	b int
}

// New creates a new input handler.
func New(histfile, root string, interactive, cygwin bool) (*Handler, error) {
	return &Handler{
		histfile:    histfile,
		root:        root,
		wd:          root,
		interactive: interactive,
		cygwin:      cygwin,
		buf:         new(buf.Buf),
	}, nil
}

// ForceInteractive forces the interactive mode.
func (h *Handler) ForceInteractive(interactive bool) {
	h.interactive = interactive
}

// HistoryFile returns the history file name for the handler.
func (h *Handler) HistoryFile() string {
	return h.histfile
}

// SetPrompt sets the prompt on a readline instance.
func (h *Handler) SetPrompt(l *readline.Instance) {
	if !h.interactive {
		return
	}

	s := notConnected

	if h.db != nil {
		s = h.u.Short()
	}

	state := "="
	switch {
	case h.q && h.qdollar:
		state = "$"

	case h.q && h.qdbl:
		state = `"`

	case h.q:
		state = "'"

	case h.mc:
		state = "*"

	case h.b != 0:
		state = "("

	case h.buf.Len != 0:
		state = "-"
	}

	l.SetPrompt(s + state + "> ")
}

// Open handles opening a specified database URL.
func (h *Handler) Open(urlstr string) error {
	if urlstr == "" {
		return nil
	}

	// parse dsn
	var err error
	h.u, err = dburl.Parse(urlstr)
	switch {
	case err == dburl.ErrInvalidDatabaseScheme:
		fi, err := os.Stat(urlstr)
		switch {
		case os.IsNotExist(err):
			return nil
		case err != nil:
			return err
		}

		// TODO: add support for postgres unix domain sockets
		if fi.Mode()&os.ModeSocket != 0 {
			return h.Open("mysql+unix:" + urlstr)
		}

		// it is a file, so reattempt to open it with sqlite3
		return h.Open("sqlite3:" + urlstr)

	case err != nil:
		return err
	}

	// check driver
	if _, ok := drivers[h.u.Driver]; !ok {
		return &Error{
			Driver: h.u.Driver,
			Err:    ErrDriverNotAvailable,
		}
	}

	// add connection parameters for databases
	dsn := h.u.DSN
	dsn = addQueryParam(h.u.Driver, "mysql", dsn, "parseTime", "true")
	dsn = addQueryParam(h.u.Driver, "mysql", dsn, "loc", "Local")
	dsn = addQueryParam(h.u.Driver, "mysql", dsn, "sql_mode", "ansi")
	dsn = addQueryParam(h.u.Driver, "sqlite3", dsn, "loc", "auto")

	// connect
	h.db, err = sql.Open(h.u.Driver, dsn)
	if err != nil {
		return err
	}

	isPG := h.u.Driver == "postgres"
	h.allowdollar, h.allowmc = isPG, isPG

	return nil
}

// Execute executes a sql query against the connected database.
func (h *Handler) Execute(w io.Writer, sqlstr string, auto, forceExec bool) error {
	if h.db == nil {
		return ErrNotConnected
	}

	if h.u.Driver == "ora" {
		sqlstr = strings.TrimSuffix(sqlstr, ";")
	}

	// select
	if s := strings.TrimLeftFunc(sqlstr, unicode.IsSpace); len(s) >= 5 {
		i := strings.IndexFunc(s, unicode.IsSpace)
		if i == -1 {
			i = len(s)
		}

		z := strings.ToUpper(s[:i])
		if z == "SELECT" ||
			(h.u.Driver == "sqlite3" && z == "PRAGMA" && !strings.ContainsRune(s[i:], '=')) {
			err := h.Query(w, sqlstr)
			if err != nil {
				return h.WrapError(err)
			}

			return nil
		}
	}

	// exec
	res, err := h.db.Exec(sqlstr)
	if err != nil {
		return err
	}

	// get count
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	// print name
	name := "EXEC"
	if i := strings.Index(sqlstr, " "); i >= 0 {
		name = strings.ToUpper(sqlstr[:i])
	}
	fmt.Fprint(w, name)

	// print count
	if count > 0 {
		fmt.Fprintf(w, " %d", count)
	}

	fmt.Fprint(w, "\n")

	return nil
}

var allcapsRE = regexp.MustCompile(`^[A-Z_]+$`)

// Query executes a query against the database.
func (h *Handler) Query(w io.Writer, sqlstr string) error {
	// run query
	q, err := h.db.Query(sqlstr)
	if err != nil {
		return err
	}
	defer q.Close()

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
						row[n] = sqlite3Parse(x)
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
	fmt.Fprintf(w, "(%d rows)\n\n", rows)

	return nil
}

// DisplayHelp displays the help message.
func (h *Handler) DisplayHelp(w io.Writer) {
	io.WriteString(w, helpDesc)
}

// Close closes the database connection if it is open.
func (h *Handler) Close() error {
	if h.db != nil {
		err := h.db.Close()

		h.allowdollar, h.allowmc = false, false
		h.db, h.u = nil, nil

		return err
	}

	return nil
}

// Reset resets the line parser state.
func (h *Handler) Reset() {
	h.buf.Reset()

	// quote state
	h.q = false
	h.qdbl = false
	h.qdollar = false
	h.qid = ""

	// multicomment state
	h.mc = false

	// balance state
	h.b = 0
}

var lineend = []rune{'\n'}

// Process reads line commands from stdin, writing output to stdout and stderr.
func (h *Handler) Process(stdin io.Reader, stdout, stderr io.Writer) error {
	var err error

	// create readline instance
	l, err := readline.NewEx(&readline.Config{
		HistoryFile:            h.HistoryFile(),
		DisableAutoSaveHistory: true,
		InterruptPrompt:        "^C",
		HistorySearchFold:      true,
		Stdin:                  stdin,
		Stdout:                 stdout,
		Stderr:                 stderr,
		FuncIsTerminal: func() bool {
			return h.interactive
		},
		FuncFilterInputRune: func(r rune) (rune, bool) {
			if r == readline.CharCtrlZ {
				return r, false
			}
			return r, true
		},
	})
	if err != nil {
		return err
	}
	defer l.Close()

	// display welcome info
	if h.interactive {
		fmt.Fprint(l.Stdout(), welcomeDesc)
	}

	var r []rune
	var rlen, i int
	var stmt string
	for {
		if rlen == 0 {
			// reset prompt and grab input
			h.SetPrompt(l)
			r, err = l.Operation.Runes()
			switch {
			case err == readline.ErrInterrupt:
				h.Reset()
				continue
			case err != nil:
				return err
			}

			rlen, i = len(r), 0

			// special intercept for "help"
			if h.interactive && h.buf.Len == 0 && rlen >= 4 && startsWithHelp(r, 0, rlen) {
				h.DisplayHelp(l.Stdout())
				r, rlen = r[:rlen], 0
				continue
			}

			// save history
			if h.interactive {
				l.SaveHistory(string(r))
			}
		}

		var execute bool
		var cmd string
		var params []string

		// process
	parse:
		for ; i < rlen; i++ {
			// grab c, next
			c, next := r[i], grab(r, i+1, rlen)
			switch {
			// find end of string quote
			case h.q:
				pos, ok := readString(r, i, rlen, h)
				i = pos
				if ok {
					h.q, h.qdbl, h.qdollar, h.qid = false, false, false, ""
				}

			// find end of multiline comment
			case h.mc:
				pos, ok := readMultilineComment(r, i, rlen)
				i, h.mc = pos, !ok

			// start of single quoted string
			case c == '\'':
				h.q = true

			// start of double quoted string
			case c == '"':
				h.q, h.qdbl = true, true

			// start of dollar quoted string literal (postgres)
			case h.allowdollar && c == '$':
				id, pos, ok := readDollarAndTag(r, i, rlen)
				if ok {
					h.q, h.qdollar, h.qid = true, true, id
				}
				i = pos

			// start of sql comment, skip to end of line
			case c == '-' && next == '-':
				i = rlen

			// start of multiline comment (postgres)
			case h.allowmc && c == '/' && next == '*':
				h.mc = true
				i++

			// unbalance
			case c == '(':
				h.b++

			// balance
			case c == ')':
				h.b = max(0, h.b-1)

			// continue processing
			case h.q || h.mc || h.b != 0:
				continue

			// start of command
			case c == '\\':
				// extract command from r
				var pos int
				cmd, params, pos = readCommand(r, i, rlen)
				r = append(r[:i], r[pos:]...)
				rlen = len(r)

				break parse

			// execute
			case c == ';':
				// set execute and skip trailing whitespace
				execute = true
				i, _ = findNonSpace(r, i+1, rlen)

				break parse
			}
		}

		// fix i
		i = min(i, rlen)

		// determine appending to buf
		empty := isEmptyLine(r, 0, i)
		appendLine := h.q || !empty
		if cmd != "" && empty {
			appendLine = false
		}
		if appendLine {
			// skip leading space when empty
			st := 0
			if h.buf.Len == 0 {
				st, _ = findNonSpace(r, 0, i)
			}

			//log.Printf(">> appending: `%s`", string(r[st:i]))
			h.buf.Append(r[st:i], lineend)
		}

		// reset r
		r = r[i:]
		rlen = len(r)
		i = 0

		// process command
		if cmd != "" {
			var quit bool
			switch cmd {
			case "q":
				quit = true

			case "c", "connect":
				if len(params) == 0 {
					cmdErr(l, cmd, missingRequiredArg)
				} else {
					err := h.Open(params[0])
					if err != nil {
						fmt.Fprintf(l.Stderr(), "error: %v\n", err)
					}
					params = params[1:]
				}

			case "Z":
				err := h.Close()
				if err != nil {
					fmt.Fprint(l.Stderr(), "error: %v\n", err)
				}

			case "copyright":
				fmt.Fprintf(l.Stdout(), copyright)

			case "errverbose":

			case "g", "gexec", "gset":
				execute = true
			/*case "crosstabview": // not likely to ever implement this*/

			case "?", "h":
			case "e", "edit", "ef", "ev":

			case "p", "print":
				// build
				s := stmt
				if h.buf.Len != 0 {
					s = h.buf.String()
				}
				if s == "" {
					s = queryBufferEmpty
				}

				// print
				fmt.Fprintf(l.Stdout(), "%s\n", s)

			case "r", "reset":
				h.Reset()
				fmt.Fprintf(l.Stdout(), queryBufferReset)

			case "w", "write":

			case "o", "out":

			case "i", "include", "ir", "include_relative":
				fmt.Fprintf(l.Stdout(), "include %v\n", params)

			// invalid command
			default:
				fmt.Fprintf(l.Stderr(), invalidCommand, cmd)
				params = nil
			}

			// print unused command parameters
			for _, p := range params {
				fmt.Fprintf(l.Stdout(), extraArgumentIgnored, cmd, p)
			}

			if quit {
				return nil
			}

			// clear
			cmd, params = "", nil
		}

		if execute {
			// clear
			if h.buf.Len != 0 {
				stmt = h.buf.String()
				h.buf.Reset()
			}

			//log.Printf("executing: `%s`", stmt)
			if stmt != "" && stmt != ";" {
				err = h.Execute(l.Stdout(), stmt, false, false)
				if err != nil {
					fmt.Fprintf(l.Stderr(), "error: %v\n", err)
				}
			}

			// clear
			execute = false
		}
	}
}

// WrapError conditionally wraps an error if the error occurs while connected
// to a database.
func (h *Handler) WrapError(err error) error {
	if h.db != nil {
		// attempt to clean up and standardize errors
		driver := h.u.Driver
		if s, ok := drivers[driver]; ok {
			driver = s
		}

		return &Error{driver, err}
	}

	return err
}

// RunCommands processes command line arguments.
func (h *Handler) RunCommands(cmds []string) error {
	h.interactive = false

	var err error
	for _, c := range cmds {
		err = h.Process(strings.NewReader(c), os.Stdout, os.Stderr)
		if err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}

// RunReadline processes input.
func (h *Handler) RunReadline(in, out string) error {
	var err error

	// configure input
	var stdin *os.File
	stdout, stderr := readline.Stdout, readline.Stderr

	// set file as stdin
	if in != "" {
		stdin, err = os.OpenFile(in, os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer stdin.Close()

		h.interactive = false
	}

	// set out as stdout
	if out != "" {
		stdout, err = os.OpenFile(out, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer stdout.Close()

		h.interactive = false
	}

	// set stdin if not set
	var r io.ReadCloser = stdin
	if stdin == nil {
		c := readline.NewCancelableStdin(readline.Stdin)
		defer c.Close()
		r = c
	}

	return h.Process(r, stdout, stderr)
}
