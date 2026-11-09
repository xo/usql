package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/xo/rline"
	"github.com/xo/usql/dump"
	"github.com/xo/usql/env"
	"github.com/xo/usql/feed"
	"github.com/xo/usql/handler"
	"github.com/xo/usql/metacmd"
	"github.com/xo/usql/text"
)

// ContextExecutor is the command context.
type ContextExecutor interface {
	ExecuteContext(context.Context) error
}

// New builds the command context.
func New(cliargs []string) ContextExecutor {
	args := &Args{}
	var (
		bashCompletion       bool
		zshCompletion        bool
		fishCompletion       bool
		powershellCompletion bool
		noDescriptions       bool
		badHelp              bool
	)
	v := viper.New()
	c := &cobra.Command{
		Use:                text.CommandName + " [flags]... [DSN]",
		Short:              text.Short(),
		Version:            text.CommandVersion,
		SilenceErrors:      true,
		SilenceUsage:       true,
		DisableAutoGenTag:  true,
		DisableSuggestions: true,
		Args: func(_ *cobra.Command, cliargs []string) error {
			if len(cliargs) > 1 {
				return text.ErrWrongNumberOfArguments
			}
			return nil
		},
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			commandUpper := text.CommandUpper()
			args.ConfigFile = strings.TrimSpace(os.Getenv(commandUpper + "_CONFIG"))
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if s := strings.TrimSpace(f.Value.String()); f.Name == "config" && s != "" {
					args.ConfigFile = s
				}
			})
			// viper gives an option the value of its USQL_ variable. The
			// configuration file holds no options of the command line.
			v.SetEnvPrefix(commandUpper)
			v.AutomaticEnv()
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				// a dump reads its options from the command line only, so
				// that a variable cannot change what it writes
				if f.Name == "config" || f.Annotations[dumpAnnotation] != nil {
					return
				}
				_ = v.BindEnv(f.Name, commandUpper+"_"+strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_")))
				if !f.Changed && v.IsSet(f.Name) {
					_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", v.Get(f.Name)))
				}
			})
			// unhide params
			switch {
			case bashCompletion,
				zshCompletion,
				fishCompletion,
				powershellCompletion,
				cmd.Name() == "__complete":
				flags := cmd.Root().Flags()
				for _, name := range []string{"no-psqlrc", "no-" + text.CommandName + "rc", "var", "variable"} {
					flags.Lookup(name).Hidden = false
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, cliargs []string) error {
			// completions and short circuits
			switch {
			case bashCompletion:
				return cmd.GenBashCompletionV2(os.Stdout, !noDescriptions)
			case zshCompletion:
				if noDescriptions {
					return cmd.GenZshCompletionNoDesc(os.Stdout)
				}
				return cmd.GenZshCompletion(os.Stdout)
			case fishCompletion:
				return cmd.GenFishCompletion(os.Stdout, !noDescriptions)
			case powershellCompletion:
				if noDescriptions {
					return cmd.GenPowerShellCompletion(os.Stdout)
				}
				return cmd.GenPowerShellCompletionWithDesc(os.Stdout)
			case badHelp:
				return errors.New("unknown shorthand flag: 'h' in -h")
			}
			// run
			args.DSN = dsnFromEnv()
			if len(cliargs) > 0 {
				args.DSN = cliargs[0]
			}
			return Run(cmd.Context(), args)
		},
	}

	c.SetVersionTemplate("{{ .Name }} {{ .Version }}\n")
	c.SetArgs(cliargs[1:])
	c.SetUsageTemplate(text.UsageTemplate)
	text.UsageString = c.UsageString

	flags := c.Flags()
	flags.SortFlags = false

	// completions / short circuits
	flags.BoolVar(&bashCompletion, "completion-script-bash", false, "output bash completion script and exit")
	flags.BoolVar(&zshCompletion, "completion-script-zsh", false, "output zsh completion script and exit")
	flags.BoolVar(&fishCompletion, "completion-script-fish", false, "output fish completion script and exit")
	flags.BoolVar(&powershellCompletion, "completion-script-powershell", false, "output powershell completion script and exit")
	flags.BoolVar(&noDescriptions, "no-descriptions", false, "disable descriptions in completion scripts")
	flags.BoolVarP(&badHelp, "bad-help", "h", false, "bad help")

	// command / file flags
	flags.VarP(commandOrFile{args, true}, "command", "c", "run only single command (SQL or internal) and exit")
	flags.VarP(commandOrFile{args, false}, "file", "f", "execute commands from file and exit")

	// general flags
	flags.BoolVarP(&args.NoPassword, "no-password", "w", false, "never prompt for password")
	flags.BoolVarP(&args.NoInit, "no-init", "X", false, "do not execute initialization scripts (aliases: --no-rc --no-psqlrc --no-"+text.CommandName+"rc)")
	flags.BoolVar(&args.NoInit, "no-rc", false, "do not read startup file")
	flags.BoolVar(&args.NoInit, "no-psqlrc", false, "do not read startup file")
	flags.BoolVar(&args.NoInit, "no-"+text.CommandName+"rc", false, "do not read startup file")
	flags.VarP(filevar{&args.Out}, "out", "o", "output file")
	flags.BoolVarP(&args.ForcePassword, "password", "W", false, "force password prompt (should happen automatically)")
	flags.BoolVarP(&args.SingleTransaction, "single-transaction", "1", false, "execute as a single transaction (if non-interactive)")
	flags.BoolVarP(&args.Dump, "dump", "D", false, "write the database as a "+text.CommandName+" script to standard output or to the file of -o, then exit")

	// the options of --dump come from the table that \dump reads too. They
	// are long options only, and hidden, because the short options of psql's
	// output use -t, -T and -N.
	for _, df := range dump.Flags {
		f := flags.VarPF(dumpFlag{df, args}, df.Name, "", df.Usage)
		if df.Value == "" {
			f.NoOptDefVal = "true"
		}
		f.Hidden = true
		_ = flags.SetAnnotation(df.Name, dumpAnnotation, []string{"true"})
	}

	// every setting of the configuration, by its full name
	sf(flags, &args.Options, "option", "O", "", `set configuration option NAME (SECTION.NAME) to VALUE (see \config command)`, "NAME=VALUE")
	// set
	sf(flags, &args.Options, "set", "v", env.SectionVars, `set script variable NAME to VALUE (see \set command, aliases: --var --variable)`, "NAME=VALUE")
	sf(flags, &args.Options, "var", "", env.SectionVars, "set variable NAME to VALUE", "NAME=VALUE")
	sf(flags, &args.Options, "variable", "", env.SectionVars, "set variable NAME to VALUE", "NAME=VALUE")
	// uset
	sf(flags, &args.Options, "uset", "U", env.SectionUsql, `set setting NAME of `+text.CommandName+` to VALUE (see \uset command)`, "NAME=VALUE")
	// cset
	sf(flags, &args.Options, "cset", "N", env.SectionConnections, `set named connection NAME to DSN (see \cset command)`, "NAME=DSN")
	// pset
	sf(flags, &args.Options, "pset", "P", env.SectionPrint, `set printing option VAR to ARG (see \pset command)`, "VAR=ARG")
	// pset flags
	sf(flags, &args.Options, "field-separator", "F", "", `field separator for unaligned and CSV output (default "|" and ",")`, "FIELD-SEPARATOR", "print.fieldsep=%q", "print.csv_fieldsep=%q")
	sf(flags, &args.Options, "record-separator", "R", "", `record separator for unaligned and CSV output (default \n)`, "RECORD-SEPARATOR", "print.recordsep=%q")
	sf(flags, &args.Options, "table-attr", "T", "", "set HTML table tag attributes (e.g., width, border)", "TABLE-ATTR", "print.tableattr=%q")
	// pset bools
	sf(flags, &args.Options, "no-align", "A", "", "unaligned table output mode", "", "print.format=unaligned")
	sf(flags, &args.Options, "html", "H", "", "HTML table output mode", "", "print.format=html")
	sf(flags, &args.Options, "tuples-only", "t", "", "print rows only", "", "print.tuples_only=on")
	sf(flags, &args.Options, "expanded", "x", "", "turn on expanded table output", "", "print.expanded=on")
	sf(flags, &args.Options, "field-separator-zero", "z", "", "set field separator for unaligned and CSV output to zero byte", "", "print.fieldsep_zero=on")
	sf(flags, &args.Options, "record-separator-zero", "0", "", "set record separator for unaligned and CSV output to zero byte", "", "print.recordsep_zero=on")
	sf(flags, &args.Options, "json", "J", "", "JSON output mode", "", "print.format=json")
	sf(flags, &args.Options, "csv", "C", "", "CSV output mode", "", "print.format=csv")
	sf(flags, &args.Options, "vertical", "G", "", "vertical output mode", "", "print.format=vertical")
	sf(flags, &args.Options, "markdown", "M", "", "markdown output mode", "", "print.format=markdown")
	// set bools
	sf(flags, &args.Options, "quiet", "q", "", "run quietly (no messages, only query output)", "", "usql.quiet=on")

	// app config
	_ = flags.StringP("config", "", "", "config file")

	// manually set --version, see github.com/spf13/cobra/command.go
	_ = flags.BoolP("version", "V", false, "output version information, then exit")
	_ = flags.SetAnnotation("version", cobra.FlagSetByCobraAnnotation, []string{"true"})

	// manually set --help, see github.com/spf13/cobra/command.go
	_ = flags.BoolP("help", "?", false, "show this help, then exit")
	_ = c.Flags().SetAnnotation("help", cobra.FlagSetByCobraAnnotation, []string{"true"})

	// mark hidden
	for _, name := range []string{
		"no-rc", "no-psqlrc", "no-" + text.CommandName + "rc", "var", "variable",
		"completion-script-bash", "completion-script-zsh", "completion-script-fish",
		"completion-script-powershell", "no-descriptions",
		"bad-help",
	} {
		flags.Lookup(name).Hidden = true
	}

	return c
}

