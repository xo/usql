package main

import (
	"github.com/knq/usql/text"
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
	Verbose       bool     `arg:"-v,--verbose,help:toggle verbose"`
}

// Description provides the go-arg description.
func (a *Args) Description() string {
	return text.Banner + ".\n"
}

// Version returns the version string for the app.
func (a *Args) Version() string {
	return text.CommandName + " " + text.CommandVersion
}
