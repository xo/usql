package feed

import "strings"

// Buffer is psql's query buffer: the statements that are not sent yet.
//
// A statement that ends with a ';' joins the buffer, and the buffer is sent.
// A statement with no ';' stays in the buffer until a ';' or a meta command
// such as \g sends it.
type Buffer struct {
	items []Item
}

// Add adds a statement to the buffer.
func (b *Buffer) Add(it Item) {
	b.items = append(b.items, it)
}

// Empty reports whether the buffer holds no statement.
func (b *Buffer) Empty() bool {
	return len(b.items) == 0
}

// Reset empties the buffer.
func (b *Buffer) Reset() {
	b.items = nil
}

// Set replaces the buffer with the statements of items, as after \e.
func (b *Buffer) Set(items []Item) {
	b.items = nil
	for _, it := range items {
		if it.Kind == Statement {
			b.items = append(b.items, it)
		}
	}
}

// Raw returns the buffer as it was written, with its variables as written.
func (b *Buffer) Raw() string {
	v := make([]string, len(b.items))
	for i, it := range b.items {
		v[i] = strings.TrimSpace(it.Text)
	}
	return strings.Join(v, "\n")
}

// Expand returns the buffer with its variables replaced, as it is sent.
func (b *Buffer) Expand(lookup Lookup) string {
	v := make([]string, len(b.items))
	for i, it := range b.items {
		v[i] = strings.TrimSpace(it.Expand(lookup))
	}
	return strings.Join(v, "\n")
}

// Prefix returns the prefix of the first statement of the buffer. See
// Item.Prefix.
func (b *Buffer) Prefix() string {
	if len(b.items) == 0 {
		return ""
	}
	return b.items[0].Prefix()
}
