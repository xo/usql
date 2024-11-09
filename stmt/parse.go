package stmt

import (
	"regexp"
	"unicode"
)

// prefixCount is the number of words to extract from a prefix.
const prefixCount = 6

// maxVarNameLen is the maximum var name length.
const maxVarNameLen = 128

// grab returns the i'th rune from r when i < end, otherwise 0.
func grab(r []rune, i, end int) rune {
	if i < end {
		return r[i]
	}
	return 0
}

// findSpace finds first space rune in r, returning end if not found.
func findSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if isSpaceOrControl(r[i]) {
			return i, true
		}
	}
	return i, false
}

// findNonSpace finds first non space rune in r, returning end if not found.
func findNonSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if !isSpaceOrControl(r[i]) {
			return i, true
		}
	}
	return i, false
}

// findRune finds the next rune c in r, returning end if not found.
func findRune(r []rune, i, end int, c rune) (int, bool) {
	for ; i < end; i++ {
		if r[i] == c {
			return i, true
		}
	}
	return i, false
}

// isEmptyLine returns true when r is empty or composed of only whitespace.
func isEmptyLine(r []rune, i, end int) bool {
	_, ok := findNonSpace(r, i, end)
	return !ok
}

// identifierRE is a regexp that matches dollar tag identifiers ($tag$).
var identifierRE = regexp.MustCompile(`(?i)^[a-z_][a-z0-9_]{0,127}$`)

// readDollarAndTag reads a dollar style $tag$ in r, starting at i, returning
// the enclosed "tag" and position, or -1 if the dollar and tag was invalid.
func readDollarAndTag(r []rune, i, end int) (string, int, bool) {
	start, found := i, false
	i++
	for ; i < end; i++ {
		if r[i] == '$' {
			found = true
			break
		}
		if i-start > 128 {
			break
		}
	}
	if !found {
		return "", i, false
	}
	// check valid identifier
	id := string(r[start+1 : i])
	if id != "" && !identifierRE.MatchString(id) {
		return "", i, false
	}
	return id, i, true
}

// readString seeks to the end of a string returning the position and whether
// or not the string's end was found.
//
// If the string's terminator was not found, then the result will be the passed
// end.
func readString(r []rune, i, end int, quote rune, tag string) (int, bool) {
	var prev, c, next rune
	for ; i < end; i++ {
		c, next = r[i], grab(r, i+1, end)
		switch {
		case quote == '\'' && c == '\\':
			i++
			prev = 0
			continue
		case quote == '\'' && c == '\'' && next == '\'':
			i++
			continue
		case quote == '\'' && c == '\'' && prev != '\'',
			quote == '"' && c == '"',
			quote == '`' && c == '`':
			return i, true
		case quote == '$' && c == '$':
			if id, pos, ok := readDollarAndTag(r, i, end); ok && tag == id {
				return pos, true
			}
		}
		prev = c
	}
	return end, false
}

// readMultilineComment finds the end of a multiline comment (ie, '*/').
func readMultilineComment(r []rune, i, end int) (int, bool) {
	i++
	for ; i < end; i++ {
		if r[i-1] == '*' && r[i] == '/' {
			return i, true
		}
	}
	return end, false
}

// readVar reads variable from r in the form of :var_name :'var_name'
// :"var_name" and :{?var_name}.
func readVar(r []rune, i, end int, next rune) *Var {
	o := 1
	switch next {
	case '\'', '"':
		o = 2
	case '{':
		if grab(r, i+2, end) != '?' {
			return nil
		}
		o, next = 3, '}'
	default:
		next = 0
	}
	n, q := readVarName(r, i+o, end, next)
	v := n
	switch {
	case n-i < o+1, next != 0 && next != q:
		return nil
	case next != 0:
		v++
	}
	if next == '}' {
		next = '?'
	}
	return &Var{
		I:     i,
		End:   v,
		Quote: next,
		Name:  string(r[i+o : n]),
	}
}

// readVarName reads a variable name up to maxVarNameLen.
func readVarName(r []rune, i, end int, q rune) (int, rune) {
	c := rune(-1)
	for n := 0; i < end && n < maxVarNameLen && c != q; i, n = i+1, n+1 {
		switch c = grab(r, i, end); {
		case c == 0, c == q, c != '_' && !unicode.IsLetter(c) && !unicode.IsNumber(c):
			return i, c
		}
	}
	return i, 0
}

