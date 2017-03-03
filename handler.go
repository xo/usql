package main

import (
	"database/sql"
	"fmt"
	"io"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/knq/dburl"
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
	s := notConnected

	if h.db != nil {
		s = h.u.Short()
	}

	return s + "=> "
}

// Cont returns the continuation prompt.
func (h *Handler) Cont() string {
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
	if err != nil {
		return err
	}

	// check driver
	if !drivers[h.u.Driver] {
		if h.u.Driver == "ora" {
			return ErrOracleDriverNotAvailable
		}
		return fmt.Errorf("driver '%s' is not available for '%s'", h.u.Driver, h.u.Scheme)
	}

	// connect
	h.db, err = sql.Open(h.u.Driver, h.u.DSN)
	return err
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

				switch x := (*j).(type) {
				case []byte:
					row[n] = string(x)

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

	fmt.Fprint(l.Stdout(), cliDesc)

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
