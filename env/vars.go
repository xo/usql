package env

import (
	"fmt"
	"io"
	"maps"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	syslocale "github.com/jeandeaual/go-locale"
	"github.com/xo/terminfo"
	"github.com/xo/usql/text"
)

// Variables handles the standard, print, and connection variables.
type Variables struct {
	// vars holds standard variables.
	vars map[string]string
	// prnt holds print variables ("print" is a reserved word).
	prnt map[string]string
	// conn holds connection variables.
	conn map[string][]string
}

// NewVars creates a set of empty variables.
func NewVars() *Variables {
	return &Variables{
		vars: make(map[string]string),
		prnt: make(map[string]string),
		conn: make(map[string][]string),
	}
}

// NewDefaultVars creates standard, print, and connection variables, based on
// environment variables.
func NewDefaultVars() *Variables {
	cmdNameUpper := strings.ToUpper(text.CommandName)
	// get USQL_* variables
	showHostInformation := "true"
	if v, _ := Getenv(cmdNameUpper + "_SHOW_HOST_INFORMATION"); v != "" {
		showHostInformation = v
	}
	// get NO_COLOR
	noColor := false
	if s, ok := Getenv("NO_COLOR"); ok {
		noColor = s != "0" && s != "false" && s != "off"
	}
	// get color level
	colorLevel, _ := terminfo.ColorLevelFromEnv()
	enableSyntaxHL := "true"
	if noColor || colorLevel < terminfo.ColorLevelBasic {
		enableSyntaxHL = "false"
	}
	// pager
	pagerCmd, ok := Getenv(cmdNameUpper+"_PAGER", "PAGER")
	pager := "off"
	if !ok {
		for _, s := range []string{"less", "more"} {
			if _, err := exec.LookPath(s); err == nil {
				pagerCmd = s
				break
			}
		}
	}
	if pagerCmd != "" {
		pager = "on"
	}
	// editor
	editorCmd, _ := Getenv(cmdNameUpper+"_EDITOR", "EDITOR", "VISUAL")
	// sslmode
	sslmode, ok := Getenv(cmdNameUpper+"_SSLMODE", "SSLMODE")
	if !ok {
		sslmode = "retry"
	}
	// determine locale
	locale := "en-US"
	if s, err := syslocale.GetLocale(); err == nil {
		locale = s
	}
	return &Variables{
		vars: map[string]string{
			// usql related logic
			"SHOW_HOST_INFORMATION": showHostInformation,
			"PAGER":                 pagerCmd,
			"EDITOR":                editorCmd,
			"QUIET":                 "off",
			"ON_ERROR_STOP":         "off",
			// prompts
			"PROMPT1": "%S%N%m%/%R%# ",
			// syntax highlighting variables
			"SYNTAX_HL":             enableSyntaxHL,
			"SYNTAX_HL_FORMAT":      colorLevel.ChromaFormatterName(),
			"SYNTAX_HL_STYLE":       "monokai",
			"SYNTAX_HL_OVERRIDE_BG": "true",
			"SSLMODE":               sslmode,
			"TERM_GRAPHICS":         "none",
		},
		prnt: map[string]string{
			"border":                   "1",
			"columns":                  "0",
			"csv_fieldsep":             ",",
			"expanded":                 "off",
			"fieldsep":                 "|",
			"fieldsep_zero":            "off",
			"footer":                   "on",
			"format":                   "aligned",
			"linestyle":                "ascii",
			"locale":                   locale,
			"null":                     "",
			"numericlocale":            "off",
			"pager_min_lines":          "0",
			"pager":                    pager,
			"recordsep":                "\n",
			"recordsep_zero":           "off",
			"tableattr":                "",
			"time":                     "RFC3339Nano",
			"timezone":                 "",
			"title":                    "",
			"tuples_only":              "off",
			"unicode_border_linestyle": "single",
			"unicode_column_linestyle": "single",
			"unicode_header_linestyle": "single",
		},
		conn: make(map[string][]string),
	}
}

