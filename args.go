package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/xo/usql/text"
)

// CommandOrFile is a special type to deal with interspersed -c, -f,
// command-line options, to ensure proper order execution.
type CommandOrFile struct {
	Command bool
	Value   string
}

// Args are the command line arguments.
type Args struct {
	DSN               string
	CommandOrFiles    []CommandOrFile
	Out               string
	ForcePassword     bool
	NoPassword        bool
	NoRC              bool
	SingleTransaction bool
	Variables         []string
	PVariables        []string
}

func (args *Args) Next() (string, bool, error) {
	if len(args.CommandOrFiles) == 0 {
		return "", false, io.EOF
	}
	cmd := args.CommandOrFiles[0]
	args.CommandOrFiles = args.CommandOrFiles[1:]
	return cmd.Value, cmd.Command, nil
}

type commandOrFile struct {
	args    *Args
	command bool
}

func (c commandOrFile) Set(value string) error {
	c.args.CommandOrFiles = append(c.args.CommandOrFiles, CommandOrFile{
		Command: c.command,
		Value:   value,
	})
	return nil
}

func (c commandOrFile) String() string {
	return ""
}

func (c commandOrFile) IsCumulative() bool {
	return true
}

// for populating args.PVariables with user-specified options
type pset struct {
	args *Args
	vals []string
}

func (p pset) Set(value string) error {
	for i, v := range p.vals {
		if strings.ContainsRune(v, '%') {
			p.vals[i] = fmt.Sprintf(v, value)
		}
	}
	p.args.PVariables = append(p.args.PVariables, p.vals...)
	return nil
}

func (p pset) String() string {
	return ""
}

func (p pset) IsCumulative() bool {
	return true
}

func NewArgs() *Args {
	args := &Args{}
	// set usage template
	kingpin.UsageTemplate(text.UsageTemplate())
	kingpin.Arg("dsn", "database url").StringVar(&args.DSN)
	// command / file flags
	kingpin.Flag("command", "run only single command (SQL or internal) and exit").Short('c').SetValue(commandOrFile{args, true})
	kingpin.Flag("file", "execute commands from file and exit").Short('f').SetValue(commandOrFile{args, false})
	// general flags
	kingpin.Flag("no-password", "never prompt for password").Short('w').BoolVar(&args.NoPassword)
	kingpin.Flag("no-rc", "do not read start up file").Short('X').BoolVar(&args.NoRC)
	kingpin.Flag("out", "output file").Short('o').StringVar(&args.Out)
	kingpin.Flag("password", "force password prompt (should happen automatically)").Short('W').BoolVar(&args.ForcePassword)
	kingpin.Flag("single-transaction", "execute as a single transaction (if non-interactive)").Short('1').BoolVar(&args.SingleTransaction)
	kingpin.Flag("set", "set variable NAME to VALUE").Short('v').PlaceHolder(", --variable=NAME=VALUE").StringsVar(&args.Variables)
	// pset
	kingpin.Flag("pset", `set printing option VAR to ARG (see \pset command)`).Short('P').PlaceHolder("VAR[=ARG]").StringsVar(&args.PVariables)
	// pset flags
	kingpin.Flag("field-separator", `field separator for unaligned and CSV output (default "|" and ",")`).Short('F').SetValue(pset{args, []string{"fieldsep=%q", "csv_fieldsep=%q"}})
	kingpin.Flag("record-separator", `record separator for unaligned and CSV output (default \n)`).Short('R').SetValue(pset{args, []string{"recordsep=%q"}})
	kingpin.Flag("table-attr", "set HTML table tag attributes (e.g., width, border)").Short('T').SetValue(pset{args, []string{"tableattr=%q"}})
	type psetconfig struct {
		long  string
		short rune
		help  string
		vals  []string
	}
	pc := func(long string, r rune, help string, vals ...string) psetconfig {
		return psetconfig{long, r, help, vals}
	}
	for _, c := range []psetconfig{
		pc("no-align", 'A', "unaligned table output mode", "format=unaligned"),
		pc("html", 'H', "HTML table output mode", "format=html"),
		pc("tuples-only", 't', "print rows only", "tuples_only=on"),
		pc("expanded", 'x', "turn on expanded table output", "expanded=on"),
		pc("field-separator-zero", 'z', "set field separator for unaligned and CSV output to zero byte", "fieldsep_zero=on"),
		pc("record-separator-zero", '0', "set record separator for unaligned and CSV output to zero byte", "recordsep_zero=on"),
		pc("json", 'J', "JSON output mode", "format=json"),
		pc("csv", 'C', "CSV output mode", "format=csv"),
		pc("vertical", 'G', "vertical output mode", "format=vertical"),
	} {
		// make copy of values for the callback closure (see https://stackoverflow.com/q/26692844)
		vals := make([]string, len(c.vals))
		copy(vals, c.vals)
		kingpin.Flag(c.long, c.help).Short(c.short).PlaceHolder("TEXT").PreAction(func(*kingpin.ParseContext) error {
			args.PVariables = append(args.PVariables, vals...)
			return nil
		}).Bool()
	}
	kingpin.Flag("quiet", "run quietly (no messages, only query output)").Short('q').PreAction(func(*kingpin.ParseContext) error {
		args.Variables = append(args.Variables, "QUIET=on")
		return nil
	}).Bool()
	// add --set as a hidden alias for --variable
	kingpin.Flag("variable", "set variable NAME to VALUE").Hidden().StringsVar(&args.Variables)
	// add --version flag
	kingpin.Flag("version", "display version and exit").PreAction(func(*kingpin.ParseContext) error {
		fmt.Fprintln(os.Stdout, text.CommandName, text.CommandVersion)
		os.Exit(0)
		return nil
	}).Short('V').Bool()
	// hide help flag
	kingpin.HelpFlag.Short('h').Hidden()
	// parse
	kingpin.Parse()
	return args
}
