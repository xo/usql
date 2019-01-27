package env

import (
	"fmt"
	"regexp"
	"strconv"
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

var vars, pvars Vars

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

	pvars = Vars{
		"border":                   "1",
		"columns":                  "0",
		"expanded":                 "off",
		"fieldsep":                 "|",
		"fieldsep_zero":            "off",
		"footer":                   "on",
		"format":                   "aligned",
		"linestyle":                "ascii",
		"null":                     "",
		"numericlocale":            "off",
		"pager":                    "1",
		"pager_min_lines":          "0",
		"recordsep":                "\n",
		"recordsep_zero":           "off",
		"tableattr":                "",
		"title":                    "",
		"tuples_only":              "off",
		"unicode_border_linestyle": "single",
		"unicode_column_linestyle": "single",
		"unicode_header_linestyle": "single",
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

// Pall returns all p variables.
func Pall() map[string]string {
	return pvars
}

var onRE = regexp.MustCompile(`(?i)^(t|tr|tru|true|on)$`)
var offRE = regexp.MustCompile(`(?i)^(f|fa|fal|fals|false|of|off)$`)
var formatRE = regexp.MustCompile(`^(unaligned|aligned|wrapped|html|asciidoc|latex|latex-longtable|troff-ms|csv|json)$`)
var linestlyeRE = regexp.MustCompile(`^(ascii|old-ascii|unicode)$`)
var borderRE = regexp.MustCompile(`^(single|double)$`)

func Pget(name string) (string, error) {
	v, ok := pvars[name]
	if !ok {
		return "", fmt.Errorf(text.UnknownFormatFieldName, name)
	}
	return v, nil
}

// Ptoggle toggles a p variable.
func Ptoggle(name, extra string) (string, error) {
	_, ok := pvars[name]
	if !ok {
		return "", fmt.Errorf(text.UnknownFormatFieldName, name)
	}

	switch name {
	case "border", "columns", "pager", "pager_min_lines":
	case "expanded":
		switch pvars[name] {
		case "on", "auto":
			pvars[name] = "off"
		case "off":
			pvars[name] = "on"
		default:
			panic(fmt.Sprintf("invalid state for field %s", name))
		}

	case "fieldsep_zero", "footer", "numericlocale", "recordsep_zero", "tuples_only":
		switch pvars[name] {
		case "on":
			pvars[name] = "off"
		case "off":
			pvars[name] = "on"
		default:
			panic(fmt.Sprintf("invalid state for field %s", name))
		}

	case "format":
		switch {
		case extra != "" && pvars[name] != extra:
			pvars[name] = extra
		case pvars[name] == "aligned":
			pvars[name] = "unaligned"
		default:
			pvars[name] = "aligned"
		}

	case "linestyle":
	case "fieldsep", "null", "recordsep":

	case "tableattr", "title":
		pvars[name] = ""

	case "unicode_border_linestyle", "unicode_column_linestyle", "unicode_header_linestyle":

	default:
		panic(fmt.Sprintf("field %s was defined in package pvars variable, but not in switch", name))
	}

	return pvars[name], nil
}

// Pset sets a p variable.
func Pset(name, value string) (string, error) {
	_, ok := pvars[name]
	if !ok {
		return "", fmt.Errorf(text.UnknownFormatFieldName, name)
	}

	switch name {
	case "border", "columns", "pager", "pager_min_lines":
		i, _ := strconv.Atoi(value)
		pvars[name] = fmt.Sprintf("%d", i)

	case "expanded":
		switch {
		case value == "auto":
			pvars[name] = "auto"
		case onRE.MatchString(value):
			pvars[name] = "on"
		case offRE.MatchString(value):
			pvars[name] = "off"
		default:
			return "", text.ErrInvalidFormatExpandedType
		}

	case "fieldsep_zero", "footer", "numericlocale", "recordsep_zero", "tuples_only":
		switch {
		case onRE.MatchString(value):
			pvars[name] = "on"
		case offRE.MatchString(value):
			pvars[name] = "off"
		default:
			return "", fmt.Errorf(text.FormatFieldInvalidValue, value, name, "Boolean")
		}

	case "format":
		if !formatRE.MatchString(value) {
			return "", text.ErrInvalidFormatType
		}
		pvars[name] = value

	case "linestyle":
		if !linestlyeRE.MatchString(value) {
			return "", text.ErrInvalidFormatLineStyle
		}
		pvars[name] = value

	case "fieldsep", "null", "recordsep", "tableattr", "title":
		pvars[name] = value

	case "unicode_border_linestyle", "unicode_column_linestyle", "unicode_header_linestyle":
		if !borderRE.MatchString(value) {
			return "", text.ErrInvalidFormatBorderLineStyle
		}
		pvars[name] = value

	default:
		panic(fmt.Sprintf("field %s was defined in package pvars variable, but not in switch", name))
	}

	return pvars[name], nil
}

// timeConstMap is the time const name to value map.
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
