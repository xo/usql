package stmt

import (
	"unicode"

	"github.com/xo/usql/text"
)

// Params holds information about command parameters.
type Params struct {
	R   []rune
	Len int
}

// NewParams creates command parameters.
func NewParams(params string) *Params {
	r := []rune(params)
	return &Params{
		R:   r,
		Len: len(r),
	}
}

// Raw reads all remaining runes. No substitution or whitespace removal is
// performed.
func (p *Params) Raw() string {
	s := string(p.R)
	p.R, p.Len = p.R[:0], 0
	return s
}

// Next reads the next command parameter using the provided substitution func.
// True indicates there are runes remaining in the command parameters to
// process.
func (p *Params) Next(unquote func(string, bool) (string, bool, error)) (string, bool, error) {
	i, _ := findNonSpace(p.R, 0, p.Len)
	if i >= p.Len {
		return "", false, nil
	}
	var ok bool
	var quote rune
	start := i
loop:
	for ; i < p.Len; i++ {
		c, next := p.R[i], grab(p.R, i+1, p.Len)
		switch {
		case quote != 0:
			start := i - 1
			i, ok = readString(p.R, i, p.Len, quote, "")
			if !ok {
				break loop
			}
			switch z, ok, err := unquote(string(p.R[start:i+1]), false); {
			case err != nil:
				return "", false, err
			case ok:
				p.R, p.Len = substitute(p.R, start, p.Len, i-start+1, z)
				i = start + len([]rune(z)) - 1
			}
			quote = 0
		// start of single, double, or backtick string
		case c == '\'' || c == '"' || c == '`':
			quote = c
		case c == ':' && next != ':':
			if v := readVar(p.R, i, p.Len, next); v != nil {
				switch z, ok, err := unquote(v.Name, true); {
				case err != nil:
					return "", false, err
				case ok || v.Quote == '?':
					p.R, p.Len = v.Substitute(p.R, z, ok)
					i += v.Len - 1
				default:
					i = v.End - 1
				}
			}
		case unicode.IsSpace(c):
			break loop
		}
	}
	if quote != 0 {
		return "", false, text.ErrUnterminatedQuotedString
	}
	s := string(p.R[start:i])
	p.R = p.R[i:]
	p.Len = len(p.R)
	return s, true, nil
}

// All retrieves all remaining command parameters using the provided
// substitution func. Will return on the first encountered error.
func (p *Params) All(unquote func(string, bool) (string, bool, error)) ([]string, error) {
	var v []string
loop:
	for {
		switch s, ok, err := p.Next(unquote); {
		case err != nil:
			return v, err
		case !ok:
			break loop
		default:
			v = append(v, s)
		}
	}
	return v, nil
}

// Arg retrieves the next argument, without decoding.
func (p *Params) Arg() (string, bool, error) {
	return p.Next(func(s string, _ bool) (string, bool, error) {
		return s, true, nil
	})
}
