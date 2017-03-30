package metacmd

import (
	"errors"
	"os/user"

	"github.com/knq/dburl"
	"github.com/knq/usql/drivers"
	"github.com/knq/usql/rline"
	"github.com/knq/usql/stmt"
)

var (
	// ErrUnknownCommand is the unknown command error.
	ErrUnknownCommand = errors.New("unknown command")

	// ErrMissingRequiredArgument is the missing required argument error.
	ErrMissingRequiredArgument = errors.New("missing required argument")
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

	// Buf returns the current query buffer.
	Buf() *stmt.Stmt

	// Open opens a database connection.
	Open(...string) error

	// Close closes the current database connection.
	Close() error

	// Include includes a file.
	Include(string, bool) error

	// Begin begins a transaction.
	Begin() error

	// Commit commits the current transaction.
	Commit() error

	// Rollback aborts the current transaction.
	Rollback() error
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

	// ExecExec indicates execution and executing the resulting rows (\gexec).
	ExecExec

	// ExecSet indicates execution and setting the resulting columns as
	// variables (\gset).
	ExecSet
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
