package informationschema

import (
	"testing"

	"github.com/xo/usql/drivers/metadata"
)

// errUnsupported stands in for the "converting NULL to X is unsupported" error
// database/sql raises, which the null* targets deliberately avoid.
type errUnsupported struct{ what string }

func (e errUnsupported) Error() string { return "converting NULL to " + e.what + " is unsupported" }

// nativeInt mirrors database/sql's conversion for an *int destination: it
// renders the driver value with asString and then parses it. Any value this
// accepts, nullInt must accept too.
func nativeInt(v any) (int, error) {
	s, ok := scanAsString(v)
	if !ok {
		return 0, errUnsupported{"int"}
	}
	if s == "" {
		return 0, nil
	}
	var n int
	neg := false
	i := 0
	if s[0] == '-' || s[0] == '+' {
		neg = s[0] == '-'
		i++
	}
	if i == len(s) {
		return 0, errUnsupported{"int"}
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, errUnsupported{"int"}
		}
		n = n*10 + int(s[i]-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}

// nativeStr mirrors database/sql's conversion for a *string destination, which
// accepts only string and []byte.
func nativeStr(v any) (string, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	case []byte:
		return string(t), nil
	}
	return "", errUnsupported{"string"}
}

var scanValues = []any{
	nil,
	int64(42), int64(0), int64(-7),
	float64(3.5), float64(0),
	"17", "", "abc", "-3",
	[]byte("23"), []byte(nil), []byte("abc"),
	true, false,
	uint64(9), int32(5),
}

// A narrower target would silently break the other databases that share this
// information_schema reader, so nullInt/nullStr must accept exactly what
// database/sql accepts -- and additionally absorb NULL.
func TestNullTargetsMatchDatabaseSQL(t *testing.T) {
	t.Run("nullInt", func(t *testing.T) {
		for _, v := range scanValues {
			var got int
			gotErr := nullInt{&got}.Scan(v)
			want, wantErr := nativeInt(v)
			if v == nil {
				wantErr = errUnsupported{"int"}
			}
			assertSame(t, "nullInt", v, got, gotErr, want, wantErr)
		}
	})
	t.Run("nullStr", func(t *testing.T) {
		for _, v := range scanValues {
			var got string
			gotErr := nullStr{&got}.Scan(v)
			var want string
			var wantErr error
			if v == nil {
				wantErr = errUnsupported{"string"}
			} else {
				want, wantErr = nativeStr(v)
			}
			assertSame(t, "nullStr", v, got, gotErr, want, wantErr)
		}
	})
	t.Run("nullBool", func(t *testing.T) {
		for _, v := range scanValues {
			var got metadata.Bool
			gotErr := nullBool{&got}.Scan(v)
			var want metadata.Bool
			var wantErr error
			if v == nil {
				wantErr = errUnsupported{"Bool"}
			} else {
				s, err := nativeStr(v)
				want, wantErr = metadata.Bool(s), err
			}
			assertSame(t, "nullBool", v, string(got), gotErr, string(want), wantErr)
		}
	})
}

func assertSame[T comparable](t *testing.T, fn string, v any, got T, gotErr error, want T, wantErr error) {
	t.Helper()
	if (gotErr == nil) != (wantErr == nil) {
		t.Errorf("%s(%T(%v)): got err=%v, database/sql err=%v", fn, v, v, gotErr, wantErr)
		return
	}
	if gotErr == nil && got != want {
		t.Errorf("%s(%T(%v)): got %v, database/sql %v", fn, v, v, got, want)
	}
}

// TestNullTargetsAbsorbNULL is the whole point of the change: NULL must be
// absorbed rather than aborting the metadata listing.
func TestNullTargetsAbsorbNULL(t *testing.T) {
	var s string
	var i int
	var b metadata.Bool
	var bv bool
	for _, tc := range []struct {
		name string
		fn   func() error
	}{
		{"nullStr", func() error { return (nullStr{&s}).Scan(nil) }},
		{"nullInt", func() error { return (nullInt{&i}).Scan(nil) }},
		{"nullBool", func() error { return (nullBool{&b}).Scan(nil) }},
		{"nullBoolVal", func() error { return (nullBoolVal{&bv}).Scan(nil) }},
	} {
		if err := tc.fn(); err != nil {
			t.Errorf("%s(nil) = %v, want nil", tc.name, err)
		}
	}
	if s != "" || i != 0 || b != "" || bv {
		t.Errorf("NULL did not zero out: %q %d %q %v", s, i, b, bv)
	}
}

// nullBoolVal additionally has to cope with information_schema.yes_or_no, a
// domain over boolean in PostgreSQL but reported as a number by openGauss.
func TestNullBoolValAcceptsYesNo(t *testing.T) {
	for _, tc := range []struct {
		in   any
		want bool
	}{
		{true, true}, {false, false},
		{int64(1), true}, {int64(0), false},
		{[]byte("YES"), true}, {[]byte("NO"), false},
		{"YES", true}, {"NO", false},
	} {
		var got bool
		if err := (nullBoolVal{&got}).Scan(tc.in); err != nil {
			t.Errorf("nullBoolVal(%v) = %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("nullBoolVal(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
