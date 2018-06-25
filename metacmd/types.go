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

	// Processed informs the handling code how many parameters went
	// unprocessed. A value of 0 means that no parameters were processed.
	Processed int
}

// Params wraps metacmd parameters.
type Params struct {
	// Handler is the process handler.
	Handler Handler

	// Name is the name of the metacmd.
	Name string

	// Params are the passed parameters.
	Params []string

	// Result is the resulting state of the command execution.
	Result Result
}

// Get returns the next parameter, increasing p.Result.Processed by 1.
func (p *Params) Get() string {
	if len(p.Params) > p.Result.Processed {
		s, _ := env.Unquote(p.Handler.User(), p.Params[p.Result.Processed], true)
		p.Result.Processed++
		return s
	}
	return ""
}

// GetOptional returns the next parameter only if it is prefixed with a "-",
// increasing p.Result.Processed by 1 when it does, otherwise returning
// defaultVal.
func (p *Params) GetOptional(defaultVal string) string {
	if len(p.Params) > p.Result.Processed && strings.HasPrefix(p.Params[p.Result.Processed], "-") {
		s := p.Params[p.Result.Processed][1:]
		p.Result.Processed++
		return s
	}
	return defaultVal
}

// GetAll gets all remaining, unprocessed parameters, incrementing
// p.Result.processed appropriately.
func (p *Params) GetAll() []string {
	x := make([]string, len(p.Params)-p.Result.Processed)
	var j int
	for i := p.Result.Processed; i < len(p.Params); i++ {
		s, _ := env.Unquote(p.Handler.User(), p.Params[i], true)
		x[j] = s
		j++
	}
	p.Result.Processed = len(p.Params)
	return x
}
