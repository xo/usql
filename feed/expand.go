package feed

import (
	"strings"
)

// Lookup returns the value of a variable, and false when it is not set.
type Lookup func(name string) (string, bool)

// Expand returns the text of a statement with each variable replaced by its
// value, as psql does:
//
//   - :name is the value.
//   - :'name' is the value as a string literal, with each ' doubled.
//   - :"name" is the value as a quoted name, with each " doubled.
//   - :{?name} is TRUE when the variable is set and FALSE when it is not.
//
// A variable that is not set is left as it is written. A value is never
// read again as input, so a ';' or a quote in it ends nothing.
//
// The escapes \; and \: become ; and :, outside strings and comments.
func (it Item) Expand(lookup Lookup) string {
	var b strings.Builder
	pos := 0
	for _, p := range it.Parts {
		b.WriteString(unescape(it.Text[pos:p.Start]))
		pos = p.End
		if p.Kind != Variable {
			b.WriteString(it.Text[p.Start:p.End])
			continue
		}
		v, ok := lookup(p.Name)
		switch {
		case p.Quote == '?':
			if ok {
				b.WriteString("TRUE")
			} else {
				b.WriteString("FALSE")
			}
		case !ok:
			b.WriteString(it.Text[p.Start:p.End])
		case p.Quote == '\'' || p.Quote == '"':
			q := string(p.Quote)
			b.WriteString(q + strings.ReplaceAll(v, q, q+q) + q)
		default:
			b.WriteString(v)
		}
	}
	b.WriteString(unescape(it.Text[pos:]))
	return b.String()
}

// unescape turns \; and \: in SQL text into ; and :.
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	return strings.NewReplacer(`\;`, `;`, `\:`, `:`).Replace(s)
}
