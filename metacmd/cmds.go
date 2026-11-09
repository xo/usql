package metacmd

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xo/dbfwd"
	"github.com/xo/dburl"
	"github.com/xo/rline"
	"github.com/xo/transit/styles"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/dump"
	"github.com/xo/usql/env"
	"github.com/xo/usql/metacmd/describe"
	"github.com/xo/usql/text"
)

// Quit is a General meta command (\q \quit). Quits the application.
//
// Descs:
//
//	q	quit {{CommandName}}
//	quit
func Quit(p *Params) error {
	p.Option.Quit = true
	return nil
}

// Copyright is a General meta command (\copyright). Writes the
// application's copyright message to the output.
//
// Descs:
//
//	copyright	show usage and distribution terms for {{CommandName}}
func Copyright(p *Params) error {
	// the logo goes straight to the terminal, because a pager shows its
	// escape codes as text
	if typ := env.TermGraphics(); typ.Available() {
		typ.Encode(p.Handler.IO(), text.Logo)
	}
	fmt.Fprintln(p.Handler.Stdout(), text.Copyright)
	return nil
}

// ErrVerbose is a General meta command (\errverbose). Writes the error of the
// last statement that failed, with every field that the driver gives, as the
// client of the database writes them, as psql's \errverbose does.
//
// Descs:
//
//	errverbose	show most recent error message at maximum verbosity
func ErrVerbose(p *Params) error {
	err, query := p.Handler.LastError()
	if err == nil {
		fmt.Fprintln(p.Handler.Stdout(), text.NoPreviousError)
		return nil
	}
	fmt.Fprintln(p.Handler.IO().Stderr(), drivers.ErrVerbose(err, query))
	return nil
}

// Drivers is a General meta command (\drivers). Writes information about the
// database drivers the application was built with to the output.
//
// Descs:
//
//	drivers	show database drivers available to {{CommandName}}
func Drivers(p *Params) error {
	stdout := p.Handler.Stdout()
	available := drivers.Available()
	names := make([]string, len(available))
	var z int
	for k := range available {
		names[z] = k
		z++
	}
	sort.Strings(names)
	fmt.Fprintln(stdout, text.AvailableDrivers)
	for _, n := range names {
		s := "  " + n
		driver, aliases := dburl.SchemeNameAndAliases(n)
		if driver != n {
			s += " (" + driver + ")"
		}
		if len(aliases) > 0 {
			s += " [" + strings.Join(aliases, ", ") + "]"
		}
		fmt.Fprintln(stdout, s)
	}
	return nil
}

// Help is a Help meta command (\?). Writes a help message to the output.
//
// Descs:
//
//	?	show help on {{CommandName}}'s meta (backslash) commands
//	?	commands	as \?, but shows all command aliases
//	?	options	show help on {{CommandName}} command-line options
//	?	variables	show help on special {{CommandName}} variables
//	?	styles	show {{CommandName}} syntax highlighting styles
func Help(p *Params) error {
	name, err := p.Next(false)
	if err != nil {
		return err
	}
	stdout := p.Handler.Stdout()
	switch name = strings.TrimSpace(strings.ToLower(name)); {
	case name == "options":
		text.Usage(stdout, true)
	case name == "variables":
		_ = env.Listing(stdout)
	case name == "styles":
		_ = Styles(stdout, styles.Names(), env.Setting("syntax_hl_style"))
	default:
		_ = Dump(stdout, name == "commands")
	}
	return nil
}

// Execute is a Query Execute meta command (\g and variants). Executes the
// active query on the open database connection.
//
// Descs:
//
//	g	[(OPTIONS)] [FILE] or ;	execute query (and send results to file or |pipe)
//	go:g
//	G	[(OPTIONS)] [FILE]	as \g, but forces vertical output mode
//	ego:G
//	gx	[(OPTIONS)] [FILE]	as \g, but forces expanded output mode
//	gexec	execute query and execute each value of the result
//	gset	[PREFIX]	execute query and store results in {{CommandName}} variables
//	gdesc	describe result of query, without executing it
func Execute(p *Params) error {
	p.Option.Exec = ExecOnly
	switch p.Name {
	case "g", "go", "G", "ego", "gx", "gset":
		p.Option.Params = make(map[string]string)
		for _, o := range p.Options() {
			if o.Value == "" {
				return text.ErrInvalidFormatOption
			}
			v := o.Value
			if c := v[0]; c == '\'' || c == '"' {
				var err error
				if v, err = env.Unquote(v); err != nil {
					return err
				}
			}
			p.Option.Params[o.Key] = v
		}
		// the rest is the file or the pipe, or the prefix of \gset
		rest, err := p.All(true)
		if err != nil {
			return err
		}
		key := "pipe"
		if p.Name == "gset" {
			key = "prefix"
		}
		if len(rest) != 0 {
			p.Option.Params[key] = strings.Join(rest, " ")
		}
		switch p.Name {
		case "G", "ego":
			p.Option.Params["format"] = "vertical"
		case "gx":
			p.Option.Params["expanded"] = "on"
		case "gset":
			p.Option.Exec = ExecSet
		}
	case "gexec":
		p.Option.Exec = ExecExec
	case "gdesc":
		p.Option.Exec = ExecDescribe
	}
	return nil
}

// Bind is a Query Execute meta command (\bind). Sets (or unsets) variables to
// be used when executing a query.
//
// Descs:
//
//	bind	[PARAM]...	set query parameters
func Bind(p *Params) error {
	bind, err := p.All(true)
	if err != nil {
		return err
	}
	var v []interface{}
	if n := len(bind); n != 0 {
		v = make([]interface{}, len(bind))
		for i := range n {
			v[i] = bind[i]
		}
	}
	p.Handler.Bind(v)
	return nil
}

// Timing is a Query Execute meta command (\timing). Sets (or toggles) writing
// timing information for executed queries to the output.
//
// Descs:
//
//	timing	[on|off]	toggle timing of commands
func Timing(p *Params) error {
	v, err := p.Next(true)
	switch {
	case err != nil:
		return err
	case v == "":
		p.Handler.SetTiming(!p.Handler.GetTiming())
	default:
		s, err := env.ParseBool(v, `\timing`)
		if err != nil {
			stderr := p.Handler.IO().Stderr()
			fmt.Fprintf(stderr, "error: %v", err)
			fmt.Fprintln(stderr)
		}
		var b bool
		if s == "on" {
			b = true
		}
		p.Handler.SetTiming(b)
	}
	setting := "off"
	if p.Handler.GetTiming() {
		setting = "on"
	}
	p.Handler.Print(text.TimingSet, setting)
	return nil
}

