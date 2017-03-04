package buf

import "testing"

func sl(len int, c byte) string {
	b := make([]byte, len)
	for i := 0; i < len; i++ {
		b[i] = c
	}
	return string(b)
}

func TestAppend(t *testing.T) {
	a512 := sl(512, 'a')
	//b1024 := sl(1024, 'b')

	tests := []struct {
		s   []string
		exp string
		l   int
		c   int
	}{
		{[]string{""}, "", 0, 0},
		{[]string{"", ""}, "\n", 1, MinCapIncrease},
		{[]string{"", "", ""}, "\n\n", 2, MinCapIncrease},
		{[]string{"", "", "", ""}, "\n\n\n", 3, MinCapIncrease},

		{[]string{"a", ""}, "a\n", 2, 2}, // 4
		{[]string{"a", "b", ""}, "a\nb\n", 4, MinCapIncrease},
		{[]string{"a", "b", "c", ""}, "a\nb\nc\n", 6, MinCapIncrease},

		{[]string{"", "a", ""}, "\na\n", 3, MinCapIncrease}, // 7
		{[]string{"", "a", "b", ""}, "\na\nb\n", 5, MinCapIncrease},
		{[]string{"", "a", "b", "c", ""}, "\na\nb\nc\n", 7, MinCapIncrease},

		{[]string{"", "foo"}, "\nfoo", 4, MinCapIncrease}, // 10
		{[]string{"", "foo", ""}, "\nfoo\n", 5, MinCapIncrease},
		{[]string{"foo", "", "bar"}, "foo\n\nbar", 8, MinCapIncrease},
		{[]string{"", "foo", "bar"}, "\nfoo\nbar", 8, MinCapIncrease},

		{[]string{a512}, a512, 512, 512}, // 14
		{[]string{a512, a512}, a512 + "\n" + a512, 1025, 5 * MinCapIncrease},
		{[]string{a512, a512, a512}, a512 + "\n" + a512 + "\n" + a512, 1538, 5 * MinCapIncrease},

		{[]string{a512, ""}, a512 + "\n", 513, 2 * MinCapIncrease}, // 17
		{[]string{a512, "", "foo"}, a512 + "\n\nfoo", 517, 2 * MinCapIncrease},
	}
	for i, test := range tests {
		b := new(Buf)
		for _, s := range test.s {
			b.AppendString(s, "\n")
		}

		if s := b.String(); s != test.exp {
			t.Errorf("test %d expected result of `%s`, got: `%s`", i, test.exp, s)
		}

		if b.Len != test.l {
			t.Errorf("test %d expected resulting len of %d, got: %d", i, test.l, b.Len)
		}

		if c := cap(b.Buf); c != test.c {
			t.Errorf("test %d expected resulting cap of %d, got: %d", i, test.c, c)
		}

		b.Reset()
		if b.Len != 0 {
			t.Errorf("test %d expected after reset len of 0, got: %d", i, b.Len)
		}
		b.AppendString("", "\n")
		if s := b.String(); s != "" {
			t.Errorf("test %d expected after reset appending an empty string would result in empty string, got: `%s`", i, s)
		}
	}
}

func TestVariedSeparator(t *testing.T) {
	b := new(Buf)

	b.AppendString("foo", "\n")
	b.AppendString("foo", "bar")

	if b.Len != 9 {
		t.Errorf("expected len of 9, got: %d", b.Len)
	}

	if s := b.String(); s != "foobarfoo" {
		t.Errorf("expected `%s`, got: `%s`", "foobarfoo", s)
	}

	if c := cap(b.Buf); c != MinCapIncrease {
		t.Errorf("expected cap of %d, got: %d", MinCapIncrease, c)
	}
}
