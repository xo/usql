package metacmd

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenType int

const (
	TokenText TokenType = iota
	TokenSingleQuoted
	TokenDoubleQuoted
	TokenBacktick
	TokenVariable
)

type Token struct {
	Type   TokenType
	Raw    string
	Text   string
	Name   string
	Exists bool
	Parts  []Token
}

type Argument struct {
	Parts []Token
}

func (arg Argument) String() string {
	var buf strings.Builder
	for _, part := range arg.Parts {
		switch part.Type {
		case TokenVariable:
			buf.WriteString(part.Raw)
		case TokenBacktick:
			for _, p := range part.Parts {
				if p.Type == TokenVariable {
					buf.WriteString(p.Raw)
					continue
				}
				buf.WriteString(p.Text)
			}
		default:
			buf.WriteString(part.Text)
		}
	}
	return buf.String()
}

func Parse(s string) ([]string, error) {
	args, err := ParseTokens(s)
	if err != nil {
		return nil, err
	}
	v := make([]string, len(args))
	for i := range args {
		v[i] = args[i].String()
	}
	return v, nil
}

func ParseTokens(s string) ([]Argument, error) {
	p := parser{s: s}
	return p.parse()
}

type parser struct {
	s string
	i int
}

func (p *parser) parse() ([]Argument, error) {
	var args []Argument
	for {
		p.skipSpace()
		if p.done() {
			return args, nil
		}
		arg, err := p.parseArgument(0)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
}

func (p *parser) parseArgument(term rune) (Argument, error) {
	var parts []Token
	for !p.done() {
		r := p.peek()
		if term != 0 && r == term {
			break
		}
		if term == 0 && unicode.IsSpace(r) {
			break
		}
		switch r {
		case '\'':
			part, err := p.parseSingleQuoted()
			if err != nil {
				return Argument{}, err
			}
			parts = append(parts, part)
		case '"':
			part, err := p.parseDoubleQuoted()
			if err != nil {
				return Argument{}, err
			}
			parts = append(parts, part)
		case '`':
			part, err := p.parseBacktick()
			if err != nil {
				return Argument{}, err
			}
			parts = append(parts, part)
		case ':':
			part, ok, err := p.parseVariable()
			if err != nil {
				return Argument{}, err
			}
			if ok {
				parts = append(parts, part)
				continue
			}
			fallthrough
		default:
			part, err := p.parseText(term)
			if err != nil {
				return Argument{}, err
			}
			parts = append(parts, part)
		}
	}
	return Argument{Parts: parts}, nil
}

func (p *parser) parseText(term rune) (Token, error) {
		start := p.i
		var buf strings.Builder
		for !p.done() {
			r := p.peek()
			if term != 0 && r == term {
				break
			}
			if term == 0 && unicode.IsSpace(r) {
				break
			}
			switch r {
			case '\'', '"', '`':
				goto done
			case ':':
				if ok, err := p.isVariable(); err != nil {
					return Token{}, err
				} else if ok {
					goto done
				}
			case '\\':
				p.next()
				if p.done() {
					buf.WriteRune('\\')
					goto done
				}
				buf.WriteRune(p.next())
				continue
			}
			buf.WriteRune(p.next())
		}
	done:
		return Token{Type: TokenText, Raw: p.s[start:p.i], Text: buf.String()}, nil
}

func (p *parser) parseSingleQuoted() (Token, error) {
	start := p.i
	p.next()
	var buf strings.Builder
	for !p.done() {
		r := p.next()
		switch r {
		case '\\':
			if p.done() {
				return Token{}, fmt.Errorf("unterminated quoted string")
			}
			r, err := p.decodeEscape()
			if err != nil {
				return Token{}, err
			}
			buf.WriteRune(r)
		case '\'':
			if !p.done() && p.peek() == '\'' {
				p.next()
				buf.WriteRune('\'')
				continue
			}
			return Token{Type: TokenSingleQuoted, Raw: p.s[start:p.i], Text: buf.String()}, nil
		default:
			buf.WriteRune(r)
		}
	}
	return Token{}, fmt.Errorf("unterminated quoted string")
}

func (p *parser) parseDoubleQuoted() (Token, error) {
	start := p.i
	p.next()
	var buf strings.Builder
	for !p.done() {
		r := p.next()
		switch r {
		case '\\':
			if p.done() {
				return Token{}, fmt.Errorf("unterminated quoted string")
			}
			buf.WriteRune(p.next())
		case '"':
			if !p.done() && p.peek() == '"' {
				p.next()
				buf.WriteRune('"')
				continue
			}
			return Token{Type: TokenDoubleQuoted, Raw: p.s[start:p.i], Text: buf.String()}, nil
		default:
			buf.WriteRune(r)
		}
	}
	return Token{}, fmt.Errorf("unterminated quoted string")
}

func (p *parser) parseBacktick() (Token, error) {
	start := p.i
	p.next()
	arg, err := p.parseArgument('`')
	if err != nil {
		return Token{}, err
	}
	if p.done() || p.peek() != '`' {
		return Token{}, fmt.Errorf("unterminated quoted string")
	}
	p.next()
	var buf strings.Builder
	for _, part := range arg.Parts {
		if part.Type == TokenVariable {
			buf.WriteString(part.Raw)
			continue
		}
		buf.WriteString(part.Text)
	}
	return Token{Type: TokenBacktick, Raw: p.s[start:p.i], Text: buf.String(), Parts: arg.Parts}, nil
}

func (p *parser) parseVariable() (Token, bool, error) {
	start := p.i
	if p.done() || p.peek() != ':' {
		return Token{}, false, nil
	}
	p.next()
	if p.done() {
		p.i = start
		return Token{}, false, nil
	}
	switch p.peek() {
	case '{':
		p.next()
		exists := false
		if !p.done() && p.peek() == '?' {
			exists = true
			p.next()
		}
		name, ok := p.readVarName()
		if !ok || p.done() || p.peek() != '}' {
			p.i = start
			return Token{}, false, nil
		}
		p.next()
		return Token{Type: TokenVariable, Raw: p.s[start:p.i], Name: name, Exists: exists}, true, nil
	case '\'', '"':
		q := p.next()
		name, err := p.readQuotedName(q)
		if err != nil {
			return Token{}, false, err
		}
		return Token{Type: TokenVariable, Raw: p.s[start:p.i], Name: name}, true, nil
	default:
		name, ok := p.readVarName()
		if !ok {
			p.i = start
			return Token{}, false, nil
		}
		return Token{Type: TokenVariable, Raw: p.s[start:p.i], Name: name}, true, nil
	}
}

func (p *parser) isVariable() (bool, error) {
	i := p.i
	_, ok, err := p.parseVariable()
	p.i = i
	return ok, err
}

func (p *parser) readQuotedName(q rune) (string, error) {
	var buf strings.Builder
	for !p.done() {
		r := p.next()
		switch r {
		case '\\':
			if p.done() {
				return "", fmt.Errorf("unterminated quoted string")
			}
			buf.WriteRune(p.next())
		case q:
			if !p.done() && p.peek() == q {
				p.next()
				buf.WriteRune(q)
				continue
			}
			return buf.String(), nil
		default:
			buf.WriteRune(r)
		}
	}
	return "", fmt.Errorf("unterminated quoted string")
}

func (p *parser) readVarName() (string, bool) {
	start := p.i
	for !p.done() {
		r := p.peek()
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			p.next()
			continue
		}
		break
	}
	if start == p.i {
		return "", false
	}
	return p.s[start:p.i], true
}