// Crosstab is a Query View meta command (\crosstab). Executes the active query
// on the open database connection and displays results in a crosstab view.
//
// Descs:
//
//	crosstab	[(OPTIONS)] [COLUMNS]	execute query and display results in crosstab
//	crosstabview
//	xtab
func Crosstab(p *Params) error {
	p.Option.Exec = ExecCrosstab
	for i := 0; i < 4; i++ {
		col, ok, err := p.NextOK(true)
		if err != nil {
			return err
		}
		p.Option.Crosstab = append(p.Option.Crosstab, col)
		if !ok {
			break
		}
	}
	return nil
}

// Chart is a Query View meta command (\chart). Executes the active query on
// the open database connection and displays results in a chart view.
//
// Descs:
//
//	chart	CHART [(OPTIONS)]	execute query and display results as a chart
func Chart(p *Params) error {
	p.Option.Exec = ExecChart
	if p.Option.Params == nil {
		p.Option.Params = make(map[string]string, 1)
	}
	params, err := p.All(true)
	if err != nil {
		return err
	}
	for i := 0; i < len(params); i++ {
		param := params[i]
		if param == "help" {
			p.Option.Params["help"] = ""
			return nil
		}
		equal := strings.IndexByte(param, '=')
		switch {
		case equal == -1 && i >= len(params)-1:
			return text.ErrWrongNumberOfArguments
		case equal == -1:
			i++
			p.Option.Params[param] = params[i]
		default:
			p.Option.Params[param[:equal]] = param[equal+1:]
		}
	}
	return nil
}

// Watch is a Query View meta command (\watch). Executes (and re-executes) the
// active query on the open database connection until canceled by the user.
//
// Descs:
//
//	watch	[(OPTIONS)] [INTERVAL]	execute query every specified interval
func Watch(p *Params) error {
	p.Option.Exec = ExecWatch
	p.Option.Watch = 2 * time.Second
	switch s, ok, err := p.NextOK(true); {
	case err != nil:
		return err
	case ok:
		d, err := time.ParseDuration(s)
		if err != nil {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				d = time.Duration(f * float64(time.Second))
			}
		}
		if d == 0 {
			return text.ErrInvalidWatchDuration
		}
		p.Option.Watch = d
	}
	return nil
}

// Connect is a Connection meta command (\c, \connect). Opens (connects) a
// database connection.
//
// Descs:
//
//	c	DSN or \c NAME	connect to dsn or named database connection
//	c	DRIVER PARAMS...	connect to database with driver and parameters
//	connect
func Connect(p *Params) error {
	vals, err := p.All(true)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	return p.Handler.Open(ctx, vals...)
}

// ConnectionInfo is a Connection meta command (\conninfo). Writes information
// about the connection to the output, as a table.
//
// Descs:
//
//	conninfo	[-password]	display information about the current database connection, with the password when -password is given
func ConnectionInfo(p *Params) error {
	opt, ok, err := p.NextOpt(true)
	switch {
	case err != nil:
		return err
	case ok && opt != "password", !ok && opt != "":
		return fmt.Errorf(text.InvalidOption, opt)
	}
	db, u := p.Handler.DB(), p.Handler.URL()
	if db == nil || u == nil {
		fmt.Fprintln(p.Handler.Stdout(), text.NotConnected)
		return nil
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var values [][]any
	add := func(name, value string) {
		if value != "" {
			values = append(values, []any{name, value})
		}
	}
	add("Driver", u.SchemeName)
	name := strings.TrimPrefix(u.Path, "/")
	if name == "" {
		name = u.Opaque
	}
	add("Database", name)
	var user string
	var password bool
	if u.User != nil {
		user = u.User.Username()
		_, password = u.User.Password()
	}
	if user == "" {
		user, _ = drivers.User(ctx, u, db)
	}
	add("Client User", user)
	add("Host", u.Hostname())
	add("Server Port", u.Port())
	add("Password Used", strconv.FormatBool(password))
	if info, err := drivers.ConnInfo(ctx, u, db); err == nil {
		for _, v := range info {
			add(v[0], v[1])
		}
	}
	ver, err := drivers.Version(ctx, u, db)
	if err != nil {
		ver = fmt.Sprintf("<unknown, error: %v>", err)
	}
	add("Server Version", ver)
	// a forward names what it reached, and its URL holds no password. The
	// URL of the local end holds the password that the forward read.
	fwdURL, fwd := p.Handler.Forward()
	if fwd != nil {
		add("Forward", fwd.String())
		add("URL", fwdURL.String())
	}
	local := "URL"
	if fwd != nil {
		local = "Local URL"
	}
	// the password shows only when it is asked for
	if ok {
		add(local, u.String())
		add("DSN", u.DSN)
	} else {
		add(local, redacted(u))
	}
	return describe.Table(p.Handler.Stdout(), "Connection Information", []string{"Parameter", "Value"}, values)
}

// Disconnect is a Connection meta command (\Z). Closes (disconnects) the
// current database connection.
//
// Descs:
//
//	Z	close (disconnect) database connection
//	disconnect
func Disconnect(p *Params) error {
	return p.Handler.Close()
}

// Password is a Connection meta command (\password). Changes the database
// user's password.
//
// Descs:
//
//	password	[USER]	change password for user
//	passwd
func Password(p *Params) error {
	username, err := p.Next(true)
	if err != nil {
		return err
	}
	user, err := p.Handler.ChangePassword(username)
	switch {
	case err == text.ErrPasswordNotSupportedByDriver || err == text.ErrNotConnected:
		return err
	case err != nil:
		return fmt.Errorf(text.PasswordChangeFailed, user, err)
	}
	// p.Handler.Print(text.PasswordChangeSucceeded, user)
	return nil
}

// Edit is a Query Buffer meta command (\e \edit). Opens the query buffer for
// editing in an external application.
//
// Descs:
//
//	e	[-raw] [FILE] [LINE]	edit the query buffer (or file) with external editor (-raw for the non-interpolated buffer)
//	edit
func Edit(p *Params) error {
	raw, err := p.RawOpt()
	if err != nil {
		return err
	}
	path, err := p.Next(true)
	if err != nil {
		return err
	}
	s := buffer(p, raw)
	line, err := p.Next(true)
	if err != nil {
		return err
	}
	// reset if no error
	var out []byte
	err = onTerminal(p, func() error {
		var err error
		out, err = env.EditFile(p.Handler.User(), path, line, []byte(s))
		return err
	})
	if err != nil {
		return err
	}
	// the edited query is saved to the history, and read as input again
	p.Handler.IO().AddHistory(string(out))
	p.Handler.Buffer().Reset()
	p.Handler.Push(string(out))
	return nil
}

// EditSource is a Query Buffer meta command (\ef \ev). Opens the definition
// of a function or a view in an external editor, and leaves the edited
// statement in the query buffer.
//
// Descs:
//
//	ef	[FUNCNAME [LINE]]	edit function definition with external editor
//	ev	[VIEWNAME [LINE]]	edit view definition with external editor
func EditSource(p *Params) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	params, err := p.Patterns()
	if err != nil {
		return err
	}
	// a last word of digits is the line, as in \ef f(integer) 3
	var line int
	if n := len(params); n > 1 {
		if i, err := strconv.Atoi(params[n-1]); err == nil && i > 0 {
			line, params = i, params[:n-1]
		}
	}
	name := strings.Join(params, " ")
	if name == "" {
		return text.ErrMissingRequiredArgument
	}
	m, err := p.Handler.Formatter(ctx)
	if err != nil {
		return err
	}
	var def string
	if p.Name == "ef" {
		def, err = m.FunctionDefinition(ctx, name)
		// the line of \ef counts from the start of the body, as \sf+ does
		if line != 0 {
			line += describe.FunctionBody(def)
		}
	} else {
		def, err = m.ViewDefinition(ctx, name)
	}
	if err != nil {
		return err
	}
	var arg string
	if line != 0 {
		arg = strconv.Itoa(line)
	}
	var out []byte
	err = onTerminal(p, func() error {
		var err error
		out, err = env.EditFile(p.Handler.User(), "", arg, []byte(def))
		return err
	})
	if err != nil {
		return err
	}
	// an unchanged definition is dropped, as psql drops it
	if strings.TrimRight(string(out), "\n") == def {
		fmt.Fprintln(p.Handler.Stdout(), text.NoChanges)
		return nil
	}
	// the edited statement is saved to the history, and read as input
	// again, where it waits for a ; or \g as psql leaves it
	p.Handler.IO().AddHistory(string(out))
	p.Handler.Buffer().Reset()
	p.Handler.Push(string(out))
	return nil
}