// Run runs the application.
func Run(ctx context.Context, args *Args) error {
	// get user
	u, err := user.Current()
	if err != nil {
		return err
	}
	// get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// determine if interactive
	interactive := isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
	forceNonInteractive := len(args.CommandOrFiles) != 0 || args.Dump
	if args.Dump {
		switch {
		case len(args.CommandOrFiles) != 0:
			return text.ErrDumpWithCommands
		case args.SingleTransaction:
			return text.ErrDumpWithSingleTransaction
		}
	} else if args.DumpOptionSet {
		return text.ErrDumpOptionWithoutDump
	}

	// the settings apply from the first source to the last, so that the
	// last wins: the defaults, the configuration file, the variables of
	// usql's environment, then the command line
	configFile, err := applyConfig(args.ConfigFile, !forceNonInteractive && interactive)
	if err != nil {
		return err
	}
	for _, opt := range args.Options {
		if err := env.Vars().SetOption(opt); err != nil {
			return fmt.Errorf("%s: %w", opt, err)
		}
	}
	if args.Charts, err = chartsFS(configFile); err != nil {
		return err
	}
	// the history is kept at a terminal, in a directory that may not be there
	// yet
	history := env.HistoryFile(u)
	if history != "" && interactive && !forceNonInteractive {
		if err := os.MkdirAll(filepath.Dir(history), 0o700); err != nil {
			fmt.Fprintf(os.Stderr, text.HistoryNotSaved+"\n", err)
			history = ""
		}
	}
	// create the line editor, which asks the handler whether an entry needs
	// more input
	var h *handler.Handler
	opts := []rline.Option{
		rline.WithHistoryFile(history),
		rline.WithContinue(func(text string) bool {
			return h != nil && h.Continue(text)
		}),
		// a hint runs the completer as keys are typed, and the completer can
		// query the database
		rline.WithHints(false),
		// the prompt text is the whole prompt, so the editor adds no marker
		rline.WithPrompt("", ""),
		// no escape codes reach a pipe or a file
		rline.WithColor(interactive),
		// Tab opens a numbered menu of the completions
		rline.WithCompletionMenu(true),
		// the user's inputrc applies, and its $if tests the name of usql
		rline.WithUserInputrc(true),
		rline.WithAppName(text.CommandName),
		// Ctrl-C leaves the entry on the screen with ^C after it, as psql does
		rline.WithKeepInterrupted(true),
	}
	// input from a pipe or a file is decoded as a script given with -f is
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		opts = append(opts, rline.WithStdin(&decodedStdin{}))
	}
	l, err := rline.New(opts...)
	if err != nil {
		return err
	}
	defer l.Close()
	// with -c or -f the commands and files are the input, and the terminal
	// is not read
	var r feed.LineReader = l
	if forceNonInteractive {
		r = feed.NewLines(strings.NewReader(""))
	}
	if h, err = handler.New(l, r, u, wd, args.Charts, args.NoPassword); err != nil {
		return err
	}
	// closing the connection at exit stops the forward that dbfwd runs for
	// it. An open transaction is rolled back, as the server does when psql
	// exits.
	defer func() {
		if h.Close() == text.ErrPreviousTransactionExists {
			_ = h.Rollback()
			_ = h.Close()
		}
	}()
	// -o sends the results to a file, as \o does
	if args.Out != "" && !args.Dump {
		f, err := os.OpenFile(args.Out, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		h.SetOutput(f)
	}
	// force password
	dsn := args.DSN
	if args.ForcePassword {
		if dsn, err = h.Password(dsn); err != nil {
			return err
		}
	}
	// open dsn
	if err = h.Open(ctx, dsn); err != nil {
		return err
	}
	// start transaction
	if args.SingleTransaction {
		if h.Interactive() {
			return text.ErrSingleTransactionCannotBeUsedWithInteractiveMode
		}
		if err = h.BeginTx(ctx, nil); err != nil {
			return err
		}
	}
	// a dump runs no script, as pg_dump reads no psqlrc
	if args.Dump {
		return dumpTo(ctx, h, args.Out, args.DumpOptions)
	}
	// init.sql
	if !args.NoInit {
		if name := env.InitFile(u); name != "" {
			if err = h.Include(name, false); err != nil && err != text.ErrNoSuchFileOrDirectory {
				return err
			}
		}
	}
	// setup runner
	f := h.Run
	if len(args.CommandOrFiles) != 0 {
		f = runCommandOrFiles(h, args.CommandOrFiles)
	}
	// run
	if err = f(); err != nil {
		return err
	}
	// commit
	if args.SingleTransaction {
		return h.Commit()
	}
	return nil
}

