package metacmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
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
	Min     int
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
				Listing(p.H.IO().Stdout())
				return nil
			},
		},

		Quit: {
			Section: SectionGeneral,
			Name:    "q",
			Desc:    "quit " + text.CommandName,
			Aliases: map[string]string{"quit": ""},
			Process: func(p *Params) error {
				p.R.Quit = true
				return nil
			},
		},

		Copyright: {
			Section: SectionGeneral,
			Name:    "copyright",
			Desc:    "show " + text.CommandName + " usage and distribution terms",
			Process: func(p *Params) error {
				out := p.H.IO().Stdout()
				fmt.Fprintln(out, text.Copyright)
				fmt.Fprintln(out)
				return nil
			},
		},

		ConnInfo: {
			Section: SectionConnection,
			Name:    "conninfo",
			Desc:    "display information about the current database connection",
			Process: func(p *Params) error {
				out := p.H.IO().Stdout()
				if db, u := p.H.DB(), p.H.URL(); db != nil && u != nil {
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
				out := p.H.IO().Stdout()

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
			Min: 1,
			Process: func(p *Params) error {
				return p.H.Open(p.A()...)
			},
		},

		Disconnect: {
			Section: SectionConnection,
			Name:    "Z",
			Desc:    "close database connection",
			Aliases: map[string]string{"disconnect": ""},
			Process: func(p *Params) error {
				return p.H.Close()
			},
		},

		Password: {
			Section: SectionConnection,
			Name:    "password",
			Desc:    "change the password for a user,[USERNAME]",
			Aliases: map[string]string{"passwd": ""},
			Process: func(p *Params) error {
				user, err := p.H.ChangePassword(p.G())
				switch {
				case err == text.ErrPasswordNotSupportedByDriver || err == text.ErrNotConnected:
					return err
				case err != nil:
					return fmt.Errorf(text.PasswordChangeFailed, user, err)
				}

				/*fmt.Fprintf(p.H.IO().Stdout(), text.PasswordChangeSucceeded, user)
				fmt.Fprintln(p.H.IO().Stdout())*/

				return nil
			},
		},

		Exec: {
			Section: SectionGeneral,
			Name:    "g",
			Desc:    "execute query (and send results to file or |pipe),[FILE] or ;",
			Aliases: map[string]string{
				"gexec": "execute query and execute each value of the result",
				"gset":  "execute query and store results in " + text.CommandName + " variables,[PREFIX]",
			},
			Process: func(p *Params) error {
				p.R.Exec = ExecOnly

				switch p.N {
				case "g":
					p.R.ExecParam = p.G()

				case "gexec":
					p.R.Exec = ExecExec

				case "gset":
					p.R.Exec, p.R.ExecParam = ExecSet, p.G()
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
				s, buf := p.H.Last(), p.H.Buf()
				if buf.Len != 0 {
					s = buf.String()
				}

				// reset if no error
				n, err := env.EditFile(p.H.User(), p.G(), p.G(), s)
				if err == nil {
					buf.Reset(n)
				}

				return err
			},
		},

		Print: {
			Section: SectionQueryBuffer,
			Name:    "p",
			Desc:    "show the contents of the query buffer",
			Aliases: map[string]string{"print": ""},
			Process: func(p *Params) error {
				// get last statement
				s, buf := p.H.Last(), p.H.Buf()
				if buf.Len != 0 {
					s = buf.String()
				}

				if s == "" {
					s = text.QueryBufferEmpty
				} else if p.H.IO().Interactive() && env.All()["SYNTAX_HL"] == "true" {
					b := new(bytes.Buffer)
					if p.H.Highlight(b, s) == nil {
						s = b.String()
					}
				}

				fmt.Fprintln(p.H.IO().Stdout(), s)
				return nil
			},
		},

		Reset: {
			Section: SectionQueryBuffer,
			Name:    "r",
			Desc:    "reset (clear) the query buffer",
			Aliases: map[string]string{"reset": ""},
			Process: func(p *Params) error {
				p.H.Reset(nil)
				fmt.Fprintln(p.H.IO().Stdout(), text.QueryBufferReset)
				return nil
			},
		},

		Echo: {
			Section: SectionInputOutput,
			Name:    "echo",
			Desc:    "write string to standard output,[STRING]",
			Process: func(p *Params) error {
				fmt.Fprintln(p.H.IO().Stdout(), strings.Join(p.A(), " "))
				return nil
			},
		},

		Write: {
			Section: SectionQueryBuffer,
			Name:    "w",
			Min:     1,
			Desc:    "write query buffer to file,FILE",
			Aliases: map[string]string{"write": ""},
			Process: func(p *Params) error {
				// get last statement
				s, buf := p.H.Last(), p.H.Buf()
				if buf.Len != 0 {
					s = buf.String()
				}

				return ioutil.WriteFile(p.G(), []byte(strings.TrimSuffix(s, "\n")+"\n"), 0644)
			},
		},

		ChangeDir: {
			Section: SectionOperatingSystem,
			Name:    "cd",
			Desc:    "change the current working directory,[DIR]",
			Process: func(p *Params) error {
				return env.Chdir(p.H.User(), p.G())
			},
		},

		SetEnv: {
			Section: SectionOperatingSystem,
			Name:    "setenv",
			Min:     1,
			Desc:    "set or unset environment variable,NAME [VALUE]",
			Process: func(p *Params) error {
				var err error

				n := p.G()
				if len(p.P) == 1 {
					err = os.Unsetenv(n)
				} else {
					err = os.Setenv(n, strings.Join(p.A(), ""))
				}

				return err
			},
		},

		ShellExec: {
			Section: SectionOperatingSystem,
			Name:    "!",
			Desc:    "execute command in shell or start interactive shell,[COMMAND]",
			Process: func(p *Params) error {
				if len(p.P) == 0 && !p.H.IO().Interactive() {
					return text.ErrNotInteractive
				}

				p.R.Processed = len(p.P)
				v, err := env.Exec(strings.TrimSpace(strings.Join(p.P, " ")))
				if err == nil && v != "" {
					fmt.Fprintln(p.H.IO().Stdout(), v)
				}

				return nil
			},
		},

		Include: {
			Section: SectionInputOutput,
			Name:    "i",
			Min:     1,
			Desc:    "execute commands from file,FILE",
			Aliases: map[string]string{
				"ir":               `as \i, but relative to location of current script,FILE`,
				"include":          "",
				"include_relative": "",
			},
			Process: func(p *Params) error {
				path := p.G()
				err := p.H.Include(path, p.N == "ir" || p.N == "include_relative")
				if err != nil {
					err = fmt.Errorf("%s: %v", path, err)
				}
				return err
			},
		},

		Begin: {
			Section: SectionTransaction,
			Name:    "begin",
			Desc:    "begin a transaction",
			Process: func(p *Params) error {
				return p.H.Begin()
			},
		},

		Commit: {
			Section: SectionTransaction,
			Name:    "commit",
			Desc:    "commit current transaction",
			Process: func(p *Params) error {
				return p.H.Commit()
			},
		},

		Rollback: {
			Section: SectionTransaction,
			Name:    "rollback",
			Desc:    "rollback (abort) current transaction",
			Aliases: map[string]string{"abort": ""},
			Process: func(p *Params) error {
				return p.H.Rollback()
			},
		},

		Prompt: {
			Section: SectionVariables,
			Name:    "prompt",
			Min:     1,
			Desc:    "prompt user to set internal variable,[TEXT] NAME",
			Process: func(p *Params) error {
				typ, n := p.V("string"), p.G()
				if n == "" {
					return text.ErrMissingRequiredArgument
				}

				err := env.ValidIdentifier(n)
				if err != nil {
					return err
				}

				v, err := p.H.ReadVar(typ, strings.Join(p.A(), " "))
				if err != nil {
					return err
				}

				return env.Set(n, v)
			},
		},

		Set: {
			Section: SectionVariables,
			Name:    "set",
			Desc:    "set internal variable, or list all if no parameters,[NAME [VALUE]]",
			Process: func(p *Params) error {
				if len(p.P) == 0 {
					vars := env.All()
					out := p.H.IO().Stdout()
					n := make([]string, len(vars))
					var i int
					for k := range vars {
						n[i] = k
						i++
					}
					sort.Strings(n)

					for _, k := range n {
						fmt.Fprintln(out, k, "=", "'"+vars[k]+"'")
					}
					return nil
				}

				return env.Set(p.G(), strings.Join(p.A(), ""))
			},
		},

		Unset: {
			Section: SectionVariables,
			Name:    "unset",
			Min:     1,
			Desc:    "unset (delete) internal variable,NAME",
			Process: func(p *Params) error {
				return env.Unset(p.G())
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

		cmdMap[c.Name] = mc
		for alias := range c.Aliases {
			cmdMap[alias] = mc
		}

		sectMap[c.Section] = append(sectMap[c.Section], mc)
	}
}
