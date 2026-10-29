// Package pager sends output to a pager program, such as less, when the output
// does not fit on the terminal.
//
// A [Writer] holds what is written to it until the output is taller or wider
// than the terminal. Then it starts the pager, writes the held output to the
// standard input of the pager, and writes everything after that to the pager
// as well. Output that fits goes to the terminal when the Writer closes.
//
// The pager writes to the terminal itself, not through the line editor,
// because a pager such as less acts like cat when its output is not a
// terminal.
package pager

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Config says when a Writer starts the pager, and how.
type Config struct {
	// Command is the pager command, such as "less -S". The shell runs it, so
	// it can hold arguments and quotes.
	Command string
	// Shell is the shell and its argument that run Command, such as
	// {"sh", "-c"}.
	Shell []string
	// Always starts the pager for any output, as \pset pager always does.
	Always bool
	// Width and Height are the size of the terminal. Output that has a line
	// wider than Width, or that has Height lines or more, starts the pager.
	// Zero turns off that test.
	Width, Height int
	// MinLines is the number of lines the output must have before the pager
	// starts, as \pset pager_min_lines sets. It applies to both tests.
	MinLines int
	// Stdout and Stderr are where the pager writes. They are the terminal.
	Stdout, Stderr *os.File
}

// Writer is an io.WriteCloser that sends what is written to it to the pager,
// or to its output when the output fits on the terminal.
//
// A Writer is used for one command, and must be closed when the command ends.
type Writer struct {
	out io.Writer
	cfg Config

	// held is the output kept until the Writer decides.
	held bytes.Buffer
	// lines is the number of lines that ended, and col is the width of the
	// line that has not ended.
	lines, col int
	// wide is set when a line was wider than the terminal.
	wide bool
	// esc is the state of an escape sequence that is in progress.
	esc escState
	// partial holds the bytes of a rune that the last write split.
	partial []byte

	// direct is set when the output goes straight to out. That happens when
	// paging is off, and when the pager could not start.
	direct bool
	// err is the error from starting the pager, kept for Close.
	err error

	cmd   *exec.Cmd
	stdin io.WriteCloser
	// gone is set when the pager stopped reading, because the user quit it.
	gone bool
}

// New returns a Writer that writes to out, or to the pager that cfg describes.
//
// When cfg has no Command or no terminal, or when the size of the terminal is
// not known and cfg.Always is not set, the Writer writes straight to out.
func New(out io.Writer, cfg Config) *Writer {
	w := &Writer{out: out, cfg: cfg}
	w.direct = cfg.Command == "" || len(cfg.Shell) == 0 || cfg.Stdout == nil ||
		(!cfg.Always && cfg.Width <= 0 && cfg.Height <= 0)
	return w
}

// Write satisfies io.Writer.
//
// When the user quits the pager before the output ends, Write discards the
// rest of the output and does not report an error.
func (w *Writer) Write(p []byte) (int, error) {
	switch {
	case w.direct:
		return w.out.Write(p)
	case w.cmd != nil:
		return w.toPager(p)
	}
	w.held.Write(p)
	w.measure(p)
	if !w.fits() {
		// Close reports a pager that did not start.
		_ = w.start()
	}
	return len(p), nil
}

// Close writes the output when it fits, or waits for the pager to exit.
func (w *Writer) Close() error {
	switch {
	case w.direct:
		if w.held.Len() != 0 {
			_, _ = w.out.Write(w.held.Bytes())
			w.held.Reset()
		}
		return w.err
	case w.cmd == nil:
		// A last line with no newline is still a line.
		if w.col != 0 {
			w.lines, w.col = w.lines+1, 0
		}
		if w.fits() {
			w.direct = true
			return w.Close()
		}
		if err := w.start(); err != nil {
			return w.Close()
		}
	}
	_ = w.stdin.Close()
	err := w.cmd.Wait()
	w.cmd, w.direct = nil, true
	if err != nil {
		return fmt.Errorf("running the pager %q: %w", w.cfg.Command, err)
	}
	return nil
}

