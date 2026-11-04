package highlight

import (
	"context"
	"slices"
	"testing"
)

// TestKeywords checks that each dialect gets the keywords of its own grammar
// after CREATE, ALTER and DROP, so that SQLite is not offered DROP DATABASE.
func TestKeywords(t *testing.T) {
	t.Parallel()
	tests := []struct {
		dialect, src string
		has, not     []string
	}{
		{"sqlite3", "create ", []string{"INDEX", "TABLE", "TEMP", "TRIGGER", "VIEW"}, []string{"DATABASE", "SCHEMA", "SEQUENCE"}},
		{"sqlite3", "drop ", []string{"INDEX", "TABLE", "TRIGGER", "VIEW"}, []string{"DATABASE", "USER", "FUNCTION"}},
		{"sqlite3", "alter ", []string{"TABLE"}, []string{"INDEX", "VIEW"}},
		{"postgres", "create ", []string{"DATABASE", "SCHEMA", "SEQUENCE", "TABLE", "MATERIALIZED"}, nil},
		{"postgres", "create unique ", []string{"INDEX"}, []string{"TABLE"}},
	}
	for _, test := range tests {
		words, err := Keywords(context.Background(), test.dialect, []byte(test.src))
		if err != nil {
			t.Fatalf("%s %q: %v", test.dialect, test.src, err)
		}
		for _, w := range test.has {
			if !slices.Contains(words, w) {
				t.Errorf("%s %q: expected %s in %q", test.dialect, test.src, w, words)
			}
		}
		for _, w := range test.not {
			if slices.Contains(words, w) {
				t.Errorf("%s %q: did not expect %s in %q", test.dialect, test.src, w, words)
			}
		}
	}
}
