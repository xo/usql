//go:build linux || darwin

package capture

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

// Record runs the program at path under a pseudo-terminal, sends the input of
// the session, and returns the transcript.
//
// Record gives the program a fixed environment and an empty home directory, so
// that a history file, a configuration file or an environment variable from
// the machine running the test cannot change the output. It collects the
// output before the first step, so that a session does not need an empty step
// to wait for the banner.
//
// This is the capture package from github.com/xo/rline, with arguments and
// extra environment entries added for usql.
func Record(ctx context.Context, path string, s Session) (*Transcript, error) {
	if len(s.Steps) == 0 {
		return nil, fmt.Errorf("recording the session %s: %w", s.Name, ErrNoSteps)
	}
	s = s.withDefaults()
	// The recording runs in a temporary directory, so a relative path to the
	// program no longer resolves. Make it absolute first.
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving %s: %w", path, err)
	}
	ctx, cancel := context.WithTimeout(ctx, s.Limit)
	defer cancel()
	// A directory the caller named is used as it is and left alone, so that
	// two runs can see each other's files. Otherwise a fresh one is made and
	// taken away afterwards.
	home := s.Dir
	if home == "" {
		made, err := os.MkdirTemp("", "usql-capture-")
		if err != nil {
			return nil, fmt.Errorf("creating home directory: %w", err)
		}
		defer func() { _ = os.RemoveAll(made) }()
		home = made
	}
	leader, follower, err := OpenPTY()
	if err != nil {
		return nil, err
	}
	defer leader.Close()
	if err := setWinsize(leader, s.Cols, s.Rows); err != nil {
		_ = follower.Close()
		return nil, err
	}
	cmd := exec.CommandContext(ctx, path, s.Args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = follower, follower, follower
	cmd.Dir = home
	// A whitelist rather than os.Environ, so that nothing on the machine
	// running the test reaches usql. USQL_* variables in particular would
	// change the prompt and the output format.
	cmd.Env = append([]string{
		"TERM=" + s.Term,
		"HOME=" + home,
		"PATH=" + os.Getenv("PATH"),
	}, s.Env...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		Ctty:    0,
	}
	if err := cmd.Start(); err != nil {
		_ = follower.Close()
		return nil, fmt.Errorf("starting %s: %w", path, err)
	}
	// The child owns the follower side now. The parent must let go of it, or
	// the reads below never see the end of the stream.
	_ = follower.Close()
	t := &Transcript{
		Session: s.Name,
		About:   s.About,
		Term:    s.Term,
		Cols:    s.Cols,
		Rows:    s.Rows,
	}
	recErr := run(leader, s, t)
	_ = cmd.Wait()
	return t, recErr
}

// run collects the startup output, then sends every step and collects its
// answer. It appends each exchange to the transcript as it goes, so that a
// recording that fails still shows how far it reached.
func run(leader *os.File, s Session, t *Transcript) error {
	// Wait for usql to say something before the quiet rule applies, so that a
	// program that has not started yet is not mistaken for one that has
	// finished writing.
	b, done, err := drain(leader, s.Replies, s.Quiet, s.Start)
	t.Exchanges = append(t.Exchanges, Exchange{Recv: b})
	if err != nil || done {
		return err
	}
	for i, step := range s.Steps {
		if step.Send != "" {
			if _, err := leader.WriteString(step.Send); err != nil {
				return fmt.Errorf("sending step %d: %w", i, err)
			}
		}
		quiet := step.Wait
		if quiet == 0 {
			quiet = s.Quiet
		}
		b, done, err := drain(leader, s.Replies, quiet, quiet)
		t.Exchanges = append(t.Exchanges, Exchange{Send: []byte(step.Send), Recv: b})
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
	return nil
}

// drain reads until the program writes nothing for the quiet period. It
// reports whether the stream ended.
//
// first is how long to wait for the first byte, which is longer than quiet
// when the program is still starting. Once anything has arrived, the quiet
// period decides when the program has finished.
//
// drain answers terminal queries as it reads. A program that asks where the
// cursor is waits for the answer, and without one it never writes again, so
// the quiet period expires and the recording stops part way through.
func drain(leader *os.File, replies []Reply, quiet, first time.Duration) ([]byte, bool, error) {
	var out []byte
	// answered is how much of out has been searched for queries. A query can
	// straddle two reads, so the search starts a little before it.
	answered := 0
	buf := make([]byte, 4096)
	for {
		wait := quiet
		if len(out) == 0 {
			wait = first
		}
		if err := leader.SetReadDeadline(time.Now().Add(wait)); err != nil {
			return out, false, fmt.Errorf("setting read deadline: %w", err)
		}
		n, err := leader.Read(buf)
		out = append(out, buf[:n]...)
		if n != 0 {
			answered = answer(leader, replies, out, answered)
		}
		switch {
		case err == nil:
			continue
		case errors.Is(err, os.ErrDeadlineExceeded):
			return out, false, nil
		case isEnd(err):
			// The child closed the terminal, which Linux reports as EIO.
			return out, true, nil
		default:
			return out, false, fmt.Errorf("reading terminal: %w", err)
		}
	}
}

// answer writes a reply for every query in out that has not been answered
// yet, and returns how much of out has now been searched.
//
// The search starts a few bytes before the last position, because a query can
// arrive split across two reads. A query found in that overlap is answered
// twice, which is why the overlap is only as long as the longest query.
func answer(leader *os.File, replies []Reply, out []byte, from int) int {
	longest := 0
	for _, r := range replies {
		if len(r.Query) > longest {
			longest = len(r.Query)
		}
	}
	start := from - longest + 1
	if start < 0 {
		start = 0
	}
	window := out[start:]
	for _, r := range replies {
		for i := 0; ; {
			j := bytes.Index(window[i:], []byte(r.Query))
			if j < 0 {
				break
			}
			i += j + len(r.Query)
			if start+i <= from {
				// Already answered on an earlier pass.
				continue
			}
			// A failed write is not fatal: the program may have exited
			// between the query and the answer, and the read that follows
			// reports the end of the stream.
			_, _ = leader.WriteString(r.Answer)
		}
	}
	return len(out)
}

// isEnd reports whether err means the program closed the terminal.
//
// The two systems report it differently. Linux fails the read with EIO once
// the last follower is closed. macOS returns the end of the stream instead,
// which Go reports as io.EOF.
func isEnd(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, os.ErrClosed) ||
		errors.Is(err, syscall.EIO) ||
		errors.Is(err, syscall.EBADF)
}

// winsize matches the struct that the TIOCSWINSZ request expects.
type winsize struct {
	rows uint16
	cols uint16
	x    uint16
	y    uint16
}

// setWinsize fixes the terminal size, so that the output does not depend on
// the window of the person who runs the recording.
func setWinsize(leader *os.File, cols, rows int) error {
	ws := winsize{rows: uint16(rows), cols: uint16(cols)}
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, leader.Fd(),
		syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(&ws))); errno != 0 {
		return fmt.Errorf("setting terminal size: %w", errno)
	}
	return nil
}
