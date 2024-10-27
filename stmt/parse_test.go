package stmt

import (
	"reflect"
	"strconv"
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
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			z := []rune(test.s)
			r := grab(z, test.i, len(z))
			if r != test.exp {
				t.Errorf("expected %c, got: %c", test.exp, r)
			}
		})
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
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			z := []rune(test.s)
			n, b := findSpace(z, test.i, len(z))
			if n != test.exp {
				t.Errorf("expected %d, got: %d", test.exp, n)
			}
			if b != test.b {
				t.Errorf("expected %t, got: %t", test.b, b)
			}
		})
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
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			z := []rune(test.s)
			n, b := findNonSpace(z, test.i, len(z))
			if n != test.exp {
				t.Errorf("expected %d, got: %d", test.exp, n)
			}
			if b != test.b {
				t.Errorf("expected %t, got: %t", test.b, b)
			}
		})
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
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			z := []rune(test.s)
			b := isEmptyLine(z, test.i, len(z))
			if b != test.exp {
				t.Errorf("expected %t, got: %t", test.exp, b)
			}
		})
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
		{`'本門台初埼本門台初埼'`, 0, `'本門台初埼本門台初埼'`, true},
		{` '本門台初埼本門台初埼' `, 1, `'本門台初埼本門台初埼'`, true},
		{`"本門台初埼本門台初埼"`, 0, `"本門台初埼本門台初埼"`, true},
		{` "本門台初埼本門台初埼" `, 1, `"本門台初埼本門台初埼"`, true},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r := []rune(test.s)
			c, end := rune(strings.TrimSpace(test.s)[0]), len(r)
			if c != '\'' && c != '"' && c != '`' {
				t.Fatal("incorrect!")
			}
			pos, ok := readString(r, test.i+1, end, c, "")
			if ok != test.ok {
				t.Fatalf("expected ok %t, got: %t", test.ok, ok)
			}
			if !test.ok {
				return
			}
			if r[pos] != c {
				t.Fatalf("expected last character to be %c, got: %c", c, r[pos])
			}
			v := string(r[test.i : pos+1])
			if n := len(v); n < 2 {
				t.Fatalf("expected result of at least length 2, got: %d", n)
			}
			if v != test.exp {
				t.Errorf("expected %q, got: %q", test.exp, v)
			}
		})
	}
}

