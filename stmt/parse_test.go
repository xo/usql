package stmt

import (
	"reflect"
	"strings"
	"testing"
)

func TestGrab(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp rune
	}{
		{"", 0, 0},
		{"a", 0, 'a'},
		{" a", 0, ' '},
		{"a ", 1, ' '},
		{"a", 1, 0},
	}
	for i, test := range tests {
		z := []rune(test.s)
		r := grab(z, test.i, len(z))
		if r != test.exp {
			t.Errorf("test %d expected %c, got: %c", i, test.exp, r)
		}
	}
}

func TestFindSpace(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp int
		b   bool
	}{
		{"", 0, 0, false},
		{" ", 0, 0, true},
		{"a", 0, 1, false},
		{"a ", 0, 1, true},
		{" a ", 0, 0, true},
		{"aaa", 0, 3, false},
		{" a ", 1, 2, true},
		{"aaa", 1, 3, false},
		{" aaa", 1, 4, false},
	}
	for i, test := range tests {
		z := []rune(test.s)
		n, b := findSpace(z, test.i, len(z))
		if n != test.exp {
			t.Errorf("test %d expected %d, got: %d", i, test.exp, n)
		}
		if b != test.b {
			t.Errorf("test %d expected %t, got: %t", i, test.b, b)
		}
	}
}

func TestFindNonSpace(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp int
		b   bool
	}{
		{"", 0, 0, false},
		{" ", 0, 1, false},
		{"a", 0, 0, true},
		{"a ", 0, 0, true},
		{" a ", 0, 1, true},
		{"    ", 0, 4, false},
		{" a ", 1, 1, true},
		{"aaa", 1, 1, true},
		{" aaa", 1, 1, true},
		{"  aa", 1, 2, true},
		{"    ", 1, 4, false},
	}
	for i, test := range tests {
		z := []rune(test.s)
		n, b := findNonSpace(z, test.i, len(z))
		if n != test.exp {
			t.Errorf("test %d expected %d, got: %d", i, test.exp, n)
		}
		if b != test.b {
			t.Errorf("test %d expected %t, got: %t", i, test.b, b)
		}
	}
}

func TestIsEmptyLine(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp bool
	}{
		{"", 0, true},
		{"a", 0, false},
		{" a", 0, false},
		{" a ", 0, false},
		{" \na", 0, false},
		{" \n\ta", 0, false},
		{"a ", 1, true},
		{" a", 1, false},
		{" a ", 1, false},
		{" \na", 1, false},
		{" \n\t ", 1, true},
	}
	for i, test := range tests {
		z := []rune(test.s)
		b := isEmptyLine(z, test.i, len(z))
		if b != test.exp {
			t.Errorf("test %d expected %t, got: %t", i, test.exp, b)
		}
	}
}

func TestReadString(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp string
		ok  bool
	}{
		{`'`, 0, ``, false},
		{` '`, 1, ``, false},
		{`''`, 0, `''`, true},
		{`'foo' `, 0, `'foo'`, true},
		{` 'foo' `, 1, `'foo'`, true},
		{`"foo"`, 0, `"foo"`, true},
		{"`foo`", 0, "`foo`", true},
		{"`'foo'`", 0, "`'foo'`", true},
		{`'foo''foo'`, 0, `'foo''foo'`, true},
		{` 'foo''foo' `, 1, `'foo''foo'`, true},
		{` "foo''foo" `, 1, `"foo''foo"`, true},
		// escaped \" aren't allowed in strings, so the second " would be next
		// double quoted string
		{`"foo\""`, 0, `"foo\"`, true},
		{` "foo\"" `, 1, `"foo\"`, true},
		{`''''`, 0, `''''`, true},
		{` '''' `, 1, `''''`, true},
		{`''''''`, 0, `''''''`, true},
		{` '''''' `, 1, `''''''`, true},
		{`'''`, 0, ``, false},
		{` ''' `, 1, ``, false},
		{`'''''`, 0, ``, false},
		{` ''''' `, 1, ``, false},
		{`"fo'o"`, 0, `"fo'o"`, true},
		{` "fo'o" `, 1, `"fo'o"`, true},
		{`"fo''o"`, 0, `"fo''o"`, true},
		{` "fo''o" `, 1, `"fo''o"`, true},
	}
	for i, test := range tests {
		r := []rune(test.s)
		c, end := rune(strings.TrimSpace(test.s)[0]), len(r)
		if c != '\'' && c != '"' && c != '`' {
			t.Fatalf("test %d incorrect!", i)
		}
		pos, ok := readString(r, test.i+1, end, c, "")
		if ok != test.ok {
			t.Fatalf("test %d expected ok %t, got: %t", i, test.ok, ok)
		}
		if !test.ok {
			continue
		}
		if r[pos] != c {
			t.Fatalf("test %d expected last character to be %c, got: %c", i, c, r[pos])
		}
		v := string(r[test.i : pos+1])
		if n := len(v); n < 2 {
			t.Fatalf("test %d expected result of at least length 2, got: %d", i, n)
		}
		if v != test.exp {
			t.Errorf("test %d expected %q, got: %q", i, test.exp, v)
		}
	}
}

