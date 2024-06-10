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
	"golang.org/x/exp/maps"
)

// Cmd is a command implementation.
type Cmd struct {
	Section Section
	Desc    Desc
	Aliases []Desc
	Process func(*Params) error
}

// cmds is the set of commands.
var cmds []Cmd

// cmdMap is the map of commands and their aliases.
var cmdMap map[string]Metacmd

// sectMap is the map of sections to its respective commands.
var sectMap map[Section][]Metacmd

func init() {
	cmds = []Cmd{
		Question: {
			Section: SectionHelp,
			Desc:    Desc{"?", "[commands]", "show help on backslash commands"},
			Aliases: []Desc{
				{"?", "options", "show help on " + text.CommandName + " command-line options"},
				{"?", "variables", "show help on special " + text.CommandName + " variables"},
			},
			Process: func(p *Params) error {
				name, err := p.Get(false)
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
					Usage(stdout, true)
				case name == "variables":
					env.Listing(stdout)
				default:
					Listing(stdout)
				}
				if cmd != nil {
					if err := wc.Close(); err != nil {
						return err
					}
					return cmd.Wait()
				}
				return nil
			},
		},
		Quit: {
			Section: SectionGeneral,
			Desc:    Desc{"q", "", "quit " + text.CommandName},
			Aliases: []Desc{{"quit", "", ""}},
			Process: func(p *Params) error {
				p.Option.Quit = true
				return nil
			},
		},
		Copyright: {
			Section: SectionGeneral,
			Desc:    Desc{"copyright", "", "show " + text.CommandName + " usage and distribution terms"},
			Process: func(p *Params) error {
				stdout := p.Handler.IO().Stdout()
				if typ := env.TermGraphics(); typ.Available() {
					typ.Encode(stdout, text.Logo)
				}
				fmt.Fprintln(stdout, text.Copyright)
				return nil
			},
		},
		ConnectionInfo: {
			Section: SectionConnection,
			Desc:    Desc{"conninfo", "", "display information about the current database connection"},
			Process: func(p *Params) error {
				s := text.NotConnected
				if db, u := p.Handler.DB(), p.Handler.URL(); db != nil && u != nil {
					s = fmt.Sprintf(text.ConnInfo, u.Driver, u.DSN)
				}
				fmt.Fprintln(p.Handler.IO().Stdout(), s)
				return nil
			},
		},
		Drivers: {
			Section: SectionGeneral,
			Desc:    Desc{"drivers", "", "display information about available database drivers"},
			Process: func(p *Params) error {
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
						if len(aliases) > 0 {
							s += " [" + strings.Join(aliases, ", ") + "]"
						}
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
			},
		},
		Connect: {
			Section: SectionConnection,
			Desc:    Desc{"c", "DSN", "connect to database url"},
			Aliases: []Desc{
				{"c", "DRIVER PARAMS...", "connect to database with driver and parameters"},
				{"connect", "", ""},
			},
			Process: func(p *Params) error {
				vals, err := p.GetAll(true)
				if err != nil {
					return err
				}
				ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
				defer cancel()
				return p.Handler.Open(ctx, vals...)
			},
		},
		SetConnVar: {
			Section: SectionConnection,
			Desc:    Desc{"cset", "[NAME [DSN]]", "set named connection, or list all if no parameters"},
			Aliases: []Desc{
				{"cset", "NAME DRIVER PARAMS...", "define named connection for database driver"},
			},
			Process: func(p *Params) error {
				ok, n, err := p.GetOK(true)
				switch {
				case err != nil:
					return err
				case ok:
					vals, err := p.GetAll(true)
					if err != nil {
						return err
					}
					return env.Cset(n, vals...)
				}
				vals := env.Call()
				keys := maps.Keys(vals)
				sort.Strings(keys)
				out := p.Handler.IO().Stdout()
				for _, k := range keys {
					fmt.Fprintln(out, k, "=", "'"+strings.Join(vals[k], " ")+"'")
				}
				return nil
			},
		},
		Disconnect: {
			Section: SectionConnection,
			Desc:    Desc{"Z", "", "close database connection"},
			Aliases: []Desc{{"disconnect", "", ""}},
			Process: func(p *Params) error {
				return p.Handler.Close()
			},
		},
		Password: {
			Section: SectionConnection,
			Desc:    Desc{"password", "[USERNAME]", "change the password for a user"},
			Aliases: []Desc{{"passwd", "", ""}},
			Process: func(p *Params) error {
				username, err := p.Get(true)
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
			},
		},
		Exec: {
			Section: SectionQueryExecute,
			Desc:    Desc{"g", "[(OPTIONS)] [FILE] or ;", "execute query (and send results to file or |pipe)"},
			Aliases: []Desc{
				{"G", "[(OPTIONS)] [FILE]", "as \\g, but forces vertical output mode"},
				{"gx", "[(OPTIONS)] [FILE]", "as \\g, but forces expanded output mode"},
				{"gexec", "", "execute query and execute each value of the result"},
				{"gset", "[PREFIX]", "execute query and store results in " + text.CommandName + " variables"},
				{"crosstabview", "[(OPTIONS)] [COLUMNS]", "execute query and display results in crosstab"},
				{"watch", "[(OPTIONS)] [DURATION]", "execute query every specified interval"},
			},
			Process: func(p *Params) error {
				p.Option.Exec = ExecOnly
				switch p.Name {
				case "g":
					params, err := p.GetAll(true)
					if err != nil {
						return err
					}
					p.Option.ParseParams(params, "pipe")
				case "gexec":
					p.Option.Exec = ExecExec
				case "gset":
					p.Option.Exec = ExecSet
					params, err := p.GetAll(true)
					if err != nil {
						return err
					}
					p.Option.ParseParams(params, "prefix")
				case "G":
					params, err := p.GetAll(true)
					if err != nil {
						return err
					}
					p.Option.ParseParams(params, "pipe")
					p.Option.Params["format"] = "vertical"
				case "gx":
					params, err := p.GetAll(true)
					if err != nil {
						return err
					}
					p.Option.ParseParams(params, "pipe")
					p.Option.Params["expanded"] = "on"
				case "crosstabview":
					p.Option.Exec = ExecCrosstab
					for i := 0; i < 4; i++ {
						ok, col, err := p.GetOK(true)
						if err != nil {
							return err
						}
						p.Option.Crosstab = append(p.Option.Crosstab, col)
						if !ok {
							break
						}
					}
				case "watch":
					p.Option.Exec = ExecWatch
					p.Option.Watch = 2 * time.Second
					ok, s, err := p.GetOK(true)
					switch {
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
				}
				return nil
			},
		},
		Bind: {
			Section: SectionQueryExecute,
			Desc:    Desc{"bind", "[PARAM]...", "set query parameters"},
			Process: func(p *Params) error {
				bind, err := p.GetAll(true)
				if err != nil {
					return err
				}
				var v []interface{}
				if n := len(bind); n != 0 {
					v = make([]interface{}, len(bind))
					for i := 0; i < n; i++ {
						v[i] = bind[i]
					}
				}
				p.Handler.Bind(v)
				return nil
			},
		},
		Edit: {
			Section: SectionQueryBuffer,
			Desc:    Desc{"e", "[FILE] [LINE]", "edit the query buffer (or file) with external editor"},
			Aliases: []Desc{{"edit", "", ""}},
			Process: func(p *Params) error {
				// get last statement
				s, buf := p.Handler.Last(), p.Handler.Buf()
				if buf.Len != 0 {
					s = buf.String()
				}
				path, err := p.Get(true)
				if err != nil {
					return err
				}
				line, err := p.Get(true)
				if err != nil {
					return err
				}
				// reset if no error
				n, err := env.EditFile(p.Handler.User(), path, line, s)
				if err != nil {
					return err
				}
				// save edited buffer to history
				p.Handler.IO().Save(string(n))
				buf.Reset(n)
				return nil
			},
		},
		Print: {
			Section: SectionQueryBuffer,
			Desc:    Desc{"p", "", "show the contents of the query buffer"},
			Aliases: []Desc{
				{"print", "", ""},
				{"raw", "", "show the raw (non-interpolated) contents of the query buffer"},
			},
			Process: func(p *Params) error {
				// get last statement
				var s string
				if p.Name == "raw" {
					s = p.Handler.LastRaw()
				} else {
					s = p.Handler.Last()
				}
				// use current statement buf if not empty
				buf := p.Handler.Buf()
				switch {
				case buf.Len != 0 && p.Name == "raw":
					s = buf.RawString()
				case buf.Len != 0:
					s = buf.String()
				}
				switch {
				case s == "":
					s = text.QueryBufferEmpty
				case p.Handler.IO().Interactive() && env.All()["SYNTAX_HL"] == "true":
					b := new(bytes.Buffer)
					if p.Handler.Highlight(b, s) == nil {
						s = b.String()
					}
				}
				fmt.Fprintln(p.Handler.IO().Stdout(), s)
				return nil
			},
		},
		Reset: {
			Section: SectionQueryBuffer,
			Desc:    Desc{"r", "", "reset (clear) the query buffer"},
			Aliases: []Desc{{"reset", "", ""}},
			Process: func(p *Params) error {
				p.Handler.Reset(nil)
				p.Handler.Print(text.QueryBufferReset)
				return nil
			},
		},
		Echo: {
			Section: SectionInputOutput,
			Desc:    Desc{"echo", "[-n] [STRING]", "write string to standard output (-n for no newline)"},
			Aliases: []Desc{
				{"qecho", "[-n] [STRING]", "write string to \\o output stream (-n for no newline)"},
				{"warn", "[-n] [STRING]", "write string to standard error (-n for no newline)"},
			},
			Process: func(p *Params) error {
				ok, n, err := p.GetOptional(true)
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
				v, err := p.GetAll(true)
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
			},
		},
		Write: {
			Section: SectionQueryBuffer,
			Desc:    Desc{"w", "FILE", "write query buffer to file"},
			Aliases: []Desc{{"write", "", ""}},
			Process: func(p *Params) error {
				// get last statement
				s, buf := p.Handler.Last(), p.Handler.Buf()
				if buf.Len != 0 {
					s = buf.String()
				}
				file, err := p.Get(true)
				if err != nil {
					return err
				}
				return os.WriteFile(file, []byte(strings.TrimSuffix(s, "\n")+"\n"), 0o644)
			},
		},
		ChangeDir: {
			Section: SectionOperatingSystem,
			Desc:    Desc{"cd", "[DIR]", "change the current working directory"},
			Process: func(p *Params) error {
				dir, err := p.Get(true)
				if err != nil {
					return err
				}
				return env.Chdir(p.Handler.User(), dir)
			},
		},
		GetEnv: {
			Section: SectionOperatingSystem,
			Desc:    Desc{"getenv", "VARNAME ENVVAR", "fetch environment variable"},
			Process: func(p *Params) error {
				n, err := p.Get(true)
				switch {
				case err != nil:
					return err
				case n == "":
					return text.ErrMissingRequiredArgument
				}
				v, err := p.Get(true)
				switch {
				case err != nil:
					return err
				case v == "":
					return text.ErrMissingRequiredArgument
				}
				value, _ := env.Getenv(v)
				return env.Set(n, value)
			},
		},
		SetEnv: {
			Section: SectionOperatingSystem,
			Desc:    Desc{"setenv", "NAME [VALUE]", "set or unset environment variable"},
			Process: func(p *Params) error {
				n, err := p.Get(true)
				if err != nil {
					return err
				}
				v, err := p.Get(true)
				if err != nil {
					return err
				}
				return os.Setenv(n, v)
			},
		},
		Timing: {
			Section: SectionOperatingSystem,
			Desc:    Desc{"timing", "[on|off]", "toggle timing of commands"},
			Process: func(p *Params) error {
				v, err := p.Get(true)
				if err != nil {
					return err
				}
				if v == "" {
					p.Handler.SetTiming(!p.Handler.GetTiming())
				} else {
					s, err := env.ParseBool(v, "\\timing")
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
			},
		},
		Shell: {
			Section: SectionOperatingSystem,
			Desc:    Desc{"!", "[COMMAND]", "execute command in shell or start interactive shell"},
			Process: func(p *Params) error {
				return env.Shell(p.GetRaw())
			},
		},
		Out: {
			Section: SectionInputOutput,
			Desc:    Desc{"o", "[FILE]", "send all query results to file or |pipe"},
			Aliases: []Desc{{"out", "", ""}},
			Process: func(p *Params) error {
				p.Handler.SetOutput(nil)
				params, err := p.GetAll(true)
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
			},
		},
		Include: {
			Section: SectionInputOutput,
			Desc:    Desc{"i", "FILE", "execute commands from file"},
			Aliases: []Desc{
				{"ir", "FILE", "as \\i, but relative to location of current script"},
				{"include", "", ""},
				{"include_relative", "", ""},
			},
			Process: func(p *Params) error {
				path, err := p.Get(true)
				if err != nil {
					return err
				}
				relative := p.Name == "ir" || p.Name == "include_relative"
				if err := p.Handler.Include(path, relative); err != nil {
					return fmt.Errorf("%s: %v", path, err)
				}
				return nil
			},
		},
		Transact: {
			Section: SectionTransaction,
			Desc:    Desc{"begin", "", "begin a transaction"},
			Aliases: []Desc{
				{"begin", "[-read-only] [ISOLATION]", "begin a transaction with isolation level"},
				{"commit", "", "commit current transaction"},
				{"rollback", "", "rollback (abort) current transaction"},
				{"abort", "", ""},
			},
			Process: func(p *Params) error {
				switch p.Name {
				case "commit":
					return p.Handler.Commit()
				case "rollback", "abort":
					return p.Handler.Rollback()
				}
				// read begin params
				readOnly := false
				ok, n, err := p.GetOptional(true)
				if ok {
					if n != "read-only" {
						return fmt.Errorf(text.InvalidOption, n)
					}
					readOnly = true
					if n, err = p.Get(true); err != nil {
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
			},
		},
		Prompt: {
			Section: SectionVariables,
			Desc:    Desc{"prompt", "[-TYPE] VAR [PROMPT]", "prompt user to set variable"},
			Process: func(p *Params) error {
				typ := "string"
				ok, n, err := p.GetOptional(true)
				if err != nil {
					return err
				}
				if ok {
					typ = n
					n, err = p.Get(true)
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
				vals, err := p.GetAll(true)
				if err != nil {
					return err
				}
				v, err := p.Handler.ReadVar(typ, strings.Join(vals, " "))
				if err != nil {
					return err
				}
				return env.Set(n, v)
			},
		},
		SetVar: {
			Section: SectionVariables,
			Desc:    Desc{"set", "[NAME [VALUE]]", "set internal variable, or list all if no parameters"},
			Process: func(p *Params) error {
				ok, n, err := p.GetOK(true)
				switch {
				case err != nil:
					return err
				case ok:
					vals, err := p.GetAll(true)
					if err != nil {
						return err
					}
					return env.Set(n, strings.Join(vals, " "))
				}
				vals := env.All()
				keys := maps.Keys(vals)
				sort.Strings(keys)
				out := p.Handler.IO().Stdout()
				for _, k := range keys {
					fmt.Fprintln(out, k, "=", "'"+vals[k]+"'")
				}
				return nil
			},
		},
		Unset: {
			Section: SectionVariables,
			Desc:    Desc{"unset", "NAME", "unset (delete) internal variable"},
			Process: func(p *Params) error {
				n, err := p.Get(true)
				if err != nil {
					return err
				}
				return env.Unset(n)
			},
		},
		SetPrintVar: {
			Section: SectionFormatting,
			Desc:    Desc{"pset", "[NAME [VALUE]]", "set table output option"},
			Aliases: []Desc{
				{"a", "", "toggle between unaligned and aligned output mode"},
				{"C", "[STRING]", "set table title, or unset if none"},
				{"f", "[STRING]", "show or set field separator for unaligned query output"},
				{"H", "", "toggle HTML output mode"},
				{"T", "[STRING]", "set HTML <table> tag attributes, or unset if none"},
				{"t", "[on|off]", "show only rows"},
				{"x", "[on|off|auto]", "toggle expanded output"},
			},
			Process: func(p *Params) error {
				var ok bool
				var val string
				var err error
				switch p.Name {
				case "a", "H":
				default:
					ok, val, err = p.GetOK(true)
					if err != nil {
						return err
					}
				}
				// display variables
				if p.Name == "pset" && !ok {
					return env.Pwrite(p.Handler.IO().Stdout())
				}
				var field, extra string
				switch p.Name {
				case "pset":
					field = val
					ok, val, err = p.GetOK(true)
					if err != nil {
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
					if val, err = env.Ptoggle(field, extra); err != nil {
						return err
					}
				} else {
					if val, err = env.Pset(field, val); err != nil {
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
						if tfmt := env.GoTime(); tfmt != val {
							val = fmt.Sprintf("%s (%q)", val, tfmt)
						}
					}
					p.Handler.Print(mask, val)
				}
				return nil
			},
		},
		Describe: {
			Section: SectionInformational,
			Desc:    Desc{"d[S+]", "[NAME]", "list tables, views, and sequences or describe table, view, sequence, or index"},
			Aliases: []Desc{
				{"da[S+]", "[PATTERN]", "list aggregates"},
				{"df[S+]", "[PATTERN]", "list functions"},
				{"di[S+]", "[PATTERN]", "list indexes"},
				{"dm[S+]", "[PATTERN]", "list materialized views"},
				{"dn[S+]", "[PATTERN]", "list schemas"},
				{"dp[S]", "[PATTERN]", "list table, view, and sequence access privileges"},
				{"ds[S+]", "[PATTERN]", "list sequences"},
				{"dt[S+]", "[PATTERN]", "list tables"},
				{"dv[S+]", "[PATTERN]", "list views"},
				{"l[+]", "", "list databases"},
			},
			Process: func(p *Params) error {
				ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
				defer cancel()
				m, err := p.Handler.MetadataWriter(ctx)
				if err != nil {
					return err
				}
				verbose := strings.ContainsRune(p.Name, '+')
				showSystem := strings.ContainsRune(p.Name, 'S')
				name := strings.TrimRight(p.Name, "S+")
				pattern, err := p.Get(true)
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
			},
		},
		Stats: {
			Section: SectionInformational,
			Desc:    Desc{"ss[+]", "[TABLE|QUERY] [k]", "show stats for a table or a query"},
			Process: func(p *Params) error {
				ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
				defer cancel()
				m, err := p.Handler.MetadataWriter(ctx)
				if err != nil {
					return err
				}
				verbose := strings.ContainsRune(p.Name, '+')
				name := strings.TrimRight(p.Name, "+")
				pattern, err := p.Get(true)
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
				ok, val, err := p.GetOK(true)
				if err != nil {
					return err
				}
				if ok {
					verbose = true
					k, err = strconv.Atoi(val)
					if err != nil {
						return err
					}
				}
				return m.ShowStats(p.Handler.URL(), name, pattern, verbose, k)
			},
		},
		Copy: {
			Section: SectionInputOutput,
			Desc:    Desc{"copy", "SRC DST QUERY TABLE", "copy query from source url to table on destination url"},
			Aliases: []Desc{
				{"copy", "SRC DST QUERY TABLE(A,...)", "copy query from source url to columns of table on destination url"},
			},
			Process: func(p *Params) error {
				ctx := context.Background()
				stdout, stderr := p.Handler.IO().Stdout, p.Handler.IO().Stderr
				srcDsn, err := p.Get(true)
				if err != nil {
					return err
				}
				srcURL, err := dburl.Parse(srcDsn)
				if err != nil {
					return err
				}
				destDsn, err := p.Get(true)
				if err != nil {
					return err
				}
				destURL, err := dburl.Parse(destDsn)
				if err != nil {
					return err
				}
				query, err := p.Get(true)
				if err != nil {
					return err
				}
				table, err := p.Get(true)
				if err != nil {
					return err
				}
				src, err := drivers.Open(ctx, srcURL, stdout, stderr)
				if err != nil {
					return err
				}
				defer src.Close()
				dest, err := drivers.Open(ctx, destURL, stdout, stderr)
				if err != nil {
					return err
				}
				defer dest.Close()
				ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
				defer cancel()
				// get the result set
				r, err := src.QueryContext(ctx, query)
				if err != nil {
					return err
				}
				defer r.Close()
				n, err := drivers.Copy(ctx, destURL, stdout, stderr, r, table)
				if err != nil {
					return err
				}
				p.Handler.Print("COPY %d", n)
				return nil
			},
		},
	}
	// set up map
	cmdMap = make(map[string]Metacmd, len(cmds))
	sectMap = make(map[Section][]Metacmd, len(SectionOrder))
	for i, c := range cmds {
		mc := Metacmd(i)
		if mc == None {
			continue
		}
		name := c.Desc.Name
		if pos := strings.IndexRune(name, '['); pos != -1 {
			mods := strings.TrimRight(name[pos+1:], "]")
			name = name[:pos]
			cmdMap[name+mods] = mc
			if len(mods) > 1 {
				for _, r := range mods {
					cmdMap[name+string(r)] = mc
				}
			}
		}
		cmdMap[name] = mc
		for _, d := range c.Aliases {
			if pos := strings.IndexRune(d.Name, '['); pos != -1 {
				mods := strings.TrimRight(d.Name[pos+1:], "]")
				d.Name = d.Name[:pos]
				cmdMap[d.Name+mods] = mc
				if len(mods) > 1 {
					for _, r := range mods {
						cmdMap[d.Name+string(r)] = mc
					}
				}
			}
			cmdMap[d.Name] = mc
		}
		sectMap[c.Section] = append(sectMap[c.Section], mc)
	}
}

// Usage is used by the [Question] command to display command line options.
var Usage = func(io.Writer, bool) {
}
