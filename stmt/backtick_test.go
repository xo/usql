package stmt

import (
	"io"
	"strings"
	"testing"
)

// sp from stmt_test.go — local copy for this file's focused cases.
func backtickSrc(s string) func() ([]rune, error) {
	ch := make(chan []rune, 1)
	ch <- []rune(s)
	close(ch)
	return func() ([]rune, error) {
		r, ok := <-ch
		if !ok {
			return nil, io.EOF
		}
		return r, nil
	}
}

func TestBacktickIdentifierKeepsStringEscapes(t *testing.T) {
	// Shape from xo/usql#587: a MySQL identifier that contains a single quote
	// must not open a phantom string, or the following literal's \\ collapses
	// to \ before the server sees it (\\b → backspace).
	const q = "select 1 as `we'ird`, hex('a\\\\b');"

	nop := func(s string, _ bool) (string, bool, error) {
		return s, false, nil
	}

	parse := func(allowBacktick bool) string {
		opts := []Option{
			WithAllowMultilineComments(true),
			WithAllowHashComments(true),
			WithAllowDollar(false),
			WithAllowBacktick(allowBacktick),
		}
		b := New(backtickSrc(q), opts...)
		for {
			_, _, err := b.Next(nop)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Next: %v", err)
			}
			if b.Ready() {
				s := b.String()
				b.Reset(nil)
				return s
			}
		}
		return b.String()
	}

	got := parse(true)
	if !strings.Contains(got, `a\\b`) {
		t.Fatalf("with AllowBacktick: expected statement to keep a\\\\b, got %q", got)
	}
	if strings.Contains(got, "a\\b") && !strings.Contains(got, `a\\b`) {
		t.Fatalf("with AllowBacktick: backslash collapsed unexpectedly: %q", got)
	}

	// Without the option, the known buggy path collapses \\ → \ via the
	// outside-quote \\ substitution (see Stmt.Next).
	buggy := parse(false)
	if strings.Contains(buggy, `a\\b`) {
		t.Fatalf("without AllowBacktick: expected collapsed escapes reproducing #587, got %q", buggy)
	}
	if !strings.Contains(buggy, `a\b`) {
		t.Fatalf("without AllowBacktick: expected a\\b after collapse, got %q", buggy)
	}
}

func TestBacktickIdentifierPlainStillWorks(t *testing.T) {
	const q = "select 1 as `plain`, hex('a\\\\b');"
	nop := func(s string, _ bool) (string, bool, error) { return s, false, nil }
	b := New(backtickSrc(q),
		WithAllowMultilineComments(true),
		WithAllowHashComments(true),
		WithAllowBacktick(true),
	)
	var got string
	for {
		_, _, err := b.Next(nop)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if b.Ready() {
			got = b.String()
			break
		}
	}
	if !strings.Contains(got, `a\\b`) {
		t.Fatalf("plain backtick identifier should keep escapes, got %q", got)
	}
}
