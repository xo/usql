package main

import "testing"

func TestStartsWith(t *testing.T) {
	tests := []struct {
		s      string
		prefix string
		res    string
		exp    bool
	}{
		{``, ``, ``, true},
		{` `, ``, ` `, true},
		{`\c`, ``, `\c`, true},
		{`\c `, ``, `\c `, true},

		{``, ` `, ``, false},
		{``, `\c `, ``, false},
		{`\c `, `\c `, ``, false},

		{`\c `, `\c`, ``, true},
		{`\c `, `\c`, ``, true},

		{`\c blah`, `\c`, `blah`, true},
		{`\c blah `, `\c`, `blah`, true},
		{` \c blah`, `\c`, `blah`, true},
		{` \c blah `, `\c`, `blah`, true},
		{` \c  blah `, `\c`, `blah`, true},
		{" \\c\tblah ", `\c`, `blah`, true},
		{" \\c\t blah ", `\c`, `blah`, true},
		{" \\c \t blah ", `\c`, `blah`, true},

		{`\ca blah`, `\c`, ``, false},
		{`\ca blah `, `\c`, ``, false},
		{` \ca blah`, `\c`, ``, false},
		{` \ca blah `, `\c`, ``, false},
		{` \ca  blah `, `\c`, ``, false},
	}

	for i, test := range tests {
		res, ok := startsWith(test.s, test.prefix)
		if ok != test.exp {
			t.Errorf("test %d startsWith(`%s`, `%s`) expected %t, got: %t", i, test.s, test.prefix, test.exp, ok)
		}
		if res != test.res {
			t.Errorf("test %d startsWith(`%s`, `%s`) expected `%s`, got: `%s`", i, test.s, test.prefix, test.res, res)
		}
	}
}
