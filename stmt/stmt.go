// Package stmt contains a statement buffer implementation.
package stmt

import (
	"bytes"
)

// MinCapIncrease is the minimum amount by which to grow a Stmt.Buf.
const MinCapIncrease = 512

// Var holds information about a variable.
type Var struct {
	// I is where the variable starts (ie, ':') in Stmt.Buf.
	I int
	// End is where the variable ends in Stmt.Buf.
	End int
	// Quote is the quote character used if the variable was quoted, 0
	// otherwise.
	Quote rune
	// Name is the actual variable name excluding ':' and any enclosing quote
	// characters.
	Name string
	// Len is the length of the replaced variable.
	Len int
	// Exist indicates whether the variable has been defined.
	Exist bool
}

// String satisfies the fmt.Stringer interface.
func (v *Var) String() string {
	var q string
	switch {
	case v.Quote == '\\':
		return "\\" + v.Name
	case v.Quote != 0:
		q = string(v.Quote)
	}
	return ":" + q + v.Name + q
}

// Stmt is a reusable statement buffer that handles reading and parsing
// SQL-like statements.
type Stmt struct {
	// f is the rune source.
	f func() ([]rune, error)
	// allowDollar allows dollar quoted strings (ie, $$ ... $$ or $tag$ ... $tag$).
	allowDollar bool
	// allowMultilineComments allows multiline comments (ie, /* ... */)
	allowMultilineComments bool
	// allowCComments allows C-style comments (ie, // ... )
	allowCComments bool
	// allowHashComments allows hash comments (ie, # ... )
	allowHashComments bool
	// Buf is the statement buffer
	Buf []rune
	// Len is the current len of any statement in Buf.
	Len int
	// Prefix is the detected prefix of the statement.
	Prefix string
	// Vars is the list of encountered variables.
	Vars []*Var
	// r is the unprocessed runes.
	r []rune
	// rlen is the number of unprocessed runes.
	rlen int
	// quote indicates currently parsing a quoted string.
	quote rune
	// quoteDollarTag is the parsed tag of a dollar quoted string
	quoteDollarTag string
	// multilineComment is state of multiline comment processing
	multilineComment bool
	// balanceCount is the balanced paren count
	balanceCount int
	// ready indicates that a complete statement has been parsed
	ready bool
}

// New creates a new Stmt using the supplied rune source f.
func New(f func() ([]rune, error), opts ...Option) *Stmt {
	b := &Stmt{
		f: f,
	}
	// apply opts
	for _, o := range opts {
		o(b)
	}
	return b
}

// String satisfies fmt.Stringer.
func (b *Stmt) String() string {
	return string(b.Buf)
}

// RawString returns the non-interpolated version of the statement buffer.
func (b *Stmt) RawString() string {
	if b.Len == 0 {
		return ""
	}
	s, z := string(b.Buf), new(bytes.Buffer)
	var i int
	// deinterpolate vars
	for _, v := range b.Vars {
		if !v.Exist {
			continue
		}
		if len(s) > i {
			z.WriteString(s[i:v.I])
		}
		if v.Quote != '\\' {
			z.WriteRune(':')
		}
		if v.Quote != 0 {
			z.WriteRune(v.Quote)
		}
		z.WriteString(v.Name)
		if v.Quote != 0 && v.Quote != '\\' {
			z.WriteRune(v.Quote)
		}
		i = v.I + v.Len
	}
	// add remaining
	if len(s) > i {
		z.WriteString(s[i:])
	}
	return z.String()
}

// Ready returns true when the statement buffer contains a non empty, balanced
// statement that has been properly terminated (ie, ended with a semicolon).
func (b *Stmt) Ready() bool {
	return b.ready
}

// Reset resets the statement buffer.
func (b *Stmt) Reset(r []rune) {
	// reset buf
	b.Buf, b.Len, b.Prefix, b.Vars = nil, 0, "", nil
	// quote state
	b.quote, b.quoteDollarTag = 0, ""
	// multicomment state
	b.multilineComment = false
	// balance state
	b.balanceCount = 0
	// ready state
	b.ready = false
	if r != nil {
		b.r, b.rlen = r, len(r)
	}
}

// lineend is the slice to use when appending a line.
var lineend = []rune{'\n'}

