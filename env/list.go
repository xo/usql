package env

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"github.com/xo/usql/text"
	"github.com/yookoala/realpath"
)

// Listing writes a formatted listing of the special environment variables to
// w, separated in sections based on variable type.
func Listing(w io.Writer) error {
	varsWithDesc := make([]string, len(varNames))
	for i, v := range varNames {
		varsWithDesc[i] = v.String()
	}
	settingsWithDesc := make([]string, len(settingNames))
	for i, v := range settingNames {
		settingsWithDesc[i] = v.String()
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
	fmt.Fprintf(
		w,
		variableTpl,
		text.CommandName,
		strings.TrimRightFunc(strings.Join(varsWithDesc, ""), unicode.IsSpace),
		strings.TrimRightFunc(strings.Join(pvarsWithDesc, ""), unicode.IsSpace),
		strings.TrimRightFunc(strings.Join(envVarsWithDesc, ""), unicode.IsSpace),
		configDir,
		configExtra,
		strings.TrimRightFunc(strings.Join(settingsWithDesc, ""), unicode.IsSpace),
	)
	return nil
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

type varName struct {
	name string
	desc string
}

func (v varName) String() string {
	return fmt.Sprintf("  %s\n    %s\n", v.name, v.desc)
}

var varNames = []varName{
	{
		`ROW_COUNT`,
		`number of rows returned or affected by last query, or 0`,
	},
}

// settingNames are the settings of the application, which \uset sets.
var settingNames = []varName{
	{
		`echo_hidden`,
		`if set, display internal queries executed by backslash commands; if set to "noexec", shows queries without execution`,
	},
	{
		`editor`,
		`editor used by the \e, \ef, and \ev commands`,
	},
	{
		`on_error_stop`,
		`stop batch execution after error`,
	},
	{
		`pager`,
		`name of external pager program`,
	},
	{
		`prompt1`,
		`specifies the standard ` + text.CommandName + ` prompt`,
	},
	{
		`prompt2`,
		`specifies the prompt used when a statement continues on another row`,
	},
	{
		`quiet`,
		`run quietly (same as -q option)`,
	},
	{
		`show_host_information`,
		`display host information when connecting to a database`,
	},
	{
		`sslmode`,
		`when set to 'retry', allows connections to attempt to reconnect when no ?sslmode= was specified on the url`,
	},
	{
		`syntax_hl`,
		`enable syntax highlighting (not used yet)`,
	},
	{
		`syntax_hl_format`,
		`syntax highlighting formatter name (not used yet)`,
	},
	{
		`syntax_hl_override_bg`,
		`overrides the background color of the highlighting style (not used yet)`,
	},
	{
		`syntax_hl_style`,
		`syntax highlighting style name (not used yet)`,
	},
	{
		`term_graphics`,
		`use the specified terminal graphics`,
	},
}

var (
	formatRE    = regexp.MustCompile(`^(unaligned|aligned|wrapped|html|asciidoc|latex|latex-longtable|troff-ms|csv|json|vertical)$`)
	linestlyeRE = regexp.MustCompile(`^(ascii|old-ascii|unicode)$`)
	borderRE    = regexp.MustCompile(`^(single|double)$`)
)

var pvarNames = []varName{
	{
		`border`,
		`border style (number)`,
	},
	{
		`columns`,
		`target width for the wrapped format`,
	},
	{
		`csv_fieldsep`,
		`field separator for CSV output (default ",")`,
	},
	{
		`expanded`,
		`expanded output [on, off, auto]`,
	},
	{
		`fieldsep`,
		`field separator for unaligned output (default "|")`,
	},
	{
		`fieldsep_zero`,
		`set field separator for unaligned output to a zero byte`,
	},
	{
		`footer`,
		`enable or disable display of the table footer [on, off]`,
	},
	{
		`format`,
		`set output format [unaligned, aligned, wrapped, vertical, html, asciidoc, csv, json, ...]`,
	},
	{
		`linestyle`,
		`set the border line drawing style [ascii, old-ascii, unicode]`,
	},
	{
		`null`,
		`set the string to be printed in place of a null value`,
	},
	{
		`numericlocale`,
		`enable display of a locale-specific character to separate groups of digits`,
	},
	{
		`pager_min_lines`,
		`minimum number of lines required in the output to use a pager, 0 to disable (default 0)`,
	},
	{
		`pager`,
		`control when an external pager is used [on, off, always]`,
	},
	{
		`recordsep`,
		`record (line) separator for unaligned output`,
	},
	{
		`recordsep_zero`,
		`set record separator for unaligned output to a zero byte`,
	},
	{
		`tableattr`,
		`specify attributes for table tag in html format, or proportional column widths for left-aligned data types in latex-longtable format`,
	},
	{
		`time`,
		`format used to display time/date column values (default RFC3339Nano)`,
	},
	{
		`timezone`,
		`the timezone to display dates in (default "")`,
	},
	{
		`title`,
		`set the table title for subsequently printed tables`,
	},
	{
		`tuples_only`,
		`if set, only actual table data is shown`,
	},
	{
		`unicode_border_linestyle`,
		`set the style of Unicode line drawing [single, double]`,
	},
	{
		`unicode_column_linestyle`,
		`set the style of Unicode line drawing [single, double]`,
	},
	{
		`unicode_header_linestyle`,
		`set the style of Unicode line drawing [single, double]`,
	},
}

var envVarNames = []varName{
	{
		text.CommandUpper() + `_EDITOR, EDITOR, VISUAL`,
		`editor used by the \e, \ef, and \ev commands`,
	},
	{
		text.CommandUpper() + `_EDITOR_LINENUMBER_ARG`,
		`how to specify a line number when invoking the editor`,
	},
	{
		text.CommandUpper() + `_HISTORY`,
		`alternative location for the command history file`,
	},
	{
		text.CommandUpper() + `_PAGER, PAGER`,
		`name of external pager program`,
	},
	{
		text.CommandUpper() + `_SHOW_HOST_INFORMATION`,
		`display host information when connecting to a database`,
	},
	{
		text.CommandUpper() + `RC`,
		`alternative location for the user's .usqlrc file`,
	},
	{
		text.CommandUpper() + `_SSLMODE, SSLMODE`,
		`when set to 'retry', allows connections to attempt to reconnect when no ?sslmode= was specified on the url`,
	},
	{
		`TERM_GRAPHICS`,
		`use the specified terminal graphics`,
	},
	{
		`SHELL`,
		`shell used by the \! command`,
	},
}

const variableTpl = `List of specially treated variables

%s script variables, for scripts and queries:
Usage:
  %[1]s --set=NAME=VALUE
  or \set NAME VALUE inside %[1]s

%[2]s

Settings of %[1]s:
Usage:
  %[1]s --uset=NAME=VALUE
  or \uset NAME VALUE inside %[1]s

%[7]s

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
