package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/knq/dburl"
	"github.com/knq/xoutil"
	"github.com/olekukonko/tablewriter"
)

const (
	notConnected = "(not connected)"
)

// Handler is a cli input handler.
type Handler struct {
	args        *Args
	interactive bool

	u  *dburl.URL
	db *sql.DB
}

// Prompt returns the base input prompt.
func (h *Handler) Prompt() string {
	if !h.interactive {
		return ""
	}

	s := notConnected

	if h.db != nil {
		s = h.u.Short()
	}

	return s + "=> "
}

// Cont returns the continuation prompt.
func (h *Handler) Cont() string {
	if !h.interactive {
		return ""
	}

	s := notConnected

	if h.db != nil {
		s = h.u.Short()
	}

	return s + "-> "
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
		if fi, err := os.Stat(urlstr); err == nil {
			// TODO: add support for postgres unix domain sockets
			if fi.Mode()&os.ModeSocket != 0 {
				return h.Open("mysql+unix:" + urlstr)
			}

			// it is a file, so reattempt to open it with sqlite3
			return h.Open("sqlite3:" + urlstr)
		}

		return err

	case err != nil:
		return err
	}

	// check driver
	if !drivers[h.u.Driver] {
		if h.u.Driver == "ora" {
			return ErrOracleDriverNotAvailable
		}
		return fmt.Errorf("driver '%s' is not available for '%s'", h.u.Driver, h.u.Scheme)
	}

	// add connection parameters for databases
	dsn := h.u.DSN
	dsn = h.addQueryParam(dsn, "mysql", "parseTime", "true")
	dsn = h.addQueryParam(dsn, "mysql", "loc", "Local")
	dsn = h.addQueryParam(dsn, "sqlite3", "loc", "auto")

	//log.Printf(">>> dsn: %s", dsn)

	// connect
	h.db, err = sql.Open(h.u.Driver, dsn)
	return err
}

// addQueryParam conditionally adds a ?name=val style query parameter to the
// end of a DSN if the connected database driver matches the supplied driver
// name.
func (h *Handler) addQueryParam(dsn, driver, name, val string) string {
	if h.u.Driver == driver {
		if !strings.Contains(dsn, name+"=") {
			s := "?"
			if strings.Contains(dsn, "?") {
				s = "&"
			}
			return dsn + s + name + "=" + val
		}
	}

	return dsn
}

// Execute executes a sql query against the connected database.
func (h *Handler) Execute(w io.Writer, sqlstr string) error {
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
				return h.wrapError(err)
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
						row[n] = h.sqlite3Parse(x)
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

// sqlite3Parse will convert buf matching a time format to a time, and will
// format it according to the handler time settings.
//
// TODO: only do this if the type of the column is a timestamp type.
func (h *Handler) sqlite3Parse(buf []byte) string {
	s := string(buf)
	if s != "" && strings.TrimSpace(s) != "" {
		t := &xoutil.SqTime{}
		err := t.Scan(buf)
		if err == nil {
			return t.Format(time.RFC3339Nano)
		}
	}

	return s
}

// HistoryFile returns the name of the history file.
func (h *Handler) HistoryFile() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	return filepath.Join(u.HomeDir, ".usql_history")
}

// DisplayHelp displays the help message.
func (h *Handler) DisplayHelp(w io.Writer) {
	io.WriteString(w, helpDesc)
}

// Close closes the database connection if it is open.
func (h *Handler) Close() error {
	if h.db != nil {
		err := h.db.Close()

		h.db = nil
		h.u = nil

		return err
	}

	return nil
}

