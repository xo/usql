package stmt

import (
	"io"
	"os/user"
	"reflect"
	"strings"
	"testing"

	"github.com/xo/usql/env"
)

func sl(n int, r rune) string {
	z := make([]rune, n)
	for i := 0; i < n; i++ {
		z[i] = r
	}
	return string(z)
}

func TestAppend(t *testing.T) {
	a512 := sl(512, 'a')
	// b1024 := sl(1024, 'b')
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
	u, err := user.Current()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	unquote := env.Unquote(u, false, env.Vars{})
	tests := []struct {
		s     string
		stmts []string
		cmds  []string
		state string
		vars  []string
	}{
		{``, nil, []string{`|`}, `=`, nil}, // 0
		{`;`, []string{`;`}, []string{`|`}, `=`, nil},
		{` ; `, []string{`;`}, []string{`|`, `|`}, `=`, nil},
		{` \v `, nil, []string{`\v| `}, `=`, nil},
		{` \v \p`, nil, []string{`\v| `, `\p|`}, `=`, nil},
		{` \v   foo   \p`, nil, []string{`\v|   foo   `, `\p|`}, `=`, nil}, // 5
		{` \v   foo   bar  \p   zz`, nil, []string{`\v|   foo   bar  `, `\p|   zz`}, `=`, nil},
		{` \very   foo   bar  \print   zz`, nil, []string{`\very|   foo   bar  `, `\print|   zz`}, `=`, nil},
		{`select 1;`, []string{`select 1;`}, []string{`|`}, `=`, nil},
		{`select 1\g`, []string{`select 1`}, []string{`\g|`}, `=`, nil},
		{`select 1 \g`, []string{`select 1 `}, []string{`\g|`}, `=`, nil}, // 10
		{` select 1 \g`, []string{`select 1 `}, []string{`\g|`}, `=`, nil},
		{` select 1   \g  `, []string{`select 1   `}, []string{`\g|  `}, `=`, nil},
		{`select 1; select 1\g`, []string{`select 1;`, `select 1`}, []string{`|`, `\g|`}, `=`, nil},
		{"select 1\n\\g", []string{`select 1`}, []string{`|`, `\g|`}, `=`, nil},
		{"select 1 \\g\n\n\n\n\\v", []string{`select 1 `}, []string{`\g|`, `|`, `|`, `|`, `\v|`}, `=`, nil}, // 15
		{"select 1 \\g\n\n\n\n\\v aoeu \\p zzz \n\n", []string{`select 1 `}, []string{`\g|`, `|`, `|`, `|`, `\v| aoeu `, `\p| zzz `, `|`, `|`}, `=`, nil},
		{" select 1 \\g \\p \n select (15)\\g", []string{`select 1 `, `select (15)`}, []string{`\g| `, `\p| `, `\g|`}, `=`, nil},
		{" select 1 (  \\g ) \n ;", []string{"select 1 (  \\g ) \n ;"}, []string{`|`, `|`}, `=`, nil},
		{ // 19
			" select 1\n;select 2\\g  select 3;  \\p   \\z  foo bar ",
			[]string{"select 1\n;", "select 2"},
			[]string{`|`, `|`, `\g|  select 3;  `, `\p|   `, `\z|  foo bar `},
			"=", nil,
		},
		{ // 20
			" select 1\\g\n\n\tselect 2\\g\n select 3;  \\p   \\z  foo bar \\p\\p select * from;  \n\\p",
			[]string{`select 1`, `select 2`, `select 3;`},
			[]string{`\g|`, `|`, `\g|`, `|`, `\p|   `, `\z|  foo bar `, `\p|`, `\p| select * from;  `, `\p|`},
			"=", nil,
		},
		{"select '';", []string{"select '';"}, []string{"|"}, "=", nil}, // 21
		{"select 'a''b\nz';", []string{"select 'a''b\nz';"}, []string{"|", "|"}, "=", nil},
		{"select 'a' 'b\nz';", []string{"select 'a' 'b\nz';"}, []string{"|", "|"}, "=", nil},
		{"select \"\";", []string{"select \"\";"}, []string{"|"}, "=", nil},
		{"select \"\n\";", []string{"select \"\n\";"}, []string{"|", "|"}, "=", nil}, // 25
		{"select $$$$;", []string{"select $$$$;"}, []string{"|"}, "=", nil},
		{"select $$\naoeu(\n$$;", []string{"select $$\naoeu(\n$$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$$tag$;", []string{"select $tag$$tag$;"}, []string{"|"}, "=", nil},
		{"select $tag$\n\n$tag$;", []string{"select $tag$\n\n$tag$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$\n(\n$tag$;", []string{"select $tag$\n(\n$tag$;"}, []string{"|", "|", "|"}, "=", nil}, // 30
		{"select $tag$\n\\v(\n$tag$;", []string{"select $tag$\n\\v(\n$tag$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$\n\\v(\n$tag$\\g", []string{"select $tag$\n\\v(\n$tag$"}, []string{"|", "|", `\g|`}, "=", nil},
		{"select $$\n\\v(\n$tag$$zz$$\\g$$\\g", []string{"select $$\n\\v(\n$tag$$zz$$\\g$$"}, []string{"|", "|", `\g|`}, "=", nil},
		{"select * --\n\\v", nil, []string{"|", `\v|`}, "-", nil}, // 34
		{"select--", nil, []string{"|"}, "-", nil},
		{"select --", nil, []string{"|"}, "-", nil},
		{"select /**/", nil, []string{"|"}, "-", nil},
		{"select/* */", nil, []string{"|"}, "-", nil},
		{"select/*", nil, []string{"|"}, "*", nil},
		{"select /*", nil, []string{"|"}, "*", nil},
		{"select * /**/", nil, []string{"|"}, "-", nil},
		{"select * /* \n\n\n--*/\n;", []string{"select * /* \n\n\n--*/\n;"}, []string{"|", "|", "|", "|", "|"}, "=", nil},
		{"select * /* \n\n\n--*/\n", nil, []string{"|", "|", "|", "|", "|"}, "-", nil}, // 43
		{"select * /* \n\n\n--\n", nil, []string{"|", "|", "|", "|", "|"}, "*", nil},
		{"\\p \\p\nselect (", nil, []string{`\p| `, `\p|`, "|"}, "(", nil}, // 45
		{"\\p \\p\nselect ()", nil, []string{`\p| `, `\p|`, "|"}, "-", nil},
		{"\n             \t\t               \n", nil, []string{"|", "|", "|"}, "=", nil},
		{"\n   aoeu      \t\t               \n", nil, []string{"|", "|", "|"}, "-", nil},
		{"$$", nil, []string{"|"}, "$", nil},
		{"$$foo", nil, []string{"|"}, "$", nil}, // 50
		{"'", nil, []string{"|"}, "'", nil},
		{"(((()()", nil, []string{"|"}, "(", nil},
		{"\"", nil, []string{"|"}, "\"", nil},
		{"\"foo", nil, []string{"|"}, "\"", nil},
		{":a :b", nil, []string{"|"}, "-", []string{"a", "b"}}, // 55
		{`select :'a b' :"foo bar"`, nil, []string{"|"}, "-", []string{"a b", "foo bar"}},
		{`select :a:b;`, []string{"select :a:b;"}, []string{"|"}, "=", []string{"a", "b"}},
		{"select :'a\n:foo:bar", nil, []string{"|", "|"}, "'", nil}, // 58
		{"select :''\n:foo:bar\\g", []string{"select :''\n:foo:bar"}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n:foo :bar\\g", []string{"select :''\n:foo :bar"}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}}, // 60
		{"select :''\n :foo :bar \\g", []string{"select :''\n :foo :bar "}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :'a\n:'foo':\"bar\"", nil, []string{"|", "|"}, "'", nil}, // 62
		{"select :''\n:'foo':\"bar\"\\g", []string{"select :''\n:'foo':\"bar\""}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n:'foo' :\"bar\"\\g", []string{"select :''\n:'foo' :\"bar\""}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n :'foo' :\"bar\" \\g", []string{"select :''\n :'foo' :\"bar\" "}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{`select 1\echo 'pg://':foo'/':bar`, nil, []string{`\echo| 'pg://':foo'/':bar`}, "-", nil}, // 66
		{`select :'foo'\echo 'pg://':bar'/' `, nil, []string{`\echo| 'pg://':bar'/' `}, "-", []string{"foo"}},
		{`select 1\g '\g`, []string{`select 1`}, []string{`\g| '\g`}, "=", nil},
		{`select 1\g "\g`, []string{`select 1`}, []string{`\g| "\g`}, "=", nil},
		{"select 1\\g `\\g", []string{`select 1`}, []string{"\\g| `\\g"}, "=", nil}, // 70
		{`select 1\g '\g `, []string{`select 1`}, []string{`\g| '\g `}, "=", nil},
		{`select 1\g "\g `, []string{`select 1`}, []string{`\g| "\g `}, "=", nil},
		{"select 1\\g `\\g ", []string{`select 1`}, []string{"\\g| `\\g "}, "=", nil},
	}
	for i, test := range tests {
		b := New(sp(test.s, "\n"), WithAllowDollar(true), WithAllowMultilineComments(true), WithAllowCComments(true))
		var stmts, cmds, aparams []string
		var vars []*Var
	loop:
		for {
			cmd, params, err := b.Next(unquote)
			switch {
			case err == io.EOF:
				break loop
			case err != nil:
				t.Fatalf("test %d did not expect error, got: %v", i, err)
			}
			vars = append(vars, b.Vars...)
			if b.Ready() || cmd == `\g` {
				stmts = append(stmts, b.String())
				b.Reset(nil)
			}
			cmds = append(cmds, cmd)
			aparams = append(aparams, params)
		}
		if len(stmts) != len(test.stmts) {
			t.Logf(">> %#v // %#v", test.stmts, stmts)
			t.Fatalf("test %d expected %d statements, got: %d", i, len(test.stmts), len(stmts))
		}
		if !reflect.DeepEqual(stmts, test.stmts) {
			t.Logf(">> %#v // %#v", test.stmts, stmts)
			t.Fatalf("test %d expected statements %s, got: %s", i, jj(test.stmts), jj(stmts))
		}
		if cz := cc(cmds, aparams); !reflect.DeepEqual(cz, test.cmds) {
			t.Logf(">> cmds: %#v, aparams: %#v, cz: %#v, test.cmds: %#v", cmds, aparams, cz, test.cmds)
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
			t.Fatalf("test %d after reset should have len(vars) == 0, got: %d", i, len(b.Vars))
		}
		if b.Prefix != "" {
			t.Fatalf("test %d after reset should have empty prefix, got: %s", i, b.Prefix)
		}
		if b.quote != 0 || b.quoteDollarTag != "" || b.multilineComment || b.balanceCount != 0 {
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

func TestEmptyVariablesRawString(t *testing.T) {
	stmt := new(Stmt)
	stmt.AppendString("select ", "\n")
	stmt.Prefix = "SELECT"
	v := &Var{
		I:    7,
		End:  9,
		Name: "a",
		Len:  0,
	}
	stmt.Vars = append(stmt.Vars, v)

	if exp, got := "select ", stmt.RawString(); exp != got {
		t.Fatalf("Defined=false, expected: %s, got: %s", exp, got)
	}

	v.Defined = true
	if exp, got := "select :a", stmt.RawString(); exp != got {
		t.Fatalf("Defined=true, expected: %s, got: %s", exp, got)
	}
}

// cc combines commands with params.
func cc(cmds []string, params []string) []string {
	if len(cmds) == 0 {
		return []string{"|"}
	}
	z := make([]string, len(cmds))
	if len(cmds) != len(params) {
		panic("length of params should be same as cmds")
	}
	for i := 0; i < len(cmds); i++ {
		z[i] = cmds[i] + "|" + params[i]
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
		if v.Name == n {
			return true
		}
	}
	return false
}
