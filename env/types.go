package env

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	syslocale "github.com/jeandeaual/go-locale"
	"github.com/xo/terminfo"
	"github.com/xo/usql/text"
	"github.com/yookoala/realpath"
)

type varName struct {
	name string
	desc string
}

func (v varName) String() string {
	return fmt.Sprintf("  %s\n    %s\n", v.name, v.desc)
}

var varNames = []varName{
	{
		"ECHO_HIDDEN",
		"if set, display internal queries executed by backslash commands; if set to \"noexec\", just show them without execution",
	},
	{
		"ON_ERROR_STOP",
		"stop batch execution after error",
	},
	{
		"PROMPT1",
		"specifies the standard " + text.CommandName + " prompt",
	},
	{
		"QUIET",
		"run quietly (same as -q option)",
	},
	{
		"ROW_COUNT",
		"number of rows returned or affected by last query, or 0",
	},
}

var pvarNames = []varName{
	{
		"border",
		"border style (number)",
	},
	{
		"columns",
		"target width for the wrapped format",
	},
	{
		"csv_fieldsep",
		`field separator for CSV output (default ",")`,
	},
	{
		"expanded",
		"expanded output [on, off, auto]",
	},
	{
		"fieldsep",
		`field separator for unaligned output (default "|")`,
	},
	{
		"fieldsep_zero",
		"set field separator for unaligned output to a zero byte",
	},
	{
		"footer",
		"enable or disable display of the table footer [on, off]",
	},
	{
		"format",
		"set output format [unaligned, aligned, wrapped, vertical, html, asciidoc, csv, json, ...]",
	},
	{
		"linestyle",
		"set the border line drawing style [ascii, old-ascii, unicode]",
	},
	{
		"null",
		"set the string to be printed in place of a null value",
	},
	{
		"numericlocale",
		"enable display of a locale-specific character to separate groups of digits",
	},
	{
		"pager_min_lines",
		"minimum number of lines required in the output to use a pager, 0 to disable (default)",
	},
	{
		"pager",
		"control when an external pager is used [on, off, always]",
	},
	{
		"recordsep",
		"record (line) separator for unaligned output",
	},
	{
		"recordsep_zero",
		"set record separator for unaligned output to a zero byte",
	},
	{
		"tableattr",
		"specify attributes for table tag in html format, or proportional column widths for left-aligned data types in latex-longtable format",
	},
	{
		"time",
		`format used to display time/date column values (default "RFC3339Nano")`,
	},
	{
		"timezone",
		`the timezone to display dates in (default '')`,
	},
	{
		"title",
		"set the table title for subsequently printed tables",
	},
	{
		"tuples_only",
		"if set, only actual table data is shown",
	},
	{
		"unicode_border_linestyle",
		"set the style of Unicode line drawing [single, double]",
	},
	{
		"unicode_column_linestyle",
		"set the style of Unicode line drawing [single, double]",
	},
	{
		"unicode_header_linestyle",
		"set the style of Unicode line drawing [single, double]",
	},
}

var envVarNames = []varName{
	{
		text.CommandUpper() + "_EDITOR, EDITOR, VISUAL",
		"editor used by the \\e, \\ef, and \\ev commands",
	},
	{
		text.CommandUpper() + "_EDITOR_LINENUMBER_ARG",
		"how to specify a line number when invoking the editor",
	},
	{
		text.CommandUpper() + "_HISTORY",
		"alternative location for the command history file",
	},
	{
		text.CommandUpper() + "_PAGER, PAGER",
		"name of external pager program",
	},
	{
		text.CommandUpper() + "_SHOW_HOST_INFORMATION",
		"display host information when connecting to a database",
	},
	{
		text.CommandUpper() + "RC",
		"alternative location for the user's .usqlrc file",
	},
	{
		text.CommandUpper() + "_SSLMODE, SSLMODE",
		"when set to 'retry', allows connections to attempt to reconnect when no ?sslmode= was specified on the url",
	},
	{
		"SYNTAX_HL",
		"enable syntax highlighting",
	},
	{
		"SYNTAX_HL_FORMAT",
		"chroma library formatter name",
	},
	{
		"SYNTAX_HL_STYLE",
		`chroma library style name (default "monokai")`,
	},
	{
		"SYNTAX_HL_OVERRIDE_BG",
		"enables overriding the background color of the chroma styles",
	},
	{
		"TERM_GRAPHICS",
		`use the specified terminal graphics`,
	},
	{
		"SHELL",
		"shell used by the \\! command",
	},
}