func (p *parser) decodeEscape() (rune, error) {
	r := p.next()
	switch r {
	case 'b':
		return '\b', nil
	case 'f':
		return '\f', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	case 'v':
		return '\v', nil
	case '\\':
		return '\\', nil
	case '\'':
		return '\'', nil
	case '"':
		return '"', nil
	case 'x':
		v, n := p.readBaseN(16, 2)
		if n == 0 {
			return 'x', nil
		}
		return rune(v), nil
	case '0', '1', '2', '3', '4', '5', '6', '7':
		p.backup(r)
		v, _ := p.readBaseN(8, 3)
		return rune(v), nil
	default:
		return r, nil
	}
}

func (p *parser) readBaseN(base, max int) (int64, int) {
	var v int64
	var n int
	for !p.done() && n < max {
		r := p.peek()
		d := digitValue(r)
		if d < 0 || d >= base {
			break
		}
		p.next()
		v = v*int64(base) + int64(d)
		n++
	}
	return v, n
}

func digitValue(r rune) int {
	switch {
	case '0' <= r && r <= '9':
		return int(r - '0')
	case 'a' <= r && r <= 'f':
		return int(r-'a') + 10
	case 'A' <= r && r <= 'F':
		return int(r-'A') + 10
	default:
		return -1
	}
}

func (p *parser) skipSpace() {
	for !p.done() && unicode.IsSpace(p.peek()) {
		p.next()
	}
}

func (p *parser) done() bool {
	return p.i >= len(p.s)
}

func (p *parser) peek() rune {
	r, _ := utf8.DecodeRuneInString(p.s[p.i:])
	return r
}

func (p *parser) next() rune {
	r, n := utf8.DecodeRuneInString(p.s[p.i:])
	p.i += n
	return r
}

func (p *parser) backup(r rune) {
	p.i -= utf8.RuneLen(r)
}
