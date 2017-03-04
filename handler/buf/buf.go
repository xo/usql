package buf

const (
	// MinCapIncrease is the minimum amount by which to grow a Buf.
	MinCapIncrease = 512
)

// Buf is a reusable rune buffer.
type Buf struct {
	Buf []rune
	Len int
}

// String satisfies fmt.Stringer.
func (b *Buf) String() string {
	return string(b.Buf)
}

// Reset resets b.
func (b *Buf) Reset() {
	b.Buf, b.Len = nil, 0
}

// Append appends r to b.Buf separated by sep when b.Buf is not already empty.
//
// Append dynamically grows b.Buf as necessary to accommodate r and the
// separator. Specifically, when b.Buf is not empty, b.Buf will grow by
// increments of MinCapIncrease.
//
// After a call to Append, b.Len will be len(b.Buf)+len(r)+len(sep) Call Reset to
// reset the Buf.
func (b *Buf) Append(r, sep []rune) {
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

// AppendString is a utility func wrapping Append.
func (b *Buf) AppendString(s, sep string) {
	b.Append([]rune(s), []rune(sep))
}