// Print is a Query Buffer meta command (\p, \print, \raw, \exec). Writes the
// query buffer to the output.
//
// Descs:
//
//	p	[-raw]	show the contents of the query buffer (-raw for the non-interpolated buffer)
//	print
//	exec:p
//	raw		as \p, but forces raw mode
func Print(p *Params) error {
	raw, err := p.RawOpt()
	if err != nil {
		return err
	}
	s := buffer(p, raw || p.Name == "raw")
	if s == "" {
		s = text.QueryBufferEmpty
	} else {
		s = p.Handler.Highlight(s)
	}
	fmt.Fprintln(p.Handler.Stdout(), s)
	return nil
}

// Write is a Query Buffer meta command (\w \write). Writes the query buffer to
// a file.
//
// Descs:
//
//	w	[-raw] FILE	write the query buffer to file (-raw for the non-interpolated buffer)
//	write
func Write(p *Params) error {
	raw, err := p.RawOpt()
	if err != nil {
		return err
	}
	name, err := p.Next(true)
	switch {
	case err != nil:
		return err
	case name == "":
		return text.ErrMissingRequiredArgument
	}
	s := buffer(p, raw)
	return os.WriteFile(name, []byte(strings.TrimSuffix(s, "\n")+"\n"), 0o644)
}

// buffer returns the query buffer, or the last query that ran when the buffer
// is empty. Its variables are replaced, as the query is sent, unless raw is
// set, which returns it as it was written.
func buffer(p *Params, raw bool) string {
	switch buf := p.Handler.Buffer(); {
	case !buf.Empty() && raw:
		return buf.Raw()
	case !buf.Empty():
		return buf.Expand(p.Handler.Lookup)
	case raw:
		return p.Handler.LastRaw()
	}
	return p.Handler.LastExec()
}

// Reset is a Query Buffer meta command (\r, \reset). Clears (resets) the query
// buffer.
//
// Descs:
//
//	r	reset (clear) the query buffer
//	reset
func Reset(p *Params) error {
	p.Handler.Reset()
	p.Handler.Print(text.QueryBufferReset)
	return nil
}

// SaveHistory is a Query Buffer meta command (\s). Writes the history of the
// line editor, oldest first, as psql does: to the output, or to a file when
// one is named.
//
// Descs:
//
//	s	[FILE]	display history or save it to file
func SaveHistory(p *Params) error {
	name, err := p.Next(true)
	if err != nil {
		return err
	}
	entries := p.Handler.IO().History()
	slices.Reverse(entries)
	if name == "" {
		w := p.Handler.Stdout()
		for _, entry := range entries {
			if _, err := fmt.Fprintln(w, entry); err != nil {
				return err
			}
		}
		return nil
	}
	var b strings.Builder
	for _, entry := range entries {
		b.WriteString(entry + "\n")
	}
	if err := os.WriteFile(name, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("\\s: %w", err)
	}
	p.Handler.Print(text.HistoryWritten, name)
	return nil
}

// History is a Query Buffer meta command (\hist, \history). Writes the history of
// the line editor, oldest first, with the number of each entry, as the
// history command of bash does. A count shows only the last entries.
//
// Descs:
//
//	hist	[COUNT]	show the history, or its last COUNT entries
//	history
func History(p *Params) error {
	s, err := p.Next(true)
	if err != nil {
		return err
	}
	entries := p.Handler.IO().History()
	slices.Reverse(entries)
	start := 0
	if s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			return fmt.Errorf(text.HistoryInvalidCount, p.Name, s)
		}
		start = max(len(entries)-n, 0)
	}
	w := p.Handler.Stdout()
	for i := start; i < len(entries); i++ {
		// the rows after the first of an entry line up under its text
		entry := strings.ReplaceAll(entries[i], "\n", "\n       ")
		if _, err := fmt.Fprintf(w, "%5d  %s\n", i+1, entry); err != nil {
			return err
		}
	}
	return nil
}

// Echo is a Input/Output meta command (\echo, \warn, \qecho). Writes a message
// to the output.
//
// Descs:
//
//	echo	[-n] [MESSAGE]...	write message to standard output (-n for no newline)
//	qecho	[-n] [MESSAGE]...	write message to \o output stream (-n for no newline)
//	warn	[-n] [MESSAGE]...	write message to standard error (-n for no newline)
func Echo(p *Params) error {
	n, ok, err := p.NextOpt(true)
	if err != nil {
		return err
	}
	f := fmt.Fprintln
	var vals []string
	switch {
	case ok && n == "n":
		f = fmt.Fprint
	case ok:
		vals = append(vals, "-"+n)
	default:
		vals = append(vals, n)
	}
	v, err := p.All(true)
	if err != nil {
		return err
	}
	out := p.Handler.Stdout()
	switch o := p.Handler.GetOutput(); {
	case p.Name == "qecho" && o != nil:
		out = o
	case p.Name == "warn":
		out = p.Handler.IO().Stderr()
	}
	f(out, strings.Join(append(vals, v...), " "))
	return nil
}

