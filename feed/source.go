package feed

import (
	"bufio"
	"errors"
	"io"
	"strings"

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
	return s.p.SetLanguage(lang)
}

// Push puts text in front of the input that is not read yet, as when \e hands
// back an edited query.
func (s *Source) Push(text string) {
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	s.text, s.grown = text+s.text, false
}

// Reset drops the text that is not fed yet, as after an interrupt.
func (s *Source) Reset() {
	s.text, s.grown = "", false
}

// ReadLine reads one line of raw input, as \prompt does. The line comes from
// the text that is read and not fed yet, if any is left after the end of the
// line that held the meta command, and otherwise from the reader.
func (s *Source) ReadLine(prompt string) (string, error) {
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
// Each call parses the text that is not fed yet, feeds its first item, and
// cuts the text after it. So a change of language, a pushed text or a line
// taken by ReadLine applies to the next item, with nothing parsed ahead.
//
// A line can end an item only when canEnd says so, so the text is parsed
// again only after such a line, or at the end of the input. A
// statement of many lines is then not parsed once for each of its lines.
func (s *Source) Next() (Item, error) {
	for {
		first, ok, err := s.p.first(s.text, s.grown, func(n int, last Item) bool {
			return s.ready(n, last) != 0
		})
		if err != nil {
			return Item{}, err
		}
		if ok {
			s.text, s.grown = s.text[first.End:], false
			return first, nil
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
			if s.eof || canEnd(line) {
				break
			}
		}
		s.text += s.more.String()
		s.more.Reset()
		s.grown = true
	}
}

// canEnd reports whether a line can end an item: it holds a ';' or a '\', or
// it is a line with only '/', which ends a statement of Oracle.
func canEnd(line string) bool {
	return strings.ContainsAny(line, ";\\") || strings.TrimSpace(line) == "/"
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

// read reads one more line or entry, and returns it.
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
