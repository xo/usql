package metacmd

import (
	"slices"
	"testing"

	"github.com/xo/usql/feed"
	"github.com/xo/usql/metacmd/describe"
)

func TestNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		exp  []string
	}{
		{`q`, []string{`q`}},
		{`l[+]`, []string{`l`, `l+`}},
		{`d[S+]`, []string{`d`, `dS`, `d+`, `dS+`}},
		{`df[an][S]`, []string{`df`, `dfa`, `dfn`, `dfan`, `dfS`, `dfaS`, `dfnS`, `dfanS`}},
	}
	for _, test := range tests {
		if got := (desc{Name: test.name}).Names(); !slices.Equal(got, test.exp) {
			t.Errorf("%s: expected %q, got %q", test.name, test.exp, got)
		}
	}
}

func TestRelations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, exp string
	}{
		{`dvt`, `dtv`},
		{`dEvS+`, `dvES+`},
		{`dsv+`, `dvs+`},
		{`dt`, `dt`},
		{`dT`, `dT`},
		{`d`, `d`},
		{`dx`, `dx`},
		{`q`, `q`},
	}
	for _, test := range tests {
		if got := relations(test.name); got != test.exp {
			t.Errorf("%s: expected %q, got %q", test.name, test.exp, got)
		}
		if _, err := Decode(feed.Item{Kind: feed.Command, Name: test.name}); err != nil && test.name != `q` {
			t.Errorf("%s: expected no error, got: %v", test.name, err)
		}
	}
}

// TestEveryListIsACommand checks that each list the metadata package prints
// is reachable as a command, with its S and + forms where psql has them.
func TestEveryListIsACommand(t *testing.T) {
	t.Parallel()
	for _, name := range describe.Commands() {
		if cmds[name] == nil {
			t.Errorf("\\%s is printed by the metadata package, and no command reaches it", name)
		}
	}
}