func TestReadCommand(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp string
	}{
		{`\c foo bar z`, 0, `\c| foo bar z|`}, // 0
		{`\c foo bar z `, 0, `\c| foo bar z |`},
		{`\c foo bar z  `, 0, `\c| foo bar z  |`},
		{`\c    foo    bar    z  `, 0, `\c|    foo    bar    z  |`},
		{`\c    pg://blah    bar    z  `, 0, `\c|    pg://blah    bar    z  |`},
		{`\foo    pg://blah    bar    z  `, 0, `\foo|    pg://blah    bar    z  |`}, // 5
		{`\a\b`, 0, `\a||\b`},
		{`\a \b`, 0, `\a| |\b`},
		{"\\a \n\\b", 0, "\\a| |\n\\b"},
		{` \ab \bc \cd `, 5, `\bc| |\cd `},
		{`\p foo \p`, 0, `\p| foo |\p`}, // 10
		{`\p foo   \p bar`, 0, `\p| foo   |\p bar`},
		{`\p\p`, 0, `\p||\p`},
		{`\p \r foo`, 0, `\p| |\r foo`},
		{`\print   \reset    foo`, 0, `\print|   |\reset    foo`},
		{`\print   \reset    foo`, 9, `\reset|    foo|`}, // 15
		{`\print   \reset    foo  `, 9, `\reset|    foo  |`},
		{`\print   \reset    foo  bar  `, 9, `\reset|    foo  bar  |`},
		{`\c 'foo bar' z`, 0, `\c| 'foo bar' z|`},
		{`\c foo "bar " z `, 0, `\c| foo "bar " z |`},
		{"\\c `foo bar z  `  ", 0, "\\c| `foo bar z  `  |"}, // 20
		{`\c 'aoeu':foo:bar'test'  `, 0, `\c| 'aoeu':foo:bar'test'  |`},
		{"\\a \n\\b\\c\n", 0, "\\a| |\n\\b\\c\n"},
		{`\a'aoeu' \b`, 0, `\a'aoeu'| |\b`},
		{`\foo 'test' "bar"\print`, 0, `\foo| 'test' "bar"|\print`}, // 25
		{`\foo 'test' "bar"  \print`, 0, `\foo| 'test' "bar"  |\print`},
		{`\aaoeu' \b`, 0, `\aaoeu'| |\b`},
		{`\aaoeu' '\b  `, 0, `\aaoeu'| '\b  |`},
		{`\aaoeu' '\b  '\print`, 0, `\aaoeu'| '\b  '|\print`},
		{`\aaoeu' '\b  ' \print`, 0, `\aaoeu'| '\b  ' |\print`}, // 30
		{`\aaoeu' '\b  ' \print `, 0, `\aaoeu'| '\b  ' |\print `},
		{"\\foo `aoeu'aoeu'\\print", 0, "\\foo| `aoeu'aoeu'\\print|"},
		{"\\foo `aoeu'aoeu'  \\print", 0, "\\foo| `aoeu'aoeu'  \\print|"},
		{`\foo "aoeu'aoeu'\\print`, 0, `\foo| "aoeu'aoeu'\\print|`},
		{`\foo "aoeu'aoeu'  \\print`, 0, `\foo| "aoeu'aoeu'  \\print|`}, // 35
		{`\foo "\""\print`, 0, `\foo| "\""|\print`},
		{`\foo "\"'"\print`, 0, `\foo| "\"'"|\print`},
		{`\foo "\"''"\print`, 0, `\foo| "\"''"|\print`},
	}
	for i, test := range tests {
		z := []rune(test.s)
		if !strings.Contains(test.exp, "|") {
			t.Fatalf("test %d expected value is invalid (missing |): %q", i, test.exp)
		}
		v := strings.Split(test.exp, "|")
		if len(v) != 3 {
			t.Fatalf("test %d should have 3 expected values, has: %d", i, len(v))
		}
		cmd, params := readCommand(z, test.i, len(z))
		if s := string(z[test.i:cmd]); s != v[0] {
			t.Errorf("test %d expected command to be `%s`, got: `%s` [%d, %d]", i, v[0], s, cmd, params)
		}
		if s := string(z[cmd:params]); s != v[1] {
			t.Errorf("test %d expected params to be `%s`, got: `%s` [%d, %d]", i, v[1], s, cmd, params)
		}
		if s := string(z[params:]); s != v[2] {
			t.Errorf("test %d expected remaining to be `%s`, got: `%s`", i, v[2], s)
		}
	}
}

