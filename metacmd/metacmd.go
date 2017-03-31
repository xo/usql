package metacmd

import (
	"github.com/knq/usql/env"
	"github.com/knq/usql/text"
)

// Metacmd represents a command and associated meta information about it.
type Metacmd uint

// Decode converts a command name (or alias) into a Runner.
func Decode(name string, params []string) (Runner, error) {
	mc, ok := cmdMap[name]
	if !ok {
		return nil, text.ErrUnknownCommand
	}

	cmd := cmds[mc]
	if cmd.Min > len(params) {
		return nil, text.ErrMissingRequiredArgument
	}

	return RunnerFunc(func(h Handler) (Res, error) {
		for i, s := range params {
			v, err := env.Unquote(h.User(), s, true)
			if err != nil {
				return Res{Processed: len(params)}, err
			}
			params[i] = v
		}

		p := &Params{h, name, params, Res{}}
		err := cmd.Process(p)
		return p.R, err
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

	// Connect is the connect meta command (\c).
	Connect

	// Disconnect is the disconnect meta command (\Z).
	Disconnect

	// Password is the change password meta command (\password).
	Password

	// ConnInfo is the connection info meta command (\conninfo).
	ConnInfo

	// Drivers is the driver info meta command (\drivers).
	Drivers

	// Describe is the describe meta command (\d and variants).
	Describe

	// Exec is the execute meta command (\g and variants).
	Exec

	// Edit is the edit query buffer meta command (\e).
	Edit

	// Print is the print query buffer meta command (\p).
	Print

	// Reset is the reset query buffer meta command (\r).
	Reset

	// Echo is the echo meta command (\echo).
	Echo

	// Write is the write meta command (\w).
	Write

	// ChangeDir is the system change directory meta command (\cd).
	ChangeDir

	// SetEnv is the system set environment variable meta command (\setenv).
	SetEnv

	// Include is the system include file meta command (\i and variants).
	Include

	// Begin is the transaction begin meta command (\begin).
	Begin

	// Commit is the transaction commit meta command (\commit).
	Commit

	// Rollback is the transaction rollback (abort) meta command (\rollback).
	Rollback

	// Prompt is the variable prompt meta command (\prompt).
	Prompt

	// Set is the variable set meta command (\set).
	Set

	// Unset is the variable unset meta command (\unset).
	Unset
)