// decodedStdin is standard input as UTF-8. It is decoded on its first read,
// so that the encoding setting of the command line or of init.sql applies.
type decodedStdin struct {
	r io.Reader
}

// Read satisfies io.Reader.
func (s *decodedStdin) Read(p []byte) (int, error) {
	if s.r == nil {
		r, err := feed.Decode(os.Stdin, env.Setting("encoding"))
		if err != nil {
			return 0, err
		}
		s.r = r
	}
	return s.r.Read(p)
}

// Args are the command line arguments.
type Args struct {
	DSN               string
	CommandOrFiles    []CommandOrFile
	Out               string
	ForcePassword     bool
	NoPassword        bool
	NoInit            bool
	SingleTransaction bool
	// Options are the settings of the command line, as SECTION.NAME=VALUE,
	// in the order that they were given.
	Options []string
	// ConfigFile is the configuration file that --config or USQL_CONFIG
	// names.
	ConfigFile string
	Charts     billy.Filesystem
	// Dump writes the database as a script, with the options of DumpOptions.
	// DumpOptionSet is set when the command line gives an option of a dump.
	Dump          bool
	DumpOptions   dump.Options
	DumpOptionSet bool
}

// dumpAnnotation marks the options of --dump.
const dumpAnnotation = "usql_dump"

