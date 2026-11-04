package metacmd

import (
	"slices"
	"testing"

	"github.com/xo/usql/env"
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

// TestCompletions checks the names that the completer offers for \ and Tab.
func TestCompletions(t *testing.T) {
	t.Parallel()
	got := Completions()
	names := make(map[string]bool)
	for _, name := range got {
		if names[name] {
			t.Errorf("%s is listed twice", name)
		}
		names[name] = true
	}
	for _, name := range []string{
		`\sf`, `\sf+`, `\sv`, `\ss`, `\dt`, `\dtS+`, `\dfa`, `\dfS+`, `\dPi`, `\dP+`,
		`\gexec`, `\crosstabview`, `\conninfo`, `\uset`, `\quit`, `\lo_list`,
	} {
		if !names[name] {
			t.Errorf("%s is missing", name)
		}
	}
	// a kind of object is offered alone, not in every combination
	for _, name := range []string{`\dfan`, `\dfanptw`, `\dtvS`} {
		if names[name] {
			t.Errorf("%s is offered", name)
		}
	}
	if !slices.IsSorted(got) {
		t.Error("the names are not sorted")
	}
}

// TestExpandShell checks that a variable in the text of a command between
// backticks is replaced as psql replaces it: :name by its value, :'name' by
// its value quoted for the shell, and a variable that is not set not at all.
func TestExpandShell(t *testing.T) {
	_ = env.Vars().Set("shellx", "it's")
	tests := []struct {
		s, exp string
	}{
		{"echo :shellx", "echo it's"},
		{"echo :'shellx'", `echo 'it'\''s'`},
		{"echo :nope :'nope'", "echo :nope :'nope'"},
		{"echo a:b", "echo a:b"},
	}
	for _, test := range tests {
		if got := expandShell(test.s); got != test.exp {
			t.Errorf("expandShell(%q): expected %q, got %q", test.s, test.exp, got)
		}
	}
}
