package stmt

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

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
		b := new(Stmt)
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

		b.Reset(nil)
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
	b := new(Stmt)

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

func TestNextResetState(t *testing.T) {
	tests := []struct {
		s     string
		stmts []string
		cmds  []string
		state string
		vars  []string
	}{
		{"", nil, []string{""}, "=", nil}, // 0
		{";", []string{";"}, []string{""}, "=", nil},
		{" ; ", []string{";"}, []string{"", ""}, "=", nil},
		{" \\v ", nil, []string{"v"}, "=", nil},
		{" \\v \\p", nil, []string{"v", "p"}, "=", nil},
		{" \\v   foo   \\p", nil, []string{"v foo", "p"}, "=", nil},
		{" \\v   foo   bar  \\p   zz", nil, []string{"v foo|bar", "p zz"}, "=", nil},
		{" \\very   foo   bar  \\print   zz", nil, []string{"very foo|bar", "print zz"}, "=", nil},

		{"select 1;", []string{"select 1;"}, []string{""}, "=", nil}, // 8
		{"select 1\\g", []string{"select 1"}, []string{"g"}, "=", nil},
		{"select 1 \\g", []string{"select 1 "}, []string{"g"}, "=", nil},
		{" select 1 \\g", []string{"select 1 "}, []string{"g"}, "=", nil},
		{" select 1   \\g  ", []string{"select 1   "}, []string{"g"}, "=", nil},

		{"select 1; select 1\\g", []string{"select 1;", "select 1"}, []string{"", "g"}, "=", nil}, // 13
		{"select 1\n\\g", []string{"select 1"}, []string{"", "g"}, "=", nil},
		{"select 1 \\g\n\n\n\n\\v", []string{"select 1 "}, []string{"g", "", "", "", "v"}, "=", nil},
		{"select 1 \\g\n\n\n\n\\v aoeu \\p zzz \n\n", []string{"select 1 "}, []string{"g", "", "", "", "v aoeu", "p zzz", "", ""}, "=", nil},
		{" select 1 \\g \\p \n select (15)\\g", []string{"select 1 ", "select (15)"}, []string{"g", "p", "g"}, "=", nil},
		{" select 1 (  \\g ) \n ;", []string{"select 1 (  \\g ) \n ;"}, []string{"", ""}, "=", nil},

		{ // 19
			" select 1\n;select 2\\g  select 3;  \\p   \\z  foo bar ",
			[]string{"select 1\n;", "select 2"},
			[]string{"", "", "g select|3;", "p", "z foo|bar"},
			"=", nil,
		},

		{ // 20
			" select 1\\g\n\n\tselect 2\\g\n select 3;  \\p   \\z  foo bar \\p\\p select * from;  \n\\p",
			[]string{"select 1", "select 2", "select 3;"},
			[]string{"g", "", "g", "", "p", "z foo|bar", "p\\p select|*|from;", "p"},
			"=", nil,
		},

		{"select '';", []string{"select '';"}, []string{""}, "=", nil}, // 21
		{"select 'a''b\nz';", []string{"select 'a''b\nz';"}, []string{"", ""}, "=", nil},
		{"select 'a' 'b\nz';", []string{"select 'a' 'b\nz';"}, []string{"", ""}, "=", nil},
		{"select \"\";", []string{"select \"\";"}, []string{""}, "=", nil},
		{"select \"\n\";", []string{"select \"\n\";"}, []string{"", ""}, "=", nil},
		{"select $$$$;", []string{"select $$$$;"}, []string{""}, "=", nil},
		{"select $$\naoeu(\n$$;", []string{"select $$\naoeu(\n$$;"}, []string{"", "", ""}, "=", nil},
		{"select $tag$$tag$;", []string{"select $tag$$tag$;"}, []string{""}, "=", nil},
		{"select $tag$\n\n$tag$;", []string{"select $tag$\n\n$tag$;"}, []string{"", "", ""}, "=", nil},
		{"select $tag$\n(\n$tag$;", []string{"select $tag$\n(\n$tag$;"}, []string{"", "", ""}, "=", nil},
		{"select $tag$\n\\v(\n$tag$;", []string{"select $tag$\n\\v(\n$tag$;"}, []string{"", "", ""}, "=", nil},
		{"select $tag$\n\\v(\n$tag$\\g", []string{"select $tag$\n\\v(\n$tag$"}, []string{"", "", "g"}, "=", nil},
		{"select $$\n\\v(\n$tag$$zz$$\\g$$\\g", []string{"select $$\n\\v(\n$tag$$zz$$\\g$$"}, []string{"", "", "g"}, "=", nil},

		{"select * --\n\\v", nil, []string{"", "v"}, "-", nil}, // 34
		{"select * /* \n\n\n--*/\n;", []string{"select * /* \n\n\n--*/\n;"}, []string{"", "", "", "", ""}, "=", nil},

		{"select * /* \n\n\n--*/\n", nil, []string{"", "", "", "", ""}, "-", nil}, // 36
		{"select * /* \n\n\n--\n", nil, []string{"", "", "", "", ""}, "*", nil},
		{"\\p \\p\nselect (", nil, []string{"p", "p", ""}, "(", nil},
		{"\\p \\p\nselect ()", nil, []string{"p", "p", ""}, "-", nil},
		{"\n             \t\t               \n", nil, []string{"", "", ""}, "=", nil},
		{"\n   aoeu      \t\t               \n", nil, []string{"", "", ""}, "-", nil},
		{"$$", nil, []string{""}, "$", nil},
		{"$$foo", nil, []string{""}, "$", nil},
		{"'", nil, []string{""}, "'", nil},
		{"(((()()", nil, []string{""}, "(", nil},
		{"\"", nil, []string{""}, "\"", nil},
		{"\"foo", nil, []string{""}, "\"", nil},

		{":a :b", nil, []string{""}, "-", []string{"a", "b"}}, // 48
		{`select :'a b' :"foo bar"`, nil, []string{""}, "-", []string{"a b", "foo bar"}},
		{`select :a:b;`, []string{"select :a:b;"}, []string{""}, "=", []string{"a", "b"}},

		{"select :'a\n:foo:bar", nil, []string{"", ""}, "'", nil}, // 51
		{"select :''\n:foo:bar\\g", []string{"select :''\n:foo:bar"}, []string{"", "g"}, "=", []string{"foo", "bar"}},
		{"select :''\n:foo :bar\\g", []string{"select :''\n:foo :bar"}, []string{"", "g"}, "=", []string{"foo", "bar"}},
		{"select :''\n :foo :bar \\g", []string{"select :''\n :foo :bar "}, []string{"", "g"}, "=", []string{"foo", "bar"}},

		{"select :'a\n:'foo':\"bar\"", nil, []string{"", ""}, "'", nil}, // 55
		{"select :''\n:'foo':\"bar\"\\g", []string{"select :''\n:'foo':\"bar\""}, []string{"", "g"}, "=", []string{"foo", "bar"}},
		{"select :''\n:'foo' :\"bar\"\\g", []string{"select :''\n:'foo' :\"bar\""}, []string{"", "g"}, "=", []string{"foo", "bar"}},
		{"select :''\n :'foo' :\"bar\" \\g", []string{"select :''\n :'foo' :\"bar\" "}, []string{"", "g"}, "=", []string{"foo", "bar"}},
	}

	for i, test := range tests {
		b := New(sp(test.s, "\n"), AllowDollar(true), AllowMultilineComments(true))

		var stmts, cmds []string
		var aparams [][]string
		var vars []*Var
		for {
			cmd, params, err := b.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Fatalf("test %d did not expect error, got: %v", i, err)
			}
			vars = append(vars, b.Vars...)

			if b.Ready() || cmd == "g" {
				stmts = append(stmts, b.String())
				b.Reset(nil)
			}
			cmds = append(cmds, cmd)
			aparams = append(aparams, params)
		}
		if len(stmts) != len(test.stmts) {
			t.Logf(">> %v // %v", test.stmts, stmts)
			t.Fatalf("test %d expected %d statements, got: %d", i, len(test.stmts), len(stmts))
		}

		if !reflect.DeepEqual(stmts, test.stmts) {
			t.Fatalf("test %d expected statements %s, got: %s", i, jj(test.stmts), jj(stmts))
		}

		if cz := cc(t, cmds, aparams); !reflect.DeepEqual(cz, test.cmds) {
			t.Fatalf("test %d expected commands %v, got: %v", i, jj(test.cmds), jj(cz))
		}

		if st := b.State(); st != test.state {
			t.Fatalf("test %d expected end parse state `%s`, got: `%s`", i, test.state, st)
		}

		if len(vars) != len(test.vars) {
			t.Fatalf("test %d expected %d vars, got: %d", i, len(test.vars), len(vars))
		}

		for _, n := range test.vars {
			if !hasVar(vars, n) {
				t.Fatalf("test %d missing variable `%s`", i, n)
			}
		}

		b.Reset(nil)
		if len(b.Buf) != 0 {
			t.Fatalf("test %d after reset b.Buf should have len %d, got: %d", i, 0, len(b.Buf))
		}
		if b.Len != 0 {
			t.Fatalf("test %d after reset should have len %d, got: %d", i, 0, b.Len)
		}
		if len(b.Vars) != 0 {
			t.Fatalf("test %d after reset should have len(vars) == 0, got: %d", len(b.Vars))
		}
		if b.Prefix != "" {
			t.Fatalf("test %d after reset should have empty prefix, got: %s", i, b.Prefix)
		}
		if b.q || b.qdbl || b.qdollar || b.qid != "" || b.mc || b.b != 0 {
			t.Fatalf("test %d after reset should have a cleared parse state", i)
		}
		if st := b.State(); st != "=" {
			t.Fatalf("test %d after reset should have state `=`, got: `%s`", i, st)
		}
		if b.ready {
			t.Fatalf("test %d after reset should not be ready", i)
		}
	}
}

func cc(t *testing.T, cmds []string, params [][]string) []string {
	var z []string
	for i, c := range cmds {
		p := strings.Join(params[i], "|")
		if p != "" {
			c += " " + p
		}
		z = append(z, c)
	}
	return z
}

func jj(s []string) string {
	return "[`" + strings.Join(s, "`,`") + "`]"
}

func sp(a, sep string) func() ([]rune, error) {
	s := strings.Split(a, sep)
	return func() ([]rune, error) {
		if len(s) > 0 {
			z := s[0]
			s = s[1:]
			return []rune(z), nil
		}

		return nil, io.EOF
	}
}

func hasVar(vars []*Var, n string) bool {
	for _, v := range vars {
		if v.N == n {
			return true
		}
	}

	return false
}
