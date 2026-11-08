package feed

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/xo/transit"
)

// LineReader reads one line or one entry of input. A *rline.Prompt is one:
// at a terminal it returns a whole entry, which may hold several rows, and
// otherwise it returns one line.
type LineReader interface {
	ReadLine(prompt string) (string, error)
}

// Source feeds the items of an input to the handler, one at a time.
//
// It keeps the text that it has read and not yet fed, parses that text, and
// feeds an item once nothing more can change it. A statement that ends with a
// ';' is complete, and so is a meta command once its line has ended. The last
// item of the text waits for more input when it is a statement with no ';'.
//
// At a terminal the line editor hands back a whole entry that is finished, so
// every item of the entry is fed. A script, a pipe or an included file is read
// one line at a time, and the same rule decides when to feed.
type Source struct {
	r      LineReader
	p      *Parser
	prompt func() string

	// text is what was read and not fed yet.
	text string
	// more holds the lines read since text was last parsed. Next adds them
	// to text once before it parses, because adding each line to text
	// copies all of text.
	more strings.Builder
	// grown says that text only grew since it was last parsed, so that the
	// parser reads only what was added.
	grown bool
	// eof is set once the reader has no more input.
	eof bool
	// line is how many lines were read, for an error that names a line.
	line int
	// queue holds the items of the last parse that can be fed and are not
	// fed yet. Their offsets are in the text that was parsed, and cut is how
	// much of that text was fed since.
	queue []Item
	cut   int
	// alone reports whether a line, read when nothing waits to be fed, is
	// fed at once as a statement with no ';', as psql reads help, quit and
	// exit typed alone. It is nil unless SetAlone set it.
	alone func(line string) bool
}

// SetAlone sets the test for a line that is fed at once when nothing else
// waits to be fed, such as help typed alone at a terminal.
func (s *Source) SetAlone(alone func(line string) bool) {
	s.alone = alone
}

// NewSource returns a Source that reads r, and parses it in the language.
// prompt returns the prompt for the next read, and may be nil.
func NewSource(r LineReader, lang *transit.Language, prompt func() string) (*Source, error) {
	p, err := NewParser(lang)
	if err != nil {
		return nil, err
	}
	return &Source{r: r, p: p, prompt: prompt}, nil
}

// Parser returns the parser of the source.
func (s *Source) Parser() *Parser {
	return s.p
}

// SetLanguage changes the language of the source. The text that is not fed
// yet is parsed again in the new language.
func (s *Source) SetLanguage(lang *transit.Language) error {
	s.queue = nil
	return s.p.SetLanguage(lang)
}

// Push puts text in front of the input that is not read yet, as when \e hands
// back an edited query.
func (s *Source) Push(text string) {
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	s.text, s.grown, s.queue = text+s.text, false, nil
}

// Reset drops the text that is not fed yet, as after an interrupt.
func (s *Source) Reset() {
	s.text, s.grown, s.queue = "", false, nil
}

// ReadLine reads one line of raw input, as \prompt does. The line comes from
// the text that is read and not fed yet, if any is left after the end of the
// line that held the meta command, and otherwise from the reader.
func (s *Source) ReadLine(prompt string) (string, error) {
	s.queue = nil
	if t := strings.TrimPrefix(s.text, "\n"); t != s.text {
		s.text, s.grown = t, false
	}
	if s.text != "" {
		line, rest, _ := strings.Cut(s.text, "\n")
		s.text, s.grown = rest, false
		return line, nil
	}
	if s.eof {
		return "", io.EOF
	}
	return s.r.ReadLine(prompt)
}

// Next returns the next item. At the end of the input it returns io.EOF. A
// statement with no ';' at the end of the input is returned before io.EOF.
//
// A parse of the text that is not fed yet finds each item that can be fed.
// Next feeds them one for each call, and cuts the text after each one. A
// change of language, a pushed text or a line taken by ReadLine drops the
// items that are not fed, so it applies to the next item. A line of many
// statements is then parsed once, and not once for each statement.
//
// A line can end an item only when canEnd says so, so the text is parsed
// again only after such a line, or at the end of the input. A
// statement of many lines is then not parsed once for each of its lines.
func (s *Source) Next() (Item, error) {
	if len(s.queue) != 0 {
		it := s.queue[0]
		s.queue = s.queue[1:]
		return s.feed(it), nil
	}
	for {
		items, err := s.p.readyItems(s.text, s.grown, s.ready)
		if err != nil {
			return Item{}, err
		}
		if len(items) != 0 {
			s.queue, s.cut = items[1:], 0
			return s.feed(items[0]), nil
		}
		for {
			if s.eof {
				s.text = ""
				return Item{}, io.EOF
			}
			line, err := s.read()
			if err != nil {
				return Item{}, err
			}
			if s.alone != nil && !s.eof && s.text == "" && s.more.Len() == len(line)+1 && s.alone(line) {
				s.more.Reset()
				return Item{Kind: Statement, Text: line}, nil
			}
			if s.eof || canEnd(line) {
				break
			}
		}
		s.text += s.more.String()
		s.more.Reset()
		s.grown = true
	}
}

// feed cuts the text after it, an item of the last parse, and returns it with
// its offsets in the text that is left, as a parse of that text gives them.
func (s *Source) feed(it Item) Item {
	it.Start, it.End = it.Start-s.cut, it.End-s.cut
	s.text, s.grown, s.cut = s.text[it.End:], false, s.cut+it.End
	return it
}

// canEnd reports whether a line can end an item: it holds a ';' or a '\'.
func canEnd(line string) bool {
	return strings.ContainsAny(line, ";\\")
}

// ready returns how many of n items can be fed now, where last is the last of
// them.
func (s *Source) ready(n int, last Item) int {
	if n == 0 {
		return 0
	}
	switch {
	case s.eof, last.Kind == Statement && last.Terminated:
		return n
	case last.Kind == Command && strings.Contains(s.text[last.End:], "\n"):
		return n
	}
	// the last item can still grow
	return n - 1
}

// read reads one more line or entry, and returns it. A line that is not valid
// UTF-8 is an error that wraps ErrInvalidUTF8 and names the line.
func (s *Source) read() (string, error) {
	var prompt string
	if s.prompt != nil {
		prompt = s.prompt()
	}
	line, err := s.r.ReadLine(prompt)
	switch {
	case errors.Is(err, io.EOF):
		s.eof = true
		return "", nil
	case err != nil:
		return "", err
	}
	s.line += strings.Count(line, "\n") + 1
	if !utf8.ValidString(line) {
		return "", fmt.Errorf("line %d: %w", s.line, ErrInvalidUTF8)
	}
	s.more.WriteString(line)
	s.more.WriteByte('\n')
	return line, nil
}

// Lines is a LineReader of the lines of r, as for a script or an included
// file. It writes no prompt.
type Lines struct {
	r *bufio.Reader
}

// NewLines returns a LineReader of the lines of r.
func NewLines(r io.Reader) *Lines {
	return &Lines{r: bufio.NewReader(r)}
}

// ReadLine returns the next line, without its line ending.
func (l *Lines) ReadLine(string) (string, error) {
	line, err := l.r.ReadString('\n')
	if err != nil && (!errors.Is(err, io.EOF) || line == "") {
		return "", err
	}
	line = strings.TrimSuffix(line, "\n")
	return strings.TrimSuffix(line, "\r"), nil
}