func TestFindPrefix(t *testing.T) {
	tests := []struct {
		s   string
		w   int
		exp string
	}{
		{"", 4, ""},
		{"  ", 4, ""},
		{"  ", 4, ""},
		{" select ", 4, "SELECT"},
		{" select to ", 4, "SELECT TO"},
		{" select to ", 4, "SELECT TO"}, // 5
		{" select   to   ", 4, "SELECT TO"},
		{"select into from", 2, "SELECT INTO"},
		{"select into * from", 4, "SELECT INTO"},
		{" select into  *   from  ", 4, "SELECT INTO"},
		{" select   \t  into \n *  \t\t\n\n\n  from     ", 4, "SELECT INTO"}, // 10
		{"  select\n\n\tb\t\tzfrom j\n\n  ", 2, "SELECT B"},
		{"select/* aoeu  */into", 4, "SELECTINTO"}, // 12
		{"select/* aoeu  */\tinto", 4, "SELECT INTO"},
		{"select/* aoeu  */ into", 4, "SELECT INTO"},
		{"select/* aoeu  */ into ", 4, "SELECT INTO"},
		{"select /* aoeu  */ into ", 4, "SELECT INTO"},
		{"   select /* aoeu  */ into ", 4, "SELECT INTO"},
		{" select * --test\n from where \n\nfff", 4, "SELECT"},
		{"/*idreamedital*/foo//bar\n/*  nothing */test\n\n\nwe made /*\n\n\n\n*/   \t   it    ", 5, "FOO TEST WE MADE IT"},
		{" --yes\n//no\n\n\t/*whatever*/ ", 4, ""}, // 20
		{"/*/*test*/*/ select ", 4, ""},
		{"/*/*test*/*/ select ", 4, ""},
		{"//", 4, ""},
		{"-", 4, ""},
		{"* select", 4, ""},
		{"/**/", 4, ""},
		{"--\n\t\t\thello,\t--", 4, "HELLO"},
		{"/*   */\n\n\n\tselect/*--\n*/\t\b\bzzz", 4, "SELECT ZZZ"}, // 28
		{"n\nn\n\nn\tn", 7, "N N N N"},
		{"n\nn\n\nn\tn", 1, "N"},
		{"--\n/* */n/* */\nn\n--\nn\tn", 7, "N N N N"},
		{"--\n/* */n\n/* */\nn\n--\nn\tn", 7, "N N N N"},
		{"\n\n/* */\nn n", 7, "N N"},
		{"\n\n/* */\nn/* */n", 7, "NN"},
		{"\n\n/* */\nn /* */n", 7, "N N"},
		{"\n\n/* */\nn/* */\nn", 7, "N N"},
		{"\n\n/* */\nn/* */ n", 7, "N N"},
		{"*/aoeu", 7, ""},
		{"*/ \n --\naoeu", 7, ""},
		{"--\n\n--\ntest", 7, "TEST"}, // 40
		{"\b\btest", 7, "TEST"},
		{"select/*\r\n\r\n*/blah", 7, "SELECTBLAH"},
		{"\r\n\r\nselect from where", 8, "SELECT FROM WHERE"},
		{"\r\n\b\bselect 1;create 2;", 8, "SELECT"},
		{"\r\n\bbegin transaction;\ncreate x where;", 8, "BEGIN TRANSACTION"}, // 45
		{"begin;test;create;awesome", 3, "BEGIN"},
		{" /* */ ; begin; ", 5, ""},
		{" /* foo */ test; test", 5, "TEST"},
		{";test", 5, ""},
		{"\b\b\t;test", 5, ""},
		{"\b\t; test", 5, ""},
		{"\b\taoeu; test", 5, "AOEU"},
		{"  TEST /*\n\t\b*/\b\t;aoeu", 10, "TEST"},
		{"begin transaction\n\tinsert into x;\ncommit;", 6, "BEGIN TRANSACTION INSERT INTO X"},
		{"--\nbegin /* */transaction/* */\n/* */\tinsert into x;--/* */\ncommit;", 6, "BEGIN TRANSACTION INSERT INTO X"},
		{"#\nbegin /* */transaction/* */\n/* */\t#\ninsert into x;#\n--/* */\ncommit;", 6, "BEGIN TRANSACTION INSERT INTO X"},
	}
	for i, test := range tests {
		if p := findPrefix([]rune(test.s), test.w, true, true, true); p != test.exp {
			t.Errorf("test %d %q expected %q, got: %q", i, test.s, test.exp, p)
		}
	}
}

