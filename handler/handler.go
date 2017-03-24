package handler

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/knq/dburl"
	"github.com/olekukonko/tablewriter"

	"github.com/knq/usql/drivers"
	"github.com/knq/usql/stmt"
	"github.com/knq/usql/text"
)

// Handler is a input process handler.
type Handler struct {
	histfile    string
	homedir     string
	wd          string
	interactive bool
	cygwin      bool

	u  *dburl.URL
	db *sql.DB
}

// New creates a new input handler.
func New(histfile, homedir, wd string, interactive, cygwin bool) (*Handler, error) {
	return &Handler{
		histfile:    histfile,
		homedir:     homedir,
		wd:          wd,
		interactive: interactive,
		cygwin:      cygwin,
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
func (h *Handler) SetPrompt(l *readline.Instance, state string) {
	if !h.interactive {
		return
	}

	s := text.NotConnected

	if h.db != nil {
		s = h.u.Short()
		if s == "" {
			s = "(" + h.u.Driver + ")"
		}
	}

	l.SetPrompt(s + state + "> ")
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

	// connect
	h.db, err = f(h.u.Driver, h.u.DSN)
	if err != nil {
		return err
	}

	// ping
	err = h.WrapError(h.db.Ping())
	if err != nil {
		h.Close()
	}
	return err
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

var beginTransactionRE = regexp.MustCompile(`(?i)^BEGIN\s*TRANSACTION;?$`)

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

	//var tx *sql.Tx

	switch h.u.Driver {
	case "ora":
		sqlstr = strings.TrimSuffix(sqlstr, ";")

	case "ql":
		if typ == "BEGIN" && beginTransactionRE.MatchString(sqlstr) {
			log.Printf("GOT BEGIN TRANSACTION")
			//tx, err := h.db.Begin()
			/*if err != nil {
				return err
			}*/
		}
		if typ == "COMMIT" {

		}
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
	for q.NextResultSet() {
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
		fmt.Fprintln(l.Stdout(), text.WelcomeDesc)
		fmt.Fprintln(l.Stdout())
	}

	// set help intercept
	f := l.Operation.Runes
	if h.interactive {
		f = func() ([]rune, error) {
			// next line
			r, err := l.Operation.Runes()
			if err != nil {
				return nil, err
			}

			// check if line starts with help
			rlen := len(r)
			if rlen >= 4 && stmt.StartsWith(r, 0, rlen, text.HelpPrefix) {
				h.DisplayHelp(l.Stdout())
				return nil, nil
			}

			// save history
			l.SaveHistory(string(r))

			return r, nil
		}
	}

	// create stmt
	var opts []stmt.Option
	if h.db != nil && h.u.Driver == "postgres" {
		opts = append(opts,
			stmt.AllowDollar(true),
			stmt.AllowMultilineComments(true),
		)
	}

	// statement buf
	var lastPrefix, last string
	buf := stmt.New(f, opts...)
	for {
		var execute bool
		var exitWithErr error

		// set prompt
		h.SetPrompt(l, buf.State())

		// get next
		cmd, params, err := buf.Next()
		switch {
		case !h.interactive && err == io.EOF:
			execute, exitWithErr = true, io.EOF

		case err == readline.ErrInterrupt:
			buf.Reset()
			continue

		case err != nil:
			return err
		}

		// grab ready state
		execute = execute || buf.Ready()

		// process command
		if cmd != "" {
			switch cmd {
			case "q", "quit":
				return nil

			case "c", "connect":
				writeErr(l, h.Close())

				if len(params) < 1 {
					cmdErr(l, cmd, text.MissingRequiredArg)
				} else {
					writeErr(l, h.Open(params[0]))
					params = params[1:]
				}

			case "cdsn", "connect_dsn":
				writeErr(l, h.Close())

				if len(params) < 2 {
					cmdErr(l, cmd, text.MissingRequiredArg)
				} else {
					writeErr(l, h.Open(params...))
					params = nil
				}

			case "Z", "disconnect":
				writeErr(l, h.Close())

			case "copyright":
				fmt.Fprintln(l.Stdout(), text.Copyright)

			case "conninfo":
				if h.u != nil {
					fmt.Fprintf(l.Stdout(), text.ConnInfo, h.u.Driver, h.u.DSN)
					fmt.Fprintln(l.Stdout())
				}

			case "dsn":
				if h.u != nil {
					fmt.Fprintln(l.Stdout(), h.u.Driver, "("+h.u.DSN+")")
				}

			case "drivers":
				names := make([]string, len(drivers.Drivers))
				var z int
				for k := range drivers.Drivers {
					names[z] = k
					z++
				}
				sort.Strings(names)

				fmt.Fprintln(l.Stdout(), text.AvailableDrivers)
				for _, n := range names {
					s := "  " + n

					driver, aliases := dburl.SchemeDriverAndAliases(n)
					if driver != n {
						s += " (" + driver + ")"
					}
					if len(aliases) > 0 {
						if len(aliases) > 0 {
							s += " [" + strings.Join(aliases, ", ") + "]"
						}
					}
					fmt.Fprintln(l.Stdout(), s)
				}

			case "errverbose":
				notImpl(l, cmd)

			case "g":
				execute = true

			case "gexec", "gset":
				notImpl(l, cmd)

			case "?", "h":
				notImpl(l, cmd)

			case "e", "edit":
				var path, line string
				params, path = pop(params, "")
				params, line = pop(params, "")

				s := last
				if buf.Len != 0 {
					s = buf.String()
				}

				n, err := h.LaunchEditor(path, line, s)
				if err != nil {
					writeErr(l, err)
					break
				}

				buf.Reset()
				buf.Feed(n)

			case "ef", "ev":
				notImpl(l, cmd)

			case "p", "print":
				// build
				s := last
				if buf.Len != 0 {
					s = buf.String()
				}
				if s == "" {
					s = text.QueryBufferEmpty
				}

				// print
				fmt.Fprintln(l.Stdout(), s)

			case "r", "reset":
				buf.Reset()
				fmt.Fprintln(l.Stdout(), text.QueryBufferReset)

			case "echo":
				// this could be done to echo the actual input (by using pos
				// above), but the implementation here remains faithful to the
				// psql implementation
				fmt.Fprintln(l.Stdout(), strings.Join(params, " "))
				params = nil

			case "w", "write":
				if len(params) == 0 {
					cmdErr(l, cmd, text.MissingRequiredArg)
				} else {
					s := last
					if buf.Len != 0 {
						s = buf.String()
					}
					writeErr(l, ioutil.WriteFile(params[0], []byte(strings.TrimSuffix(s, "\n")+"\n"), 0644))
					params = params[1:]
				}

			case "o", "out":
				notImpl(l, cmd)

			case "i", "include", "ir", "include_relative":
				if len(params) == 0 {
					cmdErr(l, cmd, text.MissingRequiredArg)
				} else {
					var fname string
					params, fname = pop(params, "")
					relative := cmd == "ir" || cmd == "include_relative"
					writeErr(l, h.IncludeFile(fname, relative), fname+": ")
				}

			case "!":
				if len(params) == 0 {
					cmdErr(l, cmd, text.MissingRequiredArg)
				} else {

				}

			case "cd":
				var path string
				params, path = pop(params, h.homedir)
				if strings.HasPrefix(path, "~/") {
					path = filepath.Join(h.homedir, strings.TrimPrefix(path, "~/"))
				}
				writeErr(l, os.Chdir(path))

			case "setenv":
				if len(params) == 0 {
					cmdErr(l, cmd, text.MissingRequiredArg)
				} else {
					var key, val string
					params, key = pop(params, "")
					params, val = pop(params, "")
					writeErr(l, os.Setenv(key, val))
				}

			// invalid command
			default:
				fmt.Fprintf(l.Stderr(), text.InvalidCommand, cmd)
				fmt.Fprintln(l.Stderr())
				params = nil
			}

			// print unused command parameters
			for _, p := range params {
				fmt.Fprintf(l.Stdout(), text.ExtraArgumentIgnored, cmd, p)
				fmt.Fprintln(l.Stdout())
			}
		}

		if execute {
			if buf.Len != 0 {
				lastPrefix, last = buf.Prefix, buf.String()
				buf.Reset()
			}

			//log.Printf(">> PROCESS EXECUTE: `%s`", last)
			if last != "" && last != ";" {
				writeErr(l, h.Execute(l.Stdout(), lastPrefix, last))
			}

			execute = false
		}

		if exitWithErr != nil {
			return exitWithErr
		}
	}
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

// IncludeFile includes the specified path.
func (h *Handler) IncludeFile(path string, relative bool) error {
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

	log.Printf(">>>> path: %s", path)

	return nil
}

// LaunchEditor launches an editor using the current query buffer.
func (h *Handler) LaunchEditor(path, line, s string) ([]rune, error) {
	var err error

	ed := getenv("USQL_EDITOR", "EDITOR", "VISUAL")
	if ed == "" {
		return nil, ErrNoEditorDefined
	}

	if path == "" {
		f, err := ioutil.TempFile("", "usql")
		if err != nil {
			return nil, err
		}

		err = f.Close()
		if err != nil {
			return nil, err
		}

		path = f.Name()
		err = ioutil.WriteFile(path, []byte(strings.TrimSuffix(s, "\n")+"\n"), 0644)
		if err != nil {
			return nil, err
		}
	}

	// setup args
	args := []string{path}
	if line != "" {
		args = append(args, "+"+line)
	}

	// create command
	c := exec.Command(ed, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// run
	err = c.Run()
	if err != nil {
		return nil, err
	}

	// read
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return []rune(strings.TrimSuffix(string(buf), "\n")), nil
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

	// configure stdin
	var stdin io.ReadCloser
	if in != "" {
		stdin, err = os.OpenFile(in, os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer stdin.Close()

		h.interactive = false
	} else if h.cygwin {
		stdin = os.Stdin
	} else if h.interactive {
		stdin = readline.Stdin
	}

	// configure stdout
	var stdout io.WriteCloser
	if out != "" {
		stdout, err = os.OpenFile(out, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer stdout.Close()

		h.interactive = false
	} else if h.cygwin {
		stdout = os.Stdout
	} else if h.interactive {
		stdin = readline.Stdin
	}

	// configure stderr
	var stderr io.Writer = os.Stderr
	if !h.cygwin {
		stderr = readline.Stderr
	}

	// wrap it with cancelable stdin
	if h.interactive {
		stdin = readline.NewCancelableStdin(stdin)
	}

	return h.Process(stdin, stdout, stderr)
}

// DisplayHelp displays the help message.
func (h *Handler) DisplayHelp(w io.Writer) {
	io.WriteString(w, text.HelpDesc)
}
