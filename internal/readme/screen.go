package readme

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Render returns the text that b leaves on a terminal, as a person reads it.
//
// b is what a program wrote to a terminal: text, carriage returns, line feeds
// and escape sequences. Render keeps every line that scrolled past, so the
// result is the whole session and not only the last screen. It follows the
// sequences that a line editor uses to redraw a line: the cursor movements, and
// erasing to the end of a line or of the screen. It drops the rest, such as
// colors and the titles of windows.
func Render(b []byte) string {
	s := &screen{lines: [][]rune{nil}}
	for len(b) != 0 {
		switch c := b[0]; {
		case c == 0x1b:
			b = s.escape(b)
			continue
		case c == '\r':
			s.col = 0
		case c == '\n':
			s.down(1)
		case c == '\b':
			s.col = max(0, s.col-1)
		case c == '\t':
			s.col += 8 - s.col%8
		case c < ' ' || c == 0x7f:
		default:
			r, n := utf8.DecodeRune(b)
			s.put(r)
			b = b[n:]
			continue
		}
		b = b[1:]
	}
	var lines []string
	for _, l := range s.lines {
		var sb strings.Builder
		for _, r := range l {
			if r != wide {
				sb.WriteRune(r)
			}
		}
		lines = append(lines, strings.TrimRight(sb.String(), " "))
	}
	for len(lines) != 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// wide fills the second column of a character that takes two.
const wide = rune(-1)

// screen is the text of a terminal, one rune for each column, and the place of
// its cursor.
type screen struct {
	lines    [][]rune
	row, col int
	saved    [2]int
}

// put writes r at the cursor, and moves the cursor past it.
func (s *screen) put(r rune) {
	cells := []rune{r}
	if runewidth.RuneWidth(r) == 2 {
		cells = append(cells, wide)
	}
	l := s.lines[s.row]
	for _, c := range cells {
		for len(l) <= s.col {
			l = append(l, ' ')
		}
		l[s.col] = c
		s.col++
	}
	s.lines[s.row] = l
}

// down moves the cursor down n rows, and adds the rows that do not exist.
func (s *screen) down(n int) {
	s.row += n
	for len(s.lines) <= s.row {
		s.lines = append(s.lines, nil)
	}
}

// escape follows the escape sequence at the start of b, and returns the rest.
func (s *screen) escape(b []byte) []byte {
	if len(b) < 2 {
		return nil
	}
	switch b[1] {
	case '[':
		return s.csi(b[2:])
	case ']', 'P', '_', '^', 'X':
		// a string, which ends with BEL or with ESC \
		for i := 2; i < len(b); i++ {
			switch {
			case b[i] == '\a':
				return b[i+1:]
			case b[i] == 0x1b && i+1 < len(b) && b[i+1] == '\\':
				return b[i+2:]
			}
		}
		return nil
	case '(', ')', '*', '+':
		// the choice of a character set
		return b[min(3, len(b)):]
	case '7':
		s.saved = [2]int{s.row, s.col}
	case '8':
		s.row, s.col = s.saved[0], s.saved[1]
	}
	return b[2:]
}

// csi follows a control sequence, whose parameters start b.
func (s *screen) csi(b []byte) []byte {
	i := 0
	for i < len(b) && (b[i] < 0x40 || b[i] > 0x7e) {
		i++
	}
	if i == len(b) {
		return nil
	}
	params, final := string(b[:i]), b[i]
	if strings.HasPrefix(params, "?") || strings.HasPrefix(params, ">") {
		// a private mode, such as showing the cursor
		return b[i+1:]
	}
	n := 1
	if v, err := strconv.Atoi(strings.Split(params, ";")[0]); err == nil && v > 0 {
		n = v
	}
	switch final {
	case 'A':
		s.row = max(0, s.row-n)
	case 'B':
		s.down(n)
	case 'C':
		s.col += n
	case 'D':
		s.col = max(0, s.col-n)
	case 'G':
		s.col = n - 1
	case 'K':
		s.eraseLine(params)
	case 'J':
		s.eraseLine(params)
		if params == "" || params == "0" {
			s.lines = s.lines[:s.row+1]
		}
	case 's':
		s.saved = [2]int{s.row, s.col}
	case 'u':
		s.row, s.col = s.saved[0], s.saved[1]
	}
	return b[i+1:]
}

// eraseLine erases the line of the cursor: to its end, to its start (1), or
// all of it (2).
func (s *screen) eraseLine(params string) {
	l := s.lines[s.row]
	switch params {
	case "1":
		for i := 0; i <= s.col && i < len(l); i++ {
			l[i] = ' '
		}
	case "2":
		l = nil
	default:
		if s.col < len(l) {
			l = l[:s.col]
		}
	}
	s.lines[s.row] = l
}
