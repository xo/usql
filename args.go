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
	DSN string

	CommandOrFiles    []CommandOrFile
	Out               string
	ForcePassword     bool
	NoPassword        bool
	NoRC              bool
	SingleTransaction bool

	Variables  []string
	PVariables []string
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
		pc("field-separator", 'F', `field separator for unaligned output (default, "|")`, "fieldsep=%q", "fieldsep_zero=off"),
		pc("html", 'H', "HTML table output mode", "format=html"),
		pc("record-separator", 'R', `record separator for unaligned output (default, \n)`, "recordsep=%q", "recordsep_zero=off"),
		pc("tuples-only", 't', "print rows only", "tuples_only=on"),
		pc("table-attr", 'T', "set HTML table tag attributes (e.g., width, border)", "tableattr=%q"),
		pc("expanded", 'x', "turn on expanded table output", "expanded=on"),
		pc("field-separator-zero", 'z', "set field separator for unaligned output to zero byte", "fieldsep=''", "fieldsep_zero=on"),
		pc("record-separator-zero", '0', "set record separator for unaligned output to zero byte", "recordsep=''", "recordsep_zero=on"),
		pc("json", 'J', "JSON output mode", "format=json"),
		pc("csv", 'C', "CSV output mode", "format=csv"),
	} {
		a := kingpin.Flag(c.long, c.help).Short(c.short).PlaceHolder("TEXT")
		if strings.Contains(c.vals[0], "%q") {
			a.PreAction(func(ctxt *kingpin.ParseContext) error {
				if len(ctxt.Elements) != 1 {
					return fmt.Errorf("--%s must be passed a value", c.long)
				}
				vals := make([]string, len(c.vals))
				copy(vals, c.vals)
				vals[0] = fmt.Sprintf(vals[0], *ctxt.Elements[0].Value)
				args.PVariables = append(args.PVariables, vals...)
				return nil
			}).String()
		} else {
			a.PreAction(func(*kingpin.ParseContext) error {
				args.PVariables = append(args.PVariables, c.vals...)
				return nil
			}).Bool()
		}
	}

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
