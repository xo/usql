package stmt

import (
	"os/user"
	"reflect"
	"testing"

	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

func TestDecodeParamsGetAll(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	tests := []struct {
		s   string
		exp []string
		err bool
	}{
		{``, nil, false},
		{` `, nil, false},
		{` :foo`, []string{`bar`}, false},
		{` :'foo`, nil, true},
		{`:'foo'`, []string{`'bar'`}, false},
		{`:'foo':foo`, []string{`'bar'bar`}, false},
		{`:'foo':foo:"foo"`, []string{`'bar'bar"bar"`}, false},
		{`:'foo':foo:foo`, []string{`'bar'barbar`}, false},
		{` :'foo':foo:foo`, []string{`'bar'barbar`}, false},
		{` :'foo':yes:foo`, []string{`'bar':yesbar`}, false},
		{` :foo `, []string{`bar`}, false},
		{`:foo:foo`, []string{`barbar`}, false},
		{` :foo:foo `, []string{`barbar`}, false},
		{`  :foo:foo  `, []string{`barbar`}, false},
		{`'hello'`, []string{`hello`}, false}, // 14
		{`  'hello''yes'  `, []string{`hello'yes`}, false},
		{`  'hello':'yes'  `, []string{`hello:'yes'`}, false},
		{` :'foo `, nil, true},
		{` :'foo bar`, nil, true},
		{` :'foo  bar`, nil, true},
		{` :'foo  bar `, nil, true},
		{" `foo", nil, true},
		{" `foo bar`", []string{"foo bar"}, false},
		{" `foo  :foo`", []string{"foo  :foo"}, false},
		{` :'foo':"foo"`, []string{`'bar'"bar"`}, false},
		{` :'foo' :"foo" `, []string{`'bar'`, `"bar"`}, false},
		{` :'foo' :"foo"`, []string{`'bar'`, `"bar"`}, false},
		{` :'foo'  :"foo"`, []string{`'bar'`, `"bar"`}, false},
		{` :'foo'  :"foo" `, []string{`'bar'`, `"bar"`}, false},
		{` :'foo'  :"foo"  :foo `, []string{`'bar'`, `"bar"`, `bar`}, false},
		{` :'foo':foo:"foo" `, []string{`'bar'bar"bar"`}, false}, // 30
		{` :'foo''yes':'foo' `, []string{`'bar'yes'bar'`}, false},
		{` :'foo' 'yes' :'foo' `, []string{`'bar'`, `yes`, `'bar'`}, false},
		{` 'yes':'foo':"foo"'blah''no' "\ntest" `, []string{`yes'bar'"bar"blah'no`, `\ntest`}, false},
	}
	for i, test := range tests {
		vals, err := DecodeParams(test.s).GetAll(testUnquote(u, t, i, test.s))
		if test.err && err != text.ErrUnterminatedQuotedString {
			t.Fatalf("test %d expected unterminated quoted string error, got: %v", i, err)
		}
		if !reflect.DeepEqual(vals, test.exp) {
			t.Errorf("test %d expected %v, got: %v", i, test.exp, vals)
		}
	}
}

func testUnquote(u *user.User, t *testing.T, i int, teststr string) func(string, bool) (bool, string, error) {
	f := env.Unquote(u, false, env.Vars{
		"foo": "bar",
	})
	return func(s string, isvar bool) (bool, string, error) {
		// t.Logf("test %d %q s: %q, isvar: %t", i, teststr, s, isvar)
		return f(s, isvar)
	}
}
