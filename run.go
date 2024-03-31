package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/xo/usql/env"
	"github.com/xo/usql/handler"
	"github.com/xo/usql/metacmd"
	"github.com/xo/usql/rline"
	"github.com/xo/usql/text"
)

// Run processes args, processing args.CommandOrFiles if non-empty, if
// specified, otherwise launch an interactive readline from stdin.
func Run(ctx context.Context, cliargs []string) error {
	args := &Args{}
	v := viper.New()
	c := &cobra.Command{
		Use:     text.CommandName + " [flags]... [DSN]",
		Short:   text.Short(),
		Version: text.CommandVersion,
		Args:    cobra.RangeArgs(0, 1),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			commandUpper := text.CommandUpper()
			configFile := strings.TrimSpace(os.Getenv(commandUpper + "_CONFIG"))
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if s := strings.TrimSpace(f.Value.String()); f.Name == "config" && s != "" {
					configFile = s
				}
			})
			if configFile != "" {
				v.SetConfigFile(configFile)
			} else {
				v.SetConfigName(text.ConfigName)
				if configDir, err := os.UserConfigDir(); err == nil {
					v.AddConfigPath(filepath.Join(configDir, text.CommandName))
				}
			}
			if err := v.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return err
				}
			}
			v.SetEnvPrefix(commandUpper)
			v.AutomaticEnv()
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Name == "config" {
					return
				}
				_ = v.BindEnv(f.Name, commandUpper+"_"+strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_")))
				if !f.Changed && v.IsSet(f.Name) {
					_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", v.Get(f.Name)))
				}
			})
			return nil
		},
		RunE: func(cmd *cobra.Command, cliargs []string) error {
			if len(cliargs) > 0 {
				args.DSN = cliargs[0]
			}
			return run(cmd.Context(), args)
		},
	}

	c.SetVersionTemplate("{{ .Name }} {{ .Version }}\n")
	c.SetArgs(cliargs[1:])
	c.SilenceErrors = true
	c.SetUsageTemplate(text.UsageTemplate)

	flags := c.Flags()
	flags.SortFlags = false
	// command / file flags
	flags.VarP(commandOrFile{args, true}, "command", "c", "run only single command (SQL or internal) and exit")
	flags.VarP(commandOrFile{args, false}, "file", "f", "execute commands from file and exit")
	// general flags
	flags.BoolVarP(&args.NoPassword, "no-password", "w", false, "never prompt for password")
	flags.BoolVarP(&args.NoRC, "no-rc", "X", false, "do not read start up file (aliases: --no-psqlrc --no-usqlrc)")
	flags.BoolVar(&args.NoRC, "no-psqlrc", false, "do not read startup file")
	flags.BoolVar(&args.NoRC, "no-usqlrc", false, "do not read startup file")
	flags.VarP(filevar{&args.Out}, "out", "o", "output file")
	flags.BoolVarP(&args.ForcePassword, "password", "W", false, "force password prompt (should happen automatically)")
	flags.BoolVarP(&args.SingleTransaction, "single-transaction", "1", false, "execute as a single transaction (if non-interactive)")

	ss := func(v *[]string, name, short, usage, placeholder string, vals ...string) {
		f := flags.VarPF(vs{v, vals, placeholder}, name, short, usage)
		if placeholder == "" {
			f.DefValue, f.NoOptDefVal = "true", "true"
		}
	}
	// set
	ss(&args.Variables, "set", "v", `set variable NAME to VALUE (see \set command, aliases: --var --variable)`, "NAME=VALUE")
	ss(&args.Variables, "var", "", "set variable NAME to VALUE", "NAME=VALUE")
	ss(&args.Variables, "variable", "", "set variable NAME to VALUE", "NAME=VALUE")
	// pset
	ss(&args.PVariables, "pset", "P", `set printing option VAR to ARG (see \pset command)`, "VAR=ARG")
	// pset flags
	ss(&args.PVariables, "field-separator", "F", `field separator for unaligned and CSV output (default "|" and ",")`, "FIELD-SEPARATOR", "fieldsep=%q", "csv_fieldsep=%q")
	ss(&args.PVariables, "record-separator", "R", `record separator for unaligned and CSV output (default \n)`, "RECORD-SEPARATOR", "recordsep=%q")
	ss(&args.PVariables, "table-attr", "T", "set HTML table tag attributes (e.g., width, border)", "TABLE-ATTR", "tableattr=%q")
	// pset bools
	ss(&args.PVariables, "no-align", "A", "unaligned table output mode", "", "format=unaligned")
	ss(&args.PVariables, "html", "H", "HTML table output mode", "", "format=html")
	ss(&args.PVariables, "tuples-only", "t", "print rows only", "", "tuples_only=on")
	ss(&args.PVariables, "expanded", "x", "turn on expanded table output", "", "expanded=on")
	ss(&args.PVariables, "field-separator-zero", "z", "set field separator for unaligned and CSV output to zero byte", "", "fieldsep_zero=on")
	ss(&args.PVariables, "record-separator-zero", "0", "set record separator for unaligned and CSV output to zero byte", "", "recordsep_zero=on")
	ss(&args.PVariables, "json", "J", "JSON output mode", "", "format=json")
	ss(&args.PVariables, "csv", "C", "CSV output mode", "", "format=csv")
	ss(&args.PVariables, "vertical", "G", "vertical output mode", "", "format=vertical")
	// set bools
	ss(&args.Variables, "quiet", "q", "run quietly (no messages, only query output)", "", "QUIET=on")
	// add config
	_ = flags.StringP("config", "", "", "config file")
	// manually set --version, see github.com/spf13/cobra/command.go
	flags.BoolP("version", "V", false, "output version information, then exit")
	_ = flags.SetAnnotation("version", cobra.FlagSetByCobraAnnotation, []string{"true"})
	// manually set --help, see github.com/spf13/cobra/command.go
	flags.Bool("help", false, "show this help, then exit")
	_ = c.Flags().SetAnnotation("help", cobra.FlagSetByCobraAnnotation, []string{"true"})
	// expose to metacmd
	metacmd.Usage = func(w io.Writer) {
		_, _ = w.Write([]byte(text.Short() + "\n\n" + c.UsageString()))
	}
	// mark hidden
	for _, s := range []string{"no-psqlrc", "no-usqlrc", "var", "variable"} {
		if err := flags.MarkHidden(s); err != nil {
			return err
		}
	}
	return c.ExecuteContext(ctx)
}

