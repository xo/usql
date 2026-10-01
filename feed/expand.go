package feed

import (
	"strings"
	"unicode"
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

// maxPrefixWords is the most words that Prefix returns.
const maxPrefixWords = 6

// Prefix returns the first words of a statement, in upper case and joined by
// spaces, such as "CREATE TABLE". It skips comments, and it stops at the
// first character that is not a letter or white space, or after six words.
// usql uses it to decide whether a statement returns rows.
func (it Item) Prefix() string {
	var words []string
	var word strings.Builder
	flush := func() bool {
		if word.Len() != 0 {
			words = append(words, strings.ToUpper(word.String()))
			word.Reset()
		}
		return len(words) < maxPrefixWords
	}
	pos := 0
	for _, p := range append(it.Parts, Part{Start: len(it.Text), End: len(it.Text)}) {
		for _, r := range it.Text[pos:p.Start] {
			switch {
			case unicode.IsLetter(r):
				word.WriteRune(r)
			case unicode.IsSpace(r):
				if !flush() {
					return strings.Join(words, " ")
				}
			default:
				flush()
				return strings.Join(words, " ")
			}
		}
		if p.Kind != Comment && p.Start != p.End {
			// a string, a variable or a name ends the prefix
			flush()
			return strings.Join(words, " ")
		}
		// a comment separates words
		if !flush() {
			return strings.Join(words, " ")
		}
		pos = p.End
	}
	flush()
	return strings.Join(words, " ")
}