// Next reads the next statement from the rune source, returning when either
// the statement has been terminated, or a meta command has been read from the
// rune source. After a call to Next, the collected statement is available in
// Stmt.Buf, or call Stmt.String() to convert it to a string.
//
// After a call to Next, Reset should be called if the extracted statement was
// executed (ie, processed). Note that the rune source supplied to New will be
// called again only after any remaining collected runes have been processed.
//
// Example:
//     buf := stmt.New(runeSrc)
//     for {
//         cmd, params, err := buf.Next(unquoteFunc)
//         if err { /* ... */ }
//
//         execute, quit := buf.Ready() || cmd == "g", cmd == "q"
//
//         // process command ...
//         switch cmd {
//             /* ... */
//         }
//
//         if quit {
//             break
//         }
//
//         if execute {
//            s := buf.String()
//            res, err := db.Query(s)
//            /* handle database ... */
//            buf.Reset(nil)
//         }
//     }
func (b *Stmt) Next(unquote func(string, bool) (bool, string, error)) (string, string, error) {
	var err error
	var i int
	// no runes to process, grab more
	if b.rlen == 0 {
		b.r, err = b.f()
		if err != nil {
			return "", "", err
		}
		b.rlen = len(b.r)
	}
	var cmd, params string
	var ok bool
parse:
	for ; i < b.rlen; i++ {
		// log.Printf(">> (%c) %d", b.r[i], i)
		// grab c, next
		c, next := b.r[i], grab(b.r, i+1, b.rlen)
		switch {
		// find end of string
		case b.quote != 0:
			i, ok = readString(b.r, i, b.rlen, b.quote, b.quoteDollarTag)
			if ok {
				b.quote, b.quoteDollarTag = 0, ""
			}
		// find end of multiline comment
		case b.multilineComment:
			i, ok = readMultilineComment(b.r, i, b.rlen)
			b.multilineComment = !ok
		// start of single or double quoted string
		case c == '\'' || c == '"':
			b.quote = c
		// start of dollar quoted string literal (postgres)
		case b.allowDollar && c == '$':
			var id string
			id, i, ok = readDollarAndTag(b.r, i, b.rlen)
			if ok {
				b.quote, b.quoteDollarTag = '$', id
			}
		// start of sql comment, skip to end of line
		case c == '-' && next == '-':
			i = b.rlen
		// start of c-style comment, skip to end of line
		case b.allowCComments && c == '/' && next == '/':
			i = b.rlen
		// start of hash comment, skip to end of line
		case b.allowHashComments && c == '#':
			i = b.rlen
		// start of multiline comment
		case b.allowMultilineComments && c == '/' && next == '*':
			b.multilineComment = true
			i++
		// variable declaration
		case c == ':' && next != ':':
			if v := readVar(b.r, i, b.rlen); v != nil {
				var q string
				if v.Quote != 0 {
					q = string(v.Quote)
				}
				b.Vars = append(b.Vars, v)
				if ok, z, _ := unquote(q+v.Name+q, true); ok {
					v.Exist = true
					b.r, b.rlen = substituteVar(b.r, v, z)
					i--
				}
				if b.Len != 0 {
					v.I += b.Len + 1
				}
			}
		// unbalance
		case c == '(':
			b.balanceCount++
		// balance
		case c == ')':
			b.balanceCount = max(0, b.balanceCount-1)
		// continue processing quoted string, multiline comment, or unbalanced statements
		case b.quote != 0 || b.multilineComment || b.balanceCount != 0:
		// skip escaped backslash, semicolon, colon
		case c == '\\' && (next == '\\' || next == ';' || next == ':'):
			// FIXME: the below works, but it may not make sense to keep this enabled.
			// FIXME: also, the behavior is slightly different than psql
			v := &Var{
				I:     i,
				End:   i + 2,
				Quote: '\\',
				Name:  string(next),
			}
			b.Vars = append(b.Vars, v)
			b.r, b.rlen = substituteVar(b.r, v, string(next))
			if b.Len != 0 {
				v.I += b.Len + 1
			}
		// start of command
		case c == '\\':
			// parse command and params end positions
			cend, pend := readCommand(b.r, i, b.rlen)
			cmd, params = string(b.r[i:cend]), string(b.r[cend:pend])
			// remove command and params from r
			b.r = append(b.r[:i], b.r[pend:]...)
			b.rlen = len(b.r)
			break parse
		// terminated
		case c == ';':
			b.ready = true
			i++
			break parse
		}
	}
	// fix i -- i will be +1 when passing the length, which is a problem as the
	// '\n' will get copied from the source.
	i = min(i, b.rlen)
	// append line to buf when:
	// 1. in a quoted string (ie, ', ", or $)
	// 2. in a multiline comment
	// 3. line is not empty
	//
	// DO NOT append to buf when:
	// 1. line is empty/whitespace and not in a string/multiline comment
	empty := isEmptyLine(b.r, 0, i)
	appendLine := b.quote != 0 || b.multilineComment || !empty
	if !b.multilineComment && cmd != "" && empty {
		appendLine = false
	}
	if appendLine {
		// skip leading space when empty
		st := 0
		if b.Len == 0 {
			st, _ = findNonSpace(b.r, 0, i)
		}
		// log.Printf(">> appending: `%s`", string(r[st:i]))
		b.Append(b.r[st:i], lineend)
	}
	// set prefix
	b.Prefix = findPrefix(b.Buf, prefixCount)
	// reset r
	b.r = b.r[i:]
	b.rlen = len(b.r)
	// log.Printf("returning from NEXT: `%s`", string(b.Buf))
	// log.Printf(">>>>>>>>>>>> REMAIN: `%s`", string(b.r))
	// log.Printf(">>>>>>>>>>>>    CMD: `%s`", cmd)
	// log.Printf(">>>>>>>>>>>> PARAMS: %v", params)
	return cmd, params, nil
}

