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
	"github.com/xo/rline"
	"github.com/xo/usql/feed"
	"github.com/xo/usql/text"
)

// Handler is the shared interface for a command handler.
type Handler interface {
	// IO returns the session of the line editor, which is also where the
	// handler writes.
	IO() *rline.Prompt
	// User returns the current user.
	User() *user.User
	// URL returns the current database URL.
	URL() *dburl.URL
	// DB returns the current database connection.
	DB() drivers.DB
	// LastExec returns the last executed query.
	LastExec() string
	// LastRaw returns the last raw (non-interpolated) query.
	LastRaw() string
	// Buffer returns the query buffer.
	Buffer() *feed.Buffer
	// Lookup returns the value of a variable, for expanding the buffer.
	Lookup(string) (string, bool)
	// Reset empties the query buffer and forgets the last query.
	Reset()
	// Push puts text in front of the input, as \e does with the edited query.
	Push(string)
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
	// GetTiming mode.
	GetTiming() bool
	// SetTiming mode.
	SetTiming(bool)
	// GetOutput writer.
	GetOutput() io.Writer
	// SetOutput writer.
	SetOutput(io.WriteCloser)
	// MetadataWriter retrieves the metadata writer for the handler.
	MetadataWriter(context.Context) (*metadata.Writer, error)
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

// Decode returns the function that runs the meta command it. The function
// returns the options of the command, and each argument that the command did
// not read.
func Decode(it feed.Item) (func(Handler) (Option, []string, error), error) {
	name := it.Name
	f, ok := cmds[name]
	if !ok {
		f, ok = cmds[relations(name)]
	}
	if !ok || name == "" {
		return nil, text.ErrUnknownCommand
	}
	return func(h Handler) (Option, []string, error) {
		p := &Params{
			Handler: h,
			Name:    name,
			item:    it,
		}
		err := f(p)
		return p.Option, p.Unread(), err
	}, nil
}

// relations returns name with the class letters of \d in the order that
// they are registered in, so that \dvt finds \dtv, as psql allows.
func relations(name string) string {
	base := strings.TrimRight(name, "S+")
	if len(base) < 2 || base[0] != 'd' || strings.Trim(base[1:], "tvmsE") != "" {
		return name
	}
	var b strings.Builder
	b.WriteByte('d')
	for _, c := range "tvmsE" {
		if strings.ContainsRune(base[1:], c) {
			b.WriteRune(c)
		}
	}
	return b.String() + name[len(base):]
}

// Params are the name and the arguments of one meta command, as the usql
// grammar parsed them. A command reads its arguments in order, and the
// handler reports the ones that it did not read.
type Params struct {
	// Handler is the process handler.
	Handler Handler
	// Name is the name of the metacmd.
	Name string
	// Option contains resulting command execution options.
	Option Option

	// item is the meta command.
	item feed.Item
	// next is the index of the next argument to read.
	next int
	// options is set once the option list of \g has been read.
	options bool
}

// arg returns the next argument that is not an option list, and moves past
// it. It returns false when no argument is left.
func (p *Params) arg() (feed.Arg, bool) {
	for p.next < len(p.item.Args) {
		a := p.item.Args[p.next]
		p.next++
		if a.Parts[0].Kind != feed.Options {
			return a, true
		}
	}
	return feed.Arg{}, false
}

// value returns the value of an argument.
//
// A string and a quoted name lose their quotes, and a variable is replaced by
// its value: :'name' as a quoted string and :"name" as a quoted name, as in
// psql. A shell command between backticks runs when exec is set. A quoted
// name keeps its quotes when pattern is set, so that a psql pattern can tell
// it from a name that the database folds.
func (p *Params) value(a feed.Arg, exec, pattern bool) (string, error) {
	var b strings.Builder
	for _, part := range a.Parts {
		raw := p.item.Text[part.Start:part.End]
		switch part.Kind {
		case feed.String:
			s, err := env.Unquote(raw)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		case feed.QuotedIdentifier:
			if pattern {
				b.WriteString(raw)
				continue
			}
			s, err := env.Unquote(raw)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		case feed.Variable:
			v, ok := env.Vars().Get(part.Name)
			switch {
			case part.Quote == '?' && ok:
				b.WriteString("TRUE")
			case part.Quote == '?':
				b.WriteString("FALSE")
			case !ok:
				b.WriteString(raw)
			case part.Quote == '\'' || part.Quote == '"':
				q := string(part.Quote)
				b.WriteString(q + strings.ReplaceAll(v, q, q+q) + q)
			default:
				b.WriteString(v)
			}
		case feed.Backtick:
			if !exec {
				b.WriteString(raw)
				continue
			}
			out, err := env.Exec(part.Value)
			if err != nil {
				return "", err
			}
			b.WriteString(out)
		case feed.Pipe:
			b.WriteString("|" + part.Value)
		case feed.Shell:
			b.WriteString(part.Value)
		default:
			b.WriteString(raw)
		}
	}
	return b.String(), nil
}

// Next returns the value of the next argument, or an empty string when none
// is left. A shell command between backticks runs when exec is set.
func (p *Params) Next(exec bool) (string, error) {
	v, _, err := p.NextOK(exec)
	return v, err
}

// NextOK returns the value of the next argument, and false when none is left.
func (p *Params) NextOK(exec bool) (string, bool, error) {
	a, ok := p.arg()
	if !ok {
		return "", false, nil
	}
	v, err := p.value(a, exec, false)
	return v, err == nil, err
}

// NextOpt returns the value of the next argument. It returns true when the
// value starts with a "-", along with the value without the "-".
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

// All returns the values of the arguments that are left.
func (p *Params) All(exec bool) ([]string, error) {
	var v []string
	for {
		s, ok, err := p.NextOK(exec)
		switch {
		case err != nil:
			return nil, err
		case !ok:
			return v, nil
		}
		v = append(v, s)
	}
}

// NextPattern returns the value of the next argument as a psql pattern, in
// which a quoted name keeps its quotes.
func (p *Params) NextPattern() (string, error) {
	a, ok := p.arg()
	if !ok {
		return "", nil
	}
	return p.value(a, true, true)
}

// Patterns returns the values of the arguments that are left, as psql
// patterns. See NextPattern.
func (p *Params) Patterns() ([]string, error) {
	var v []string
	for {
		a, ok := p.arg()
		if !ok {
			return v, nil
		}
		s, err := p.value(a, true, true)
		if err != nil {
			return nil, err
		}
		v = append(v, s)
	}
}

// Raw returns the shell command of \!, as it is written.
func (p *Params) Raw() string {
	a, ok := p.arg()
	if !ok {
		return ""
	}
	s, _ := p.value(a, false, false)
	return s
}

// Options returns the keys and values of the option list of \g and its forms,
// such as (format=csv), in order.
func (p *Params) Options() []feed.Option {
	p.options = true
	for _, a := range p.item.Args {
		if a.Parts[0].Kind == feed.Options {
			return a.Parts[0].Options
		}
	}
	return nil
}

// Unread returns each argument that the command did not read, as it is
// written, so that the handler can report it as ignored.
func (p *Params) Unread() []string {
	var v []string
	for _, a := range p.item.Args[p.next:] {
		if a.Parts[0].Kind == feed.Options && p.options {
			continue
		}
		first, last := a.Parts[0], a.Parts[len(a.Parts)-1]
		v = append(v, p.item.Text[first.Start:last.End])
	}
	return v
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

// Names returns the names for the command. Each group of letters between
// brackets adds every combination of its letters, in their order, so d[S+]
// is d, dS, d+ and dS+, and df[an][S+] adds dfa, dfan, dfaS+ and the rest.
func (d desc) Names() []string {
	i := strings.Index(d.Name, "[")
	if i == -1 {
		return []string{d.Name}
	}
	names := []string{d.Name[:i]}
	for s := d.Name[i:]; s != ""; {
		j := strings.Index(s, "]")
		if s[0] != '[' || j == -1 {
			panic(fmt.Sprintf("invalid command %q", d.Name))
		}
		for _, c := range s[1:j] {
			for _, n := range names {
				names = append(names, n+string(c))
			}
		}
		s = s[j+1:]
	}
	return names
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