// Out is a Input/Output meta command (\o \out). Sets (redirects) the output to
// a file or a command.
//
// Descs:
//
//	o	[FILE]	send all query results to file or |pipe
//	out
func Out(p *Params) error {
	p.Handler.SetOutput(nil)
	params, err := p.All(true)
	if err != nil {
		return err
	}
	pipe := strings.Join(params, " ")
	if pipe == "" {
		return nil
	}
	var out io.WriteCloser
	if pipe[0] == '|' {
		out, _, err = env.Pipe(p.Handler.IO(), p.Handler.IO().Stderr(), pipe[1:])
	} else {
		out, err = os.OpenFile(pipe, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	}
	if err != nil {
		return err
	}
	p.Handler.SetOutput(out)
	return nil
}

// Copy is a Input/Output meta command (\copy). Copies data between databases.
//
// Descs:
//
//	copy	SRC DST QUERY TABLE	copy results of query from source database into table on destination database
//	copy	SRC DST QUERY TABLE(A,...)	copy results of query from source database into table's columns on destination database
func Copy(p *Params) error {
	srcstr, err := p.Next(true)
	if err != nil {
		return err
	}
	src, err := dburl.Parse(srcstr)
	if err != nil {
		return err
	}
	deststr, err := p.Next(true)
	if err != nil {
		return err
	}
	dest, err := dburl.Parse(deststr)
	if err != nil {
		return err
	}
	query, err := p.Next(true)
	if err != nil {
		return err
	}
	table, err := p.Next(true)
	if err != nil {
		return err
	}
	ctx := context.Background()
	stdout := p.Handler.Stdout
	stderr := p.Handler.IO().Stderr
	srcDb, err := drivers.Open(ctx, src, stdout, stderr)
	if err != nil {
		return err
	}
	defer srcDb.Close()
	destDb, err := drivers.Open(ctx, dest, stdout, stderr)
	if err != nil {
		return err
	}
	defer destDb.Close()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	// get the result set
	r, err := srcDb.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer r.Close()
	n, err := drivers.Copy(ctx, dest, stdout, stderr, r, table)
	if err != nil {
		return err
	}
	p.Handler.Print("COPY %d", n)
	return nil
}

// DumpDatabase is a Input/Output meta command (\dump). Writes the database of
// the connection as a script, as --dump does, to the output, to a file, or to
// a command. It takes the options of --dump, written as -NAME, from the same
// table, so that a \dump and a --dump with the same options write the same
// script. A file is written whole or not at all.
//
// Descs:
//
//	dump	[OPTIONS] [FILE]	write the database as a {{CommandName}} script to the output, a file or |pipe (OPTIONS: -target DRIVER, -schema-only, -data-only, -schema PATTERN, -exclude-schema PATTERN, -table PATTERN, -exclude-table PATTERN, -allow-lossy, -allow-text, -restrict-key KEY)
func DumpDatabase(p *Params) error {
	var opts dump.Options
	var dest string
	for dest == "" {
		a, ok, err := p.NextOK(true)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		name, isOpt := dumpOption(a)
		if !isOpt {
			dest = a
			break
		}
		f, ok := dump.LookupFlag(name)
		if !ok {
			return fmt.Errorf(text.InvalidOption, name)
		}
		var value string
		if f.Value != "" {
			if value, ok, err = p.NextOK(true); err != nil {
				return err
			} else if !ok {
				return fmt.Errorf(text.DumpOptionNeedsValue, name, f.Value)
			}
		}
		if err := f.Set(&opts, value); err != nil {
			return err
		}
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	write := func(out io.Writer) error {
		w := bufio.NewWriter(out)
		if err := p.Handler.Dump(ctx, w, opts); err != nil {
			_ = w.Flush()
			return err
		}
		return w.Flush()
	}
	switch {
	case dest == "" || dest == "-":
		if out := p.Handler.GetOutput(); out != nil {
			return write(out)
		}
		return write(p.Handler.Stdout())
	case dest[0] == '|':
		// the command is the rest of the arguments, as for \o
		rest, err := p.All(true)
		if err != nil {
			return err
		}
		out, cmd, err := env.Pipe(p.Handler.IO(), p.Handler.IO().Stderr(), strings.Join(append([]string{dest[1:]}, rest...), " "))
		if err != nil {
			return err
		}
		err = write(out)
		if cerr := out.Close(); err == nil {
			err = cerr
		}
		if werr := cmd.Wait(); err == nil {
			err = werr
		}
		return err
	}
	return dump.WriteFile(dest, write)
}

// dumpOption returns the name of an option of \dump, which starts with one
// dash or two, as -schema-only or --schema-only. A lone dash names standard
// output, and is not an option.
func dumpOption(s string) (string, bool) {
	if len(s) < 2 || s[0] != '-' {
		return "", false
	}
	return strings.TrimPrefix(s[1:], "-"), true
}

// Set is a Control/Conditional meta command (\set). Sets (or shows) the script
// variables, the section vars of \config, by their names without the section.
// The settings of the application are set with \uset.
//
// Descs:
//
//	set	[NAME [VALUE]]	set script variable, or show all script variables if no parameters
func Set(p *Params) error {
	switch n, ok, err := p.NextOK(true); {
	case err != nil:
		return err
	case ok:
		vals, err := p.All(true)
		if err != nil {
			return err
		}
		return env.Vars().Set(n, strings.Join(vals, " "))
	}
	return env.Vars().Dump(p.Handler.Stdout())
}

// Unset is a Control/Conditional meta command (\unset). Unsets a script variable.
//
// Descs:
//
//	unset	NAME	unset (delete) script variable
func Unset(p *Params) error {
	n, err := p.Next(true)
	if err != nil {
		return err
	}
	return env.Vars().Unset(n)
}

// Prompt is a Control/Conditional meta command (\prompt). Prompts the user for
// input, and sets a script variable to the answer.
//
// Descs:
//
//	prompt	[-TYPE] VAR [PROMPT]	prompt user to set script variable
func Prompt(p *Params) error {
	typ := "string"
	n, ok, err := p.NextOpt(true)
	if err != nil {
		return err
	}
	if ok {
		typ = n
		n, err = p.Next(true)
		if err != nil {
			return err
		}
	}
	if n == "" {
		return text.ErrMissingRequiredArgument
	}
	if err := env.ValidIdentifier(n); err != nil {
		return err
	}
	vals, err := p.All(true)
	if err != nil {
		return err
	}
	v, err := p.Handler.ReadVar(typ, strings.Join(vals, " "))
	if err != nil {
		return err
	}
	return env.Vars().Set(n, v)
}

// Include is a Control/Conditional meta command (\i, \include and variants).
// Includes (runs) the specified file in the current execution environment.
//
// Descs:
//
//	i	FILE	execute commands from file
//	include:i
//	ir	FILE	as \i, but relative to location of current script
//	include_relative:ir
func Include(p *Params) error {
	path, err := p.Next(true)
	if err != nil {
		return err
	}
	relative := p.Name == "ir" || p.Name == "include_relative"
	if err := p.Handler.Include(path, relative); err != nil {
		return fmt.Errorf("%s: %v", path, err)
	}
	return nil
}

// Conditional is a Control/Conditional meta command (\if, \elif, \else,
// \endif). Starts, closes, and ends a conditional block within the
// application.
//
// The expression of \if and \elif is a boolean, read as psql reads it: its
// arguments, with their variables replaced, joined by a space. A branch that
// does not run reads no expression, so it replaces no variable and runs no
// command between backticks.
//
// Descs:
//
//	if	EXPR	begin conditional block
//	elif	EXPR	alternative within current conditional block
//	else	final alternative within current conditional block
//	endif	end conditional block
func Conditional(p *Params) error {
	switch p.Name {
	case "if":
		defer p.skip()
		return p.Handler.If(p.boolean)
	case "elif":
		defer p.skip()
		return p.Handler.ElseIf(p.boolean)
	case "else":
		return p.Handler.Else()
	case "endif":
		return p.Handler.EndIf()
	}
	return nil
}

// boolean returns the value of the expression of \if or \elif.
func (p *Params) boolean() (bool, error) {
	args, err := p.All(true)
	if err != nil {
		return false, err
	}
	s := strings.Join(args, " ")
	v, ok := env.ParseVariableBool(s)
	if !ok {
		return false, fmt.Errorf(text.UnrecognizedValueForCond, s, p.Name)
	}
	return v, nil
}

// skip marks the arguments that are left as read, so that the handler does
// not report them as ignored.
func (p *Params) skip() {
	p.next = len(p.item.Args)
}

// Restrict is a Control/Conditional meta command (\restrict, \unrestrict). Enters and
// leaves restricted mode, in which no meta command runs but \unrestrict with
// the key, as psql does. pg_dump writes \restrict at the start of a dump and
// \unrestrict at its end, so that the data of the dump cannot run a command.
//
// Descs:
//
//	restrict	RESTRICT_KEY	enter restricted mode with provided key
//	unrestrict	RESTRICT_KEY	exit restricted mode if key matches
func Restrict(p *Params) error {
	key, err := p.Next(true)
	if err != nil {
		return err
	}
	if p.Name == "unrestrict" {
		err = p.Handler.Unrestrict(key)
	} else {
		err = p.Handler.Restrict(key)
	}
	if err != nil {
		return fmt.Errorf(`\%s: %w`, p.Name, err)
	}
	return nil
}

// Transact is a Transaction meta command (\begin, \commit, \rollback). Begins,
// commits, or aborts (rollback) the current database transaction on the open
// database connection.
//
// Descs:
//
//	begin	[-read-only [ISOLATION]]	begin transaction, with optional isolation level
//	commit	commit current transaction
//	rollback	rollback (abort) current transaction
//	abort:rollback
func Transact(p *Params) error {
	switch p.Name {
	case "commit":
		return p.Handler.Commit()
	case "rollback", "abort":
		return p.Handler.Rollback()
	}
	// read begin params
	readOnly := false
	n, ok, err := p.NextOpt(true)
	if ok {
		if n != "read-only" {
			return fmt.Errorf(text.InvalidOption, n)
		}
		readOnly = true
		if n, err = p.Next(true); err != nil {
			return err
		}
	}
	// build tx options
	var txOpts *sql.TxOptions
	if readOnly || n != "" {
		isolation := sql.LevelDefault
		switch strings.ToLower(n) {
		case "default", "":
		case "read-uncommitted":
			isolation = sql.LevelReadUncommitted
		case "read-committed":
			isolation = sql.LevelReadCommitted
		case "write-committed":
			isolation = sql.LevelWriteCommitted
		case "repeatable-read":
			isolation = sql.LevelRepeatableRead
		case "snapshot":
			isolation = sql.LevelSnapshot
		case "serializable":
			isolation = sql.LevelSerializable
		case "linearizable":
			isolation = sql.LevelLinearizable
		default:
			return text.ErrInvalidIsolationLevel
		}
		txOpts = &sql.TxOptions{
			Isolation: isolation,
			ReadOnly:  readOnly,
		}
	}
	// begin
	return p.Handler.Begin(txOpts)
}

// Config is a Settings meta command (\config). Sets, or shows, the
// configuration of the application. A setting has the name SECTION.NAME, as in
// print.border. \pset, \uset, \cset, \set and \setenv set the sections
// print, usql, connections, vars and env, and show them by the names without
// the section. Only the configuration file sets the section aliases.
//
// Descs:
//
//	config	[OPTIONS] [NAME [VALUE]]	set configuration option NAME (SECTION.NAME), or show the settings of usql and print if no parameters (OPTIONS: -all shows every section, -password shows passwords)
func Config(p *Params) error {
	opts, err := p.flags("all", "password")
	if err != nil {
		return err
	}
	stdout := p.Handler.Stdout()
	name, ok, err := p.NextOK(true)
	switch {
	case err != nil:
		return err
	case !ok && opts["all"]:
		return env.Vars().DumpConfig(stdout, opts["password"], env.Sections...)
	case !ok:
		return env.Vars().DumpConfig(stdout, opts["password"])
	}
	section, key, err := env.Vars().SplitName(name)
	switch {
	case err != nil:
		return err
	case key == "":
		return env.Vars().DumpConfig(stdout, opts["password"], section)
	}
	vals, err := p.All(true)
	switch {
	case err != nil:
		return err
	case len(vals) != 0:
		return setConfig(p, name, section, key, vals)
	}
	value, ok := env.Vars().Config(section, key, opts["password"])
	if !ok {
		return fmt.Errorf(text.ConfigNotSet, name)
	}
	fmt.Fprintln(stdout, strings.TrimRight(name+" "+value, " "))
	return nil
}

// setConfig sets a setting, and writes the message of \pset or \uset for
// their sections.
func setConfig(p *Params, name, section, key string, vals []string) error {
	value := strings.Join(vals, " ")
	switch section {
	case env.SectionPrint:
		value, err := env.Vars().SetPrint(key, value)
		if err != nil {
			return err
		}
		printSet(p, key, value)
		return nil
	case env.SectionUsql:
		if err := env.Vars().SetSetting(key, value); err != nil {
			return err
		}
		p.Handler.Print(text.SettingIs, name, env.Vars().SettingText(key))
		return nil
	case env.SectionConnections:
		return env.Vars().SetConn(key, vals...)
	}
	return env.Vars().SetConfig(name, value)
}

// SetPrint is a Settings meta command (\pset, \a, \C, \f, \H, \t, \T, \x).
// Sets, toggles, or displays the application's print formatting variables.
//
// Descs:
//
//	pset	[NAME [VALUE]]	set table print formatting option, or show all print formatting options if no parameters
//	a		toggle between unaligned and aligned output mode	DEPRECATED
//	C	[TITLE]	set table title, or unset if none	DEPRECATED
//	f	[SEPARATOR]	show or set field separator for unaligned query output	DEPRECATED
//	H		toggle HTML output mode	DEPRECATED
//	html:H
//	T	[ATTRIBUTES]	set HTML <table> tag attributes, or unset if none	DEPRECATED
//	t	[on|off]	show only rows	DEPRECATED
//	x	[on|off|auto]	toggle expanded output	DEPRECATED
func SetPrint(p *Params) error {
	var ok bool
	var val string
	var err error
	// \html is psql's long name of \H
	if p.Name == "html" {
		p.Name = "H"
	}
	switch p.Name {
	case "a", "H":
	default:
		if val, ok, err = p.NextOK(true); err != nil {
			return err
		}
	}
	// display variables
	if p.Name == "pset" && !ok {
		return env.Vars().DumpPrint(p.Handler.Stdout())
	}
	var field, extra string
	switch p.Name {
	case "pset":
		field = val
		if val, ok, err = p.NextOK(true); err != nil {
			return err
		}
	case "a":
		field = "format"
	case "C":
		field = "title"
	case "f":
		field = "fieldsep"
	case "H":
		field, extra = "format", "html"
	case "t":
		field = "tuples_only"
	case "T":
		field = "tableattr"
	case "x":
		field = "expanded"
	}
	if !ok {
		if val, err = env.Vars().TogglePrint(field, extra); err != nil {
			return err
		}
	} else {
		if val, err = env.Vars().SetPrint(field, val); err != nil {
			return err
		}
	}
	printSet(p, field, val)
	return nil
}

// printSet writes the message of \pset for a print setting that changed.
func printSet(p *Params, field, val string) {
	// a number of xheader_width is written as a number, as psql writes it
	if _, err := strconv.Atoi(val); field == "xheader_width" && err == nil {
		field = "xheader_width_n"
	}
	// special replacement name for expanded field, when 'auto'
	if field == "expanded" && val == "auto" {
		field = "expanded_auto"
	}
	// format output
	mask := text.FormatFieldNameSetMap[field]
	unsetMask := text.FormatFieldNameUnsetMap[field]
	switch {
	case strings.Contains(mask, "%d"):
		i, _ := strconv.Atoi(val)
		p.Handler.Print(mask, i)
	case unsetMask != "" && val == "":
		p.Handler.Print(unsetMask)
	case !strings.Contains(mask, "%"):
		p.Handler.Print(mask)
	default:
		if field == "time" {
			val = fmt.Sprintf("%q", val)
			if tfmt := env.Vars().PrintTimeFormat(); tfmt != val {
				val = fmt.Sprintf("%s (%q)", val, tfmt)
			}
		}
		p.Handler.Print(mask, val)
	}
}

// Uset is a Settings meta command (\uset). Sets, or shows, the settings of
// the application, such as pager and on_error_stop, as \pset does for the
// print settings.
//
// Descs:
//
//	uset	[NAME [VALUE]]	set {{CommandName}} configuration option, or show all {{CommandName}} settings if no parameters
func Uset(p *Params) error {
	name, ok, err := p.NextOK(true)
	switch {
	case err != nil:
		return err
	case !ok:
		return env.Vars().DumpSettings(p.Handler.Stdout())
	}
	vals, err := p.All(true)
	if err != nil {
		return err
	}
	if len(vals) != 0 {
		if err := env.Vars().SetSetting(name, strings.Join(vals, " ")); err != nil {
			return err
		}
	}
	if _, ok := env.Vars().Setting(name); !ok {
		return fmt.Errorf(text.UnknownSetting, name)
	}
	p.Handler.Print(text.SettingIs, name, env.Vars().SettingText(name))
	return nil
}

// SetConn is a Settings meta command (\cset). Sets a connection variable.
//
// Descs:
//
//	cset	[NAME [URL]]	set named connection, or show all named connections if no parameters
//	cset	NAME DRIVER PARAMS...	set named connection for driver and parameters
func SetConn(p *Params) error {
	switch n, ok, err := p.NextOK(true); {
	case err != nil:
		return err
	case ok:
		vals, err := p.All(true)
		if err != nil {
			return err
		}
		return env.Vars().SetConn(n, vals...)
	}
	return env.Vars().DumpConn(p.Handler.Stdout())
}

// Describe is a Informational meta command (\d and variants). Queries the open
// database connection for information about the database schema and writes the
// information to the output.
//
// Descs:
//
//	d[Sx+]	[NAME]	list tables, views, and sequences or describe table, view, sequence, or index
//	d[tvmsEi][Sx+]
//	da[Sx+]	[PATTERN]	list aggregates
//	dA[x+]	[PATTERN]	list access methods
//	dAc[x+]	[AMPTRN [PATTERN]]	list operator classes
//	dAf[x+]	[AMPTRN [PATTERN]]	list operator families
//	dAo[x+]	[AMPTRN [PATTERN]]	list operators of operator families
//	dAp[x+]	[AMPTRN [PATTERN]]	list support functions of operator families
//	db[x+]	[PATTERN]	list tablespaces
//	dc[Sx+]	[PATTERN]	list conversions
//	dconfig[x+]	[PATTERN]	list configuration parameters
//	dC[x+]	[PATTERN]	list casts
//	dd[Sx]	[PATTERN]	show object descriptions not displayed elsewhere
//	dD[Sx+]	[PATTERN]	list domains
//	ddp[x]	[PATTERN]	list default privileges
//	dE[Sx+]	[PATTERN]	list foreign tables
//	des[x+]	[PATTERN]	list foreign servers
//	det[x+]	[PATTERN]	list foreign tables and their servers
//	deu[x+]	[PATTERN]	list user mappings
//	dew[x+]	[PATTERN]	list foreign-data wrappers
//	df[anptw][Sx+]	[NAME [TYPES]]	list [only agg/normal/procedure/trigger/window] functions
//	dF[x+]	[PATTERN]	list text search configurations
//	dFd[x+]	[PATTERN]	list text search dictionaries
//	dFp[x+]	[PATTERN]	list text search parsers
//	dFt[x+]	[PATTERN]	list text search templates
//	dg[Sx+]	[PATTERN]	list roles
//	di[Sx+]	[PATTERN]	list indexes
//	dl[x+]	list large objects, same as \lo_list
//	dL[Sx+]	[PATTERN]	list procedural languages
//	dm[Sx+]	[PATTERN]	list materialized views
//	dn[Sx+]	[PATTERN]	list schemas
//	do[Sx+]	[NAME [TYPES]]	list operators
//	dO[Sx+]	[PATTERN]	list collations
//	dp[Sx]	[PATTERN]	list table, view, and sequence access privileges
//	dP[itn][x+]	[PATTERN]	list [only index/table] partitioned relations [n=nested]
//	drds[x]	[ROLEPTRN [DBPTRN]]	list per-database role settings
//	drg[Sx]	[PATTERN]	list role grants
//	dRp[x+]	[PATTERN]	list replication publications
//	dRs[x+]	[PATTERN]	list replication subscriptions
//	ds[Sx+]	[PATTERN]	list sequences
//	dt[Sx+]	[PATTERN]	list tables
//	dT[Sx+]	[PATTERN]	list data types
//	du[Sx+]	[PATTERN]	list roles
//	dv[Sx+]	[PATTERN]	list views
//	dx[x+]	[PATTERN]	list extensions
//	dX[x]	[PATTERN]	list extended statistics
//	dy[x+]	[PATTERN]	list event triggers
//	l[x+]	[PATTERN]	list databases
//	list[x+]:l
//	lo_list[x+]:dl
//	z[Sx]	[PATTERN]	same as \dp
func Describe(p *Params) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	m, err := p.Handler.Formatter(ctx)
	if err != nil {
		return err
	}
	verbose, showSystem := p.Modifier('+'), p.Modifier('S')
	// x prints this command in expanded mode, as in psql 18
	m.SetExpanded(p.Modifier('x'))
	name := p.Base()
	// \list is psql's long name of \l
	if name == "list" {
		name = "l"
	}
	patterns, err := p.Patterns()
	if err != nil {
		return err
	}
	var pattern string
	if len(patterns) != 0 {
		pattern = patterns[0]
		// psql hides the objects of the system only when neither S nor a
		// pattern is given
		showSystem = true
	}
	switch {
	case name == "d":
		if pattern != "" {
			return m.DescribeTableDetails(ctx, pattern, verbose, showSystem)
		}
		return m.ListTables(ctx, "tvmsE", pattern, verbose, showSystem)
	case name == "da", strings.HasPrefix(name, "df"):
		// psql's \df takes the types of the arguments after the name, and
		// its \da takes none
		var types []string
		if strings.HasPrefix(name, "df") && len(patterns) > 1 {
			types = patterns[1:]
		}
		return m.DescribeFunctions(ctx, name, pattern, types, verbose, showSystem)
	case name == "dp", name == "z":
		return m.ListPrivilegeSummaries(ctx, pattern, showSystem)
	case name[0] == 'd' && strings.Trim(name[1:], "tvmsEi") == "":
		// any combination of the classes, such as \dtv and \dti
		return m.ListTables(ctx, name, pattern, verbose, showSystem)
	case name == "dn":
		return m.ListSchemas(ctx, pattern, verbose, showSystem)
	case name == "l":
		return m.ListAllDbs(ctx, pattern, verbose)
	case name == "lo_list":
		name = "dl"
	}
	// the letters after dP choose indexes, tables and nested relations
	var flags string
	if strings.HasPrefix(name, "dP") {
		name, flags = "dP", name[2:]
	}
	_, err = m.Run(ctx, describe.Call{Name: name, Flags: flags, Patterns: patterns, Verbose: verbose, System: showSystem})
	return err
}

// ShowSource is a Informational meta command (\sf, \sv). Writes the
// definition of a function or a view.
//
// Descs:
//
//	sf[+]	FUNCNAME	show a function's definition
//	sv[+]	VIEWNAME	show a view's definition
func ShowSource(p *Params) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	m, err := p.Handler.Formatter(ctx)
	if err != nil {
		return err
	}
	// a function's argument types can hold spaces, as in f(integer, text)
	params, err := p.Patterns()
	name := strings.Join(params, " ")
	switch {
	case err != nil:
		return err
	case name == "":
		return text.ErrMissingRequiredArgument
	}
	verbose := strings.HasSuffix(p.Name, "+")
	if strings.HasPrefix(p.Name, "sf") {
		return m.ShowFunction(ctx, name, verbose)
	}
	return m.ShowView(ctx, name, verbose)
}

