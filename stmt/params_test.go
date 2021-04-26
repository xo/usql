package stmt

import (
	"os/user"
	"reflect"
	"testing"

	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

func TestDecodeParamsGetRaw(t *testing.T) {
	const exp = `  'a string'  "another string"   `
	p := DecodeParams(exp)
	s := p.GetRaw()
	if s != exp {
		t.Errorf("expected %q, got: %q", exp, s)
	}
	u, err := user.Current()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	unquote := testUnquote(u, t, 0, exp)
	ok, s, err := p.Get(unquote)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if s != "" {
		t.Errorf("expected empty string, got: %q", s)
	}
	if ok {
		t.Errorf("expected ok=false, got: %t", ok)
	}
	v, err := p.GetAll(unquote)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected v to have length 0, got: %d", len(v))
	}
}

func TestDecodeParamsGetAll(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	tests := []struct {
		s   string
		exp []string
		err error
	}{
		{``, nil, nil},
		{` `, nil, nil},
		{` :foo`, []string{`bar`}, nil},
		{` :'foo`, nil, text.ErrUnterminatedQuotedString},
		{`:'foo'`, []string{`'bar'`}, nil},
		{`:'foo':foo`, []string{`'bar'bar`}, nil},
		{`:'foo':foo:"foo"`, []string{`'bar'bar"bar"`}, nil},
		{`:'foo':foo:foo`, []string{`'bar'barbar`}, nil},
		{` :'foo':foo:foo`, []string{`'bar'barbar`}, nil},
		{` :'foo':yes:foo`, []string{`'bar':yesbar`}, nil},
		{` :foo `, []string{`bar`}, nil},
		{`:foo:foo`, []string{`barbar`}, nil},
		{` :foo:foo `, []string{`barbar`}, nil},
		{`  :foo:foo  `, []string{`barbar`}, nil},
		{`'hello'`, []string{`hello`}, nil}, // 14
		{`  'hello''yes'  `, []string{`hello'yes`}, nil},
		{`  'hello\'...\'yes'  `, []string{`hello'...'yes`}, nil},
		{`  "hello\'...\'yes"  `, nil, text.ErrInvalidQuotedString},
		{`  "hello\"...\"yes"  `, nil, text.ErrInvalidQuotedString},
		{`  'hello':'yes'  `, []string{`hello:'yes'`}, nil},
		{` :'foo `, nil, text.ErrUnterminatedQuotedString},
		{` :'foo bar`, nil, text.ErrUnterminatedQuotedString},
		{` :'foo  bar`, nil, text.ErrUnterminatedQuotedString},
		{` :'foo  bar `, nil, text.ErrUnterminatedQuotedString},
		{" `foo", nil, text.ErrUnterminatedQuotedString},
		{" `foo bar`", []string{"foo bar"}, nil},
		{" `foo  :foo`", []string{"foo  :foo"}, nil},
		{` :'foo':"foo"`, []string{`'bar'"bar"`}, nil},
		{` :'foo' :"foo" `, []string{`'bar'`, `"bar"`}, nil},
		{` :'foo' :"foo"`, []string{`'bar'`, `"bar"`}, nil},
		{` :'foo'  :"foo"`, []string{`'bar'`, `"bar"`}, nil},
		{` :'foo'  :"foo" `, []string{`'bar'`, `"bar"`}, nil},
		{` :'foo'  :"foo"  :foo `, []string{`'bar'`, `"bar"`, `bar`}, nil},
		{` :'foo':foo:"foo" `, []string{`'bar'bar"bar"`}, nil}, // 30
		{` :'foo''yes':'foo' `, []string{`'bar'yes'bar'`}, nil},
		{` :'foo' 'yes' :'foo' `, []string{`'bar'`, `yes`, `'bar'`}, nil},
		{` 'yes':'foo':"foo"'blah''no' "\ntest" `, []string{`yes'bar'"bar"blah'no`, "\ntest"}, nil},
	}
	for i, test := range tests {
		vals, err := DecodeParams(test.s).GetAll(testUnquote(u, t, i, test.s))
		if err != test.err {
			t.Fatalf("test %d for %q expected err %v, got: %v", i, test.s, test.err, err)
		}
		if !reflect.DeepEqual(vals, test.exp) {
			t.Errorf("test %d for %q expected %v, got: %v", i, test.s, test.exp, vals)
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
