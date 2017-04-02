package stmt

import (
	"regexp"
	"strings"
	"unicode"
)

// grab grabs i from r, or returns 0 if i >= end.
func grab(r []rune, i, end int) rune {
	if i < end {
		return r[i]
	}
	return 0
}

// findSpace finds first space rune in r, returning end if not found.
func findSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if unicode.IsSpace(r[i]) {
			return i, true
		}
	}
	return i, false
}

// findNonSpace finds first non space rune in r, returning end if not found.
func findNonSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if !unicode.IsSpace(r[i]) {
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

// StartsWith determines if r starts with s, ignoring case, and skipping
// initial whitespace and returning -1 if r does not start with s.
//
// Note: assumes s contains at least one non space.
func StartsWith(r []rune, i, end int, s string) bool {
	slen := len(s)

	// find start
	var found bool
	i, found = findNonSpace(r, i, end)
	if !found || i+slen > end {
		return false
	}

	// check
	if strings.ToLower(string(r[i:i+slen])) == s {
		return true
	}

	return false
}

// trimSplit splits r by whitespace into a string slice.
func trimSplit(r []rune, i, end int) []string {
	var a []string

	for i < end {
		n, found := findNonSpace(r, i, end)
		if !found || n == end {
			// empty
			return a
		}

		var m int
		if c := r[n]; c == '\'' || c == '"' || c == '`' {
			m, _ = findRune(r, n+1, end, c)
			m++
		} else {
			m, _ = findSpace(r, n, end)
		}

		a = append(a, string(r[n:m]))
		i = m
	}

	return a
}

var identifierRE = regexp.MustCompile(`(?i)^[a-z][a-z0-9_]{0,127}$`)

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

// readString seeks to the end of a string (depending on the state of b)
// returning the position and whether or not the string's end was found.
//
// If the string's terminator was not found, then the result will be the passed
// end.
func readString(r []rune, i, end int, b *Stmt) (int, bool) {
	var prev, c rune
	for ; i < end; i++ {
		c = r[i]
		switch {
		case b.allowDollar && b.qdollar && c == '$':
			if id, pos, ok := readDollarAndTag(r, i, end); ok && b.qid == id {
				return pos, true
			}

		case b.qdbl && c == '"':
			return i, true

		case !b.qdbl && !b.qdollar && c == '\'' && prev != '\'':
			return i, true
		}
		prev = r[i]
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

// readStringVar reads a string quoted variable.
func readStringVar(r []rune, i, end int) *Var {
	start, q := i, grab(r, i+1, end)
	for i += 2; i < end; i++ {
		c := grab(r, i, end)
		if c == q {
			if i-start < 3 {
				return nil
			}

			return &Var{
				I:   start,
				End: i + 1,
				Q:   q,
				N:   string(r[start+2 : i]),
			}
		} /*
			// this is commented out, because need to determine what should be
			// the "right" behavior ... should we only allow "identifiers"?
			else if c != '_' && !unicode.IsLetter(c) && !unicode.IsNumber(c) {
				return nil
			}
		*/
	}

	return nil
}

// readVar reads the variable.
func readVar(r []rune, i, end int) *Var {
	if grab(r, i, end) != ':' || grab(r, i+1, end) == ':' {
		return nil
	}

	if end-i < 2 {
		return nil
	}

	if c := grab(r, i+1, end); c == '"' || c == '\'' {
		return readStringVar(r, i, end)
	}

	start := i
	i++
	for ; i < end; i++ {
		if c := grab(r, i, end); c != '_' && !unicode.IsLetter(c) && !unicode.IsNumber(c) {
			break
		}
	}

	if i-start < 2 {
		return nil
	}

	return &Var{
		I:   start,
		End: i,
		N:   string(r[start+1 : i]),
	}
}

// readCommand reads the command and any parameters from r.
func readCommand(r []rune, i, end int) (string, []string, int) {
	i++

	// find end (either end of r, or the next command)
	start, found := i, false
	for ; i < end-1; i++ {
		if unicode.IsSpace(r[i]) && r[i+1] == '\\' {
			found = true
			break
		}
	}

	// fix i
	if found {
		i++
	} else {
		i = end
	}

	// split values
	a := trimSplit(r, start, i)
	if len(a) == 0 {
		return "", nil, i
	}

	return a[0], a[1:], i
}

// findEndOfWords finds the end of the first n words in r, and returns the
// end place and count of the words found.
func findEndOfWords(r []rune, i, end, n int) (int, int) {
	w := 0
	for ; w < n; w++ {
		z, ok := findNonSpace(r, i, end)
		if !ok {
			break
		}
		i, _ = findSpace(r, z, end)
	}
	return i, w
}

var spaceRE = regexp.MustCompile(`\s+`)

// findPrefix finds the prefix in r up to n words.
func findPrefix(r []rune, i, end, n int) string {
	plen, w := findEndOfWords(r, i, end, n)
	if w != 0 {
		st, _ := findNonSpace(r, i, plen)
		for i = st; i < plen; i++ {
			if !unicode.IsSpace(r[i]) && !unicode.IsLetter(r[i]) {
				break
			}
		}
		return strings.ToUpper(spaceRE.ReplaceAllString(strings.TrimSpace(string(r[st:i])), " "))
	}

	return ""
}

// FindPrefix finds the prefix in s up to n words.
func FindPrefix(s string, n int) string {
	r := []rune(s)
	return findPrefix(r, 0, len(r), n)
}

// substituteVar substitutes part of r, based on v, with s.
func substituteVar(r []rune, v *Var, s string) ([]rune, int) {
	v.Len = len(s)

	// grow ...
	tlen := len(r) + v.Len - (v.End - v.I)
	if tlen > cap(r) {
		z := make([]rune, tlen)
		copy(z, r)
		r = z
	}

	// substitute
	copy(r[v.I+v.Len:], r[v.End:])
	copy(r[v.I:v.I+v.Len], []rune(s))

	return r[:tlen], tlen
}