// dumpFlag provides a [pflag.Value] for an option of a dump.
type dumpFlag struct {
	f    dump.Flag
	args *Args
}

// Set satisfies the [pflag.Value] interface.
func (d dumpFlag) Set(value string) error {
	d.args.DumpOptionSet = true
	return d.f.Set(&d.args.DumpOptions, value)
}

// String satisfies the [pflag.Value] interface.
func (dumpFlag) String() string {
	return ""
}

// Type satisfies the [pflag.Value] interface.
func (d dumpFlag) Type() string {
	if d.f.Value == "" {
		return "bool"
	}
	return d.f.Value
}

// dumpTo writes the dump of the database to the file name, or to standard
// output when name is empty or -. A file is written whole or not at all (see
// [dump.WriteFile]).
func dumpTo(ctx context.Context, h *handler.Handler, name string, opts dump.Options) error {
	write := func(out io.Writer) error {
		w := bufio.NewWriter(out)
		if err := h.Dump(ctx, w, opts); err != nil {
			_ = w.Flush()
			return err
		}
		return w.Flush()
	}
	if name == "" || name == "-" {
		return write(os.Stdout)
	}
	return dump.WriteFile(name, write)
}

// CommandOrFile is a special type to deal with interspersed -c, -f,
// command-line options, to ensure proper order execution.
type CommandOrFile struct {
	Command bool
	Value   string
}