// Stats is a Informational meta command (\ss and variants). Queries the open
// database connection for stats and writes it to the output.
//
// Descs:
//
//	ss[+]	TABLE [k]	show stats for the columns of a table
func Stats(p *Params) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	m, err := p.Handler.Formatter(ctx)
	if err != nil {
		return err
	}
	verbose := strings.ContainsRune(p.Name, '+')
	// a table is required, because the statistics of every column of every
	// table is a long list that says little
	pattern, err := p.NextPattern()
	switch {
	case err != nil:
		return err
	case pattern == "":
		return text.ErrMissingRequiredArgument
	}
	k := 0
	if verbose {
		k = 3
	}
	val, ok, err := p.NextOK(true)
	switch {
	case err != nil:
		return err
	case ok:
		verbose = true
		if k, err = strconv.Atoi(val); err != nil {
			return err
		}
	}
	return m.ShowStats(ctx, pattern, verbose, k)
}

// Kubernetes is a Informational meta command (\k, \kube). Lists the databases that
// dbfwd can reach in a Kubernetes cluster, with the URL that \c opens.
//
// The scope is a k8s URL whose k:// is optional, such as * for the current
// namespace, pg*.prod, *.* for every namespace, or cnpg://* for the
// CloudNativePG clusters. The options -A, -n NAMESPACE and --context NAME fill
// in what the scope leaves out, as kubectl reads them, and dbfwd reports an
// option that disagrees with it.
//
// Descs:
//
//	k[+]	[OPTIONS] [SCOPE]	list databases in Kubernetes, with the URL for \c (OPTIONS: -A, -n NAMESPACE, --context NAME)
//	kube[+]:k
func Kubernetes(p *Params) error {
	var all bool
	var namespace, kubeContext, scope string
	for {
		v, ok, err := p.NextOK(true)
		switch {
		case err != nil:
			return err
		case !ok:
			return listKubernetes(p, scope, all, namespace, kubeContext)
		}
		name, value, hasValue := strings.Cut(v, "=")
		switch {
		case v == "-A" || v == "--all-namespaces":
			all = true
		case name == "-n" || name == "--namespace" || name == "--context":
			if !hasValue {
				if value, err = p.Next(true); err != nil {
					return err
				}
			}
			if value == "" {
				return fmt.Errorf(text.KubernetesMissingValue, name)
			}
			if name == "--context" {
				kubeContext = value
			} else {
				namespace = value
			}
		case strings.HasPrefix(v, "-") && v != "-":
			return fmt.Errorf(text.InvalidOption, v)
		case scope != "":
			return fmt.Errorf(text.KubernetesExtraArgument, v, scope)
		default:
			scope = v
		}
	}
}

