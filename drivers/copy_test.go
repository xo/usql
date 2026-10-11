package drivers

import (
	"slices"
	"testing"
)

func TestSplitName(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		want []string
	}{
		{"t", []string{"t"}},
		{"s.t", []string{"s", "t"}},
		{`"Mixed Case"`, []string{"Mixed Case"}},
		{`s."Odd.Table"`, []string{"s", "Odd.Table"}},
		{`"a""b".c`, []string{`a"b`, "c"}},
		{"Mixed.Case", []string{"Mixed", "Case"}},
	} {
		if got := splitName(tt.name); !slices.Equal(got, tt.want) {
			t.Errorf("splitName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
