package metacmd

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

// Cmd is a command implementation.
type Cmd struct {
	Section Section
	Name    string
	Desc    Desc
	Aliases map[string]Desc
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
			Name:    "?",
			Desc:    Desc{"show help on backslash commands", "[commands]"},
			Aliases: map[string]Desc{
				"?":  {"show help on " + text.CommandName + " command-line options", "options"},
				"? ": {"show help on special variables", "variables"},
			},
			Process: func(p *Params) error {
				name, err := p.Get(false)
				if err != nil {
					return err
				}
				switch name {
				default:
					Listing(p.Handler.IO().Stdout())
				case "commands":
					Listing(p.Handler.IO().Stdout())
				case "options":
					// FIXME: decouple
					kingpin.Usage()
				case "variables":
					env.Listing(p.Handler.IO().Stdout())
				}
				return nil
			},
		},
		Quit: {
			Section: SectionGeneral,
			Name:    "q",
			Desc:    Desc{"quit " + text.CommandName, ""},
			Aliases: map[string]Desc{"quit": {}},
			Process: func(p *Params) error {
				p.Option.Quit = true
				return nil
			},
		},
		Copyright: {
			Section: SectionGeneral,
			Name:    "copyright",
			Desc:    Desc{"show " + text.CommandName + " usage and distribution terms", ""},
			Process: func(p *Params) error {
				if typ := env.TermGraphics(); typ.Available() {
					typ.Encode(p.Handler.IO().Stdout(), text.Logo)
				}
				p.Handler.Print(text.Copyright)
				return nil
			},
		},
		ConnectionInfo: {
			Section: SectionConnection,
			Name:    "conninfo",
			Desc:    Desc{"display information about the current database connection", ""},
			Process: func(p *Params) error {
				if db, u := p.Handler.DB(), p.Handler.URL(); db != nil && u != nil {
					p.Handler.Print(text.ConnInfo, u.Driver, u.DSN)
				} else {
					p.Handler.Print(text.NotConnected)
				}
				return nil
			},
		},
		Drivers: {
			Section: SectionGeneral,
			Name:    "drivers",
			Desc:    Desc{"display information about available database drivers", ""},
			Process: func(p *Params) error {
				out := p.Handler.IO().Stdout()
				available := drivers.Available()
				names := make([]string, len(available))
				var z int
				for k := range available {
					names[z] = k
					z++
				}
				sort.Strings(names)
				fmt.Fprintln(out, text.AvailableDrivers)
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
					fmt.Fprintln(out, s)
				}
				return nil
			},
		},
		Connect: {
			Section: SectionConnection,
			Name:    "c",
			Desc:    Desc{"connect to database url", "DSN"},
			Aliases: map[string]Desc{
				"c":       {"connect to database with driver and parameters", "DRIVER PARAMS..."},
				"connect": {},
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
		Disconnect: {
			Section: SectionConnection,
			Name:    "Z",
			Desc:    Desc{"close database connection", ""},
			Aliases: map[string]Desc{"disconnect": {}},
			Process: func(p *Params) error {
				return p.Handler.Close()
			},
		},
		Password: {
			Section: SectionConnection,
			Name:    "password",
			Desc:    Desc{"change the password for a user", "[USERNAME]"},
			Aliases: map[string]Desc{"passwd": {}},
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
			Name:    "g",
			Desc:    Desc{"execute query (and send results to file or |pipe)", "[(OPTIONS)] [FILE] or ;"},
			Aliases: map[string]Desc{
				"gexec":        {"execute query and execute each value of the result", ""},
				"gset":         {"execute query and store results in " + text.CommandName + " variables", "[PREFIX]"},
				"gx":           {`as \g, but forces expanded output mode`, `[(OPTIONS)] [FILE]`},
				"G":            {`as \g, but forces vertical output mode`, `[(OPTIONS)] [FILE]`},
				"crosstabview": {"execute query and display results in crosstab", "[(OPTIONS)] [COLUMNS]"},
				"watch":        {"execute query every specified interval", "[(OPTIONS)] [DURATION]"},
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
		Edit: {
			Section: SectionQueryBuffer,
			Name:    "e",
			Desc:    Desc{"edit the query buffer (or file) with external editor", "[FILE] [LINE]"},
			Aliases: map[string]Desc{"edit": {}},
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
			Name:    "p",
			Desc:    Desc{"show the contents of the query buffer", ""},
			Aliases: map[string]Desc{
				"print": {},
				"raw":   {"show the raw (non-interpolated) contents of the query buffer", ""},
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
			Name:    "r",
			Desc:    Desc{"reset (clear) the query buffer", ""},
			Aliases: map[string]Desc{"reset": {}},
			Process: func(p *Params) error {
				p.Handler.Reset(nil)
				fmt.Fprintln(p.Handler.IO().Stdout(), text.QueryBufferReset)
				return nil
			},
		},
		Echo: {
			Section: SectionInputOutput,
			Name:    "echo",
			Desc:    Desc{"write string to standard output (-n for no newline)", "[-n] [STRING]"},
			Aliases: map[string]Desc{
				"qecho": {"write string to \\o output stream (-n for no newline)", "[-n] [STRING]"},
				"warn":  {"write string to standard error (-n for no newline)", "[-n] [STRING]"},
			},
			Process: func(p *Params) error {
				nl := "\n"
				var vals []string
				ok, n, err := p.GetOptional(true)
				if err != nil {
					return err
				}
				if ok && n == "n" {
					nl = ""
				} else if ok {
					vals = append(vals, "-"+n)
				} else {
					vals = append(vals, n)
				}
				v, err := p.GetAll(true)
				if err != nil {
					return err
				}
				out := io.Writer(p.Handler.IO().Stdout())
				if o := p.Handler.GetOutput(); p.Name == "qecho" && o != nil {
					out = o
				} else if p.Name == "warn" {
					out = p.Handler.IO().Stderr()
				}
				fmt.Fprint(out, strings.Join(append(vals, v...), " ")+nl)
				return nil
			},
		},
		Write: {
			Section: SectionQueryBuffer,
			Name:    "w",
			Desc:    Desc{"write query buffer to file", "FILE"},
			Aliases: map[string]Desc{"write": {}},
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
			Name:    "cd",
			Desc:    Desc{"change the current working directory", "[DIR]"},
			Process: func(p *Params) error {
				dir, err := p.Get(true)
				if err != nil {
					return err
				}
				return env.Chdir(p.Handler.User(), dir)
			},
		},
		SetEnv: {
			Section: SectionOperatingSystem,
			Name:    "setenv",
			Desc:    Desc{"set or unset environment variable", "NAME [VALUE]"},
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
			Name:    "timing",
			Desc:    Desc{"toggle timing of commands", "[on|off]"},
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
			Name:    "!",
			Desc:    Desc{"execute command in shell or start interactive shell", "[COMMAND]"},
			Process: func(p *Params) error {
				return env.Shell(p.GetRaw())
			},
		},
		Out: {
			Section: SectionInputOutput,
			Name:    "o",
			Desc:    Desc{"send all query results to file or |pipe", "[FILE]"},
			Aliases: map[string]Desc{"out": {}},
			Process: func(p *Params) error {
				if out := p.Handler.GetOutput(); out != nil {
					p.Handler.SetOutput(nil)
				}
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
					out, _, err = env.Pipe(pipe[1:])
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
			Name:    "i",
			Desc:    Desc{"execute commands from file", "FILE"},
			Aliases: map[string]Desc{
				"ir":               {`as \i, but relative to location of current script`, `FILE`},
				"include":          {},
				"include_relative": {},
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
			Name:    "begin",
			Desc:    Desc{"begin a transaction", ""},
			Aliases: map[string]Desc{
				"begin":    {"begin a transaction with isolation level", "[-read-only] [ISOLATION]"},
				"commit":   {"commit current transaction", ""},
				"rollback": {"rollback (abort) current transaction", ""},
				"abort":    {},
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
			Name:    "prompt",
			Desc:    Desc{"prompt user to set variable", "[-TYPE] <VAR> [PROMPT]"},
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
			Name:    "set",
			Desc:    Desc{"set internal variable, or list all if no parameters", "[NAME [VALUE]]"},
			Process: func(p *Params) error {
				ok, n, err := p.GetOK(true)
				if err != nil {
					return err
				}
				if !ok {
					vals := env.All()
					out := p.Handler.IO().Stdout()
					n := make([]string, len(vals))
					var i int
					for k := range vals {
						n[i] = k
						i++
					}
					sort.Strings(n)
					for _, k := range n {
						fmt.Fprintln(out, k, "=", "'"+vals[k]+"'")
					}
					return nil
				}
				vals, err := p.GetAll(true)
				if err != nil {
					return err
				}
				return env.Set(n, strings.Join(vals, ""))
			},
		},
		Unset: {
			Section: SectionVariables,
			Name:    "unset",
			Desc:    Desc{"unset (delete) internal variable", "NAME"},
			Process: func(p *Params) error {
				n, err := p.Get(true)
				if err != nil {
					return err
				}
				return env.Unset(n)
			},
		},
		SetFormatVar: {
			Section: SectionFormatting,
			Name:    "pset",
			Desc:    Desc{"set table output option", "[NAME [VALUE]]"},
			Aliases: map[string]Desc{
				"a": {"toggle between unaligned and aligned output mode", ""},
				"C": {"set table title, or unset if none", "[STRING]"},
				"f": {"show or set field separator for unaligned query output", "[STRING]"},
				"H": {"toggle HTML output mode", ""},
				"T": {"set HTML <table> tag attributes, or unset if none", "[STRING]"},
				"t": {"show only rows", "[on|off]"},
				"x": {"toggle expanded output", "[on|off|auto]"},
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
			Name:    "d[S+]",
			Desc:    Desc{"list tables, views, and sequences or describe table, view, sequence, or index", "[NAME]"},
			Aliases: map[string]Desc{
				"da[S+]": {"list aggregates", "[PATTERN]"},
				"df[S+]": {"list functions", "[PATTERN]"},
				"dm[S+]": {"list materialized views", "[PATTERN]"},
				"dv[S+]": {"list views", "[PATTERN]"},
				"ds[S+]": {"list sequences", "[PATTERN]"},
				"dn[S+]": {"list schemas", "[PATTERN]"},
				"dt[S+]": {"list tables", "[PATTERN]"},
				"di[S+]": {"list indexes", "[PATTERN]"},
				"dp[S]":  {"list table, view, and sequence access privileges", "[PATTERN]"},
				"l[+]":   {"list databases", ""},
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
			Name:    "ss[+]",
			Desc:    Desc{"show stats for a table or a query", "[TABLE|QUERY] [k]"},
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
			Name:    "copy",
			Desc:    Desc{"copy query from source url to table on destination url", "SRC DST QUERY TABLE"},
			Aliases: map[string]Desc{
				"copy": {"copy query from source url to columns of table on destination url", "SRC DST QUERY TABLE(A,...)"},
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
		name := c.Name
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
		for alias := range c.Aliases {
			if pos := strings.IndexRune(alias, '['); pos != -1 {
				mods := strings.TrimRight(alias[pos+1:], "]")
				alias = alias[:pos]
				cmdMap[alias+mods] = mc
				if len(mods) > 1 {
					for _, r := range mods {
						cmdMap[alias+string(r)] = mc
					}
				}
			}
			cmdMap[alias] = mc
		}
		sectMap[c.Section] = append(sectMap[c.Section], mc)
	}
}
