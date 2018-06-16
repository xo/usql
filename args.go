package main

import (
	"github.com/xo/usql/text"
)

// Args are the command line arguments.
type Args struct {
	DSN               string   `arg:"positional,help:database url"`
	Commands          []string `arg:"-c,--command,separate,help:run only single command (SQL or internal) and exit"`
	File              string   `arg:"-f,--file,help:execute commands from file and exit"`
	Out               string   `arg:"-o,--output,help:output file"`
	Username          string   `arg:"-U,--username,help:database user name"`
	ForcePassword     bool     `arg:"-W,--password,help:force password prompt (should happen automatically)"`
	NoPassword        bool     `arg:"-w,--no-password,help:never prompt for password"`
	NoRC              bool     `arg:"-X,--no-rc,help:do not read start up file"`
	SingleTransaction bool     `arg:"-1,--single-transaction,help:execute as a single transaction (if non-interactive)"`
	Variables         []string `arg:"-v,--set,--variable,separate,help:set variable NAME=VALUE"`
}

// Description provides the go-arg description.
func (a *Args) Description() string {
	return text.CommandName + " - " + text.Banner + ".\n"
}

// Version returns the version string for the app.
func (a *Args) Version() string {
	return text.CommandName + " " + text.CommandVersion
}

// Usage returns the usage line for the app.
func (a *Args) Usage() string {
	return "Usage:\n  " + text.CommandName + " [OPTION]... DSN"
}

// Title returns a replacement title.
func (a *Args) Title(name string) string {
	if name == "Positional arguments" {
		return "Arguments"
	}
	return name
}