// listKubernetes lists the databases of a scope, and writes them as a table.
func listKubernetes(p *Params, scope string, all bool, namespace, kubeContext string) error {
	var where []dbfwd.Field
	switch {
	case all && namespace != "":
		return fmt.Errorf(text.KubernetesAllAndNamespace, namespace)
	case all:
		namespace = "*"
	}
	if namespace != "" {
		where = append(where, dbfwd.Field{Name: "namespace", Value: namespace})
	}
	if kubeContext != "" {
		where = append(where, dbfwd.Field{Name: "context", Value: kubeContext})
	}
	u, err := dbfwd.ListURL("k8s", scope, where...)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var opts []dbfwd.Option
	if l := p.Handler.IO(); l.Interactive() {
		opts = append(opts, dbfwd.WithInteractive(l.Release))
	}
	listing, err := dbfwd.List(ctx, u, opts...)
	switch {
	case errors.Is(err, dbfwd.ErrForbidden) && namespace == "*":
		return fmt.Errorf(text.KubernetesAllForbidden, err)
	case err != nil:
		return err
	}
	w, scopeName := p.Handler.Stdout(), scopeText(listing.Scope)
	if len(listing.Entries) == 0 {
		fmt.Fprintf(w, text.KubernetesNotFound+"\n", scopeName)
	} else {
		plus := strings.HasSuffix(p.Name, "+")
		columns := []string{"Namespace", "Name", "Type", "Kind", "Operator"}
		if plus {
			columns = append(columns, "Roles", "Ready", "Databases", "Credential")
		}
		columns = append(columns, "URL")
		values := make([][]any, 0, len(listing.Entries))
		for _, e := range listing.Entries {
			row := []any{fieldValue(e.Target.Scope, "namespace"), e.Target.Name, e.Target.Type, e.Target.Kind, e.Operator}
			if plus {
				row = append(row,
					strings.Join(e.Roles, ", "),
					strconv.Itoa(e.Ready)+"/"+strconv.Itoa(e.Pods),
					strings.Join(e.Databases, ", "),
					e.Credential,
				)
			}
			values = append(values, append(row, e.URL))
		}
		if err := describe.Table(w, fmt.Sprintf(text.KubernetesTitle, scopeName), columns, values); err != nil {
			return err
		}
	}
	// what RBAC kept from the listing goes to stderr after the table, as psql
	// writes its notes
	for _, s := range listing.Skipped {
		fmt.Fprintln(p.Handler.IO().Stderr(), fmt.Sprintf(text.KubernetesSkipped, s.Kind, scopeText(s.Scope), s.Reason))
	}
	return nil
}