func TestReadVar(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp *Var
	}{
		{``, 0, nil},
		{`:`, 0, nil},
		{` :`, 0, nil},
		{`a:`, 0, nil},
		{`a:a`, 0, nil},
		{`: `, 0, nil},
		{`: a `, 0, nil},
		{`:a`, 0, v(0, 2, `a`)}, // 7
		{`:ab`, 0, v(0, 3, `ab`)},
		{`:a `, 0, v(0, 2, `a`)},
		{`:a_ `, 0, v(0, 3, `a_`)},
		{":a_\t ", 0, v(0, 3, `a_`)},
		{":a_\n ", 0, v(0, 3, `a_`)},
		{`:a9`, 0, v(0, 3, `a9`)}, // 13
		{`:ab9`, 0, v(0, 4, `ab9`)},
		{`:a 9`, 0, v(0, 2, `a`)},
		{`:a_9 `, 0, v(0, 4, `a_9`)},
		{":a_9\t ", 0, v(0, 4, `a_9`)},
		{":a_9\n ", 0, v(0, 4, `a_9`)},
		{`:a_;`, 0, v(0, 3, `a_`)}, // 19
		{`:a_\`, 0, v(0, 3, `a_`)},
		{`:a_$`, 0, v(0, 3, `a_`)},
		{`:a_'`, 0, v(0, 3, `a_`)},
		{`:a_"`, 0, v(0, 3, `a_`)},
		{`:ab `, 0, v(0, 3, `ab`)}, // 24
		{`:ab123 `, 0, v(0, 6, `ab123`)},
		{`:ab123`, 0, v(0, 6, `ab123`)},
		{`:'`, 0, nil}, // 27
		{`:' `, 0, nil},
		{`:' a`, 0, nil},
		{`:' a `, 0, nil},
		{`:"`, 0, nil},
		{`:" `, 0, nil},
		{`:" a`, 0, nil},
		{`:" a `, 0, nil},
		{`:''`, 0, nil}, // 35
		{`:'' `, 0, nil},
		{`:'' a`, 0, nil},
		{`:""`, 0, nil},
		{`:"" `, 0, nil},
		{`:"" a`, 0, nil},
		{`:'     `, 0, nil}, // 41
		{`:'       `, 0, nil},
		{`:"     `, 0, nil},
		{`:"       `, 0, nil},
		{`:'a'`, 0, v(0, 4, `a`, `'`)}, // 45
		{`:'a' `, 0, v(0, 4, `a`, `'`)},
		{`:'ab'`, 0, v(0, 5, `ab`, `'`)},
		{`:'ab' `, 0, v(0, 5, `ab`, `'`)},
		{`:'ab  ' `, 0, v(0, 7, `ab  `, `'`)},
		{`:"a"`, 0, v(0, 4, `a`, `"`)}, // 50
		{`:"a" `, 0, v(0, 4, `a`, `"`)},
		{`:"ab"`, 0, v(0, 5, `ab`, `"`)},
		{`:"ab" `, 0, v(0, 5, `ab`, `"`)},
		{`:"ab  " `, 0, v(0, 7, `ab  `, `"`)},
		{`:型`, 0, v(0, 2, "型")}, // 55
		{`:'型'`, 0, v(0, 4, "型", `'`)},
		{`:"型"`, 0, v(0, 4, "型", `"`)},
		{` :型 `, 1, v(1, 3, "型")},
		{` :'型' `, 1, v(1, 5, "型", `'`)},
		{` :"型" `, 1, v(1, 5, "型", `"`)},
		{`:型示師`, 0, v(0, 4, "型示師")}, // 61
		{`:'型示師'`, 0, v(0, 6, "型示師", `'`)},
		{`:"型示師"`, 0, v(0, 6, "型示師", `"`)},
		{` :型示師 `, 1, v(1, 5, "型示師")},
		{` :'型示師' `, 1, v(1, 7, "型示師", `'`)},
		{` :"型示師" `, 1, v(1, 7, "型示師", `"`)},
	}
	for i, test := range tests {
		z := []rune(test.s)
		v := readVar(z, test.i, len(z))
		if !reflect.DeepEqual(v, test.exp) {
			t.Errorf("test %d expected %#v, got: %#v", i, test.exp, v)
		}
		if test.exp != nil && v != nil {
			n := string(z[v.I+1 : v.End])
			if v.Quote != 0 {
				if c := rune(n[0]); c != v.Quote {
					t.Errorf("test %d expected var to start with quote %c, got: %c", i, c, v.Quote)
				}
				if c := rune(n[len(n)-1]); c != v.Quote {
					t.Errorf("test %d expected var to end with quote %c, got: %c", i, c, v.Quote)
				}
				n = n[1 : len(n)-1]
			}
			if n != test.exp.Name {
				t.Errorf("test %d expected var name of `%s`, got: `%s`", i, test.exp.Name, n)
			}
		}
	}
}

