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

// DecodeParams decodes command parameters.
func DecodeParams(params string) *Params {
	r := []rune(params)
	return &Params{
		R:   r,
		Len: len(r),
	}
}

// GetRaw reads all remaining runes. No substitution or whitespace removal is
// performed.
func (p *Params) GetRaw() string {
	s := string(p.R)
	p.R, p.Len = p.R[:0], 0
	return s
}

// Get reads the next command parameter using the provided substitution func.
// True indicates there are runes remaining in the command parameters to
// process.
func (p *Params) Get(f func(string, bool) (bool, string, error)) (bool, string, error) {
	var ok bool
	var quote rune
	start, _ := findNonSpace(p.R, 0, p.Len)
	i := start
	if i >= p.Len {
		return false, "", nil
	}
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
			ok, z, err := f(string(p.R[start:i+1]), false)
			switch {
			case err != nil:
				return false, "", err
			case ok:
				p.R, p.Len = substitute(p.R, start, p.Len, i-start+1, z)
				i = start + len(z) - 1
			}
			quote = 0
		// start of single, double, or backtick string
		case c == '\'' || c == '"' || c == '`':
			quote = c
		case c == ':' && next != ':':
			if v := readVar(p.R, i, p.Len); v != nil {
				n := v.String()
				ok, z, err := f(n[1:], true)
				switch {
				case err != nil:
					return false, "", err
				case ok:
					p.R, p.Len = substitute(p.R, v.I, p.Len, len(n), z)
					i = v.I + len(z) - 1
				default:
					i += len(n) - 1
				}
			}
		case unicode.IsSpace(c):
			break loop
		}
	}
	if quote != 0 {
		return false, "", text.ErrUnterminatedQuotedString
	}
	v := string(p.R[start:i])
	p.R = p.R[i:]
	p.Len = len(p.R)
	return true, v, nil
}

// GetAll retrieves all remaining command parameters using the provided
// substitution func. Will return on the first encountered error.
func (p *Params) GetAll(f func(string, bool) (bool, string, error)) ([]string, error) {
	var s []string
	for {
		ok, v, err := p.Get(f)
		if err != nil {
			return s, err
		}
		if !ok {
			break
		}
		s = append(s, v)
	}
	return s, nil
}