// scopeText returns a scope of dbfwd as text, such as "context kind-prod,
// namespace prod". Every namespace reads as "every namespace".
func scopeText(scope []dbfwd.Field) string {
	parts := make([]string, 0, len(scope))
	for _, f := range scope {
		switch {
		case f.Value == "":
		case f.Name == "namespace" && f.Value == "*":
			parts = append(parts, "every namespace")
		default:
			parts = append(parts, f.Name+" "+f.Value)
		}
	}
	return strings.Join(parts, ", ")
}

// fieldValue returns the value of the field name of a scope of dbfwd.
func fieldValue(scope []dbfwd.Field, name string) string {
	for _, f := range scope {
		if f.Name == name {
			return f.Value
		}
	}
	return ""
}

// redacted returns u with its password written as xxxxx, as the Redacted
// method of net/url writes it.
func redacted(u *dburl.URL) string {
	if u.User == nil {
		return u.String()
	}
	if _, ok := u.User.Password(); !ok {
		return u.String()
	}
	v := *u
	v.User = url.UserPassword(u.User.Username(), "xxxxx")
	return v.String()
}

// Shell is a Operating System/Environment meta command (\!). Executes a
// command using the Operating System/Environment's shell.
//
// Descs:
//
//	!	[COMMAND]	execute command in shell or start interactive shell
func Shell(p *Params) error {
	return onTerminal(p, func() error {
		return env.Shell(p.Raw())
	})
}