// Vars returns a copy of the standard variables.
func (v *Variables) Vars() map[string]string {
	return maps.Clone(v.vars)
}

// Print returns a copy of the print variables.
func (v *Variables) Print() map[string]string {
	return maps.Clone(v.prnt)
}

// Conn returns a copy of the connection variables.
func (v *Variables) Conn() map[string][]string {
	return maps.Clone(v.conn)
}

// Get retrieves a standard variable.
func (v *Variables) Get(name string) (string, bool) {
	value, ok := v.vars[name]
	return value, ok
}

// Set sets a standard variable.
func (v *Variables) Set(name, value string) error {
	if err := ValidIdentifier(name); err != nil {
		return err
	}
	switch name {
	case "ON_ERROR_STOP", "QUIET":
		if value == "" {
			value = "on"
		} else {
			var err error
			if value, err = ParseBool(value, name); err != nil {
				return err
			}
		}
	}
	v.vars[name] = value
	return nil
}

// Unset unsets a standard variable.
func (v *Variables) Unset(name string) error {
	if err := ValidIdentifier(name); err != nil {
		return err
	}
	delete(v.vars, name)
	return nil
}

// Dump dumps the standard variables to w.
func (v *Variables) Dump(w io.Writer) error {
	for _, k := range slices.Sorted(maps.Keys(v.vars)) {
		_, _ = fmt.Fprintln(w, k, "=", Quote(v.vars[k]))
	}
	return nil
}

// GetPrint returns a print variable.
func (v *Variables) GetPrint(name string) (string, error) {
	if val, ok := v.prnt[name]; ok {
		return val, nil
	}
	return "", fmt.Errorf(text.UnknownFormatFieldName, name)
}

// SetPrint sets a print variable.
func (v *Variables) SetPrint(name, value string) (string, error) {
	if _, ok := v.prnt[name]; !ok {
		return "", fmt.Errorf(text.UnknownFormatFieldName, name)
	}
	switch name {
	case "border", "columns", "pager_min_lines":
		i, _ := strconv.Atoi(value)
		v.prnt[name] = fmt.Sprintf("%d", i)
	case "pager":
		s, err := ParseKeywordBool(value, name, "always")
		if err != nil {
			return "", text.ErrInvalidFormatPagerType
		}
		v.prnt[name] = s
	case "expanded":
		s, err := ParseKeywordBool(value, name, "auto")
		if err != nil {
			return "", text.ErrInvalidFormatExpandedType
		}
		v.prnt[name] = s
	case "fieldsep_zero", "footer", "numericlocale", "recordsep_zero", "tuples_only":
		s, err := ParseBool(value, name)
		if err != nil {
			return "", err
		}
		v.prnt[name] = s
	case "format":
		if !formatRE.MatchString(value) {
			return "", text.ErrInvalidFormatType
		}
		v.prnt[name] = value
	case "linestyle":
		if !linestlyeRE.MatchString(value) {
			return "", text.ErrInvalidFormatLineStyle
		}
		v.prnt[name] = value
	case "csv_fieldsep", "fieldsep", "null", "recordsep", "tableattr", "time", "title", "locale":
		v.prnt[name] = value
	case "timezone":
		if _, err := time.LoadLocation(value); err != nil {
			return "", text.ErrInvalidTimezoneLocation
		}
		v.prnt[name] = value
	case "unicode_border_linestyle", "unicode_column_linestyle", "unicode_header_linestyle":
		if !borderRE.MatchString(value) {
			return "", text.ErrInvalidFormatBorderLineStyle
		}
		v.prnt[name] = value
	default:
		panic(fmt.Sprintf("field %s was defined in the print variables, but not in switch", name))
	}
	return v.prnt[name], nil
}