// commandOrFile provides a [pflag.Value] to wrap the command or file value in
// [Args].
type commandOrFile struct {
	args    *Args
	command bool
}

// Set satisfies the [pflag.Value] interface.
func (c commandOrFile) Set(value string) error {
	c.args.CommandOrFiles = append(c.args.CommandOrFiles, CommandOrFile{
		Command: c.command,
		Value:   value,
	})
	return nil
}

// String satisfies the [pflag.Value] interface.
func (c commandOrFile) String() string {
	return ""
}

// Type satisfies the [pflag.Value] interface.
func (c commandOrFile) Type() string {
	if c.command {
		return "COMMAND"
	}
	return "FILE"
}

// vs handles setting vars with predefined values.
type vs struct {
	vars *[]string
	// section is the section of the value, which goes before its name.
	section string
	vals    []string
	typ     string
}

// Set satisfies the [pflag.Value] interface.
func (p vs) Set(value string) error {
	if len(p.vals) != 0 {
		for _, v := range p.vals {
			if strings.Contains(v, "%") {
				*p.vars = append(*p.vars, fmt.Sprintf(v, value))
			} else {
				*p.vars = append(*p.vars, v)
			}
		}
	} else if p.section != "" {
		*p.vars = append(*p.vars, p.section+"."+value)
	} else {
		*p.vars = append(*p.vars, value)
	}
	return nil
}

// String satisfies the [pflag.Value] interface.
func (vs) String() string {
	return ""
}

// Type satisfies the [pflag.Value] interface.
func (p vs) Type() string {
	if p.typ == "" {
		return "bool"
	}
	return p.typ
}

// filevar is a file var.
type filevar struct {
	v *string
}

// Set satisfies the [pflag.Value] interface.
func (p filevar) Set(value string) error {
	*p.v = value
	return nil
}

// String satisfies the [pflag.Value] interface.
func (filevar) String() string {
	return ""
}

// Type satisfies the [pflag.Value] interface.
func (filevar) Type() string {
	return "FILE"
}

