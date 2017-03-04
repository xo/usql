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

	var err error

	// parse dsn
	h.u, err = dburl.Parse(urlstr)
	switch {
	case err == dburl.ErrInvalidDatabaseScheme:
		if _, err = os.Stat(urlstr); err == nil {
			// it is a file, so reattempt to open it with sqlite3
			return h.Open("sqlite3:" + urlstr)
		}

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
			return h.Query(w, sqlstr)
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

// Run handles stuff
func (h *Handler) Run() error {
	var err error

	// open
	err = h.Open(h.args.DSN)
	if err != nil {
		return err
	}

	//log.Printf(">>> commands: %v", h.args.Commands)

	// short circuit if commands provided
	if len(h.args.Commands) > 0 {
		return h.RunCommands()
	}

	// create readline instance
	l, err := readline.NewEx(&readline.Config{
		Prompt:                 h.Prompt(),
		HistoryFile:            h.HistoryFile(),
		DisableAutoSaveHistory: true,
		InterruptPrompt:        "^C",
		HistorySearchFold:      true,
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
		if !multi {
			switch {
			case z == "help":
				h.DisplayHelp(l.Stdout())
				continue

			case z == `\q`:
				l.SaveHistory(line)
				return nil

			case strings.HasPrefix(line, `\c `):
				l.SaveHistory(line)

				err = h.Close()
				if err != nil {
					return err
				}

				urlstr := strings.TrimSpace(line[2:])
				err = h.Open(urlstr)
				if err != nil {
					fmt.Fprintf(l.Stderr(), "error: could not connect to `%s`: %v\n", urlstr, err)
				}

				l.SetPrompt(h.Prompt())
				continue

			case z == `\Z`:
				l.SaveHistory(line)

				err = h.Close()
				if err != nil {
					return err
				}

				l.SetPrompt(h.Prompt())
				continue
			}
		}

		stmt = append(stmt, line)

		if len(z) == 0 {
			continue
		}
		if !strings.HasSuffix(z, ";") {
			multi = true
			l.SetPrompt(h.Cont())
			continue
		}

		s := strings.Join(stmt, "\n")
		l.SaveHistory(s)
		l.SetPrompt(h.Prompt())

		err = h.Execute(l.Stdout(), s)
		if err != nil {
			fmt.Fprintf(l.Stderr(), "error: %v\n", err)
		}

		stmt = stmt[:0]
		multi = false
	}
}

// RunCommands runs the argument commands.
func (h *Handler) RunCommands() error {
	var err error
	for _, c := range h.args.Commands {
		err = h.Execute(os.Stdout, c)
		if err != nil {
			return err
		}
	}

	return nil
}
