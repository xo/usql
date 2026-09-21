// Package capture records terminal sessions from usql running under a
// pseudo-terminal. A pseudo-terminal is a pair of file descriptors that acts
// as a terminal for another program.
//
// usql behaves differently when it does not own a terminal. It turns off the
// prompt, the line editor, completion and colour, so a test that pipes input
// into it exercises none of that. Recording under a pseudo-terminal is the
// only way to test what a person actually sees.
//
// A recording is a Transcript, which pairs each chunk of input with everything
// usql wrote in answer. Transcripts are stored as golden files.
//
// A golden file here records what usql does today, not what it ought to do.
// That is a weaker thing than a golden checked against a reference, and it
// fails in one specific way: the first recording of wrong behaviour freezes
// it, and the test then defends the bug. Before trusting a new golden, break
// the code it covers on purpose and check that the golden fails. A break that
// does not fail it names either a missing case or a line nothing reads.
//
// Linux and macOS record. Every other platform returns ErrUnsupported, because
// Windows has no pseudo-terminal device file and needs a pseudo console built
// with CreatePseudoConsole instead.
//
// The design, and the pseudo-terminal code, come from the capture package in
// github.com/xo/rline.
package capture

import (
	"regexp"
	"time"
)

// Error is an error.
//
// These are constants rather than variables so that nothing can reassign one.
// The text of each is its own name with the Err prefix taken off. Context
// belongs in the wrapping where the error is returned.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrNoSteps is returned when a session carries no input.
	ErrNoSteps Error = "no steps"

	// ErrUnsupported is returned by Record on a platform that cannot record.
	ErrUnsupported Error = "unsupported"
)

// Key sequences that a terminal sends. These name the bytes that the sessions
// send, so that a session reads as a description of what a person types.
const (
	KeyEnter     = "\r"
	KeyTab       = "\t"
	KeyShiftTab  = "\x1b[Z"
	KeyEscape    = "\x1b"
	KeyBackspace = "\x7f"
	KeyUp        = "\x1b[A"
	KeyDown      = "\x1b[B"
	KeyRight     = "\x1b[C"
	KeyLeft      = "\x1b[D"
	KeyHome      = "\x1b[H"
	KeyEnd       = "\x1b[F"

	CtrlA = "\x01" // start of line
	CtrlC = "\x03" // cancel the line
	CtrlD = "\x04" // end of input
	CtrlE = "\x05" // end of line
	CtrlK = "\x0b" // delete to end of line
	CtrlR = "\x12" // search the history
	CtrlU = "\x15" // delete to start of line
	CtrlW = "\x17" // delete the word before the cursor
)

// Default limits, used when a Session leaves them at zero.
const (
	// DefaultQuiet is how long usql must write nothing before the recording
	// moves on. usql answers a query before it prompts again, so this has to
	// outlast a query against a local database.
	DefaultQuiet = 400 * time.Millisecond

	// DefaultLimit is the longest a recording can run.
	DefaultLimit = 30 * time.Second

	// DefaultStart is the longest to wait for usql to write its first byte.
	// It is far longer than DefaultQuiet, because starting is slower than
	// answering: usql links a very large driver set, and it connects to a
	// database before it prompts.
	DefaultStart = 15 * time.Second

	DefaultTerm = "xterm-256color"
	DefaultCols = 80
	DefaultRows = 24
)

// Session describes one terminal session to record.
type Session struct {
	// Name identifies the session and names its golden file.
	Name string

	// About says what the session exercises. It is written into the
	// transcript, so that a reader knows why the recording exists.
	About string

	// Args are the command line arguments given to usql.
	Args []string

	// Env holds extra environment entries, each "NAME=value". Record passes
	// a fixed environment and adds these, so a session cannot depend on the
	// environment of the person who runs it.
	Env []string

	// Terminal settings. These stay fixed so that the output does not change
	// between runs.
	Term string
	Cols int
	Rows int

	// Dir is the directory usql runs in, and the home directory it is given.
	// Empty means a fresh temporary one, so that a history file or a
	// configuration file from an earlier run cannot change the output.
	Dir string

	// Steps are the input chunks, sent in order. Record always collects the
	// output before the first step, so a session does not need an empty step
	// to wait for the banner and the first prompt.
	Steps []Step

	// Quiet is how long usql must write nothing before the next step.
	// Limit is the longest a recording can run.
	Quiet time.Duration
	Limit time.Duration

	// Replies answer queries the program sends to its terminal. Empty means
	// DefaultReplies, which is what every session uses.
	Replies []Reply

	// Start is the longest to wait for the first byte, before the quiet rule
	// takes over.
	//
	// Reading until a program goes quiet cannot tell one that has finished
	// writing from one that has not started. usql is usually still connecting
	// when a short quiet period runs out, and the recording would then be
	// taken to have begun: the first input goes into a terminal nobody reads
	// yet, the terminal echoes it, and usql discards it when it turns raw
	// mode on. Waiting for the first byte closes that.
	Start time.Duration
}

