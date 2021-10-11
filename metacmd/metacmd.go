// Package metacmd contains meta information and implentation for usql's
// backslash (\) commands.
package metacmd

import (
	"github.com/xo/usql/stmt"
	"github.com/xo/usql/text"
)

// Metacmd represents a command and associated meta information about it.
type Metacmd uint

// Decode converts a command name (or alias) into a Runner.
func Decode(name string, params *stmt.Params) (Runner, error) {
	mc, ok := cmdMap[name]
	if !ok || name == "" {
		return nil, text.ErrUnknownCommand
	}
	cmd := cmds[mc]
	return RunnerFunc(func(h Handler) (Option, error) {
		p := &Params{
			Handler: h,
			Name:    name,
			Params:  params,
		}
		err := cmd.Process(p)
		return p.Option, err
	}), nil
}

// Command types.
const (
	// None is an empty command.
	None Metacmd = iota
	// Question is question meta command (\?)
	Question
	// Quit is the quit meta command (\?).
	Quit
	// Copyright is the copyright meta command (\copyright).
	Copyright
	// Connect is the connect meta command (\c, \connect).
	Connect
	// Copy is the copy meta command (\copy).
	Copy
	// Disconnect is the disconnect meta command (\Z).
	Disconnect
	// Password is the change password meta command (\password).
	Password
	// ConnectionInfo is the connection info meta command (\conninfo).
	ConnectionInfo
	// Drivers is the driver info meta command (\drivers).
	Drivers
	// Describe is the describe meta command (\d and variants).
	Describe
	// Exec is the execute meta command (\g and variants).
	Exec
	// Edit is the edit query buffer meta command (\e).
	Edit
	// Print is the print query buffer meta command (\p, \print, \raw).
	Print
	// Reset is the reset query buffer meta command (\r, \reset).
	Reset
	// Echo is the echo meta command (\echo, \warn, \qecho).
	Echo
	// Write is the write meta command (\w).
	Write
	// ChangeDir is the system change directory meta command (\cd).
	ChangeDir
	// SetEnv is the system set environment variable meta command (\setenv).
	SetEnv
	// Shell is the system shell exec meta command (\!).
	Shell
	// Out is the switch output meta command (\o).
	Out
	// Include is the system include file meta command (\i and variants).
	Include
	// Transact is the transaction meta command (\begin, \commit, \rollback).
	Transact
	// Prompt is the variable prompt meta command (\prompt).
	Prompt
	// SetVar is the set variable meta command (\set).
	SetVar
	// Unset is the variable unset meta command (\unset).
	Unset
	// SetFormatVar is the set format variable meta commands (\pset, \a, \C, \f, \H, \t, \T, \x).
	SetFormatVar
	// Timing is the timing meta command (\timing).
	Timing
	// Stats is the show stats meta command (\ss and variants).
	Stats
)