func TestSubstitute(t *testing.T) {
	a512 := sl(512, 'a')
	b512 := sl(512, 'a')
	b512 = b512[:1] + "b" + b512[2:]
	if len(b512) != 512 {
		t.Fatalf("b512 should be length 512, is: %d", len(b512))
	}
	tests := []struct {
		s   string
		i   int
		n   int
		t   string
		exp string
	}{
		{"", 0, 0, "", ""},
		{"a", 0, 1, "b", "b"},
		{"ab", 1, 1, "cd", "acd"},
		{"", 0, 0, "ab", "ab"},
		{"abc", 1, 2, "d", "ad"},
		{a512, 1, 1, "b", b512},
		{"foo", 0, 1, "bar", "baroo"},
	}
	for i, test := range tests {
		r := []rune(test.s)
		r, rlen := substitute(r, test.i, len(r), test.n, test.t)
		if rlen != len(test.exp) {
			t.Errorf("test %d expected length %d, got: %d", i, len(test.exp), rlen)
		}
		if s := string(r); s != test.exp {
			t.Errorf("test %d expected %q, got %q", i, test.exp, s)
		}
	}
}

func TestSubstituteVar(t *testing.T) {
	a512 := sl(512, 'a')
	tests := []struct {
		s   string
		v   *Var
		sub string
		exp string
	}{
		{`:a`, v(0, 2, `a`), `x`, `x`},
		{` :a`, v(1, 3, `a`), `x`, ` x`},
		{`:a `, v(0, 2, `a`), `x`, `x `},
		{` :a `, v(1, 3, `a`), `x`, ` x `},
		{` :'a' `, v(1, 5, `a`, `'`), `'x'`, ` 'x' `},
		{` :"a" `, v(1, 5, "a", `"`), `"x"`, ` "x" `},
		{`:a`, v(0, 2, `a`), ``, ``}, // 6
		{` :a`, v(1, 3, `a`), ``, ` `},
		{`:a `, v(0, 2, `a`), ``, ` `},
		{` :a `, v(1, 3, `a`), ``, `  `},
		{` :'a' `, v(1, 5, `a`, `'`), ``, `  `},
		{` :"a" `, v(1, 5, "a", `"`), "", `  `},
		{` :aaa `, v(1, 5, "aaa"), "", "  "}, // 12
		{` :aaa `, v(1, 5, "aaa"), a512, " " + a512 + " "},
		{` :` + a512 + ` `, v(1, len(a512)+2, a512), "", "  "},
		{`:foo`, v(0, 4, "foo"), "这是一个", `这是一个`}, // 15
		{`:foo `, v(0, 4, "foo"), "这是一个", `这是一个 `},
		{` :foo`, v(1, 5, "foo"), "这是一个", ` 这是一个`},
		{` :foo `, v(1, 5, "foo"), "这是一个", ` 这是一个 `},
		{`:'foo'`, v(0, 6, `foo`, `'`), `'这是一个'`, `'这是一个'`}, // 19
		{`:'foo' `, v(0, 6, `foo`, `'`), `'这是一个'`, `'这是一个' `},
		{` :'foo'`, v(1, 7, `foo`, `'`), `'这是一个'`, ` '这是一个'`},
		{` :'foo' `, v(1, 7, `foo`, `'`), `'这是一个'`, ` '这是一个' `},
		{`:"foo"`, v(0, 6, `foo`, `"`), `"这是一个"`, `"这是一个"`}, // 23
		{`:"foo" `, v(0, 6, `foo`, `"`), `"这是一个"`, `"这是一个" `},
		{` :"foo"`, v(1, 7, `foo`, `"`), `"这是一个"`, ` "这是一个"`},
		{` :"foo" `, v(1, 7, `foo`, `"`), `"这是一个"`, ` "这是一个" `},
		{`:型`, v(0, 2, `型`), `x`, `x`}, // 27
		{` :型`, v(1, 3, `型`), `x`, ` x`},
		{`:型 `, v(0, 2, `型`), `x`, `x `},
		{` :型 `, v(1, 3, `型`), `x`, ` x `},
		{` :'型' `, v(1, 5, `型`, `'`), `'x'`, ` 'x' `},
		{` :"型" `, v(1, 5, "型", `"`), `"x"`, ` "x" `},
		{`:型`, v(0, 2, `型`), ``, ``}, // 33
		{` :型`, v(1, 3, `型`), ``, ` `},
		{`:型 `, v(0, 2, `型`), ``, ` `},
		{` :型 `, v(1, 3, `型`), ``, `  `},
		{` :'型' `, v(1, 5, `型`, `'`), ``, `  `},
		{` :"型" `, v(1, 5, "型", `"`), "", `  `},
		{`:型示師`, v(0, 4, `型示師`), `本門台初埼本門台初埼`, `本門台初埼本門台初埼`}, // 39
		{` :型示師`, v(1, 5, `型示師`), `本門台初埼本門台初埼`, ` 本門台初埼本門台初埼`},
		{`:型示師 `, v(0, 4, `型示師`), `本門台初埼本門台初埼`, `本門台初埼本門台初埼 `},
		{` :型示師 `, v(1, 5, `型示師`), `本門台初埼本門台初埼`, ` 本門台初埼本門台初埼 `},
		{` :型示師 `, v(1, 5, `型示師`), `本門台初埼本門台初埼`, ` 本門台初埼本門台初埼 `},
		{` :'型示師' `, v(1, 7, `型示師`), `'本門台初埼本門台初埼'`, ` '本門台初埼本門台初埼' `},
		{` :"型示師" `, v(1, 7, `型示師`), `"本門台初埼本門台初埼"`, ` "本門台初埼本門台初埼" `},
	}
	for i, test := range tests {
		z := []rune(test.s)
		y, l := substituteVar(z, test.v, test.sub)
		if sl := len([]rune(test.sub)); test.v.Len != sl {
			t.Errorf("test %d, expected v.Len to be %d, got: %d", i, sl, test.v.Len)
		}
		if el := len([]rune(test.exp)); l != el {
			t.Errorf("test %d expected l==%d, got: %d", i, el, l)
		}
		if s := string(y); s != test.exp {
			t.Errorf("test %d expected `%s`, got: `%s`", i, test.exp, s)
		}
	}
}

func v(i, end int, n string, x ...string) *Var {
	z := &Var{
		I:    i,
		End:  end,
		Name: n,
	}
	if len(x) != 0 {
		z.Quote = []rune(x[0])[0]
	}
	return z
}
