package metacmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/knq/dburl"
	"github.com/knq/usql/drivers"
	"github.com/knq/usql/env"
	"github.com/knq/usql/text"
)

// Cmd is a command implementation.
type Cmd struct {
	Section Section
	Name    string
	Desc    string
	Min     int
	Aliases map[string]string
	Process func(Handler, string, []string) (Res, error)
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
			Process: func(h Handler, _ string, _ []string) (Res, error) {
				Listing(h.IO().Stdout())
				return Res{}, nil
			},
		},

		Quit: {
			Section: SectionGeneral,
			Name:    "q",
			Desc:    "quit " + text.CommandName,
			Aliases: map[string]string{"quit": ""},
			Process: func(Handler, string, []string) (Res, error) {
				return Res{Quit: true}, nil
			},
		},

		Copyright: {
			Section: SectionGeneral,
			Name:    "copyright",
			Desc:    "show " + text.CommandName + " usage and distribution terms",
			Process: func(h Handler, _ string, _ []string) (Res, error) {
				fmt.Fprintln(h.IO().Stdout(), text.Copyright)
				return Res{}, nil
			},
		},

		ConnInfo: {
			Section: SectionConnection,
			Name:    "conninfo",
			Desc:    "display information about the current database connection",
			Process: func(h Handler, _ string, _ []string) (Res, error) {
				if u := h.URL(); u != nil {
					out := h.IO().Stdout()
					fmt.Fprintf(out, text.ConnInfo, u.Driver, u.DSN)
					fmt.Fprintln(out)
				}
				return Res{}, nil
			},
		},

		Drivers: {
			Section: SectionGeneral,
			Name:    "drivers",
			Desc:    "display information about available database drivers",
			Process: func(h Handler, _ string, _ []string) (Res, error) {
				out := h.IO().Stdout()

				names := make([]string, len(drivers.Drivers))
				var z int
				for k := range drivers.Drivers {
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

				return Res{}, nil
			},
		},

		Connect: {
			Section: SectionConnection,
			Name:    "c",
			Desc:    "connect to database with url,URL",
			Aliases: map[string]string{
				"c":       "connect to database with SQL driver and parameters,DRIVER [PARAMS]",
				"connect": "",
			},
			Min: 1,
			Process: func(h Handler, _ string, params []string) (Res, error) {
				return Res{Processed: len(params)}, h.Open(params...)
			},
		},

		Disconnect: {
			Section: SectionConnection,
			Name:    "Z",
			Desc:    "close database connection",
			Aliases: map[string]string{"disconnect": ""},
			Process: func(h Handler, _ string, _ []string) (Res, error) {
				return Res{}, h.Close()
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
			Process: func(h Handler, cmd string, params []string) (Res, error) {
				res := Res{
					Exec: ExecOnly,
				}

				switch cmd {
				case "g":
					if len(params) > 0 {
						res.ExecParam = params[0]
						res.Processed++
					}

				case "gexec":
					res.Exec = ExecExec

				case "gset":
					res.Exec = ExecSet
					if len(params) > 0 {
						res.ExecParam = params[0]
						res.Processed++
					}
				}

				return res, nil
			},
		},

		Edit: {
			Section: SectionQueryBuffer,
			Name:    "e",
			Desc:    "edit the query buffer (or file) with external editor,[FILE] [LINE]",
			Aliases: map[string]string{"edit": ""},
			Process: func(h Handler, _ string, params []string) (Res, error) {
				var res Res
				var path, line string

				// get path, line params
				if len(params) > 0 {
					path = env.Expand(params[0], h.User().HomeDir)
					res.Processed++
				}
				if len(params) > 1 {
					line = params[1]
					res.Processed++
				}

				// get last statement
				s, buf := h.Last(), h.Buf()
				if buf.Len != 0 {
					s = buf.String()
				}

				n, err := env.EditFile(path, line, s)

				// reset if no error
				if err == nil {
					buf.Reset()
					buf.Feed(n)
				}

				return res, err
			},
		},

		Print: {
			Section: SectionQueryBuffer,
			Name:    "p",
			Desc:    "show the contents of the query buffer",
			Aliases: map[string]string{"print": ""},
			Process: func(h Handler, _ string, _ []string) (Res, error) {
				// get last statement
				s, buf := h.Last(), h.Buf()
				if buf.Len != 0 {
					s = buf.String()
				}

				if s == "" {
					s = text.QueryBufferEmpty
				}

				fmt.Fprintln(h.IO().Stdout(), s)
				return Res{}, nil
			},
		},

		Reset: {
			Section: SectionQueryBuffer,
			Name:    "r",
			Desc:    "reset (clear) the query buffer",
			Aliases: map[string]string{"reset": ""},
			Process: func(h Handler, _ string, _ []string) (Res, error) {
				h.Buf().Reset()
				fmt.Fprintln(h.IO().Stdout(), text.QueryBufferReset)
				return Res{}, nil
			},
		},

		Echo: {
			Section: SectionInputOutput,
			Name:    "echo",
			Desc:    "write string to standard output,[STRING]",
			Process: func(h Handler, _ string, params []string) (Res, error) {
				fmt.Fprintln(h.IO().Stdout(), strings.Join(params, " "))
				return Res{Processed: len(params)}, nil
			},
		},

		Write: {
			Section: SectionQueryBuffer,
			Name:    "w",
			Min:     1,
			Desc:    "write query buffer to file,FILE",
			Aliases: map[string]string{"write": ""},
			Process: func(h Handler, _ string, params []string) (Res, error) {
				// get last statement
				s, buf := h.Last(), h.Buf()
				if buf.Len != 0 {
					s = buf.String()
				}

				return Res{Processed: 1}, ioutil.WriteFile(
					params[0],
					[]byte(strings.TrimSuffix(s, "\n")+"\n"),
					0644,
				)
			},
		},

		ChangeDir: {
			Section: SectionOperatingSystem,
			Name:    "cd",
			Desc:    "change the current working directory,[DIR]",
			Process: func(h Handler, _ string, params []string) (Res, error) {
				var res Res

				home, path := h.User().HomeDir, ""
				if len(params) > 0 {
					path = env.Expand(params[0], home)
					res.Processed++
				}

				return res, os.Chdir(path)
			},
		},

		SetEnv: {
			Section: SectionOperatingSystem,
			Name:    "setenv",
			Min:     1,
			Desc:    "set or unset environment variable,NAME [VALUE]",
			Process: func(h Handler, _ string, params []string) (Res, error) {
				var err error

				n := params[0]
				if len(params) == 1 {
					err = os.Unsetenv(n)
				} else {
					err = os.Setenv(n, strings.Join(params, " "))
				}

				return Res{Processed: len(params)}, err
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
			Process: func(h Handler, cmd string, params []string) (Res, error) {
				err := h.Include(
					env.Expand(params[0], h.User().HomeDir),
					cmd == "ir" || cmd == "include_relative",
				)
				if err != nil {
					err = fmt.Errorf("%s: %v", params[0], err)
				}
				return Res{Processed: 1}, err
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
		for alias, _ := range c.Aliases {
			cmdMap[alias] = mc
		}

		sectMap[c.Section] = append(sectMap[c.Section], mc)
	}
}