// applyConfig applies the configuration file, then the variables of usql's
// environment, such as USQL_PAGER, over the defaults. It reads the file that
// name names, or else config.yaml or config.yml in the configuration
// directory. The errors of the file and of the variables are written out, and
// do not stop usql, but a file that name names must exist. It returns the
// file that it read.
//
// The terminal graphics are off unless the output is a terminal, so that no
// control characters reach a pipe or a file. Only the command line turns them
// on there.
func applyConfig(name string, interactive bool) (string, error) {
	// at a terminal, the terminal graphics are found out, unless
	// TERM_GRAPHICS names them
	if interactive {
		s, _ := env.Getenv("TERM_GRAPHICS")
		if err := env.Vars().SetSetting("term_graphics", s); err != nil {
			fmt.Fprintln(os.Stderr, "error: TERM_GRAPHICS:", err)
		}
	}
	if name == "" {
		name = defaultConfigFile()
	} else if _, err := os.Stat(name); err != nil {
		return "", err
	}
	if name != "" {
		for _, err := range env.Vars().LoadConfig(name, metacmd.Has) {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}
	if err := env.Vars().ApplyEnviron(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	if s, ok := env.Getenv(text.CommandUpper() + "_TERM_GRAPHICS"); ok && interactive {
		if err := env.Vars().SetSetting("term_graphics", s); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s_TERM_GRAPHICS: %v\n", text.CommandUpper(), err)
		}
	}
	if !interactive {
		_ = env.Vars().SetSetting("term_graphics", "none")
	}
	return name, nil
}

// defaultConfigFile returns config.yaml or config.yml in the configuration
// directory, whichever exists first, or an empty string.
func defaultConfigFile() string {
	dir, err := env.ConfigDir()
	if err != nil {
		return ""
	}
	for _, ext := range []string{".yaml", ".yml"} {
		name := filepath.Join(dir, text.ConfigName+ext)
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}
	return ""
}

// chartsFS creates a filesystem for charts, in the setting charts.path, which
// is relative to the directory of the configuration file.
func chartsFS(configFile string) (billy.Filesystem, error) {
	configDir := filepath.Dir(configFile)
	if configFile == "" {
		var err error
		if configDir, err = env.ConfigDir(); err != nil {
			return nil, err
		}
	}
	chartsPath := env.Vars().Chart("path")
	if chartsPath == "" {
		chartsPath = "charts"
	}
	fs := osfs.New(configDir, osfs.WithBoundOS())
	switch fi, err := fs.Stat(chartsPath); {
	case err != nil && os.IsNotExist(err) && chartsPath == "charts":
		return memfs.New(), nil
	case err != nil && os.IsNotExist(err):
		fmt.Fprintln(os.Stderr, fmt.Sprintf(text.ChartsPathDoesNotExist, chartsPath))
		return memfs.New(), nil
	case err != nil:
		return nil, err
	case !fi.IsDir():
		fmt.Fprintln(os.Stderr, fmt.Sprintf(text.ChartsPathIsNotADirectory, chartsPath))
		return memfs.New(), nil
	}
	return fs.Chroot(chartsPath)
}

// dsnFromEnv returns the value of USQL_DSN, the DSN that usql opens when the
// command line names none. It removes the variable from the environment of
// usql, so that a password in it does not reach a shell, a pager or an editor
// that usql starts. A DSN on the command line wins, and \c never reads the
// variable.
func dsnFromEnv() string {
	name := text.CommandUpper() + "_DSN"
	dsn := strings.TrimSpace(os.Getenv(name))
	_ = os.Unsetenv(name)
	return dsn
}

// runCommandOrFiles processes all the supplied commands or files.
func runCommandOrFiles(h *handler.Handler, commandsOrFiles []CommandOrFile) func() error {
	return func() error {
		for _, c := range commandsOrFiles {
			if c.Command {
				if err := h.RunCommand(c.Value); err != nil {
					return err
				}
			} else {
				if err := h.Include(c.Value, false); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// sf sets a flag. The value of the flag goes in section, unless vals give
// the full names that the flag sets.
func sf(flags *pflag.FlagSet, v *[]string, name, short, section, usage, placeholder string, vals ...string) {
	f := flags.VarPF(vs{v, section, vals, placeholder}, name, short, usage)
	if placeholder == "" {
		// a flag with no value is a switch, and is off until it is given
		f.NoOptDefVal = "true"
	}
}