func run(ctx context.Context, args *Args) error {
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
	cygwin := isatty.IsCygwinTerminal(os.Stdout.Fd()) && isatty.IsCygwinTerminal(os.Stdin.Fd())
	forceNonInteractive := len(args.CommandOrFiles) != 0
	// enable term graphics
	/*
		if !forceNonInteractive && interactive && !cygwin {
			// NOTE: this is done here and not in the env.init() package, because
			// NOTE: we need to determine if it is interactive first, otherwise it
			// NOTE: could mess up the non-interactive output with control characters
			var typ string
			if s, _ := env.Getenv(commandUpper+"_TERM_GRAPHICS", "TERM_GRAPHICS"); s != "" {
				typ = s
			}
			if err := env.Set("TERM_GRAPHICS", typ); err != nil {
				return err
			}
		}
	*/

	// fmt.Fprintf(os.Stdout, "\n\nVARS: %v\n\nPVARS: %v\n\n\n", args.Variables, args.PVariables)

	// handle variables
	for _, v := range args.Variables {
		if i := strings.Index(v, "="); i != -1 {
			_ = env.Set(v[:i], v[i+1:])
		} else {
			_ = env.Unset(v)
		}
	}
	for _, v := range args.PVariables {
		if i := strings.Index(v, "="); i != -1 {
			vv := v[i+1:]
			if c := vv[0]; c == '\'' || c == '"' {
				var err error
				vv, err = env.Dequote(vv, c)
				if err != nil {
					return err
				}
			}
			if _, err = env.Pset(v[:i], vv); err != nil {
				return err
			}
		} else {
			if _, err = env.Ptoggle(v, ""); err != nil {
				return err
			}
		}
	}
	// create input/output
	l, err := rline.New(interactive, cygwin, forceNonInteractive, args.Out, env.HistoryFile(u))
	if err != nil {
		return err
	}
	defer l.Close()
	// create handler
	h := handler.New(l, u, wd, args.NoPassword)
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
		if h.IO().Interactive() {
			return text.ErrSingleTransactionCannotBeUsedWithInteractiveMode
		}
		if err = h.BeginTx(ctx, nil); err != nil {
			return err
		}
	}
	// rc file
	if rc := env.RCFile(u); !args.NoRC && rc != "" {
		if err = h.Include(rc, false); err != nil && err != text.ErrNoSuchFileOrDirectory {
			return err
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

// vs handles setting vars with predefined var names.
type vs struct {
	vars *[]string
	vals []string
	typ  string
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
	if p.isBool() {
		return "bool"
	}
	return p.typ
}

// isBool satisfies the pflag.boolFlag interface.
func (p vs) isBool() bool {
	return len(p.vals) != 0 && !strings.Contains(p.vals[0], "%")
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

// runCommandOrFiles processes all the supplied commands or files.
func runCommandOrFiles(h *handler.Handler, commandsOrFiles []CommandOrFile) func() error {
	return func() error {
		for _, c := range commandsOrFiles {
			h.SetSingleLineMode(c.Command)
			if c.Command {
				h.Reset([]rune(c.Value))
				if err := h.Run(); err != nil {
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
