package rline

import (
	"errors"
	"io"
	"os"

	"github.com/gohxs/readline"
	isatty "github.com/mattn/go-isatty"
)

var (
	// ErrInterrupt is the interrupt error.
	ErrInterrupt = readline.ErrInterrupt

	// ErrPasswordNotAvailable is the password not available error.
	ErrPasswordNotAvailable = errors.New("password not available")
)

// IO is the common input/output interface.
type IO interface {
	// Next returns the next line of runes (excluding '\n') from the input.
	Next() ([]rune, error)

	// Close closes the IO.
	Close() error

	// Stdout is the IO's standard out.
	Stdout() io.Writer

	// Stderr is the IO's standard error out.
	Stderr() io.Writer

	// Interactive determines if the IO is an interactive terminal.
	Interactive() bool

	// Cygwin determines if the IO is a Cygwin interactive terminal.
	Cygwin() bool

	// Prompt sets the prompt for the next interactive line read.
	Prompt(string)

	// Save saves a line of history.
	Save(string) error

	// Password prompts for a password.
	Password(string) (string, error)

	// SetOutput sets the output filter func.
	SetOutput(func(string) string)
}

// Rline provides a type compatible with the IO interface.
type Rline struct {
	Inst *readline.Instance

	N func() ([]rune, error)
	C func() error

	Out io.Writer
	Err io.Writer

	Int bool
	Cyg bool
	P   func(string)
	S   func(string) error
	Pw  func(string) (string, error)
}

// Next returns the next line of runes (excluding '\n') from the input.
func (l *Rline) Next() ([]rune, error) {
	if l.N != nil {
		return l.N()
	}

	return nil, io.EOF
}

// Close closes the IO.
func (l *Rline) Close() error {
	if l.C != nil {
		return l.C()
	}
	return nil
}

// Stdout is the IO's standard out.
func (l *Rline) Stdout() io.Writer {
	return l.Out
}

// Stderr is the IO's standard error out.
func (l *Rline) Stderr() io.Writer {
	return l.Err
}

// Interactive determines if the IO is an interactive terminal.
func (l *Rline) Interactive() bool {
	return l.Int
}

// Cygwin determines if the IO is a Cygwin interactive terminal.
func (l *Rline) Cygwin() bool {
	return l.Cyg
}

// Prompt sets the prompt for the next interactive line read.
func (l *Rline) Prompt(s string) {
	if l.P != nil {
		l.P(s)
	}
}

// Save saves a line of history.
func (l *Rline) Save(s string) error {
	if l.S != nil {
		return l.S(s)
	}

	return nil
}

// Password prompts for a password.
func (l *Rline) Password(prompt string) (string, error) {
	if l.Pw != nil {
		return l.Pw(prompt)
	}

	return "", ErrPasswordNotAvailable
}

// SetOutput sets the output format func.
func (l *Rline) SetOutput(f func(string) string) {
	l.Inst.Config.Output = f
}

// New creates a new readline input/output handler.
func New(forceNonInteractive bool, out, histfile string) (IO, error) {
	var err error

	// determine if interactive
	interactive := isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
	cygwin := isatty.IsCygwinTerminal(os.Stdout.Fd()) && isatty.IsCygwinTerminal(os.Stdin.Fd())

	var closers []func() error

	// configure stdin
	var stdin io.ReadCloser
	if forceNonInteractive {
		interactive, cygwin = false, false
	} else if cygwin {
		stdin = os.Stdin
	} else {
		stdin = readline.Stdin
	}

	// configure stdout
	var stdout io.WriteCloser
	if out != "" {
		stdout, err = os.OpenFile(out, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		closers = append(closers, stdout.Close)

		interactive = false
	} else if cygwin {
		stdout = os.Stdout
	} else {
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
	if forceNonInteractive {
		n = nil
	}

	return &Rline{
		Inst: l,
		N:    n,
		C: func() error {
			for _, f := range closers {
				f()
			}
			return nil
		},
		Out: stdout,
		Err: stderr,
		Int: interactive || cygwin,
		Cyg: cygwin,
		P:   l.SetPrompt,
		S:   l.SaveHistory,
		Pw: func(prompt string) (string, error) {
			buf, err := l.ReadPassword(prompt)
			if err != nil {
				return "", err
			}
			return string(buf), nil
		},
	}, nil
}
