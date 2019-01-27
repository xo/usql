package text

import (
	"regexp"
	"strings"
)

// Various usql text bits.
var (
	CommandName = `usql`

	CommandVersion = `0.0.0-dev`

	Banner = `the universal command-line interface for SQL databases`

	NotConnected = `(not connected)`

	HelpPrefix = `help`

	WelcomeDesc = `Type "` + HelpPrefix + `" for help.`

	QueryBufferEmpty = `Query buffer is empty.`

	QueryBufferReset = `Query buffer reset (cleared).`

	InvalidCommand = `Invalid command \%s. Try \? for help.`

	ExtraArgumentIgnored = `\%s: extra argument %q ignored`

	MissingRequiredArg = `\%s: missing required argument`

	Copyright = CommandName + ", " + Banner + ".\n\n" + License

	RowCount = `(%d rows)`

	AvailableDrivers = `Available Drivers:`

	ConnInfo = `Connected with driver %s (%s)`

	BadPassFile = `could not open %q, not a file`

	BadPassFileMode = `password file %q has group or world access`

	BadPassFileLine = `line %d of password file incorrectly formatted`

	BadPassFileFieldEmpty = `line %d field %d of password file cannot be blank`

	BadPassFileUsername = `username in line %d of password file cannot contain *`

	EnterPassword = `Enter password: `

	EnterPreviousPassword = `Enter previous password: `

	PasswordsDoNotMatch = `Passwords do not match, trying again ...`

	NewPassword = `Enter new password: `

	ConfirmPassword = `Confirm password: `

	PasswordChangeFailed = `\password for %q failed: %v`

	CouldNotSetVariable = `could not set variable %q`

	//PasswordChangeSucceeded = `\password succeeded for %q`

	HelpDesc string

	HelpBanner = `You are using ` + CommandName + ", " + Banner + `.`

	HelpCommandPrefix = `Type:  `

	HelpCommands = [][]string{
		{`copyright`, `for distribution terms`},
		//[]string{`h`, `for help with SQL commands`},
		{`?`, `for help with ` + CommandName + ` commands`},
		{`g`, `or terminate with semicolon to execute query`},
		{`q`, `to quit`},
	}

	UnknownFormatFieldName = `unknown option: %s`

	FormatFieldInvalidValue = `unrecognized value %q for %q: %s expected`

	FormatFieldNameSetMap = map[string]string{
		`border`:                   `Border style is %d.`,
		`columns`:                  `Target width is %d.`,
		`expanded`:                 `Expanded display is %s.`,
		`expanded_auto`:            `Expanded display is used automatically.`,
		`fieldsep`:                 `Field separator is %q.`,
		`fieldsep_zero`:            `Field separator is zero byte.`,
		`footer`:                   `Default footer is %s.`,
		`format`:                   `Output format is %s.`,
		`linestyle`:                `Line style is %s.`,
		`null`:                     `Null display is %q.`,
		`numericlocale`:            `Locale-adjusted numeric output is %s.`,
		`pager`:                    `Pager usage is %s.`,
		`pager_min_lines`:          `Pager won't be used for less than %d line(s).`,
		`recordsep`:                `Field separator is %q.`,
		`recordsep_zero`:           `Record separator is zero byte.`,
		`tableattr`:                `Table attributes are %q.`,
		`title`:                    `Title is %q.`,
		`tuples_only`:              `Tuples only is %s.`,
		`unicode_border_linestyle`: `Unicode border line style is %q.`,
		`unicode_column_linestyle`: `Unicode column line style is %q.`,
		`unicode_header_linestyle`: `Unicode header line style is %q.`,
	}

	FormatFieldNameUnsetMap = map[string]string{
		`tableattr`: `Table attributes unset.`,
		`title`:     `Title is unset.`,
	}
)

func init() {
	// setup help description
	cmds := make([]string, len(HelpCommands))
	for i, h := range HelpCommands {
		cmds[i] = `\` + h[0] + " " + h[1]
	}

	HelpDesc = HelpBanner +
		"\n" + HelpCommandPrefix +
		strings.Join(cmds, "\n"+strings.Repeat(" ", len(HelpCommandPrefix)))
}

var spaceRE = regexp.MustCompile(`\s+`)

// Command returns the command name without spaces.
var Command = func() string {
	return spaceRE.ReplaceAllString(CommandName, "")
}

// CommandLower returns the lower case command name without spaces.
var CommandLower = func() string {
	return strings.ToLower(Command())
}

// CommandUpper returns the upper case command name without spaces.
var CommandUpper = func() string {
	return strings.ToUpper(Command())
}

// UsageTemplate returns the usage template.
var UsageTemplate = func() string {
	n := CommandLower()

	return n + `, ` + Banner + `

Usage:
  ` + n + ` [OPTIONS]... [DSN]

Arguments:
  DSN                            database url

{{if .Context.Flags}}\
Options:
{{.Context.Flags|FlagsToTwoColumns|FormatTwoColumns}}{{end}}`
}