// onTerminal runs fn while the line editor lends the terminal to the program
// that fn starts, so that Ctrl-C goes to that program and does not end usql.
// fn runs on its own when the line editor does not write to a file.
func onTerminal(p *Params, fn func() error) error {
	l := p.Handler.IO()
	if l == nil {
		return fn()
	}
	err := l.Release(func(*os.File) error {
		return fn()
	})
	if errors.Is(err, rline.ErrNotAFile) {
		return fn()
	}
	return err
}

// Chdir is a Operating System/Environment meta command (\cd). Changes the
// current directory for the Operating System/Environment.
//
// Descs:
//
//	cd	[DIR]	change the current working directory
func Chdir(p *Params) error {
	dir, err := p.Next(true)
	if err != nil {
		return err
	}
	return env.Chdir(p.Handler.User(), dir)
}

// Getenv is a Operating System/Environment meta command (\getenv). Sets a
// script variable to the value of a variable of the environment, as the
// commands that the application runs see it.
//
// Descs:
//
//	getenv	VARNAME ENVVAR	fetch environment variable
func Getenv(p *Params) error {
	n, err := p.Next(true)
	switch {
	case err != nil:
		return err
	case n == "":
		return text.ErrMissingRequiredArgument
	}
	v, err := p.Next(true)
	switch {
	case err != nil:
		return err
	case v == "":
		return text.ErrMissingRequiredArgument
	}
	value, _ := env.Vars().Getenv(v)
	return env.Vars().Set(n, value)
}

// Setenv is a Operating System/Environment meta command (\setenv). Sets (or
// unsets) a variable of the environment of the programs that the application
// starts, the section env of \config. The environment of the application does
// not change.
//
// Descs:
//
//	setenv	NAME [VALUE]	set or unset environment variable of the commands that {{CommandName}} runs
func Setenv(p *Params) error {
	n, err := p.Next(true)
	switch {
	case err != nil:
		return err
	case n == "":
		return text.ErrMissingRequiredArgument
	}
	switch v, ok, err := p.NextOK(true); {
	case err != nil:
		return err
	case ok:
		return env.Vars().SetEnv(n, v)
	}
	return env.Vars().UnsetEnv(n)
}