// Step is one chunk of input, and the quiet period that follows it.
type Step struct {
	// Send holds the bytes to write to the terminal.
	Send string

	// Wait overrides the quiet period of the session for this step.
	Wait time.Duration
}

// Reply is a canned answer to a query that the program sends to its terminal.
type Reply struct {
	// Query is the sequence the program writes.
	Query string

	// Answer is what a terminal would write back.
	Answer string
}

// DefaultReplies answer the queries usql sends when it starts.
//
// usql asks the terminal what it is and where the cursor is, then waits for
// the answers before it draws a prompt. A recording that does not answer
// leaves usql blocked until the session hits its limit, and the transcript
// ends part way through the banner. The answers are fixed rather than tracked,
// so that a recording does not depend on where the cursor really was.
var DefaultReplies = []Reply{
	// Primary device attributes. The answer says a VT100 with an advanced
	// video option, which is what most terminals report.
	{Query: "\x1b[0c", Answer: "\x1b[?1;2c"},
	{Query: "\x1b[c", Answer: "\x1b[?1;2c"},
	// Cursor position report, answered as the top left corner.
	{Query: "\x1b[6n", Answer: "\x1b[1;1R"},
}

// Scrubber replaces text that changes between runs.
type Scrubber struct {
	// Match selects the text to replace.
	Match *regexp.Regexp

	// With is the replacement, and may refer to groups as $1.
	With string
}

// Scrub applies every scrubber to b, in order.
func Scrub(b []byte, scrubbers []Scrubber) []byte {
	for _, s := range scrubbers {
		b = s.Match.ReplaceAll(b, []byte(s.With))
	}
	return b
}

// Transcript is a recorded session.
type Transcript struct {
	// The settings that produced the recording.
	Session string
	About   string
	Term    string
	Cols    int
	Rows    int

	// Exchanges are the input chunks and their answers, in order. The first
	// exchange sends nothing and holds whatever usql wrote at startup.
	Exchanges []Exchange
}

// Exchange is one chunk of input and everything usql wrote in answer.
type Exchange struct {
	// Send holds the bytes that went to usql.
	Send []byte

	// Recv holds the bytes that usql wrote before it went quiet.
	Recv []byte
}

// Bytes returns every byte that usql wrote, in order.
func (t *Transcript) Bytes() []byte {
	var out []byte
	for _, e := range t.Exchanges {
		out = append(out, e.Recv...)
	}
	return out
}

// Apply replaces the text that scrubbers select in every answer.
func (t *Transcript) Apply(scrubbers []Scrubber) {
	for i := range t.Exchanges {
		t.Exchanges[i].Recv = Scrub(t.Exchanges[i].Recv, scrubbers)
	}
}

// withDefaults returns a copy of s with every zero limit filled in.
//
// Only Record calls this, and Record records on Linux and macOS alone, so a
// build for any other system has no caller for it.
//
//nolint:unused // used by Record on the systems that can record
func (s Session) withDefaults() Session {
	if s.Term == "" {
		s.Term = DefaultTerm
	}
	if s.Cols == 0 {
		s.Cols = DefaultCols
	}
	if s.Rows == 0 {
		s.Rows = DefaultRows
	}
	if s.Quiet == 0 {
		s.Quiet = DefaultQuiet
	}
	if s.Limit == 0 {
		s.Limit = DefaultLimit
	}
	if s.Start == 0 {
		s.Start = DefaultStart
	}
	if s.Replies == nil {
		s.Replies = DefaultReplies
	}
	return s
}
