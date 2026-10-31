package readme

import (
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"text", "a\r\nb\r\n", "a\nb"},
		{"colors are dropped", "\x1b[32mgreen\x1b[39m\r\n", "green"},
		{"a redraw replaces the line", "db=> s\rdb=> se\x1b[K\rdb=> sel\x1b[K\r\n", "db=> sel"},
		{"erase to the end", "abcdef\r\x1b[3C\x1b[K\r\n", "abc"},
		{"up and over", "a\r\nb\x1b[1A\rc\r\n", "c\nb"},
		{"a title is dropped", "\x1b]0;title\aok\r\n", "ok"},
		{"private modes are dropped", "\x1b[?25lx\x1b[?25h\r\n", "x"},
		{"trailing blank lines", "a\r\n\r\n\r\n", "a"},
		{"trailing spaces", "a   \r\n", "a"},
		{"wide characters", "日本\r\x1b[2C\x1b[K\r\n", "日"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Render([]byte(test.in)); got != test.want {
				t.Errorf("expected %q, got: %q", test.want, got)
			}
		})
	}
}
