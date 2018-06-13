package env

import (
	"time"
	"unicode"

	"github.com/xo/terminfo"
	"github.com/xo/usql/text"
)

// Vars is a map of variables.
type Vars map[string]string

// Set sets a variable name.
func (v Vars) Set(name, value string) {
	v[name] = value
}

// Unset unsets a variable name.
func (v Vars) Unset(name string) {
	delete(v, name)
}

// All returns all variables as a map.
func (v Vars) All() map[string]string {
	return map[string]string(v)
}

var vars Vars

func init() {
	// get USQL_* variables
	enableHostInformation := "true"
	if v := Getenv("USQL_SHOW_HOST_INFORMATION"); v != "" {
		enableHostInformation = v
	}
	timefmt := "RFC3339Nano"
	if v := Getenv("USQL_TIME_FORMAT"); v != "" {
		timefmt = v
	}

	// get color level
	colorLevel, _ := terminfo.ColorLevelFromEnv()
	enableSyntaxHL := "true"
	if colorLevel < terminfo.ColorLevelBasic {
		enableSyntaxHL = "false"
	}

	vars = Vars{
		// usql related logic
		"SHOW_HOST_INFORMATION": enableHostInformation,
		"TIME_FORMAT":           timefmt,

		// syntax highlighting variables
		"SYNTAX_HL":             enableSyntaxHL,
		"SYNTAX_HL_FORMAT":      colorLevel.ChromaFormatterName(),
		"SYNTAX_HL_STYLE":       "monokai",
		"SYNTAX_HL_OVERRIDE_BG": "true",
	}
}

// ValidIdentifier returns an error when n is not a valid identifier.
func ValidIdentifier(n string) error {
	r := []rune(n)
	rlen := len(r)
	if rlen < 1 {
		return text.ErrInvalidIdentifier
	}
	for i := 0; i < rlen; i++ {
		if c := r[i]; c != '_' && !unicode.IsLetter(c) && !unicode.IsNumber(c) {
			return text.ErrInvalidIdentifier
		}
	}
	return nil
}

// Set sets a variable.
func Set(name, value string) error {
	if err := ValidIdentifier(name); err != nil {
		return err
	}

	vars.Set(name, value)
	return nil
}

// Unset unsets a variable.
func Unset(name string) error {
	if err := ValidIdentifier(name); err != nil {
		return err
	}

	vars.Unset(name)
	return nil
}

// All returns all variables.
func All() map[string]string {
	return vars
}

// timeConstMap is the name
var timeConstMap = map[string]string{
	"ANSIC":       time.ANSIC,
	"UnixDate":    time.UnixDate,
	"RubyDate":    time.RubyDate,
	"RFC822":      time.RFC822,
	"RFC822Z":     time.RFC822Z,
	"RFC850":      time.RFC850,
	"RFC1123":     time.RFC1123,
	"RFC1123Z":    time.RFC1123Z,
	"RFC3339":     time.RFC3339,
	"RFC3339Nano": time.RFC3339Nano,
	"Kitchen":     time.Kitchen,
	"Stamp":       time.Stamp,
	"StampMilli":  time.StampMilli,
	"StampMicro":  time.StampMicro,
	"StampNano":   time.StampNano,
}

// Timefmt returns the environment TIME_FORMAT.
func Timefmt() string {
	tfmt := vars["TIME_FORMAT"]
	if s, ok := timeConstMap[tfmt]; ok {
		return s
	}
	return tfmt
}