func TestReadCommand(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp string
	}{
		{`\c foo bar z`, 0, `\c| foo bar z|`},
		{`\c foo bar z `, 0, `\c| foo bar z |`},
		{`\c foo bar z  `, 0, `\c| foo bar z  |`},
		{`\c    foo    bar    z  `, 0, `\c|    foo    bar    z  |`},
		{`\c    pg://blah    bar    z  `, 0, `\c|    pg://blah    bar    z  |`},
		{`\foo    pg://blah    bar    z  `, 0, `\foo|    pg://blah    bar    z  |`},
		{`\a\b`, 0, `\a||\b`},
		{`\a \b`, 0, `\a| |\b`},
		{"\\a \n\\b", 0, "\\a| |\n\\b"},
		{` \ab \bc \cd `, 5, `\bc| |\cd `},
		{`\p foo \p`, 0, `\p| foo |\p`},
		{`\p foo   \p bar`, 0, `\p| foo   |\p bar`},
		{`\p\p`, 0, `\p||\p`},
		{`\p \r foo`, 0, `\p| |\r foo`},
		{`\print   \reset    foo`, 0, `\print|   |\reset    foo`},
		{`\print   \reset    foo`, 9, `\reset|    foo|`},
		{`\print   \reset    foo  `, 9, `\reset|    foo  |`},
		{`\print   \reset    foo  bar  `, 9, `\reset|    foo  bar  |`},
		{`\c 'foo bar' z`, 0, `\c| 'foo bar' z|`},
		{`\c foo "bar " z `, 0, `\c| foo "bar " z |`},
		{"\\c `foo bar z  `  ", 0, "\\c| `foo bar z  `  |"},
		{`\c 'foob':foo:bar'test'  `, 0, `\c| 'foob':foo:bar'test'  |`},
		{"\\a \n\\b\\c\n", 0, "\\a| |\n\\b\\c\n"},
		{`\a'foob' \b`, 0, `\a'foob'| |\b`},
		{`\foo 'test' "bar"\print`, 0, `\foo| 'test' "bar"|\print`},
		{`\foo 'test' "bar"  \print`, 0, `\foo| 'test' "bar"  |\print`},
		{`\afoob' \b`, 0, `\afoob'| |\b`},
		{`\afoob' '\b  `, 0, `\afoob'| '\b  |`},
		{`\afoob' '\b  '\print`, 0, `\afoob'| '\b  '|\print`},
		{`\afoob' '\b  ' \print`, 0, `\afoob'| '\b  ' |\print`},
		{`\afoob' '\b  ' \print `, 0, `\afoob'| '\b  ' |\print `},
		{"\\foo `foob'foob'\\print", 0, "\\foo| `foob'foob'\\print|"},
		{"\\foo `foob'foob'  \\print", 0, "\\foo| `foob'foob'  \\print|"},
		{`\foo "foob'foob'\\print`, 0, `\foo| "foob'foob'\\print|`},
		{`\foo "foob'foob'  \\print`, 0, `\foo| "foob'foob'  \\print|`},
		{`\foo "\""\print`, 0, `\foo| "\""|\print`},
		{`\foo "\"'"\print`, 0, `\foo| "\"'"|\print`},
		{`\foo "\"''"\print`, 0, `\foo| "\"''"|\print`},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			z := []rune(test.s)
			if !strings.Contains(test.exp, "|") {
				t.Fatalf("expected value is invalid (missing |): %q", test.exp)
			}
			v := strings.Split(test.exp, "|")
			if len(v) != 3 {
				t.Fatalf("should have 3 expected values, has: %d", len(v))
			}
			cmd, params := readCommand(z, test.i, len(z))
			if s := string(z[test.i:cmd]); s != v[0] {
				t.Errorf("expected command to be %q, got: %q [%d, %d]", v[0], s, cmd, params)
			}
			if s := string(z[cmd:params]); s != v[1] {
				t.Errorf("expected params to be %q, got: %q [%d, %d]", v[1], s, cmd, params)
			}
			if s := string(z[params:]); s != v[2] {
				t.Errorf("expected remaining to be %q, got: %q", v[2], s)
			}
		})
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
		{" select to ", 4, "SELECT TO"},
		{" select   to   ", 4, "SELECT TO"},
		{"select into from", 2, "SELECT INTO"},
		{"select into * from", 4, "SELECT INTO"},
		{" select into  *   from  ", 4, "SELECT INTO"},
		{" select   \t  into \n *  \t\t\n\n\n  from     ", 4, "SELECT INTO"},
		{"  select\n\n\tb\t\tzfrom j\n\n  ", 2, "SELECT B"},
		{"select/* foob  */into", 4, "SELECTINTO"},
		{"select/* foob  */\tinto", 4, "SELECT INTO"},
		{"select/* foob  */ into", 4, "SELECT INTO"},
		{"select/* foob  */ into ", 4, "SELECT INTO"},
		{"select /* foob  */ into ", 4, "SELECT INTO"},
		{"   select /* foob  */ into ", 4, "SELECT INTO"},
		{" select * --test\n from where \n\nfff", 4, "SELECT"},
		{"/*idreamedital*/foo//bar\n/*  nothing */test\n\n\nwe made /*\n\n\n\n*/   \t   it    ", 5, "FOO TEST WE MADE IT"},
		{" --yes\n//no\n\n\t/*whatever*/ ", 4, ""},
		{"/*/*test*/*/ select ", 4, ""},
		{"/*/*test*/*/ select ", 4, ""},
		{"//", 4, ""},
		{"-", 4, ""},
		{"* select", 4, ""},
		{"/**/", 4, ""},
		{"--\n\t\t\thello,\t--", 4, "HELLO"},
		{"/*   */\n\n\n\tselect/*--\n*/\t\b\bzzz", 4, "SELECT ZZZ"},
		{"n\nn\n\nn\tn", 7, "N N N N"},
		{"n\nn\n\nn\tn", 1, "N"},
		{"--\n/* */n/* */\nn\n--\nn\tn", 7, "N N N N"},
		{"--\n/* */n\n/* */\nn\n--\nn\tn", 7, "N N N N"},
		{"\n\n/* */\nn n", 7, "N N"},
		{"\n\n/* */\nn/* */n", 7, "NN"},
		{"\n\n/* */\nn /* */n", 7, "N N"},
		{"\n\n/* */\nn/* */\nn", 7, "N N"},
		{"\n\n/* */\nn/* */ n", 7, "N N"},
		{"*/foob", 7, ""},
		{"*/ \n --\nfoob", 7, ""},
		{"--\n\n--\ntest", 7, "TEST"},
		{"\b\btest", 7, "TEST"},
		{"select/*\r\n\r\n*/blah", 7, "SELECTBLAH"},
		{"\r\n\r\nselect from where", 8, "SELECT FROM WHERE"},
		{"\r\n\b\bselect 1;create 2;", 8, "SELECT"},
		{"\r\n\bbegin transaction;\ncreate x where;", 8, "BEGIN TRANSACTION"},
		{"begin;test;create;awesome", 3, "BEGIN"},
		{" /* */ ; begin; ", 5, ""},
		{" /* foo */ test; test", 5, "TEST"},
		{";test", 5, ""},
		{"\b\b\t;test", 5, ""},
		{"\b\t; test", 5, ""},
		{"\b\tfoob; test", 5, "FOOB"},
		{"  TEST /*\n\t\b*/\b\t;foob", 10, "TEST"},
		{"begin transaction\n\tinsert into x;\ncommit;", 6, "BEGIN TRANSACTION INSERT INTO X"},
		{"--\nbegin /* */transaction/* */\n/* */\tinsert into x;--/* */\ncommit;", 6, "BEGIN TRANSACTION INSERT INTO X"},
		{"#\nbegin /* */transaction/* */\n/* */\t#\ninsert into x;#\n--/* */\ncommit;", 6, "BEGIN TRANSACTION INSERT INTO X"},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if p := findPrefix([]rune(test.s), test.w, true, true, true); p != test.exp {
				t.Errorf("%q expected %q, got: %q", test.s, test.exp, p)
			}
		})
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
		{`:'ab  ' `, 0, nil},
		{`:"ab  " `, 0, nil},
		{`:{?ab  } `, 0, nil},
		{`:a`, 0, v(0, `a`)},
		{`:ab`, 0, v(0, `ab`)},
		{`:a `, 0, v(0, `a`)},
		{`:a_ `, 0, v(0, `a_`)},
		{":a_\t ", 0, v(0, `a_`)},
		{":a_\n ", 0, v(0, `a_`)},
		{`:a9`, 0, v(0, `a9`)},
		{`:ab9`, 0, v(0, `ab9`)},
		{`:a 9`, 0, v(0, `a`)},
		{`:a_9 `, 0, v(0, `a_9`)},
		{":a_9\t ", 0, v(0, `a_9`)},
		{":a_9\n ", 0, v(0, `a_9`)},
		{`:a_;`, 0, v(0, `a_`)},
		{`:a_\`, 0, v(0, `a_`)},
		{`:a_$`, 0, v(0, `a_`)},
		{`:a_'`, 0, v(0, `a_`)},
		{`:a_"`, 0, v(0, `a_`)},
		{`:ab `, 0, v(0, `ab`)},
		{`:ab123 `, 0, v(0, `ab123`)},
		{`:ab123`, 0, v(0, `ab123`)},
		{`:'`, 0, nil},
		{`:' `, 0, nil},
		{`:' a`, 0, nil},
		{`:' a `, 0, nil},
		{`:"`, 0, nil},
		{`:" `, 0, nil},
		{`:" a`, 0, nil},
		{`:" a `, 0, nil},
		{`:''`, 0, nil},
		{`:'' `, 0, nil},
		{`:'' a`, 0, nil},
		{`:""`, 0, nil},
		{`:"" `, 0, nil},
		{`:"" a`, 0, nil},
		{`:'     `, 0, nil},
		{`:'       `, 0, nil},
		{`:"     `, 0, nil},
		{`:"       `, 0, nil},
		{`:'a'`, 0, v(0, `a`, `'`)},
		{`:'a' `, 0, v(0, `a`, `'`)},
		{`:'ab'`, 0, v(0, `ab`, `'`)},
		{`:'ab' `, 0, v(0, `ab`, `'`)},
		{`:"a"`, 0, v(0, `a`, `"`)},
		{`:"a" `, 0, v(0, `a`, `"`)},
		{`:"ab"`, 0, v(0, `ab`, `"`)},
		{`:"ab" `, 0, v(0, `ab`, `"`)},
		{`:型`, 0, v(0, "型")},
		{`:'型'`, 0, v(0, "型", `'`)},
		{`:"型"`, 0, v(0, "型", `"`)},
		{` :型 `, 1, v(1, "型")},
		{` :'型' `, 1, v(1, "型", `'`)},
		{` :"型" `, 1, v(1, "型", `"`)},
		{`:型示師`, 0, v(0, "型示師")},
		{`:'型示師'`, 0, v(0, "型示師", `'`)},
		{`:"型示師"`, 0, v(0, "型示師", `"`)},
		{` :型示師 `, 1, v(1, "型示師")},
		{` :'型示師' `, 1, v(1, "型示師", `'`)},
		{` :"型示師" `, 1, v(1, "型示師", `"`)},
		{`:{?a}`, 0, v(0, "a", `?`)},
		{` :{?a} `, 1, v(1, "a", `?`)},
		{`:{?a_b} `, 0, v(0, "a_b", `?`)},
		{` :{?a_b} `, 1, v(1, "a_b", `?`)},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("parsing %q", test.s)
			z := []rune(test.s)
			v := readVar(z, test.i, len(z), grab(z, test.i+1, len(z)))
			if !reflect.DeepEqual(v, test.exp) {
				t.Errorf("\nexpected: %#v\n     got: %#v", test.exp, v)
			}
			if test.exp != nil && v != nil {
				n := string(z[v.I+1 : v.End])
				switch v.Quote {
				case '\'', '"':
					if c := rune(n[0]); c != v.Quote {
						t.Errorf("expected var to start with quote %c, got: %c", c, v.Quote)
					}
					if c := rune(n[len(n)-1]); c != v.Quote {
						t.Errorf("expected var to end with quote %c, got: %c", c, v.Quote)
					}
					n = n[1 : len(n)-1]
				case '?':
					if !strings.HasPrefix(n, "{?") {
						t.Errorf("expected var %q to start with {?", n)
					}
					if !strings.HasSuffix(n, "}") {
						t.Errorf("expected var %q to end with }", n)
					}
					n = n[2 : len(n)-1]
				}
				if n != test.exp.Name {
					t.Errorf("expected var name of %q, got: %q", test.exp.Name, n)
				}
			}
		})
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
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r := []rune(test.s)
			r, rlen := substitute(r, test.i, len(r), test.n, test.t)
			if rlen != len(test.exp) {
				t.Errorf("expected length %d, got: %d", len(test.exp), rlen)
			}
			if s := string(r); s != test.exp {
				t.Errorf("expected %q, got %q", test.exp, s)
			}
		})
	}
}
