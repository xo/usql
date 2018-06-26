// Package stmt contains a statement buffer implementation.
package stmt

import (
	"bytes"

	"github.com/xo/usql/env"
)

const (
	// MinCapIncrease is the minimum amount by which to grow a Stmt.Buf.
	MinCapIncrease = 512
)

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
}

// Stmt is a reusable statement buffer that handles reading and parsing
// SQL-like statements.
type Stmt struct {
	// f is the rune source.
	f func() ([]rune, error)

	// parse settings
	allowDollar, allowMultilineComments, allowCComments, allowHashComments bool

	// Buf is the statement buffer.
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

	// quoted string state
	quote       bool
	quoteDouble bool
	quoteDollar bool
	quoteTagID  string

	// multicomment state
	multilineComment bool

	// balanced paren count
	balanceCount int

	// ready is the state
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
		if v.Len == 0 {
			continue
		}
		z.WriteString(s[i:v.I])
		z.WriteRune(':')
		if v.Quote != 0 {
			z.WriteRune(v.Quote)
		}
		z.WriteString(v.Name)
		if v.Quote != 0 {
			z.WriteRune(v.Quote)
		}
		i = v.I + v.Len
	}

	// add remaining
	z.WriteString(s[i:])
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
	b.quote, b.quoteDouble, b.quoteDollar, b.quoteTagID = false, false, false, ""

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
//         cmd, params, err := buf.Next()
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
func (b *Stmt) Next() (string, []string, error) {
	var err error
	var i int

	// no runes to process, grab more
	if b.rlen == 0 {
		b.r, err = b.f()
		if err != nil {
			return "", nil, err
		}
		b.rlen = len(b.r)
	}

	var cmd string
	var params []string

parse:
	for ; i < b.rlen; i++ {
		//log.Printf(">> (%c) %d", b.r[i], i)

		// grab c, next
		c, next := b.r[i], grab(b.r, i+1, b.rlen)
		switch {
		// find end of string quote
		case b.quote:
			pos, ok := readString(b.r, i, b.rlen, b)
			i = pos
			if ok {
				b.quote, b.quoteDouble, b.quoteDollar, b.quoteTagID = false, false, false, ""
			}

		// find end of multiline comment
		case b.multilineComment:
			pos, ok := readMultilineComment(b.r, i, b.rlen)
			i, b.multilineComment = pos, !ok

		// start of single quoted string
		case c == '\'':
			b.quote = true

		// start of double quoted string
		case c == '"':
			b.quote, b.quoteDouble = true, true

		// start of dollar quoted string literal (postgres)
		case b.allowDollar && c == '$':
			id, pos, ok := readDollarAndTag(b.r, i, b.rlen)
			if ok {
				b.quote, b.quoteDollar, b.quoteTagID = true, true, id
			}
			i = pos

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
				if ok, z, _ := env.Getvar(q + v.Name + q); ok {
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

		// continue processing
		case b.quote || b.multilineComment || b.balanceCount != 0:
			continue

		// skip escaped backslash
		case c == '\\' && next == '\\':
			i++

		// start of command
		case c == '\\':
			// extract command from r
			var pos int
			cmd, params, pos = readCommand(b.r, i, b.rlen)
			b.r = append(b.r[:i], b.r[pos:]...)
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
	appendLine := b.quote || b.multilineComment || !empty
	if !b.multilineComment && cmd != "" && empty {
		appendLine = false
	}
	if appendLine {
		// skip leading space when empty
		st := 0
		if b.Len == 0 {
			st, _ = findNonSpace(b.r, 0, i)
		}

		//log.Printf(">> appending: `%s`", string(r[st:i]))
		b.Append(b.r[st:i], lineend)
	}

	// set prefix
	b.Prefix = findPrefix(b.Buf, prefixWords)

	// reset r
	b.r = b.r[i:]
	b.rlen = len(b.r)

	//log.Printf("returning from NEXT: `%s`", string(b.Buf))
	//log.Printf(">>>>>>>>>>>> REMAIN: `%s`", string(b.r))
	//log.Printf(">>>>>>>>>>>>    CMD: `%s`", cmd)
	//log.Printf(">>>>>>>>>>>> PARAMS: %v", params)

	return cmd, params, nil
}

// Append appends r to b.Buf separated by sep when b.Buf is not already empty.
//
// Append dynamically grows b.Buf as necessary to accommodate r and the
// separator. Specifically, when b.Buf is not empty, b.Buf will grow by
// increments of MinCapIncrease.
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
	case b.quote && b.quoteDollar:
		return "$"

	case b.quote && b.quoteDouble:
		return `"`

	case b.quote:
		return "'"

	case b.multilineComment:
		return "*"

	case b.balanceCount != 0:
		return "("

	case b.Len != 0:
		return "-"
	}

	return "="
}
