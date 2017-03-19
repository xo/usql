package handler

import (
	"strings"
)

// OnOff is a type that wraps a bool, for use in parsing/displaying command
// parameters.
type OnOff struct {
	Bool    bool
	Assumed bool
}

// String satisifies stringer.
func (b OnOff) String() string {
	if b.Bool {
		return "on"
	}
	return "off"
}

// MarshalText satisfies the TextMarhsaler interface.
func (b OnOff) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText satisfies the TextUnmarshaler interface.
func (b *OnOff) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch s {
	case "t", "true", "1", "on":
		b.Bool, b.Assumed = true, false
		return nil

	case "f", "false", "0", "off":
		b.Bool, b.Assumed = false, false
		return nil
	}

	v := len(s) != 0
	b.Bool, b.Assumed = v, v

	return nil
}
