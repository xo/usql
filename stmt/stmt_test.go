package stmt

import (
	"io"
	"os/user"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/xo/usql/env"
)

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
		{[]string{"", ""}, "\n", 1, minCapIncrease},
		{[]string{"", "", ""}, "\n\n", 2, minCapIncrease},
		{[]string{"", "", "", ""}, "\n\n\n", 3, minCapIncrease},
		{[]string{"a", ""}, "a\n", 2, 2},
		{[]string{"a", "b", ""}, "a\nb\n", 4, minCapIncrease},
		{[]string{"a", "b", "c", ""}, "a\nb\nc\n", 6, minCapIncrease},
		{[]string{"", "a", ""}, "\na\n", 3, minCapIncrease},
		{[]string{"", "a", "b", ""}, "\na\nb\n", 5, minCapIncrease},
		{[]string{"", "a", "b", "c", ""}, "\na\nb\nc\n", 7, minCapIncrease},
		{[]string{"", "foo"}, "\nfoo", 4, minCapIncrease},
		{[]string{"", "foo", ""}, "\nfoo\n", 5, minCapIncrease},
		{[]string{"foo", "", "bar"}, "foo\n\nbar", 8, minCapIncrease},
		{[]string{"", "foo", "bar"}, "\nfoo\nbar", 8, minCapIncrease},
		{[]string{a512}, a512, 512, 512},
		{[]string{a512, a512}, a512 + "\n" + a512, 1025, 5 * minCapIncrease},
		{[]string{a512, a512, a512}, a512 + "\n" + a512 + "\n" + a512, 1538, 5 * minCapIncrease},
		{[]string{a512, ""}, a512 + "\n", 513, 2 * minCapIncrease},
		{[]string{a512, "", "foo"}, a512 + "\n\nfoo", 517, 2 * minCapIncrease},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			b := new(Stmt)
			for _, s := range test.s {
				b.AppendString(s, "\n")
			}
			if s := b.String(); s != test.exp {
				t.Errorf("expected result of %q, got: %q", test.exp, s)
			}
			if b.Len != test.l {
				t.Errorf("expected resulting len of %d, got: %d", test.l, b.Len)
			}
			if c := cap(b.Buf); c != test.c {
				t.Errorf("expected resulting cap of %d, got: %d", test.c, c)
			}
			b.Reset(nil)
			if b.Len != 0 {
				t.Errorf("expected after reset len of 0, got: %d", b.Len)
			}
			b.AppendString("", "\n")
			if s := b.String(); s != "" {
				t.Errorf("expected after reset appending an empty string would result in empty string, got: %q", s)
			}
		})
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
		t.Errorf("expected %q, got: %q", "foobarfoo", s)
	}
	if c := cap(b.Buf); c != minCapIncrease {
		t.Errorf("expected cap of %d, got: %d", minCapIncrease, c)
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
		{``, nil, []string{`|`}, `=`, nil},
		{`;`, []string{`;`}, []string{`|`}, `=`, nil},
		{` ; `, []string{`;`}, []string{`|`, `|`}, `=`, nil},
		{` \v `, nil, []string{`\v| `}, `=`, nil},
		{` \v \p`, nil, []string{`\v| `, `\p|`}, `=`, nil},
		{` \v   foo   \p`, nil, []string{`\v|   foo   `, `\p|`}, `=`, nil},
		{` \v   foo   bar  \p   zz`, nil, []string{`\v|   foo   bar  `, `\p|   zz`}, `=`, nil},
		{` \very   foo   bar  \print   zz`, nil, []string{`\very|   foo   bar  `, `\print|   zz`}, `=`, nil},
		{`select 1;`, []string{`select 1;`}, []string{`|`}, `=`, nil},
		{`select 1\g`, []string{`select 1`}, []string{`\g|`}, `=`, nil},
		{`select 1 \g`, []string{`select 1 `}, []string{`\g|`}, `=`, nil},
		{` select 1 \g`, []string{`select 1 `}, []string{`\g|`}, `=`, nil},
		{` select 1   \g  `, []string{`select 1   `}, []string{`\g|  `}, `=`, nil},
		{`select 1; select 1\g`, []string{`select 1;`, `select 1`}, []string{`|`, `\g|`}, `=`, nil},
		{"select 1\n\\g", []string{`select 1`}, []string{`|`, `\g|`}, `=`, nil},
		{"select 1 \\g\n\n\n\n\\v", []string{`select 1 `}, []string{`\g|`, `|`, `|`, `|`, `\v|`}, `=`, nil},
		{"select 1 \\g\n\n\n\n\\v foob \\p zzz \n\n", []string{`select 1 `}, []string{`\g|`, `|`, `|`, `|`, `\v| foob `, `\p| zzz `, `|`, `|`}, `=`, nil},
		{" select 1 \\g \\p \n select (15)\\g", []string{`select 1 `, `select (15)`}, []string{`\g| `, `\p| `, `\g|`}, `=`, nil},
		{" select 1 (  \\g ) \n ;", []string{"select 1 (  \\g ) \n ;"}, []string{`|`, `|`}, `=`, nil},
		{
			" select 1\n;select 2\\g  select 3;  \\p   \\z  foo bar ",
			[]string{"select 1\n;", "select 2"},
			[]string{`|`, `|`, `\g|  select 3;  `, `\p|   `, `\z|  foo bar `},
			"=", nil,
		},
		{
			" select 1\\g\n\n\tselect 2\\g\n select 3;  \\p   \\z  foo bar \\p\\p select * from;  \n\\p",
			[]string{`select 1`, `select 2`, `select 3;`},
			[]string{`\g|`, `|`, `\g|`, `|`, `\p|   `, `\z|  foo bar `, `\p|`, `\p| select * from;  `, `\p|`},
			"=", nil,
		},
		{"select '';", []string{"select '';"}, []string{"|"}, "=", nil},
		{"select 'a''b\nz';", []string{"select 'a''b\nz';"}, []string{"|", "|"}, "=", nil},
		{"select 'a' 'b\nz';", []string{"select 'a' 'b\nz';"}, []string{"|", "|"}, "=", nil},
		{"select \"\";", []string{"select \"\";"}, []string{"|"}, "=", nil},
		{"select \"\n\";", []string{"select \"\n\";"}, []string{"|", "|"}, "=", nil},
		{"select $$$$;", []string{"select $$$$;"}, []string{"|"}, "=", nil},
		{"select $$\nfoob(\n$$;", []string{"select $$\nfoob(\n$$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$$tag$;", []string{"select $tag$$tag$;"}, []string{"|"}, "=", nil},
		{"select $tag$\n\n$tag$;", []string{"select $tag$\n\n$tag$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$\n(\n$tag$;", []string{"select $tag$\n(\n$tag$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$\n\\v(\n$tag$;", []string{"select $tag$\n\\v(\n$tag$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$\n\\v(\n$tag$\\g", []string{"select $tag$\n\\v(\n$tag$"}, []string{"|", "|", `\g|`}, "=", nil},
		{"select $$\n\\v(\n$tag$$zz$$\\g$$\\g", []string{"select $$\n\\v(\n$tag$$zz$$\\g$$"}, []string{"|", "|", `\g|`}, "=", nil},
		{"select * --\n\\v", nil, []string{"|", `\v|`}, "-", nil},
		{"select--", nil, []string{"|"}, "-", nil},
		{"select --", nil, []string{"|"}, "-", nil},
		{"select /**/", nil, []string{"|"}, "-", nil},
		{"select/* */", nil, []string{"|"}, "-", nil},
		{"select/*", nil, []string{"|"}, "*", nil},
		{"select /*", nil, []string{"|"}, "*", nil},
		{"select * /**/", nil, []string{"|"}, "-", nil},
		{"select * /* \n\n\n--*/\n;", []string{"select * /* \n\n\n--*/\n;"}, []string{"|", "|", "|", "|", "|"}, "=", nil},
		{"select * /* \n\n\n--*/\n", nil, []string{"|", "|", "|", "|", "|"}, "-", nil},
		{"select * /* \n\n\n--\n", nil, []string{"|", "|", "|", "|", "|"}, "*", nil},
		{"\\p \\p\nselect (", nil, []string{`\p| `, `\p|`, "|"}, "(", nil},
		{"\\p \\p\nselect ()", nil, []string{`\p| `, `\p|`, "|"}, "-", nil},
		{"\n             \t\t               \n", nil, []string{"|", "|", "|"}, "=", nil},
		{"\n   foob      \t\t               \n", nil, []string{"|", "|", "|"}, "-", nil},
		{"$$", nil, []string{"|"}, "$", nil},
		{"$$foo", nil, []string{"|"}, "$", nil},
		{"'", nil, []string{"|"}, "'", nil},
		{"(((()()", nil, []string{"|"}, "(", nil},
		{"\"", nil, []string{"|"}, "\"", nil},
		{"\"foo", nil, []string{"|"}, "\"", nil},
		{":a :b", nil, []string{"|"}, "-", []string{"a", "b"}},
		{":{?a_b} :{?_foo_bar_}", nil, []string{"|"}, "-", []string{"a_b", "_foo_bar_"}},
		{`select :'a_b' :"foo_bar_"`, nil, []string{"|"}, "-", []string{"a_b", "foo_bar_"}},
		{`select :a:b;`, []string{"select :a:b;"}, []string{"|"}, "=", []string{"a", "b"}},
		{"select :'a\n:foo:bar", nil, []string{"|", "|"}, "'", nil},
		{"select :''\n:foo:bar\\g", []string{"select :''\n:foo:bar"}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n:foo :bar\\g", []string{"select :''\n:foo :bar"}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n :foo :bar \\g", []string{"select :''\n :foo :bar "}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :'a\n:'foo':\"bar\"", nil, []string{"|", "|"}, "'", nil},
		{"select :''\n:'foo':\"bar\"\\g", []string{"select :''\n:'foo':\"bar\""}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n:'foo' :\"bar\"\\g", []string{"select :''\n:'foo' :\"bar\""}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n :'foo' :\"bar\" \\g", []string{"select :''\n :'foo' :\"bar\" "}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{`select 1\echo 'pg://':foo'/':bar`, nil, []string{`\echo| 'pg://':foo'/':bar`}, "-", nil},
		{`select :'foo'\echo 'pg://':bar'/' `, nil, []string{`\echo| 'pg://':bar'/' `}, "-", []string{"foo"}},
		{`select 1\g '\g`, []string{`select 1`}, []string{`\g| '\g`}, "=", nil},
		{`select 1\g "\g`, []string{`select 1`}, []string{`\g| "\g`}, "=", nil},
		{"select 1\\g `\\g", []string{`select 1`}, []string{"\\g| `\\g"}, "=", nil},
		{`select 1\g '\g `, []string{`select 1`}, []string{`\g| '\g `}, "=", nil},
		{`select 1\g "\g `, []string{`select 1`}, []string{`\g| "\g `}, "=", nil},
		{"select 1\\g `\\g ", []string{`select 1`}, []string{"\\g| `\\g "}, "=", nil},
		{"select $$\\g$$\\g", []string{`select $$\g$$`}, []string{`\g|`}, "=", nil},
		{"select $1\\bind a b c\\g", []string{`select $1`}, []string{`\bind| a b c`, `\g|`}, "=", nil},
		{"select $1 \\bind a b c \\g", []string{`select $1 `}, []string{`\bind| a b c `, `\g|`}, "=", nil},
		{"select $2, $a$ foo $a$, $1 \\bind a b \\g", []string{`select $2, $a$ foo $a$, $1 `}, []string{`\bind| a b `, `\g|`}, "=", nil},
		{"select \\;\\\\\\:\\; \n;", []string{"select ;\\:; \n;"}, []string{`|`, `|`}, `=`, []string{`;`, `\`, `:`, `;`}},
		{"select \\;\\;\\;\\; \n;", []string{"select ;;;; \n;"}, []string{`|`, `|`}, `=`, []string{`;`, `;`, `;`, `;`}},
		{`select \;\;\;\;;`, []string{`select ;;;;;`}, []string{`|`}, `=`, []string{`;`, `;`, `;`, `;`}},
		{`select \\\;\\\;\\\;\\\;;`, []string{`select \;\;\;\;;`}, []string{`|`}, `=`, []string{`\`, `;`, `\`, `;`, `\`, `;`, `\`, `;`}},
		{`select \:foo;`, []string{`select :foo;`}, []string{`|`}, `=`, []string{":"}},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("statement: %q", test.s)
			b := New(
				sp(test.s, "\n"),
				WithAllowDollar(true),
				WithAllowMultilineComments(true),
				WithAllowCComments(true),
			)
			var stmts, cmds, aparams []string
			var vars []*Var
		loop:
			for {
				cmd, params, err := b.Next(unquote)
				t.Logf("next buf:%q", string(b.Buf))
				t.Logf("next cmd:%q params:%q", cmd, params)
				switch {
				case err == io.EOF:
					break loop
				case err != nil:
					t.Fatalf("expected no error, got: %v", err)
				}
				if b.Ready() || cmd == `\g` {
					stmts = append(stmts, b.String())
					vars = append(vars, b.Vars...)
					b.Reset(nil)
				}
				cmds = append(cmds, cmd)
				aparams = append(aparams, params)
			}
			vars = append(vars, b.Vars...)
			if len(stmts) != len(test.stmts) || !reflect.DeepEqual(stmts, test.stmts) {
				t.Errorf("expected %d statements, got: %d", len(test.stmts), len(stmts))
				t.Logf("expected:")
				for _, s := range test.stmts {
					t.Logf(" %q", s)
				}
				t.Logf("got:")
				for _, s := range stmts {
					t.Logf(" %q", s)
				}
			}
			if cz := cc(cmds, aparams); !reflect.DeepEqual(cz, test.cmds) {
				t.Logf(">> cmds: %#v, aparams: %#v, cz: %#v, test.cmds: %#v", cmds, aparams, cz, test.cmds)
				t.Errorf("commands do not match")
				t.Logf("expected:")
				for _, s := range test.cmds {
					t.Logf(" %q", s)
				}
				t.Logf("got:")
				for _, s := range cz {
					t.Logf(" %q", s)
				}
			}
			if st := b.State(); st != test.state {
				t.Errorf("expected end parse state %q, got: %q", test.state, st)
			}
			if len(vars) != len(test.vars) {
				t.Errorf("expected %d vars, got: %d", len(test.vars), len(vars))
				t.Logf("expected:")
				for _, v := range test.vars {
					t.Logf(" %#v", v)
				}
				t.Logf("got:")
				for _, v := range vars {
					t.Logf(" %#v", v)
				}
			}
			for i, v := range test.vars {
				if len(vars) < i {
					t.Logf("expected var %d: %#v", i, v)
					continue
				}
				if vars[i].Name != v {
					t.Errorf("expected var %d: %q, got: %q", i, v, vars[i].Name)
				}
			}
			b.Reset(nil)
			if len(b.Buf) != 0 {
				t.Errorf("after reset b.Buf should have len %d, got: %d", 0, len(b.Buf))
			}
			if b.Len != 0 {
				t.Errorf("after reset should have len %d, got: %d", 0, b.Len)
			}
			if len(b.Vars) != 0 {
				t.Errorf("after reset should have len(vars) == 0, got: %d", len(b.Vars))
			}
			if b.Prefix != "" {
				t.Errorf("after reset should have empty prefix, got: %s", b.Prefix)
			}
			if b.quote != 0 || b.quoteDollarTag != "" || b.multilineComment || b.balanceCount != 0 {
				t.Fatal("after reset should have a cleared parse state")
			}
			if st := b.State(); st != "=" {
				t.Errorf("after reset should have state `=`, got: %q", st)
			}
			if b.ready {
				t.Fatal("after reset should not be ready")
			}
		})
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

func TestVarSubstitute(t *testing.T) {
	a512 := sl(512, 'a')
	tests := []struct {
		s   string
		v   *Var
		sub string
		exp string
	}{
		{`:a`, v(0, `a`), `x`, `x`},
		{` :a`, v(1, `a`), `x`, ` x`},
		{`:a `, v(0, `a`), `x`, `x `},
		{` :a `, v(1, `a`), `x`, ` x `},
		{` :'a' `, v(1, `a`, `'`), `x`, ` 'x' `},
		{` :"a" `, v(1, "a", `"`), `x`, ` "x" `},
		{`:a`, v(0, `a`), ``, ``},
		{` :a`, v(1, `a`), ``, ` `},
		{`:a `, v(0, `a`), ``, ` `},
		{` :a `, v(1, `a`), ``, `  `},
		{` :'a' `, v(1, `a`, `'`), ``, ` '' `},
		{` :"a" `, v(1, "a", `"`), "", ` "" `},
		{` :aaa `, v(1, "aaa"), "", "  "},
		{` :aaa `, v(1, "aaa"), a512, " " + a512 + " "},
		{` :` + a512 + ` `, v(1, a512), "", "  "},
		{`:foo`, v(0, "foo"), "这是一个", `这是一个`},
		{`:foo `, v(0, "foo"), "这是一个", `这是一个 `},
		{` :foo`, v(1, "foo"), "这是一个", ` 这是一个`},
		{` :foo `, v(1, "foo"), "这是一个", ` 这是一个 `},
		{`:'foo'`, v(0, `foo`, `'`), `这是一个`, `'这是一个'`},
		{`:'foo' `, v(0, `foo`, `'`), `这是一个`, `'这是一个' `},
		{` :'foo'`, v(1, `foo`, `'`), `这是一个`, ` '这是一个'`},
		{` :'foo' `, v(1, `foo`, `'`), `这是一个`, ` '这是一个' `},
		{`:"foo"`, v(0, `foo`, `"`), `这是一个`, `"这是一个"`},
		{`:"foo" `, v(0, `foo`, `"`), `这是一个`, `"这是一个" `},
		{` :"foo"`, v(1, `foo`, `"`), `这是一个`, ` "这是一个"`},
		{` :"foo" `, v(1, `foo`, `"`), `这是一个`, ` "这是一个" `},
		{`:型`, v(0, `型`), `x`, `x`},
		{` :型`, v(1, `型`), `x`, ` x`},
		{`:型 `, v(0, `型`), `x`, `x `},
		{` :型 `, v(1, `型`), `x`, ` x `},
		{` :'型' `, v(1, `型`, `'`), `x`, ` 'x' `},
		{` :"型" `, v(1, "型", `"`), `x`, ` "x" `},
		{`:型`, v(0, `型`), ``, ``},
		{` :型`, v(1, `型`), ``, ` `},
		{`:型 `, v(0, `型`), ``, ` `},
		{` :型 `, v(1, `型`), ``, `  `},
		{` :'型' `, v(1, `型`, `'`), ``, ` '' `},
		{` :"型" `, v(1, "型", `"`), "", ` "" `},
		{`:型示師`, v(0, `型示師`), `本門台初埼本門台初埼`, `本門台初埼本門台初埼`},
		{` :型示師`, v(1, `型示師`), `本門台初埼本門台初埼`, ` 本門台初埼本門台初埼`},
		{`:型示師 `, v(0, `型示師`), `本門台初埼本門台初埼`, `本門台初埼本門台初埼 `},
		{` :型示師 `, v(1, `型示師`), `本門台初埼本門台初埼`, ` 本門台初埼本門台初埼 `},
		{` :型示師 `, v(1, `型示師`), `本門台初埼本門台初埼`, ` 本門台初埼本門台初埼 `},
		{` :'型示師' `, v(1, `型示師`, `'`), `a`, ` 'a' `},
		{` :"型示師" `, v(1, `型示師`, `"`), `b`, ` "b" `},
		{` :'型示師' `, v(1, `型示師`, `'`), `本門台初埼本門台初埼`, ` '本門台初埼本門台初埼' `},
		{` :"型示師" `, v(1, `型示師`, `"`), `本門台初埼本門台初埼`, ` "本門台初埼本門台初埼" `},
		{`:{?foo}`, v(0, `foo`, `?`), ``, `TRUE`},
		{`:{?foo_}`, v(0, `foo_`, `?`), ``, `FALSE`},
		{` :{?foo} `, v(1, `foo`, `?`), ``, ` TRUE `},
		{` :{?foo_} `, v(1, `foo_`, `?`), ``, ` FALSE `},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("`%s` `%s`: `%s` --> `%s`", test.v.Name, test.sub, test.s, test.exp)
			t.Logf("[]byte: %d", len([]byte(test.s)))
			t.Logf("[]rune: %d", len([]rune(test.s)))
			t.Logf("v.I:%d v.End:%d", test.v.I, test.v.End)
			r, n := test.v.Substitute([]rune(test.s), test.sub, test.v.Name == "foo")
			if exp := len([]rune(test.exp)); n != exp {
				t.Errorf("expected n to be %d, got: %d", exp, n)
			}
			if s := string(r); s != test.exp {
				t.Errorf("expected %q, got: %q", test.exp, s)
			}
		})
	}
}

func v(i int, name string, x ...string) *Var {
	z := &Var{
		I:    i,
		Name: name,
		End:  i + len([]rune(name)) + 1, // :name
	}
	if len(x) != 0 {
		z.Quote = []rune(x[0])[0]
		switch z.Quote {
		case '\'', '"':
			z.End += 2 // '', ""
		case '?':
			z.End += 3 // {?}
		}
	}
	return z
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

func sl(n int, r rune) string {
	z := make([]rune, n)
	for i := 0; i < n; i++ {
		z[i] = r
	}
	return string(z)
}
