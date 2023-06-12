// Package rline provides a readline implementation for usql.
package rline

import (
	"bufio"
	"bytes"
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
	// Prompt sets the prompt for the next interactive line read.
	Prompt(string)
	// Completer sets the auto-completer.
	Completer(Completer)
	// Save saves a line of history.
	Save(string) error
	// Password prompts for a password.
	Password(string) (string, error)
	// SetOutput sets the output filter func.
	SetOutput(func(string) string)
}

// Completer returns candidates matching current input
type Completer interface {
	// Complete current input with matching commands
	Complete(line []rune, pos int) (newLine [][]rune, length int)
}

// rline provides a type compatible with the IO interface.
type rline struct {
	instance       *readline.Instance
	nextLine       func() ([]rune, error)
	close          func() error
	stdout         io.Writer
	stderr         io.Writer
	isInteractive  bool
	prompt         func(string)
	completer      func(Completer)
	saveHistory    func(string) error
	passwordPrompt passwordPrompt
}

type passwordPrompt func(string) (string, error)

// Next returns the next line of runes (excluding '\n') from the input.
func (l *rline) Next() ([]rune, error) {
	if l.nextLine != nil {
		return l.nextLine()
	}
	return nil, io.EOF
}

// Close closes the IO.
func (l *rline) Close() error {
	if l.close != nil {
		return l.close()
	}
	return nil
}

// Stdout is the IO's standard out.
func (l *rline) Stdout() io.Writer {
	return l.stdout
}

// Stderr is the IO's standard error out.
func (l *rline) Stderr() io.Writer {
	return l.stderr
}

// Interactive determines if the IO is an interactive terminal.
func (l *rline) Interactive() bool {
	return l.isInteractive
}

// Prompt sets the prompt for the next interactive line read.
func (l *rline) Prompt(s string) {
	if l.prompt != nil {
		l.prompt(s)
	}
}

// Completer sets the auto-completer.
func (l *rline) Completer(a Completer) {
	if l.completer != nil {
		l.completer(a)
	}
}

// Save saves a line of history.
func (l *rline) Save(s string) error {
	if l.saveHistory != nil {
		return l.saveHistory(s)
	}
	return nil
}

// Password prompts for a password.
func (l *rline) Password(prompt string) (string, error) {
	if l.passwordPrompt != nil {
		return l.passwordPrompt(prompt)
	}
	return "", ErrPasswordNotAvailable
}

// SetOutput sets the output format func.
func (l *rline) SetOutput(f func(string) string) {
	l.instance.Config.Output = f
}

// New creates a new readline input/output handler.
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

func NewFromReader(reader *bufio.Reader, out, err io.Writer, pw passwordPrompt) IO {
	return &rline{
		nextLine: func() ([]rune, error) {
			buf := new(bytes.Buffer)
			var b []byte
			var isPrefix bool
			var err error
			for {
				// read
				b, isPrefix, err = reader.ReadLine()
				// when not EOF
				if err != nil && err != io.EOF {
					return nil, err
				}
				// append
				if _, werr := buf.Write(b); werr != nil {
					return nil, werr
				}
				// end of line
				if !isPrefix || err != nil {
					break
				}
			}
			// peek and read possible line ending \n or \r\n
			if err != io.EOF {
				if err := peekEnding(buf, reader); err != nil {
					return nil, err
				}
			}
			return []rune(buf.String()), err
		},
		stdout:         out,
		stderr:         err,
		passwordPrompt: pw,
	}
}

// peekEnding peeks to see if the next successive bytes in r is \n or \r\n,
// writing to w if it is. Does not advance r if the next bytes are not \n or
// \r\n.
func peekEnding(w io.Writer, r *bufio.Reader) error {
	// peek first byte
	buf, err := r.Peek(1)
	switch {
	case err != nil && err != io.EOF:
		return err
	case err == nil && buf[0] == '\n':
		if _, rerr := r.ReadByte(); err != nil && err != io.EOF {
			return rerr
		}
		_, werr := w.Write([]byte{'\n'})
		return werr
	case err == nil && buf[0] != '\r':
		return nil
	}
	// peek second byte
	buf, err = r.Peek(1)
	switch {
	case err != nil && err != io.EOF:
		return err
	case err == nil && buf[0] != '\n':
		return nil
	}
	if _, rerr := r.ReadByte(); err != nil && err != io.EOF {
		return rerr
	}
	_, werr := w.Write([]byte{'\n'})
	return werr
}
