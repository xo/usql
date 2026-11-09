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

// Listing writes the settings of the configuration to w, by their full
// names, one section at a time, and then the variables of the environment.
func Listing(w io.Writer) error {
	// determine config dir name
	configDir, configExtra := buildConfigDir("config.yaml")
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
	if configExtra != "" {
		configExtra = " (" + configExtra + ")"
	}
	fmt.Fprintf(
		w,
		variableTpl,
		text.CommandName,
		configDir,
		configExtra,
		names(settingNames),
		names(pvarNames),
		names(varNames),
		names(chartNames),
		names(append(ev, envVarNames...)),
	)
	return nil
}

// names returns the names and the descriptions of a section, without the
// section, as psql lists its variables.
func names(v []varName) string {
	var b strings.Builder
	for _, n := range v {
		b.WriteString(n.String())
	}
	return strings.TrimRightFunc(b.String(), unicode.IsSpace)
}

func buildConfigDir(configName string) (string, string) {
	dir := `$HOME/.config/usql`
	switch runtime.GOOS {
	case "darwin":
		dir = `$HOME/Library/Application Support/usql`
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
		`number of rows returned or affected by the last query, or 0 if it failed`,
	},
}

// chartNames are the settings of \chart.
var chartNames = []varName{
	{
		`path`,
		`directory of the chart files, relative to the directory of the configuration file (default "charts")`,
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
		`fetch_count`,
		`the number of result rows to read and show at a time, 0 for all of them at once, as psql's FETCH_COUNT (default 0)`,
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
		`encoding`,
		`encoding of a script read with -f, \i or from a pipe, such as windows-1252 or shift_jis (default utf-8)`,
	},
	{
		`sslmode`,
		`when set to 'retry', allows connections to attempt to reconnect when no ?sslmode= was specified on the url`,
	},
	{
		`syntax_hl`,
		`enable syntax highlighting of the input`,
	},
	{
		`syntax_hl_format`,
		`color depth of the highlighting [noop, terminal, terminal256, terminal16m]`,
	},
	{
		`syntax_hl_override_bg`,
		`keep the background of the terminal, and not the background of the highlighting style`,
	},
	{
		`syntax_hl_style`,
		`syntax highlighting style name`,
	},
	{
		`term_graphics`,
		`use the specified terminal graphics`,
	},
}

var (
	formatRE    = regexp.MustCompile(`^(unaligned|aligned|wrapped|html|asciidoc|latex|latex-longtable|troff-ms|csv|json|vertical|markdown)$`)
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
		`set output format [unaligned, aligned, wrapped, vertical, html, asciidoc, csv, json, markdown, ...]`,
	},
	{
		`header_align`,
		`alignment of the column names in aligned, wrapped and markdown output [center, left, right, column] (default center)`,
	},
	{
		`linestyle`,
		`set the border line drawing style [ascii, old-ascii, unicode]`,
	},
	{
		`json_indent`,
		`number of spaces to indent JSON output, 0 for compact output [N, on, off] (default 0)`,
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
		`number of lines the output must have before the pager starts, 0 for no minimum (default 0)`,
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
	{
		`xheader_width`,
		`maximum width of the header of a record in expanded output [full, column, page, N] (default full)`,
	},
}

var envVarNames = []varName{
	{
		text.CommandUpper() + `_DSN`,
		`DSN to open when the command line names none, removed from the environment after it is read`,
	},
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
		`path of the history file (default the history file in the state directory)`,
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
		`path of the file that runs at start (default init.sql in the config directory)`,
	},
	{
		text.CommandUpper() + `_SSLMODE, SSLMODE`,
		`when set to 'retry', allows connections to attempt to reconnect when no ?sslmode= was specified on the url`,
	},
	{
		text.CommandUpper() + `_TERM_GRAPHICS, TERM_GRAPHICS`,
		`use the specified terminal graphics`,
	},
	{
		text.CommandUpper() + `_SHELL, SHELL`,
		`shell that runs the \! command, a backtick, \g |command, \o |command, the pager and the editor`,
	},
}

const variableTpl = `List of specially treated variables

Each setting belongs to a section. \config and -O name a setting as
SECTION.NAME, as in print.border. The configuration file holds each section
as a map of names and values:
  %[2]s%[3]s

%[1]s script variables (section vars):
Usage:
  %[1]s --set=NAME=VALUE
  or \set NAME VALUE inside %[1]s

%[6]s

Settings of %[1]s (section usql):
Usage:
  %[1]s --uset=NAME=VALUE
  or \uset NAME VALUE inside %[1]s

%[4]s

Display settings (section print):
Usage:
  %[1]s --pset=NAME[=VALUE]
  or \pset NAME [VALUE] inside %[1]s

%[5]s

Chart settings (section charts):
Usage:
  %[1]s --option=charts.NAME=VALUE
  or \config charts.NAME VALUE inside %[1]s

%[7]s

Environment of the commands that %[1]s runs, such as \!, the pager and the editor (section env):
Usage:
  %[1]s --option=env.NAME=VALUE
  or \setenv NAME [VALUE] inside %[1]s

Named connections (section connections):
Usage:
  %[1]s --cset=NAME=DSN
  or \cset NAME [DSN] inside %[1]s
  or \cset NAME DRIVER PARAMS... inside %[1]s

Aliases, the meta commands of the user (section aliases):
Usage:
  NAME: TEXT under aliases: in %[2]s, which \NAME runs

Environment variables:
Usage:
  NAME=VALUE [NAME=VALUE] %[1]s ...

%[8]s
`