// TogglePrint toggles a print variable.
func (v *Variables) TogglePrint(name, extra string) (string, error) {
	if _, ok := v.prnt[name]; !ok {
		return "", fmt.Errorf(text.UnknownFormatFieldName, name)
	}
	switch name {
	case "border", "columns", "pager_min_lines":
	case "pager":
		switch v.prnt[name] {
		case "on", "always":
			v.prnt[name] = "off"
		case "off":
			v.prnt[name] = "on"
		default:
			panic(fmt.Sprintf("invalid state for field %s", name))
		}
	case "expanded":
		switch v.prnt[name] {
		case "on", "auto":
			v.prnt[name] = "off"
		case "off":
			v.prnt[name] = "on"
		default:
			panic(fmt.Sprintf("invalid state for field %s", name))
		}
	case "fieldsep_zero", "footer", "numericlocale", "recordsep_zero", "tuples_only":
		switch v.prnt[name] {
		case "on":
			v.prnt[name] = "off"
		case "off":
			v.prnt[name] = "on"
		default:
			panic(fmt.Sprintf("invalid state for field %s", name))
		}
	case "format":
		switch {
		case extra != "" && v.prnt[name] != extra:
			v.prnt[name] = extra
		case v.prnt[name] == "aligned":
			v.prnt[name] = "unaligned"
		default:
			v.prnt[name] = "aligned"
		}
	case "linestyle":
	case "csv_fieldsep", "fieldsep", "null", "recordsep", "time", "timezone", "locale":
	case "tableattr", "title":
		v.prnt[name] = ""
	case "unicode_border_linestyle", "unicode_column_linestyle", "unicode_header_linestyle":
	default:
		panic(fmt.Sprintf("field %s was defined in the print variables, but not in switch", name))
	}
	return v.prnt[name], nil
}

// DumpPrint dumps the print variables to w.
func (v *Variables) DumpPrint(w io.Writer) error {
	width, keys := 0, maps.Keys(v.prnt)
	for k := range keys {
		width = max(len(k), width)
	}
	for _, k := range slices.Sorted(keys) {
		val := v.prnt[k]
		switch k {
		case "csv_fieldsep", "fieldsep", "recordsep", "null":
			val = strconv.QuoteToASCII(val)
		case "tableattr", "title":
			if val != "" {
				val = strconv.QuoteToASCII(val)
			}
		}
		fmt.Fprintf(w, "%-*s %s\n", width, k, val)
		// k+strings.Repeat(" ", width-len(k)), val)
	}
	return nil
}

// PrintTimeFormat returns the user's time format converted to Go's time.Format
// value.
func (v *Variables) PrintTimeFormat() string {
	tfmt := v.prnt["time"]
	if s, ok := timeConsts[tfmt]; ok {
		return s
	}
	return tfmt
}

// SetConn sets a named connection variable.
func (v *Variables) SetConn(name string, vals ...string) error {
	if err := ValidIdentifier(name); err != nil {
		return err
	}
	if _, ok := v.conn[name]; len(vals) == 0 || vals[0] == "" && ok {
		delete(v.conn, name)
	} else {
		v.conn[name] = slices.Clone(vals)
	}
	return nil
}

// GetConn returns a connection variable.
func (v *Variables) GetConn(name string) ([]string, bool) {
	vals, ok := v.conn[name]
	if !ok {
		return nil, false
	}
	return slices.Clone(vals), true
}

// DumpConn dumps the connection variables to w.
func (v *Variables) DumpConn(w io.Writer) error {
	for _, k := range slices.Sorted(maps.Keys(v.conn)) {
		fmt.Fprintln(w, k, "=", Quote(strings.Join(v.conn[k], " ")))
	}
	return nil
}

// timeConsts are well known time consts.
var timeConsts = map[string]string{
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

/*
// Get retrieves a standard variable.
func (v *Vars) Get(s string) (string, bool, error) {
func (v *Vars) Unquote()
	q, n := "", s
	if c := s[0]; c == '\'' || c == '"' {
		var err error
		if n, err = Unquote(s); err != nil {
			return "", false, err
		}
		q = string(c)
	}
	if val, ok := v.v[n]; ok {
		return q + val + q, true, nil
	}
	return s, false, nil
*/
