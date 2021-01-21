package metacmd

import (
	"io"
	"os/user"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
	"github.com/xo/usql/rline"
	"github.com/xo/usql/stmt"
)

// Handler is the shared interface for a command handler.
type Handler interface {
	// IO handles the handler's IO.
	IO() rline.IO
	// User returns the current user.
	User() *user.User
	// URL returns the current database URL.
	URL() *dburl.URL
	// DB returns the current database connection.
	DB() drivers.DB
	// Last returns the last executed query.
	Last() string
	// LastRaw returns the last raw (non-interpolated) query.
	LastRaw() string
	// Buf returns the current query buffer.
	Buf() *stmt.Stmt
	// Reset resets the last and current query buffer.
	Reset([]rune)
	// Open opens a database connection.
	Open(...string) error
	// Close closes the current database connection.
	Close() error
	// ChangePassword changes the password for a user.
	ChangePassword(string) (string, error)
	// ReadVar reads a variable of a specified type.
	ReadVar(string, string) (string, error)
	// Include includes a file.
	Include(string, bool) error
	// Begin begins a transaction.
	Begin() error
	// Commit commits the current transaction.
	Commit() error
	// Rollback aborts the current transaction.
	Rollback() error
	// Highlight highlights the statement.
	Highlight(io.Writer, string) error
	// GetTiming mode.
	GetTiming() bool
	// SetTiming mode.
	SetTiming(bool)
}

// Runner is a runner interface type.
type Runner interface {
	Run(Handler) (Result, error)
}

// RunnerFunc is a type wrapper for a single func satisfying Runner.Run.
type RunnerFunc func(Handler) (Result, error)

// Run satisfies the Runner interface.
func (f RunnerFunc) Run(h Handler) (Result, error) {
	return f(h)
}

// ExecType represents the type of execution requested.
type ExecType int

const (
	// ExecNone indicates no execution.
	ExecNone ExecType = iota
	// ExecOnly indicates plain execution only (\g).
	ExecOnly
	// ExecPipe indicates execution and piping results (\g |file)
	ExecPipe
	// ExecSet indicates execution and setting the resulting columns as
	// variables (\gset).
	ExecSet
	// ExecExec indicates execution and executing the resulting rows (\gexec).
	ExecExec
)

// Result is the result of metacmd execution.
type Result struct {
	// Quit instructs the handling code to quit.
	Quit bool
	// Exec informs the handling code of the type of execution.
	Exec ExecType
	// ExecParam is an accompanying parameter for execution. For ExecPipe, it
	// will be the name of a file. For ExecSet it will be the variable prefix.
	ExecParam string
	// Expanded forces expanded output.
	Expanded bool
}

// Params wraps metacmd parameters.
type Params struct {
	// Handler is the process handler.
	Handler Handler
	// Name is the name of the metacmd.
	Name string
	// Params are the passed parameters.
	Params *stmt.Params
	// Result is the resulting state of the command execution.
	Result Result
}

// Get returns the next command parameter, using env.Unquote to decode quoted
// strings.
func (p *Params) Get(exec bool) (string, error) {
	_, v, err := p.Params.Get(env.Unquote(
		p.Handler.User(),
		exec,
		env.All(),
	))
	if err != nil {
		return "", err
	}
	return v, nil
}

// GetOK returns the next command parameter, using env.Unquote to decode quoted
// strings.
func (p *Params) GetOK(exec bool) (bool, string, error) {
	return p.Params.Get(env.Unquote(
		p.Handler.User(),
		exec,
		env.All(),
	))
}

// GetOptional returns the next command parameter, using env.Unquote to decode
// quoted strings, returns true when the value is prefixed with a "-", along
// with the value sans the "-" prefix. Otherwise returns false and the value.
func (p *Params) GetOptional(exec bool) (bool, string, error) {
	v, err := p.Get(exec)
	if err != nil {
		return false, "", err
	}
	if len(v) > 0 && v[0] == '-' {
		return true, v[1:], nil
	}
	return false, v, nil
}

// GetAll gets all remaining command parameters using env.Unquote to decode
// quoted strings.
func (p *Params) GetAll(exec bool) ([]string, error) {
	return p.Params.GetAll(env.Unquote(
		p.Handler.User(),
		exec,
		env.All(),
	))
}

// GetRaw gets the remaining command parameters as a raw string.
//
// Note: no other processing is done to interpolate variables or to decode
// string values.
func (p *Params) GetRaw() string {
	return p.Params.GetRaw()
}
