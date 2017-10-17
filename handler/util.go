package handler

import (
	"runtime"
	"strings"
	"unicode"
)

// lineterm is the end of line terminal.
var lineterm string

func init() {
	lineterm = "\n"
	if runtime.GOOS == "windows" {
		lineterm = "\r\n"
	}
}

// empty reports whether s contains at least one printable, non-space character.
func empty(s string) bool {
	i := strings.IndexFunc(s, func(r rune) bool {
		return unicode.IsPrint(r) && !unicode.IsSpace(r)
	})

	return i == -1
}
