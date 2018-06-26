package main

import (
	"fmt"
	"io"
	"os"

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
	Variables         []string
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
	args := new(Args)

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
	kingpin.Flag("variable", "set variable").Short('v').PlaceHolder("NAME=VALUE").StringsVar(&args.Variables)

	// add --set as a hidden alias for --variable
	kingpin.Flag("set", "set variable").Hidden().StringsVar(&args.Variables)

	// add --version flag
	kingpin.Flag("version", "display version and exit").PreAction(func(*kingpin.ParseContext) error {
		fmt.Fprintln(os.Stdout, text.CommandName, text.CommandVersion)
		os.Exit(0)
		return nil
	}).Bool()

	// hide help flag
	kingpin.HelpFlag.Short('h').Hidden()

	// parse
	kingpin.Parse()

	return args
}
