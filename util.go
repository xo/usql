package main

import (
	"strings"
	"unicode"
)

const (
	aboutDesc = `usql is the universal command-line interface for SQL databases.
`

	welcomeDesc = `Type "help" for help.

`

	helpDesc = `You are using usql, the universal command-line interface for SQL databases.
Type: \c[onnect] <url>  connect to url
      \q                quit
`
)

// Args are the command line arguments.
type Args struct {
	DSN           string   `arg:"positional,help:database url"`
	Commands      []string `arg:"-c,--command,separate,help:run only single command (SQL or internal) and exit"`
	File          string   `arg:"-f,--file,help:execute commands from file then exit"`
	Out           string   `arg:"-o,--output,help:output file"`
	HistoryFile   string   `arg:"--hist-file,env:USQL_HISTFILE,help:history file"`
	Username      string   `arg:"-U,--username,help:database user name"`
	DisablePretty bool     `arg:"-p,--disable-pretty,help:disable pretty formatting"`
	NoRC          bool     `arg:"-X,--disable-rc,help:do not read start up file"`
}

// Description provides the go-arg description.
func (a *Args) Description() string {
	return aboutDesc
}

// startsWith checks that s begins with the specified prefix and is followed by
// at least one space, returning the remaining string trimmed of spaces.
//
// ie, a call of startsWith(`\c blah `, `\c`) should return `blah`.
func startsWith(s string, prefix string) (string, bool) {
	if prefix == "" {
		return s, true
	}

	var i int
	rs, rslen := []rune(s), len(s)
	for ; i < rslen; i++ {
		if !unicode.IsSpace(rs[i]) {
			break
		}
	}

	if i >= rslen {
		return "", false
	}

	match := true
	ps, pslen := []rune(prefix), len(prefix)
	if i+pslen+1 > rslen {
		return "", false
	}

	for j := 0; j < pslen && i+j < rslen; j++ {
		match = match && rs[i+j] == ps[j]
		if !match {
			return "", false
		}
	}

	if !unicode.IsSpace(rs[i+pslen]) {
		return "", false
	}

	return strings.TrimSpace(string(rs[i+pslen+1:])), true
}