// readCommand reads the command and any parameters from r, returning the
// offset from i for the end of command, and the end of the command parameters.
//
// A command is defined as the first non-blank text after \, followed by
// parameters up to either the next \ or a control character (for example, \n):
func readCommand(r []rune, i, end int) (int, int) {
	// find end of command
	c, next := rune(0), rune(0)
command:
	for ; i < end; i++ {
		switch next = grab(r, i+1, end); {
		case next == 0:
			return end, end
		case next == '\\' || unicode.IsControl(next):
			i++
			return i, i
		case unicode.IsSpace(next):
			i++
			break command
		}
	}
	cmd, quote := i, rune(0)
params:
	// find end of params
	for ; i < end; i++ {
		switch c, next = r[i], grab(r, i+1, end); {
		case next == 0:
			return cmd, end
		case quote == 0 && (c == '\'' || c == '"' || c == '`'):
			quote = c
		case quote != 0 && c == quote:
			quote = 0
		// skip escaped
		case quote != 0 && c == '\\' && (next == quote || next == '\\'):
			i++
		case quote == 0 && (c == '\\' || unicode.IsControl(c)):
			break params
		}
	}
	// log.Printf(">>> params: %q remaining: %q", string(r[cmd:i]), string(r[i:end]))
	return cmd, i
}

// findPrefix finds the prefix in r up to n words.
func findPrefix(r []rune, n int, allowCComments, allowHashComments, allowMultilineComments bool) string {
	var s []rune
	var words int
loop:
	for i, end := 0, len(r); i < end; i++ {
		// skip space + control characters
		if j, _ := findNonSpace(r, i, end); i != j {
			r, end, i = r[j:], end-j, 0
		}
		// grab current and next character
		c, next := grab(r, i, end), grab(r, i+1, end)
		switch {
		// do nothing
		case c == 0:
		// statement terminator
		case c == ';':
			break loop
		// single line comments '--' and '//'
		case c == '-' && next == '-', c == '/' && next == '/' && allowCComments, c == '#' && allowHashComments:
			if i != 0 {
				s, words = appendUpperRunes(s, r[:i], ' '), words+1
			}
			// find line end
			if i, _ = findRune(r, i, end, '\n'); i >= end {
				break
			}
			r, end, i = r[i+1:], end-i-1, -1
		// multiline comments '/*' '*/'
		case c == '/' && next == '*' && allowMultilineComments:
			if i != 0 {
				s, words = appendUpperRunes(s, r[:i]), words+1
			}
			// find comment end '*/'
			for i += 2; i < end; i++ {
				if grab(r, i, end) == '*' && grab(r, i+1, end) == '/' {
					r, end, i = r[i+2:], end-i-2, -1
					break
				}
			}
			// add space when remaining runes begin with space, and previous
			// captured word did not
			if sl := len(s); end > 0 && sl != 0 && isSpaceOrControl(r[0]) && !isSpaceOrControl(s[sl-1]) {
				s = append(s, ' ')
			}
		// end of statement, max words, or punctuation that can be ignored
		case words == n || !unicode.IsLetter(c):
			break loop
		// ignore remaining, as no prefix can come after
		case next != '/' && !unicode.IsLetter(next):
			s, words = appendUpperRunes(s, r[:i+1], ' '), words+1
			if next == 0 {
				break
			}
			if next == ';' {
				break loop
			}
			r, end, i = r[i+2:], end-i-2, -1
		}
	}
	// trim right ' ', if any
	if sl := len(s); sl != 0 && s[sl-1] == ' ' {
		return string(s[:sl-1])
	}
	return string(s)
}

// FindPrefix finds the first 6 prefix words in s.
func FindPrefix(s string, allowCComments, allowHashComments, allowMultilineComments bool) string {
	return findPrefix([]rune(s), prefixCount, allowCComments, allowHashComments, allowMultilineComments)
}

// substitute substitutes n runes in r starting at i with the runes in s.
// Dynamically grows r if necessary.
func substitute(r []rune, i, end, n int, s string) ([]rune, int) {
	sr, rcap := []rune(s), cap(r)
	sn := len(sr)
	// grow ...
	tlen := len(r) + sn - n
	if tlen > rcap {
		z := make([]rune, tlen)
		copy(z, r)
		r = z
	} else {
		r = r[:rcap]
	}
	// substitute
	copy(r[i+sn:], r[i+n:])
	copy(r[i:], sr)
	return r[:tlen], tlen
}

// appendUpperRunes creates a new []rune from s, with the runes in r on the end
// converted to upper case. extra runes will be appended to the final []rune.
func appendUpperRunes(s []rune, r []rune, extra ...rune) []rune {
	sl, rl, el := len(s), len(r), len(extra)
	sre := make([]rune, sl+rl+el)
	copy(sre[:sl], s)
	for i := 0; i < rl; i++ {
		sre[sl+i] = unicode.ToUpper(r[i])
	}
	copy(sre[sl+rl:], extra)
	return sre
}
