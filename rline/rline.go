// Package rline provides a readline implementation for usql.
package rline

import (
	"errors"
	"io"
	"os"

	"github.com/chzyer/readline"
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
	// Completer sets the auto-completer.
	Completer(readline.AutoCompleter)
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
	N    func() ([]rune, error)
	C    func() error
	Out  io.Writer // This is the currently active stdout writer (could be filtered)
	Err  io.Writer // This is the currently active stderr writer
	originalStdout io.WriteCloser // The base stdout writer, used for filtering
	originalStderr io.Writer      // The base stderr writer
	Int  bool
	Cyg  bool
	P    func(string)
	A    func(readline.AutoCompleter)
	S    func(string) error
	Pw   func(string) (string, error)
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

// Completer sets the auto-completer.
func (l *Rline) Completer(a readline.AutoCompleter) {
	if l.A != nil {
		l.A(a)
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

// filterWriter implements io.Writer and applies a string filter.
type filterWriter struct {
	writer io.Writer
	filter func(string) string
}

func (fw *filterWriter) Write(p []byte) (n int, err error) {
	if fw.filter == nil {
		return fw.writer.Write(p)
	}
	filtered := fw.filter(string(p))
	return fw.writer.Write([]byte(filtered))
}

// Close implements io.Closer if the underlying writer is an io.Closer.
func (fw *filterWriter) Close() error {
	if closer, ok := fw.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// SetOutput sets the output format func.
func (l *Rline) SetOutput(f func(string) string) {
	cfg := l.Inst.Config.Clone() // Get a mutable copy of the config

	var currentWriter io.WriteCloser
	if f != nil {
		// Wrap the original stdout with the filter
		currentWriter = &filterWriter{writer: l.originalStdout, filter: f}
	} else {
		// If filter is nil, revert to the original stdout
		currentWriter = l.originalStdout
	}

	// Update both the Rline's exposed writer and the readline instance's config
	l.Out = currentWriter
	cfg.Stdout = currentWriter
	l.Inst.SetConfig(cfg) // Apply the modified config to the readline instance
}

// New creates a new readline input/output handler.
func New(interactive, cygwin, forceNonInteractive bool, out, histfile string) (IO, error) {
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
		Stdout:                 stdout, // Pass the original stdout
		Stderr:                 stderr, // Pass the original stderr
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
	return &Rline{
		Inst: l,
		N:    n,
		C: func() error {
			for _, f := range closers {
				_ = f()
			}
			return nil
		},
		Out:            stdout, // Rline.Stdout() returns this original writer initially
		Err:            stderr, // Rline.Stderr() returns this original writer
		originalStdout: stdout, // Store the original for filtering
		originalStderr: stderr, // Store the original for filtering (though SetOutput only affects stdout)
		Int: interactive || cygwin,
		Cyg: cygwin,
		P:   l.SetPrompt,
		A: func(a readline.AutoCompleter) {
			cfg := l.Config.Clone()
			cfg.AutoComplete = a
			l.SetConfig(cfg)
		},
		S:  l.SaveHistory,
		Pw: pw,
	}, nil
}
