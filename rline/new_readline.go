//go:build new_readline

package rline

import (
	"github.com/reeflective/readline/inputrc"
	"golang.org/x/term"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattn/go-isatty"
	"github.com/reeflective/readline"
)

var (
	// ErrInterrupt is the interrupt error.
	ErrInterrupt = readline.ErrInterrupt
)

// baseRline should be embedded in a struct implementing the IO interface,
// as it keeps implementation specific state.
type baseRline struct {
	instance *readline.Shell
	prompt   string
}

// Prompt sets the prompt for the next interactive line read.
func (l *rline) Prompt(s string) {
	l.prompt = s
}

// Completer sets the auto-completer.
func (l *rline) Completer(a Completer) {
	l.instance.Completer = func(line []rune, cursor int) readline.Completions {
		candidates, _ := a.Complete(line, cursor)
		values := make([]string, len(candidates))
		for candidate := range candidates {
			values = append(values, string(candidate))
		}
		return readline.CompleteValues(values...)
	}
}

// SetOutput sets the output format func.
func (l *rline) SetOutput(f func(string) string) {
	l.instance.SyntaxHighlighter = func(line []rune) string {
		return f(string(line))
	}
}

// New readline input/output handler.
func New(forceNonInteractive bool, out, histfile string) (IO, error) {
	// determine if interactive
	interactive, cygwin := false, false
	if !forceNonInteractive {
		interactive = isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
		cygwin = isatty.IsCygwinTerminal(os.Stdout.Fd()) && isatty.IsCygwinTerminal(os.Stdin.Fd())
	}
	var stdout io.WriteCloser
	var closers []func() error
	switch {
	case out != "":
		var err error
		stdout, err = os.OpenFile(out, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		closers = append(closers, stdout.Close)
		interactive = false
	default:
		stdout = os.Stdout
	}
	// configure stderr
	var stderr io.Writer = os.Stderr
	// TODO handle interrupts?
	options := []inputrc.Option{inputrc.WithName("usql")}
	/*
		&readline.Config{
				HistoryFile:            histfile,
				DisableAutoSaveHistory: true,
				InterruptPrompt:        "^C",
				HistorySearchFold:      true,
				Stdin:                  stdin,
				Stdout:                 stdout,
				Stderr:                 stderr,
				FuncIsTerminal: func() bool {
					return interactive || cygwin
				},
				FuncFilterInputRune: func(r rune) (rune, bool) {
					if r == readline.CharCtrlZ {
						return r, false
					}
					return r, true
				},
			}
	*/
	// create readline instance
	shell := readline.NewShell(options...)
	var history readline.History
	if histfile != "" {
		history, err := readline.NewHistoryFromFile(histfile)
		if err != nil {
			return nil, err
		}
		shell.History.Add("default", history)
	}

	n := func() ([]rune, error) {
		line, err := shell.Readline()
		return []rune(line), err
	}
	pw := func(prompt string) (string, error) {
		_, err := shell.Printf(prompt)
		if err != nil {
			return "", err
		}
		return readPassword()
	}
	if forceNonInteractive {
		n, pw = nil, nil
	}
	result := &rline{
		baseRline: baseRline{instance: shell},
		nextLine:  n,
		close: func() error {
			for _, f := range closers {
				_ = f()
			}
			return nil
		},
		stdout:         stdout,
		stderr:         stderr,
		isInteractive:  interactive || cygwin,
		passwordPrompt: pw,
	}
	shell.Prompt.Primary(func() string {
		return result.prompt
	})
	if history != nil {
		result.saveHistory = func(input string) error {
			_, err := history.Write(input)
			return err
		}
	}
	return result, nil
}

func readPassword() (string, error) {
	stdin := syscall.Stdin
	oldState, err := term.GetState(stdin)
	if err != nil {
		return "", err
	}
	defer term.Restore(stdin, oldState)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	go func() {
		for _ = range sigch {
			term.Restore(stdin, oldState)
			os.Exit(1)
		}
	}()

	password, err := term.ReadPassword(stdin)
	if err != nil {
		return "", err
	}
	return string(password), nil
}