// Vars is a map of variables to their values.
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

// vars are the environment variables.
var vars Vars

// pvars are the environment printing variables.
var pvars Vars

// cvars are the environment named connections.
var cvars map[string][]string

func init() {
	cmdNameUpper := strings.ToUpper(text.CommandName)
	// get USQL_* variables
	enableHostInformation := "true"
	if v, _ := Getenv(cmdNameUpper + "_SHOW_HOST_INFORMATION"); v != "" {
		enableHostInformation = v
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
	vars = Vars{
		// usql related logic
		"SHOW_HOST_INFORMATION": enableHostInformation,
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
	}
	// determine locale
	locale := "en-US"
	if s, err := syslocale.GetLocale(); err == nil {
		locale = s
	}
	pvars = Vars{
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
	}
	cvars = make(map[string][]string)
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
	if name == "ON_ERROR_STOP" || name == "QUIET" {
		if value == "" {
			value = "on"
		} else {
			var err error
			if value, err = ParseBool(value, name); err != nil {
				return err
			}
		}
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
func All() Vars {
	m := make(Vars)
	for k, v := range vars {
		m[k] = v
	}
	return m
}

// Pall returns all p variables.
func Pall() Vars {
	m := make(Vars)
	for k, v := range pvars {
		m[k] = v
	}
	return m
}

// Pwrite writes the p variables to the writer.
func Pwrite(w io.Writer) error {
	keys := make([]string, len(pvars))
	var i, width int
	for k := range pvars {
		keys[i], width = k, max(len(k), width)
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		val := pvars[k]
		switch k {
		case "csv_fieldsep", "fieldsep", "recordsep", "null":
			val = strconv.QuoteToASCII(val)
		case "tableattr", "title":
			if val != "" {
				val = strconv.QuoteToASCII(val)
			}
		}
		fmt.Fprintln(w, k+strings.Repeat(" ", width-len(k)), val)
	}
	return nil
}

var (
	formatRE    = regexp.MustCompile(`^(unaligned|aligned|wrapped|html|asciidoc|latex|latex-longtable|troff-ms|csv|json|vertical)$`)
	linestlyeRE = regexp.MustCompile(`^(ascii|old-ascii|unicode)$`)
	borderRE    = regexp.MustCompile(`^(single|double)$`)
)

func ParseBool(value, name string) (string, error) {
	switch strings.ToLower(value) {
	case "1", "t", "tr", "tru", "true", "on":
		return "on", nil
	case "0", "f", "fa", "fal", "fals", "false", "of", "off":
		return "off", nil
	}
	return "", fmt.Errorf(text.FormatFieldInvalidValue, value, name, "Boolean")
}

func ParseKeywordBool(value, name string, keywords ...string) (string, error) {
	v := strings.ToLower(value)
	switch v {
	case "1", "t", "tr", "tru", "true", "on":
		return "on", nil
	case "0", "f", "fa", "fal", "fals", "false", "of", "off":
		return "off", nil
	}
	for _, k := range keywords {
		if v == k {
			return v, nil
		}
	}
	return "", fmt.Errorf(text.FormatFieldInvalid, value, name)
}

func Get(name string) string {
	return vars[name]
}

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
	case "border", "columns", "pager_min_lines":
	case "pager":
		switch pvars[name] {
		case "on", "always":
			pvars[name] = "off"
		case "off":
			pvars[name] = "on"
		default:
			panic(fmt.Sprintf("invalid state for field %s", name))
		}
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
	case "csv_fieldsep", "fieldsep", "null", "recordsep", "time", "timezone", "locale":
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
	case "border", "columns", "pager_min_lines":
		i, _ := strconv.Atoi(value)
		pvars[name] = fmt.Sprintf("%d", i)
	case "pager":
		s, err := ParseKeywordBool(value, name, "always")
		if err != nil {
			return "", text.ErrInvalidFormatPagerType
		}
		pvars[name] = s
	case "expanded":
		s, err := ParseKeywordBool(value, name, "auto")
		if err != nil {
			return "", text.ErrInvalidFormatExpandedType
		}
		pvars[name] = s
	case "fieldsep_zero", "footer", "numericlocale", "recordsep_zero", "tuples_only":
		s, err := ParseBool(value, name)
		if err != nil {
			return "", err
		}
		pvars[name] = s
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
	case "csv_fieldsep", "fieldsep", "null", "recordsep", "tableattr", "time", "title", "locale":
		pvars[name] = value
	case "timezone":
		if _, err := time.LoadLocation(value); err != nil {
			return "", text.ErrInvalidTimezoneLocation
		}
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

// Cset sets a named connection for the environment.
func Cset(name string, vals ...string) error {
	if err := ValidIdentifier(name); err != nil {
		return err
	}
	if _, ok := cvars[name]; len(vals) == 0 || vals[0] == "" && ok {
		delete(cvars, name)
	} else {
		v := make([]string, len(vals))
		copy(v, vals)
		cvars[name] = v
	}
	return nil
}

// Cget returns the environment's named connection.
func Cget(name string) ([]string, bool) {
	vals, ok := cvars[name]
	if !ok {
		return nil, false
	}
	v := make([]string, len(vals))
	copy(v, vals)
	return v, true
}

// Call returns all named connections from the environment.
func Call() map[string][]string {
	m := make(map[string][]string, len(cvars))
	for k, vals := range cvars {
		v := make([]string, len(vals))
		copy(v, vals)
		m[k] = v
	}
	return m
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

// GoTime returns the user's time format converted to Go's time.Format value.
func GoTime() string {
	tfmt := pvars["time"]
	if s, ok := timeConsts[tfmt]; ok {
		return s
	}
	return tfmt
}

// Listing writes a formatted listing of the special environment variables to
// w, separated in sections based on variable type.
func Listing(w io.Writer) {
	varsWithDesc := make([]string, len(varNames))
	for i, v := range varNames {
		varsWithDesc[i] = v.String()
	}
	pvarsWithDesc := make([]string, len(pvarNames))
	for i, v := range pvarNames {
		pvarsWithDesc[i] = v.String()
	}

	// determine config dir name
	configDir, configExtra := buildConfigDir("config.yaml")

	// environment var names
	configDesc := configDir
	if configExtra != "" {
		configDesc = configExtra
	}
	ev := []varName{
		{
			text.CommandUpper() + "_CONFIG",
			fmt.Sprintf(`config file path (default %q)`, configDesc),
		},
	}
	envVarsWithDesc := make([]string, len(envVarNames)+1)
	for i, v := range append(ev, envVarNames...) {
		envVarsWithDesc[i] = v.String()
	}

	if configExtra != "" {
		configExtra = " (" + configExtra + ")"
	}

	template := `List of specially treated variables

%s variables:
Usage:
  %[1]s --set=NAME=VALUE
  or \set NAME VALUE inside %[1]s

%[2]s

Display settings:
Usage:
  %[1]s --pset=NAME[=VALUE]
  or \pset NAME [VALUE] inside %[1]s

%[3]s

Environment variables:
Usage:
  NAME=VALUE [NAME=VALUE] %[1]s ...
  or \setenv NAME [VALUE] inside %[1]s

%[4]s

Connection variables:
Usage:
  %[1]s --cset NAME[=DSN]
  or \cset NAME [DSN] inside %[1]s
  or \cset NAME DRIVER PARAMS... inside %[1]s
  or define in %[5]s%[6]s
`
	fmt.Fprintf(
		w, template,
		text.CommandName,
		strings.TrimRightFunc(strings.Join(varsWithDesc, ""), unicode.IsSpace),
		strings.TrimRightFunc(strings.Join(pvarsWithDesc, ""), unicode.IsSpace),
		strings.TrimRightFunc(strings.Join(envVarsWithDesc, ""), unicode.IsSpace),
		configDir,
		configExtra,
	)
}

func buildConfigDir(configName string) (string, string) {
	dir := `$HOME/.config/usql`
	switch runtime.GOOS {
	case "darwin":
		dir = `$HOME/Library/Application Support`
	case "windows":
		dir = `%AppData%\usql`
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(dir, configName), ""
	}
	if configDir, err = realpath.Realpath(configDir); err != nil {
		return filepath.Join(dir, configName), ""
	}
	return filepath.Join(dir, configName), filepath.Join(configDir, "usql", configName)
}