// Append appends r to b.Buf separated by sep when b.Buf is not already empty.
//
// Dynamically grows b.Buf as necessary to accommodate r and the separator.
// Specifically, when b.Buf is not empty, b.Buf will grow by increments of
// MinCapIncrease.
//
// After a call to Append, b.Len will be len(b.Buf)+len(sep)+len(r). Call Reset
// to reset the Buf.
func (b *Stmt) Append(r, sep []rune) {
	rlen := len(r)
	// initial
	if b.Buf == nil {
		b.Buf, b.Len = r, rlen
		return
	}
	blen, seplen := b.Len, len(sep)
	tlen := blen + rlen + seplen
	// grow
	if bcap := cap(b.Buf); tlen > bcap {
		n := tlen + 2*rlen
		n += MinCapIncrease - (n % MinCapIncrease)
		z := make([]rune, blen, n)
		copy(z, b.Buf)
		b.Buf = z
	}
	b.Buf = b.Buf[:tlen]
	copy(b.Buf[blen:], sep)
	copy(b.Buf[blen+seplen:], r)
	b.Len = tlen
}

// AppendString is a util func wrapping Append.
func (b *Stmt) AppendString(s, sep string) {
	b.Append([]rune(s), []rune(sep))
}

// State returns a string representing the state of statement parsing.
func (b *Stmt) State() string {
	switch {
	case b.quote != 0:
		return string(b.quote)
	case b.multilineComment:
		return "*"
	case b.balanceCount != 0:
		return "("
	case b.Len != 0:
		return "-"
	}
	return "="
}

// Option is a statement buffer option.
type Option func(*Stmt)

// WithAllowDollar is a statement buffer option to set allowing dollar strings (ie,
// $$text$$ or $tag$text$tag$).
func WithAllowDollar(enable bool) Option {
	return func(b *Stmt) {
		b.allowDollar = enable
	}
}

// WithAllowMultilineComments is a statement buffer option to set allowing multiline comments
// (ie, /* ... */).
func WithAllowMultilineComments(enable bool) Option {
	return func(b *Stmt) {
		b.allowMultilineComments = enable
	}
}

// WithAllowCComments is a statement buffer option to set allowing C-style comments
// (ie, // ...).
func WithAllowCComments(enable bool) Option {
	return func(b *Stmt) {
		b.allowCComments = enable
	}
}

// WithAllowHashComments is a statement buffer option to set allowing hash comments
// (ie, # ...).
func WithAllowHashComments(enable bool) Option {
	return func(b *Stmt) {
		b.allowHashComments = enable
	}
}
