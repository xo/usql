package metacmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

// Cmd is a command implementation.
type Cmd struct {
	Section Section
	Name    string
	Desc    string
	Aliases map[string]string
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
			Desc:    "show help on backslash commands,[commands]",
			Aliases: map[string]string{
				"?":  "show help on " + text.CommandName + " command-line options,options",
				"? ": "show help on special variables,variables",
			},
			Process: func(p *Params) error {
				Listing(p.Handler.IO().Stdout())
				return nil
			},
		},
		Quit: {
			Section: SectionGeneral,
			Name:    "q",
			Desc:    "quit " + text.CommandName,
			Aliases: map[string]string{"quit": ""},
			Process: func(p *Params) error {
				p.Result.Quit = true
				return nil
			},
		},
		Copyright: {
			Section: SectionGeneral,
			Name:    "copyright",
			Desc:    "show " + text.CommandName + " usage and distribution terms",
			Process: func(p *Params) error {
				out := p.Handler.IO().Stdout()
				fmt.Fprintln(out, text.Copyright)
				fmt.Fprintln(out)
				return nil
			},
		},
		ConnectionInfo: {
			Section: SectionConnection,
			Name:    "conninfo",
			Desc:    "display information about the current database connection",
			Process: func(p *Params) error {
				out := p.Handler.IO().Stdout()
				if db, u := p.Handler.DB(), p.Handler.URL(); db != nil && u != nil {
					fmt.Fprintf(out, text.ConnInfo, u.Driver, u.DSN)
					fmt.Fprintln(out)
				} else {
					fmt.Fprintln(out, text.NotConnected)
				}
				return nil
			},
		},
		Drivers: {
			Section: SectionGeneral,
			Name:    "drivers",
			Desc:    "display information about available database drivers",
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
			Desc:    "connect to database with url,URL",
			Aliases: map[string]string{
				"c":       "connect to database with SQL driver and parameters,DRIVER PARAMS...",
				"connect": "",
			},
			Process: func(p *Params) error {
				vals, err := p.GetAll(true)
				if err != nil {
					return err
				}
				return p.Handler.Open(vals...)
			},
		},
		Disconnect: {
			Section: SectionConnection,
			Name:    "Z",
			Desc:    "close database connection",
			Aliases: map[string]string{"disconnect": ""},
			Process: func(p *Params) error {
				return p.Handler.Close()
			},
		},
		Password: {
			Section: SectionConnection,
			Name:    "password",
			Desc:    "change the password for a user,[USERNAME]",
			Aliases: map[string]string{"passwd": ""},
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
				/*fmt.Fprintf(p.Handler.IO().Stdout(), text.PasswordChangeSucceeded, user)
				fmt.Fprintln(p.Handler.IO().Stdout())*/
				return nil
			},
		},
		Exec: {
			Section: SectionGeneral,
			Name:    "g",
			Desc:    "execute query (and send results to file or |pipe),[FILE] or ;",
			Aliases: map[string]string{
				"gexec":        "execute query and execute each value of the result",
				"gset":         "execute query and store results in " + text.CommandName + " variables,[PREFIX]",
				"gx":           `as \g, but forces expanded output mode,`,
				"crosstabview": "execute query and display results in crosstab,[COLUMNS]",
			},
			Process: func(p *Params) error {
				p.Result.Exec = ExecOnly
				switch p.Name {
				case "g":
					params, err := p.GetAll(true)
					if err != nil {
						return err
					}
					p.Result.ParseExecParams(params, "pipe")
				case "gexec":
					p.Result.Exec = ExecExec
				case "gset":
					p.Result.Exec = ExecSet
					params, err := p.GetAll(true)
					if err != nil {
						return err
					}
					p.Result.ParseExecParams(params, "prefix")
				case "gx":
					params, err := p.GetAll(true)
					if err != nil {
						return err
					}
					p.Result.ParseExecParams(params, "pipe")
					p.Result.ExecParams["expanded"] = "on"
				case "crosstabview":
					p.Result.Exec = ExecCrosstab
					for i := 0; i < 4; i++ {
						ok, col, err := p.GetOK(true)
						if err != nil {
							return err
						}
						p.Result.Crosstab = append(p.Result.Crosstab, col)
						if !ok {
							break
						}
					}
				}
				return nil
			},
		},
		Edit: {
			Section: SectionQueryBuffer,
			Name:    "e",
			Desc:    "edit the query buffer (or file) with external editor,[FILE] [LINE]",
			Aliases: map[string]string{"edit": ""},
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
				buf.Reset(n)
				return nil
			},
		},
		Print: {
			Section: SectionQueryBuffer,
			Name:    "p",
			Desc:    "show the contents of the query buffer",
			Aliases: map[string]string{
				"print": "",
				"raw":   "show the raw (non-interpolated) contents of the query buffer",
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
			Desc:    "reset (clear) the query buffer",
			Aliases: map[string]string{"reset": ""},
			Process: func(p *Params) error {
				p.Handler.Reset(nil)
				fmt.Fprintln(p.Handler.IO().Stdout(), text.QueryBufferReset)
				return nil
			},
		},
		Echo: {
			Section: SectionInputOutput,
			Name:    "echo",
			Desc:    "write string to standard output,[STRING]",
			Process: func(p *Params) error {
				vals, err := p.GetAll(true)
				if err != nil {
					return err
				}
				fmt.Fprintln(p.Handler.IO().Stdout(), strings.Join(vals, " "))
				return nil
			},
		},
		Qecho: {
			Section: SectionInputOutput,
			Name:    "qecho",
			Desc:    "write string to \\o output stream,[STRING]",
			Process: func(p *Params) error {
				vals, err := p.GetAll(true)
				if err != nil {
					return err
				}
				var out io.Writer = p.Handler.GetOutput()
				if out == nil {
					out = p.Handler.IO().Stdout()
				}
				fmt.Fprintln(out, strings.Join(vals, " "))
				return nil
			},
		},
		Write: {
			Section: SectionQueryBuffer,
			Name:    "w",
			Desc:    "write query buffer to file,FILE",
			Aliases: map[string]string{"write": ""},
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
				return ioutil.WriteFile(file, []byte(strings.TrimSuffix(s, "\n")+"\n"), 0o644)
			},
		},
		ChangeDir: {
			Section: SectionOperatingSystem,
			Name:    "cd",
			Desc:    "change the current working directory,[DIR]",
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
			Desc:    "set or unset environment variable,NAME [VALUE]",
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
			Desc:    "toggle timing of commands,[on|off]",
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
				out := p.Handler.IO().Stdout()
				fmt.Fprintf(out, text.TimingSet, setting)
				fmt.Fprintln(out)
				return nil
			},
		},
		Shell: {
			Section: SectionOperatingSystem,
			Name:    "!",
			Desc:    "execute command in shell or start interactive shell,[COMMAND]",
			Process: func(p *Params) error {
				return env.Shell(p.GetRaw())
			},
		},
		Out: {
			Section: SectionInputOutput,
			Name:    "o",
			Desc:    "send all query results to file or |pipe,[FILE]",
			Aliases: map[string]string{"out": ""},
			Process: func(p *Params) error {
				if out := p.Handler.GetOutput(); out != nil {
					out.Close()
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
					out, err = env.Pipe(pipe[1:])
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
			Desc:    "execute commands from file,FILE",
			Aliases: map[string]string{
				"ir":               `as \i, but relative to location of current script,FILE`,
				"include":          "",
				"include_relative": "",
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
			Desc:    "begin a transaction",
			Aliases: map[string]string{
				"commit":   "commit current transaction",
				"rollback": "rollback (abort) current transaction",
			},
			Process: func(p *Params) error {
				var f func() error
				switch p.Name {
				case "begin":
					f = p.Handler.Begin
				case "commit":
					f = p.Handler.Commit
				case "rollback":
					f = p.Handler.Rollback
				}
				return f()
			},
		},
		Prompt: {
			Section: SectionVariables,
			Name:    "prompt",
			Desc:    "prompt user to set variable,[-TYPE] <VAR> [PROMPT]",
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
			Desc:    "set internal variable, or list all if no parameters,[NAME [VALUE]]",
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
			Desc:    "unset (delete) internal variable,NAME",
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
			Desc:    "set table output option,[NAME [VALUE]]",
			Aliases: map[string]string{
				"a": "toggle between unaligned and aligned output mode",
				"C": "set table title, or unset if none,[STRING]",
				"f": "show or set field separator for unaligned query output,[STRING]",
				"H": "toggle HTML output mode",
				"T": "set HTML <table> tag attributes, or unset if none,[STRING]",
				"t": "show only rows,[on|off]",
				"x": "toggle expanded output,[on|off|auto]",
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
				out := p.Handler.IO().Stdout()
				switch {
				case strings.Contains(mask, "%d"):
					i, _ := strconv.Atoi(val)
					fmt.Fprintf(out, mask, i)
				case unsetMask != "" && val == "":
					fmt.Fprint(out, unsetMask)
				case !strings.Contains(mask, "%"):
					fmt.Fprint(out, mask)
				default:
					fmt.Fprintf(out, mask, val)
				}
				fmt.Fprintln(out)
				return nil
			},
		},
		Describe: {
			Section: SectionInformational,
			Name:    "d[S+]",
			Desc:    "list tables, views, and sequences or describe table, view, sequence, or index,[NAME]",
			Aliases: map[string]string{
				"da[S+]": "list aggregates,[PATTERN]",
				"df[S+]": "list functions,[PATTERN]",
				"dm[S+]": "list materialized views,[PATTERN]",
				"dv[S+]": "list views,[PATTERN]",
				"ds[S+]": "list sequences,[PATTERN]",
				"dn[S+]": "list schemas,[PATTERN]",
				"dt[S+]": "list tables,[PATTERN]",
				"di[S+]": "list indexes,[PATTERN]",
				"l[+]":   "list databases",
			},
			Process: func(p *Params) error {
				opts := p.Handler.ReaderOptions()
				m, err := drivers.NewMetadataWriter(p.Handler.URL(), p.Handler.DB(), p.Handler.IO().Stdout(), opts...)
				if err != nil {
					return err
				}
				verbose := strings.ContainsRune(p.Name, '+')
				showSystem := strings.ContainsRune(p.Name, 'S')
				name := strings.TrimRight(p.Name, "S+")
				pattern, err := p.Get(false)
				if err != nil {
					return err
				}
				switch name {
				case "d":
					if pattern != "" {
						return m.DescribeTableDetails(pattern, verbose, showSystem)
					}
					return m.ListTables("tvmsE", pattern, verbose, showSystem)
				case "df", "da":
					return m.DescribeFunctions(name, pattern, verbose, showSystem)
				case "dt", "dtv", "dtm", "dts", "dv", "dm", "ds":
					return m.ListTables(name, pattern, verbose, showSystem)
				case "dn":
					return m.ListSchemas(pattern, verbose, showSystem)
				case "di":
					return m.ListIndexes(pattern, verbose, showSystem)
				case "l":
					return m.ListAllDbs(pattern, verbose)
				}
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
