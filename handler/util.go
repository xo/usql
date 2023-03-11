package handler

import (
	"regexp"
	"strings"
	"unicode"
)

// linetermRE is the end of line terminal.
var linetermRE = regexp.MustCompile(`(?:\r?\n)+$`)

// empty reports whether s contains at least one printable, non-space character.
func empty(s string) bool {
	i := strings.IndexFunc(s, func(r rune) bool {
		return unicode.IsPrint(r) && !unicode.IsSpace(r)
	})
	return i == -1
}

var ansiRE = regexp.MustCompile(`\x1b[[0-9]+([:;][0-9]+)*m`)

// lastcolor returns the last defined color in s, if any.
func lastcolor(s string) string {
	if i := strings.LastIndex(s, "\n"); i != -1 {
		s = s[:i]
	}
	if i := strings.LastIndex(s, "\x1b[0m"); i != -1 {
		s = s[i+4:]
	}
	return strings.Join(ansiRE.FindAllString(s, -1), "")
}