// Run processes h.args, running either the Commands if non-empty, or File if
// specified, or the input from stdin.
func (h *Handler) Run() error {
	var err error

	// open
	err = h.Open(h.args.DSN)
	if err != nil {
		return err
	}

	// short circuit if commands provided
	if len(h.args.Commands) > 0 {
		return h.RunCommands()
	}

	// configure input
	var stdin *os.File
	stdout, stderr := os.Stdout, os.Stderr

	// set file as stdin
	if h.args.File != "" {
		stdin, err = os.OpenFile(h.args.File, os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer stdin.Close()

		h.interactive = false
	}

	// set out as stdout
	if h.args.Out != "" {
		stdout, err = os.OpenFile(h.args.Out, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer stdout.Close()

		h.interactive = false
	}

	// set stdin if not set
	var r io.Reader = stdin
	if stdin == nil {
		r = readline.NewCancelableStdin(os.Stdin)
	}

	return h.Process(r, stdout, stderr)
}

// Run handles stuff
func (h *Handler) Process(stdin io.Reader, stdout, stderr io.Writer) error {
	var err error

	// create readline instance
	l, err := readline.NewEx(&readline.Config{
		Prompt:                 h.Prompt(),
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

	if h.interactive {
		fmt.Fprint(l.Stdout(), cliDesc)
	}

	// process input
	var multi bool
	var stmt []string
	for {
		line, err := l.Readline()
		switch {
		case err == readline.ErrInterrupt:
			stmt = stmt[:0]
			multi = false
			l.SetPrompt(h.Prompt())
			continue

		case err != nil:
			return nil
		}

		z := strings.TrimSpace(line)
		if len(z) == 0 {
			continue
		}

		if !multi {
			switch {
			case z == "help":
				h.DisplayHelp(l.Stdout())
				continue

			case z == `\q`:
				h.SaveHistory(l, line)
				return nil

			case strings.HasPrefix(line, `\c `) || strings.HasPrefix(line, `\connect `):
				h.SaveHistory(l, line)

				err = h.Close()
				if err != nil {
					return err
				}

				urlstr := strings.TrimSpace(line[strings.IndexRune(line, ' '):])
				err = h.Open(urlstr)
				if err != nil {
					fmt.Fprintf(l.Stderr(), "error: could not connect to database: %v\n", err)
				}

				l.SetPrompt(h.Prompt())
				continue

			case z == `\Z`:
				h.SaveHistory(l, line)

				err = h.Close()
				if err != nil {
					return err
				}

				l.SetPrompt(h.Prompt())
				continue
			}
		}

		stmt = append(stmt, line)

		if !strings.HasSuffix(z, ";") {
			multi = true
			l.SetPrompt(h.Cont())
			continue
		}

		s := strings.Join(stmt, "\n")
		h.SaveHistory(l, s)
		l.SetPrompt(h.Prompt())

		err = h.Execute(l.Stdout(), s)
		if err != nil {
			fmt.Fprintf(l.Stderr(), "error: %v\n", h.wrapError(err))
		}

		stmt = stmt[:0]
		multi = false
	}
}

// SaveHistory conditionally saves a line to the history if the session is
// interactive.
func (h *Handler) SaveHistory(l *readline.Instance, line string) error {
	if h.interactive {
		return l.SaveHistory(line)
	}

	return nil
}

// RunCommands runs the argument commands.
func (h *Handler) RunCommands() error {
	h.interactive = false

	var err error
	for _, c := range h.args.Commands {
		err = h.Process(strings.NewReader(c), os.Stdout, os.Stderr)
		if err != nil {
			return err
		}
	}

	return nil
}

// driverError is a wrapper to standardize errors.
type driverError struct {
	driver string
	err    error
}

// Error satisfies the error interface.
func (e *driverError) Error() string {
	if e.driver != "" {
		s := e.err.Error()
		return e.driver + ": " + strings.TrimLeftFunc(strings.TrimPrefix(strings.TrimSpace(s), e.driver+":"), unicode.IsSpace)
	}

	return e.err.Error()
}

// wrapError conditionally wraps an error if the error occurs while connected
// to a database.
func (h *Handler) wrapError(err error) error {
	if h.db != nil {
		// attempt to clean up and standardize errors
		driver := h.u.Driver
		if driver == "postgres" {
			driver = "pq"
		}

		return &driverError{driver, err}
	}

	return err
}
