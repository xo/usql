package metacmd

import (
	"io"
	"os/user"
	"strings"

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
}

// Runner is a runner interface type.
type Runner interface {
	Run(Handler) (Res, error)
}

// RunnerFunc is a type wrapper for a single func satisfying Runner.Run.
type RunnerFunc func(Handler) (Res, error)

// Run satisfies the Runner interface.
func (f RunnerFunc) Run(h Handler) (Res, error) {
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

// Res is the result of a meta command execution.
type Res struct {
	// Quit instructs the handling code to quit.
	Quit bool

	// Exec informs the handling code of the type of execution.
	Exec ExecType

	// ExecParam is an accompanying parameter for execution. For ExecPipe, it
	// will be the name of a file. For ExecSet it will be the variable prefix.
	ExecParam string

	// Processed informs the handling code how many parameters went
	// unprocessed. A value of 0 means that no parameters were processed.
	Processed int
}

// Params wraps command parameters.
type Params struct {
	// H is the handler.
	H Handler

	// N is the name of the command.
	N string

	// P are the passed parameters.
	P []string

	// R is the resulting state of the command execution.
	R Res
}

// G returns the next parameter, increasing p.R.Processed by 1.
func (p *Params) G() string {
	if len(p.P) > p.R.Processed {
		s, _ := env.Unquote(p.H.User(), p.P[p.R.Processed], true)
		p.R.Processed++
		return s
	}
	return ""
}

// V returns the next parameter only if it is prefixed with a "-", increasing
// p.R.Processed when it is, otherwise returning d.
func (p *Params) V(d string) string {
	if len(p.P) > p.R.Processed && strings.HasPrefix(p.P[p.R.Processed], "-") {
		s := p.P[p.R.Processed][1:]
		p.R.Processed++
		return s
	}

	return d
}

// A returns all remaining, unprocessed parameters, incrementing p.R.processed
// appropriately.
func (p *Params) A() []string {
	x := make([]string, len(p.P)-p.R.Processed)
	var j int
	for i := p.R.Processed; i < len(p.P); i++ {
		s, _ := env.Unquote(p.H.User(), p.P[i], true)
		x[j] = s
		j++
	}
	p.R.Processed = len(p.P)
	return x
}
