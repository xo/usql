package metacmd

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
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
	stdout := p.Handler.IO().Stdout()
	if typ := env.TermGraphics(); typ.Available() {
		typ.Encode(stdout, text.Logo)
	}
	fmt.Fprintln(stdout, text.Copyright)
	return nil
}

// Drivers is a General meta command (\drivers). Writes information about the
// database drivers the application was built with to the output.
//
// Descs:
//
//	drivers	show database drivers available to {{CommandName}}
func Drivers(p *Params) error {
	stdout, stderr := p.Handler.IO().Stdout(), p.Handler.IO().Stderr()
	var cmd *exec.Cmd
	var wc io.WriteCloser
	if pager := env.Get("PAGER"); p.Handler.IO().Interactive() && pager != "" {
		var err error
		if wc, cmd, err = env.Pipe(stdout, stderr, pager); err != nil {
			return err
		}
		stdout = wc
	}
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
		driver, aliases := dburl.SchemeDriverAndAliases(n)
		if driver != n {
			s += " (" + driver + ")"
		}
		if len(aliases) > 0 {
			s += " [" + strings.Join(aliases, ", ") + "]"
		}
		fmt.Fprintln(stdout, s)
	}
	if cmd != nil {
		if err := wc.Close(); err != nil {
			return err
		}
		return cmd.Wait()
	}
	return nil
}

