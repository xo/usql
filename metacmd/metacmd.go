// Package metacmd contains meta information and implementation for usql's
// backslash (\) commands.
package metacmd

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os/user"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/env"
	"github.com/xo/usql/rline"
	"github.com/xo/usql/stmt"
	"github.com/xo/usql/text"
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
	// LastExec returns the last executed query.
	LastExec() string
	// LastPrint returns the last executed printable query.
	LastPrint() string
	// LastRaw returns the last raw (non-interpolated) query.
	LastRaw() string
	// Buf returns the current query buffer.
	Buf() *stmt.Stmt
	// Reset resets the last and current query buffer.
	Reset([]rune)
	// Bind binds query parameters.
	Bind([]interface{})
	// Open opens a database connection.
	Open(context.Context, ...string) error
	// Close closes the current database connection.
	Close() error
	// ChangePassword changes the password for a user.
	ChangePassword(string) (string, error)
	// ReadVar reads a variable of a specified type.
	ReadVar(string, string) (string, error)
	// Include includes a file.
	Include(string, bool) error
	// Begin begins a transaction.
	Begin(*sql.TxOptions) error
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
	// GetOutput writer.
	GetOutput() io.Writer
	// SetOutput writer.
	SetOutput(io.WriteCloser)
	// MetadataWriter retrieves the metadata writer for the handler.
	MetadataWriter(context.Context) (metadata.Writer, error)
	// Print formats according to a format specifier and writes to handler's standard output.
	Print(string, ...interface{})
}

// Dump writes the command descriptions to w, separated by section.
func Dump(w io.Writer, hidden bool) error {
	n := 0
	for i := range sections {
		for _, desc := range descs[i] {
			if (!desc.Hidden && !desc.Deprecated) || hidden {
				n = max(n, runewidth.StringWidth(desc.Name)+1+runewidth.StringWidth(desc.Params))
			}
		}
	}
	for i, s := range sections {
		if i != 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, s)
		for _, desc := range descs[i] {
			if (!desc.Hidden && !desc.Deprecated) || hidden {
				_, _ = fmt.Fprintf(w, "  \\%- *s  %s\n", n, desc.Name+" "+desc.Params, wrap(desc.Desc, 95, n+5))
			}
		}
	}
	return nil
}

// Decode converts a command name (or alias) into a Runner.
func Decode(name string, params *stmt.Params) (func(Handler) (Option, error), error) {
	f, ok := cmds[name]
	if !ok || name == "" {
		return nil, text.ErrUnknownCommand
	}
	return func(h Handler) (Option, error) {
		p := &Params{
			Handler: h,
			Name:    name,
			Params:  params,
		}
		err := f(p)
		return p.Option, err
	}, nil
}

// Params wraps metacmd parameters.
type Params struct {
	// Handler is the process handler.
	Handler Handler
	// Name is the name of the metacmd.
	Name string
	// Params are the actual statement parameters.
	Params *stmt.Params
	// Option contains resulting command execution options.
	Option Option
}

// Next returns the next command parameter, using env.Untick.
func (p *Params) Next(exec bool) (string, error) {
	v, _, err := p.Params.Next(env.Untick(
		p.Handler.User(),
		env.Vars(),
		exec,
	))
	if err != nil {
		return "", err
	}
	return v, nil
}

// NextOK returns the next command parameter, using env.Untick.
func (p *Params) NextOK(exec bool) (string, bool, error) {
	return p.Params.Next(env.Untick(
		p.Handler.User(),
		env.Vars(),
		exec,
	))
}

// NextOpt returns the next command parameter, using env.Untick. Returns true
// when the value is prefixed with a "-", along with the value sans the "-"
// prefix. Otherwise returns false and the value.
func (p *Params) NextOpt(exec bool) (string, bool, error) {
	v, err := p.Next(exec)
	switch {
	case err != nil:
		return "", false, err
	case len(v) > 0 && v[0] == '-':
		return v[1:], true, nil
	}
	return v, false, nil
}

// All gets all remaining command parameters using env.Untick.
func (p *Params) All(exec bool) ([]string, error) {
	return p.Params.All(env.Untick(
		p.Handler.User(),
		env.Vars(),
		exec,
	))
}

// Raw returns the remaining command parameters as a raw string.
//
// Note: no other processing is done to interpolate variables or to decode
// string values.
func (p *Params) Raw() string {
	return p.Params.Raw()
}

// Option contains parsed result options of a metacmd.
type Option struct {
	// Quit instructs the handling code to quit.
	Quit bool
	// Exec informs the handling code of the type of execution.
	Exec ExecType
	// Params are accompanying string parameters for execution.
	Params map[string]string
	// Crosstab are the crosstab column parameters.
	Crosstab []string
	// Watch is the watch duration interval.
	Watch time.Duration
}

func (opt *Option) ParseParams(params []string, defaultKey string) error {
	if opt.Params == nil {
		opt.Params = make(map[string]string, len(params))
	}
	formatOpts := false
	for i, param := range params {
		if len(param) == 0 {
			continue
		}
		if !formatOpts {
			if param[0] == '(' {
				formatOpts = true
			} else {
				opt.Params[defaultKey] = strings.Join(params[i:], " ")
				return nil
			}
		}
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 1 {
			return text.ErrInvalidFormatOption
		}
		opt.Params[strings.TrimLeft(parts[0], "(")] = strings.TrimRight(parts[1], ")")
		if formatOpts && param[len(param)-1] == ')' {
			formatOpts = false
		}
	}
	return nil
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
	// ExecCrosstab indicates execution using crosstabview (\crosstabview).
	ExecCrosstab
	// ExecChart indicates execution using chart (\chart).
	ExecChart
	// ExecWatch indicates repeated execution with a fixed time interval.
	ExecWatch
)

// desc wraps a meta command description.
type desc struct {
	Func       func(*Params) error
	Name       string
	Params     string
	Desc       string
	Hidden     bool
	Deprecated bool
}

// Names returns the names for the command.
func (d desc) Names() []string {
	switch i := strings.Index(d.Name, "["); {
	case i == -1:
		return []string{d.Name}
	case !strings.HasSuffix(d.Name, "]"):
		panic(fmt.Sprintf("invalid command %q", d.Name))
	default:
		name := d.Name[:i]
		v := []string{name}
		for _, s := range d.Name[i+1 : len(d.Name)-1] {
			v = append(v, name+string(s))
		}
		return v
	}
}

// wrap wraps a line of text to the specified width, and adding the prefix to
// each wrapped line.
func wrap(s string, width, prefixWidth int) string {
	words := strings.Fields(strings.TrimSpace(s))
	if len(words) == 0 {
		return ""
	}
	prefix, wrapped := strings.Repeat(" ", prefixWidth), words[0]
	left := width - prefixWidth - len(wrapped)
	for _, word := range words[1:] {
		if left < len(word)+1 {
			wrapped += "\n" + prefix + word
			left = width - len(word)
		} else {
			wrapped += " " + word
			left -= 1 + len(word)
		}
	}
	return wrapped
}
