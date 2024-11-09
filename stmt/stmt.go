// Package stmt contains a statement buffer implementation.
package stmt

import (
	"bytes"
	"unicode"
)

// minCapIncrease is the minimum amount by which to grow a Stmt.Buf.
const minCapIncrease = 512

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

// PrintString returns a print string of the statement buffer, which is the
// statement buffer but with escaped variables re-interpolated.
func (b *Stmt) PrintString() string {
	if b.Len == 0 {
		return ""
	}
	i, s, w := 0, string(b.Buf), new(bytes.Buffer)
	// deinterpolate vars
	for _, v := range b.Vars {
		if v.Quote != '\\' {
			continue
		}
		if len(s) > i {
			w.WriteString(s[i:v.I])
		}
		w.WriteString(v.String())
		i = v.I + v.Len
	}
	// add remaining
	if len(s) > i {
		w.WriteString(s[i:])
	}
	return w.String()
}

// RawString returns the non-interpolated version of the statement buffer.
func (b *Stmt) RawString() string {
	if b.Len == 0 {
		return ""
	}
	i, s, z := 0, string(b.Buf), new(bytes.Buffer)
	// deinterpolate vars
	for _, v := range b.Vars {
		if !v.Defined && v.Quote != '\\' {
			continue
		}
		if len(s) > i {
			z.WriteString(s[i:v.I])
		}
		z.WriteString(v.String())
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
	b.Buf, b.Len, b.Prefix, b.Vars = nil, 0, "", b.Vars[:0]
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
//
//	buf := stmt.New(runeSrc)
//	for {
//	    cmd, params, err := buf.Next(unquoteFunc)
//	    if err { /* ... */ }
//
//	    execute, quit := buf.Ready() || cmd == "g", cmd == "q"
//
//	    // process command ...
//	    switch cmd {
//	        /* ... */
//	    }
//
//	    if quit {
//	        break
//	    }
//
//	    if execute {
//	       s := buf.String()
//	       res, err := db.Query(s)
//	       /* handle database ... */
//	       buf.Reset(nil)
//	    }
//	}
func (b *Stmt) Next(unquote func(string, bool) (string, bool, error)) (string, string, error) {
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
		// fmt.Fprintf(os.Stderr, "> %d: `%s`\n", i, string(b.r[i:]))
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
		case b.allowDollar && c == '$' && (next == '$' || next == '_' || unicode.IsLetter(next)):
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
			if v := readVar(b.r, i, b.rlen, next); v != nil {
				b.Vars = append(b.Vars, v)
				z, ok, _ := unquote(v.Name, true)
				if v.Defined = ok || v.Quote == '?'; v.Defined {
					b.r, b.rlen = v.Substitute(b.r, z, ok)
				}
				if b.Len != 0 {
					v.I += b.Len + 1
				}
			}
		// skip escaped backslash, semicolon, colon
		case c == '\\' && (next == '\\' || next == ';' || next == ':'):
			v := &Var{
				I:     i,
				End:   i + 2,
				Quote: '\\',
				Name:  string(next),
			}
			b.Vars = append(b.Vars, v)
			if b.r, b.rlen = v.Substitute(b.r, string(next), false); b.Len != 0 {
				v.I += b.Len + 1
			}
		// unbalance
		case c == '(':
			b.balanceCount++
		// balance
		case c == ')':
			b.balanceCount = max(0, b.balanceCount-1)
		// continue processing quoted string, multiline comment, or unbalanced statements
		case b.quote != 0 || b.multilineComment || b.balanceCount != 0:
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
	b.Prefix = findPrefix(b.Buf, prefixCount, b.allowCComments, b.allowHashComments, b.allowMultilineComments)
	// reset r
	b.r = b.r[i:]
	b.rlen = len(b.r)
	/*
		fmt.Fprintf(os.Stderr, "\n------------------------------\n")
		fmt.Fprintf(os.Stderr, "    NEXT: `%s`\n", string(b.Buf))
		fmt.Fprintf(os.Stderr, "  REMAIN: `%s`\n", string(b.r))
		fmt.Fprintf(os.Stderr, "     CMD: `%s`\n", cmd)
		fmt.Fprintf(os.Stderr, "  PARAMS: %v\n", params)
	*/
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
		n += minCapIncrease - (n % minCapIncrease)
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
	// Defined indicates whether the variable has been defined.
	Defined bool
}

// String satisfies the fmt.Stringer interface.
func (v *Var) String() string {
	switch v.Quote {
	case '\\':
		return "\\" + v.Name
	case '\'', '"':
		return ":" + string(v.Quote) + v.Name + string(v.Quote)
	case '?':
		return ":{?" + v.Name + "}"
	}
	return ":" + v.Name
}

// Substitute substitutes part of r, with s.
func (v *Var) Substitute(r []rune, s string, ok bool) ([]rune, int) {
	switch v.Quote {
	case '?':
		s = trueFalse(ok)
	case '\'', '"':
		s = string(v.Quote) + s + string(v.Quote)
	}
	// fmt.Fprintf(os.Stderr, "orig: %q repl: %q\n", string(r), s)
	sr, rcap := []rune(s), cap(r)
	v.Len = len(sr)
	// grow ...
	tlen := len(r) + v.Len - (v.End - v.I)
	if tlen > rcap {
		z := make([]rune, tlen)
		copy(z, r)
		r = z
	} else {
		r = r[:rcap]
	}
	// substitute
	copy(r[v.I+v.Len:], r[v.End:])
	copy(r[v.I:v.I+v.Len], sr)
	return r[:tlen], tlen
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

// isSpaceOrControl is a special test for either a space or a control (ie, \b)
// characters.
func isSpaceOrControl(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r)
}

// lastIndex returns the last index in r of needle, or -1 if not found.
func lastIndex(r []rune, needle rune) int {
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] == needle {
			return i
		}
	}
	return -1
}

// trueFalse returns TRUE or FALSE.
func trueFalse(ok bool) string {
	if ok {
		return "TRUE"
	}
	return "FALSE"
}

// lineend is the slice to use when appending a line.
var lineend = []rune{'\n'}