// Help is a Help meta command (\?). Writes a help message to the output.
//
// Descs:
//
//	?	[commands]	show help on {{CommandName}}'s meta (backslash) commands
//	?	options	show help on {{CommandName}} command-line options
//	?	variables	show help on special {{CommandName}} variables
func Help(p *Params) error {
	name, err := p.Next(false)
	if err != nil {
		return err
	}
	stdout, stderr := p.Handler.IO().Stdout(), p.Handler.IO().Stderr()
	var cmd *exec.Cmd
	var wc io.WriteCloser
	if pager := env.Get("PAGER"); p.Handler.IO().Interactive() && pager != "" {
		if wc, cmd, err = env.Pipe(stdout, stderr, pager); err != nil {
			return err
		}
		stdout = wc
	}
	switch name = strings.TrimSpace(strings.ToLower(name)); {
	case name == "options":
		text.Usage(stdout, true)
	case name == "variables":
		_ = env.Listing(stdout)
	default:
		_ = Dump(stdout, name == "commands")
	}
	if cmd != nil {
		if err := wc.Close(); err != nil {
			return err
		}
		return cmd.Wait()
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
func Execute(p *Params) error {
	p.Option.Exec = ExecOnly
	switch p.Name {
	case "g", "go", "G", "ego", "gx", "gset":
		params, err := p.All(true)
		switch {
		case err != nil:
			return err
		case p.Name != "gset":
			p.Option.ParseParams(params, "pipe")
		}
		switch p.Name {
		case "G", "ego":
			p.Option.Params["format"] = "vertical"
		case "gx":
			p.Option.Params["expanded"] = "on"
		case "gset":
			p.Option.Exec = ExecSet
			p.Option.ParseParams(params, "prefix")
		}
	case "gexec":
		p.Option.Exec = ExecExec
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

// ConnectionInfo is a Connection meta command (\conninfo). Writes information
// about the connection to the output.
//
// Descs:
//
//	conninfo	display information about the current database connection
func ConnectionInfo(p *Params) error {
	s := text.NotConnected
	if db, u := p.Handler.DB(), p.Handler.URL(); db != nil && u != nil {
		s = fmt.Sprintf(text.ConnInfo, u.Driver, u.DSN)
	}
	fmt.Fprintln(p.Handler.IO().Stdout(), s)
	return nil
}

// Edit is a Query Buffer meta command (\e \edit). Opens the query buffer for
// editing in an external application.
//
// Descs:
//
//	e	[-raw|-exec] [FILE] [LINE]	edit the query buffer, raw (non-interpolated) buffer, the exec buffer, or a file with external editor
//	edit
func Edit(p *Params) error {
	var exec bool
	path, ok, err := p.NextOpt(true)
	if ok {
		if path != "exec" {
			return fmt.Errorf(text.InvalidOption, path)
		}
		exec = true
		if path, err = p.Next(true); err != nil {
			return err
		}
	}
	// get last statement
	s, buf := "", p.Handler.Buf()
	switch {
	case buf.Len != 0 && exec:
		s = buf.String()
	case buf.Len != 0:
		s = buf.RawString()
	case exec:
		s = p.Handler.LastExec()
	default:
		s = p.Handler.LastRaw()
	}
	line, err := p.Next(true)
	if err != nil {
		return err
	}
	// reset if no error
	out, err := env.EditFile(p.Handler.User(), path, line, []byte(s))
	if err != nil {
		return err
	}
	// save edited buffer to history
	p.Handler.IO().Save(string(out))
	buf.Reset([]rune(string(out)))
	return nil
}

// Print is a Query Buffer meta command (\p, \print, \raw, \exec). Writes the
// query buffer to the output.
//
// Descs:
//
//	p	[-raw|-exec]	show the contents of the query buffer, the raw (non-interpolated) buffer or the exec buffer
//	print
//	raw
//	exec
func Print(p *Params) error {
	// get last statement
	var s string
	switch buf := p.Handler.Buf(); {
	case buf.Len != 0 && p.Name == "exec":
		s = buf.String()
	case buf.Len != 0 && p.Name == "raw":
		s = buf.RawString()
	case buf.Len != 0:
		s = buf.PrintString()
	case p.Name == "exec":
		s = p.Handler.LastExec()
	case p.Name == "raw":
		s = p.Handler.LastRaw()
	default:
		s = p.Handler.LastPrint()
	}
	switch {
	case s == "":
		s = text.QueryBufferEmpty
	case p.Handler.IO().Interactive() && env.Get("SYNTAX_HL") == "true":
		b := new(bytes.Buffer)
		if p.Handler.Highlight(b, s) == nil {
			s = b.String()
		}
	}
	fmt.Fprintln(p.Handler.IO().Stdout(), s)
	return nil
}

// Write is a Query Buffer meta command (\w \write). Writes the query buffer to
// a file.
//
// Descs:
//
//	w	[-raw|-exec] FILE	write the contents of the query buffer, raw (non-interpolated) buffer, or exec buffer to file
//	write
func Write(p *Params) error {
	// get last statement
	s, buf := p.Handler.LastExec(), p.Handler.Buf()
	if buf.Len != 0 {
		s = buf.String()
	}
	name, err := p.Next(true)
	if err != nil {
		return err
	}
	return os.WriteFile(name, []byte(strings.TrimSuffix(s, "\n")+"\n"), 0o644)
}

// Reset is a Query Buffer meta command (\r, \reset). Clears (resets) the query
// buffer.
//
// Descs:
//
//	r	reset (clear) the query buffer
//	reset
func Reset(p *Params) error {
	p.Handler.Reset(nil)
	p.Handler.Print(text.QueryBufferReset)
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
	out := p.Handler.IO().Stdout()
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
		out, _, err = env.Pipe(p.Handler.IO().Stdout(), p.Handler.IO().Stderr(), pipe[1:])
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
	stdout, stderr := p.Handler.IO().Stdout, p.Handler.IO().Stderr
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

// Set is a Variables meta command (\set). Sets (or shows) the application variables.
//
// Descs:
//
//	set	[NAME [VALUE]]	set {{CommandName}} application variable, or show all {{CommandName}} application variables if no parameters
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
	return env.Vars().Dump(p.Handler.IO().Stdout())
}

// Unset is a Variables meta command (\unset). Unsets a application variable.
//
// Descs:
//
//	unset	NAME	unset (delete) {{CommandName}} application variable
func Unset(p *Params) error {
	n, err := p.Next(true)
	if err != nil {
		return err
	}
	return env.Vars().Unset(n)
}

// SetPrint is a Variables meta command (\pset, \a, \C, \f, \H, \t, \T, \x).
// Sets, toggles, or displays the application's print formatting variables.
//
// Descs:
//
//	pset	[NAME [VALUE]]	set table print formatting option, or show all print formatting options if no parameters
//	a		toggle between unaligned and aligned output mode	DEPRECATED
//	C	[TITLE]	set table title, or unset if none	DEPRECATED
//	f	[SEPARATOR]	show or set field separator for unaligned query output	DEPRECATED
//	H		toggle HTML output mode	DEPRECATED
//	T	[ATTRIBUTES]	set HTML <table> tag attributes, or unset if none	DEPRECATED
//	t	[on|off]	show only rows	DEPRECATED
//	x	[on|off|auto]	toggle expanded output	DEPRECATED
func SetPrint(p *Params) error {
	var ok bool
	var val string
	var err error
	switch p.Name {
	case "a", "H":
	default:
		if val, ok, err = p.NextOK(true); err != nil {
			return err
		}
	}
	// display variables
	if p.Name == "pset" && !ok {
		return env.Vars().DumpPrint(p.Handler.IO().Stdout())
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
	return nil
}

// SetConn is a Variables meta command (\cset). Sets a connection variable.
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
	return env.Vars().DumpConn(p.Handler.IO().Stdout())
}

// Prompt is a Variables meta command (\prompt). Prompts the user for input,
// setting a application variable to the user's response.
//
// Descs:
//
//	prompt	[-TYPE] VAR [PROMPT]	prompt user to set application variable
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

// Describe is a Informational meta command (\d and variants). Queries the open
// database connection for information about the database schema and writes the
// information to the output.
//
// Descs:
//
//	d[S+]	[NAME]	list tables, views, and sequences or describe table, view, sequence, or index
//	da[S+]	[PATTERN]	list aggregates
//	df[S+]	[PATTERN]	list functions
//	di[S+]	[PATTERN]	list indexes
//	dm[S+]	[PATTERN]	list materialized views
//	dn[S+]	[PATTERN]	list schemas
//	dp[S]	[PATTERN]	list table, view, and sequence access privileges
//	ds[S+]	[PATTERN]	list sequences
//	dt[S+]	[PATTERN]	list tables
//	dv[S+]	[PATTERN]	list views
//	l[+]	list databases
func Describe(p *Params) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	m, err := p.Handler.MetadataWriter(ctx)
	if err != nil {
		return err
	}
	verbose := strings.ContainsRune(p.Name, '+')
	showSystem := strings.ContainsRune(p.Name, 'S')
	name := strings.TrimRight(p.Name, "S+")
	pattern, err := p.Next(true)
	if err != nil {
		return err
	}
	switch name {
	case "d":
		if pattern != "" {
			return m.DescribeTableDetails(p.Handler.URL(), pattern, verbose, showSystem)
		}
		return m.ListTables(p.Handler.URL(), "tvmsE", pattern, verbose, showSystem)
	case "df", "da":
		return m.DescribeFunctions(p.Handler.URL(), name, pattern, verbose, showSystem)
	case "dt", "dtv", "dtm", "dts", "dv", "dm", "ds":
		return m.ListTables(p.Handler.URL(), name, pattern, verbose, showSystem)
	case "dn":
		return m.ListSchemas(p.Handler.URL(), pattern, verbose, showSystem)
	case "di":
		return m.ListIndexes(p.Handler.URL(), pattern, verbose, showSystem)
	case "l":
		return m.ListAllDbs(p.Handler.URL(), pattern, verbose)
	case "dp":
		return m.ListPrivilegeSummaries(p.Handler.URL(), pattern, showSystem)
	}
	return nil
}

// Stats is a Informational meta command (\ss and variants). Queries the open
// database connection for stats and writes it to the output.
//
// Descs:
//
//	ss[+]	[TABLE|QUERY] [k]	show stats for a table or a query
func Stats(p *Params) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	m, err := p.Handler.MetadataWriter(ctx)
	if err != nil {
		return err
	}
	verbose := strings.ContainsRune(p.Name, '+')
	name := strings.TrimRight(p.Name, "+")
	pattern, err := p.Next(true)
	if err != nil {
		return err
	}
	k := 0
	if verbose {
		k = 3
	}
	if name == "ss" {
		name = "sswnulhmkf"
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
	return m.ShowStats(p.Handler.URL(), name, pattern, verbose, k)
}

// Conditional is a Control/Conditional meta command (\if, \elif, \else,
// \endif). Starts, closes, and ends a conditional block within the
// application.
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
	case "elif":
	case "else":
	case "endif":
	}
	return nil
}

// Shell is a Operating System/Environment meta command (\!). Executes a
// command using the Operating System/Environment's shell.
//
// Descs:
//
//	!	[COMMAND]	execute command in shell or start interactive shell
func Shell(p *Params) error {
	return env.Shell(p.Raw())
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

// Getenv is a Operating System/Environment meta command (\getenv). Sets the
// application's variable value returned from the Operating
// System/Environment's variables.
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
	value, _ := env.Getenv(v)
	return env.Vars().Set(n, value)
}

// Setenv is a Operating System/Environment meta command (\setenv). Sets (or
// unsets) a Operating System/Environment variable. Environment variables set
// this way will be passed to any child processes.
//
// Descs:
//
//	setenv	NAME [VALUE]	set or unset environment variable
func Setenv(p *Params) error {
	n, err := p.Next(true)
	if err != nil {
		return err
	}
	v, err := p.Next(true)
	if err != nil {
		return err
	}
	return os.Setenv(n, v)
}
