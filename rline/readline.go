//go:build !new_readline

package rline

import (
	"io"
	"os"

	"github.com/gohxs/readline"
	"github.com/mattn/go-isatty"
)

var (
	// ErrInterrupt is the interrupt error.
	ErrInterrupt = readline.ErrInterrupt
)

// baseRline should be embedded in a struct implementing the IO interface,
// as it keeps implementation specific state.
type baseRline struct {
	instance *readline.Instance
}

// SetOutput sets the output format func.
func (l *rline) SetOutput(f func(string) string) {
	l.instance.Config.Output = f
}

// New readline input/output handler.
func New(forceNonInteractive bool, out, histfile string) (IO, error) {
	// determine if interactive
	interactive := isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
	cygwin := isatty.IsCygwinTerminal(os.Stdout.Fd()) && isatty.IsCygwinTerminal(os.Stdin.Fd())
	var closers []func() error
	// configure stdin
	var stdin io.ReadCloser
	switch {
	case forceNonInteractive:
		interactive, cygwin = false, false
	case cygwin:
		stdin = os.Stdin
	default:
		stdin = readline.Stdin
	}
	// configure stdout
	var stdout io.WriteCloser
	switch {
	case out != "":
		var err error
		stdout, err = os.OpenFile(out, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		closers = append(closers, stdout.Close)
		interactive = false
	case cygwin:
		stdout = os.Stdout
	default:
		stdout = readline.Stdout
	}
	// configure stderr
	var stderr io.Writer = os.Stderr
	if !cygwin {
		stderr = readline.Stderr
	}
	if interactive {
		// wrap it with cancelable stdin
		stdin = readline.NewCancelableStdin(stdin)
	}
	// create readline instance
	l, err := readline.NewEx(&readline.Config{
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
	})
	if err != nil {
		return nil, err
	}
	closers = append(closers, l.Close)
	n := l.Operation.Runes
	pw := func(prompt string) (string, error) {
		buf, err := l.ReadPassword(prompt)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	if forceNonInteractive {
		n, pw = nil, nil
	}
	return &rline{
		instance: l,
		nextLine: n,
		close: func() error {
			for _, f := range closers {
				_ = f()
			}
			return nil
		},
		stdout:        stdout,
		stderr:        stderr,
		isInteractive: interactive || cygwin,
		prompt:        l.SetPrompt,
		completer: func(a Completer) {
			cfg := l.Config.Clone()
			cfg.AutoComplete = readlineCompleter{c: a}
			l.SetConfig(cfg)
		},
		saveHistory:    l.SaveHistory,
		passwordPrompt: pw,
	}, nil
}

type readlineCompleter struct {
	c Completer
}

func (r readlineCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	return r.c.Complete(line, pos)
}