// fits reports whether the output so far fits on the terminal.
func (w *Writer) fits() bool {
	switch {
	case w.cfg.Always:
		return w.held.Len() == 0
	case w.lines < w.cfg.MinLines:
		return true
	}
	return !w.wide && (w.cfg.Height <= 0 || w.lines < w.cfg.Height)
}

// start starts the pager and gives it the output held so far. When the pager
// does not start, the Writer writes the output straight to out from then on,
// and Close reports the error.
func (w *Writer) start() error {
	args := append(w.cfg.Shell[1:len(w.cfg.Shell):len(w.cfg.Shell)], w.cfg.Command)
	cmd := exec.Command(w.cfg.Shell[0], args...)
	cmd.Stdout, cmd.Stderr = w.cfg.Stdout, w.cfg.Stderr
	stdin, err := cmd.StdinPipe()
	if err == nil {
		err = cmd.Start()
	}
	if err != nil {
		w.direct = true
		w.err = fmt.Errorf("starting the pager %q: %w", w.cfg.Command, err)
		_, _ = w.out.Write(w.held.Bytes())
		w.held.Reset()
		return w.err
	}
	w.cmd, w.stdin = cmd, stdin
	_, _ = w.toPager(w.held.Bytes())
	w.held.Reset()
	return nil
}

// toPager writes p to the pager. A pager that stopped reading closes its end
// of the pipe, and the rest of the output is then thrown away.
func (w *Writer) toPager(p []byte) (int, error) {
	if w.gone {
		return len(p), nil
	}
	if _, err := w.stdin.Write(p); err != nil {
		if errors.Is(err, syscall.EPIPE) || errors.Is(err, os.ErrClosed) {
			w.gone = true
			return len(p), nil
		}
		return 0, err
	}
	return len(p), nil
}

// escState is where measure is in an escape sequence.
type escState int

// The escape states. A CSI sequence ends with a byte from 0x40 to 0x7e. An
// OSC, DCS or APC string ends with BEL or with ESC \.
const (
	escNone escState = iota
	escStart
	escCSI
	escString
	escStringEnd
)

// measure counts the lines in p and the width of the line that has not ended.
// It skips escape sequences, because they take no room on the screen.
func (w *Writer) measure(p []byte) {
	if len(w.partial) != 0 {
		p = append(w.partial, p...)
		w.partial = nil
	}
	for len(p) != 0 {
		c := p[0]
		switch w.esc {
		case escStart:
			switch c {
			case '[':
				w.esc = escCSI
			case ']', 'P', '_', '^', 'X':
				w.esc = escString
			default:
				w.esc = escNone
			}
			p = p[1:]
			continue
		case escCSI:
			if 0x40 <= c && c <= 0x7e {
				w.esc = escNone
			}
			p = p[1:]
			continue
		case escString:
			switch c {
			case '\a':
				w.esc = escNone
			case 0x1b:
				w.esc = escStringEnd
			}
			p = p[1:]
			continue
		case escStringEnd:
			w.esc = escNone
			if c != '\\' {
				w.esc = escString
			}
			p = p[1:]
			continue
		}
		switch c {
		case 0x1b:
			w.esc = escStart
		case '\n':
			w.lines, w.col = w.lines+1, 0
		case '\r':
			w.col = 0
		case '\t':
			w.col += 8 - w.col%8
		default:
			if c < utf8.RuneSelf {
				if c >= ' ' && c != 0x7f {
					w.col++
				}
				break
			}
			if !utf8.FullRune(p) {
				w.partial = append([]byte(nil), p...)
				return
			}
			r, n := utf8.DecodeRune(p)
			w.col += runewidth.RuneWidth(r)
			p = p[n:]
			w.checkWidth()
			continue
		}
		p = p[1:]
		w.checkWidth()
	}
}

// checkWidth notes a line that is wider than the terminal.
func (w *Writer) checkWidth() {
	if w.cfg.Width > 0 && w.col > w.cfg.Width {
		w.wide = true
	}
}
