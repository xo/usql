// Package text contains the text (and eventually translations) for the usql
// application.
package text

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"regexp"
	"strings"
)

// Various usql text bits.
var (
	CommandName           = `usql`
	CommandVersion        = `0.0.0-dev`
	PassfileName          = CommandName + `pass`
	Banner                = `the universal command-line interface for SQL databases`
	NotConnected          = `(not connected)`
	HelpPrefix            = `help`
	QuitPrefix            = `quit`
	ExitPrefix            = `exit`
	WelcomeDesc           = `Type "` + HelpPrefix + `" for help.`
	QueryBufferEmpty      = `Query buffer is empty.`
	QueryBufferReset      = `Query buffer reset (cleared).`
	InvalidCommand        = `Invalid command \%s. Try \? for help.`
	ExtraArgumentIgnored  = `\%s: extra argument %q ignored`
	MissingRequiredArg    = `\%s: missing required argument`
	Copyright             = CommandName + ", " + Banner + ".\n\n" + License
	RowCount              = `(%d rows)`
	AvailableDrivers      = `Available Drivers:`
	ConnInfo              = `Connected with driver %s (%s)`
	EnterPassword         = `Enter password: `
	EnterPreviousPassword = `Enter previous password: `
	PasswordsDoNotMatch   = `Passwords do not match, trying again ...`
	NewPassword           = `Enter new password: `
	ConfirmPassword       = `Confirm password: `
	PasswordChangeFailed  = `\password for %q failed: %v`
	CouldNotSetVariable   = `could not set variable %q`
	// PasswordChangeSucceeded = `\password succeeded for %q`
	HelpDesc          string
	HelpDescShort     = `Use \? for help or press control-C to clear the input buffer.`
	HelpBanner        = `You are using ` + CommandName + ", " + Banner + `.`
	HelpCommandPrefix = `Type:  `
	HelpCommands      = [][]string{
		{`copyright`, `for distribution terms`},
		//{`h`, `for help with SQL commands`},
		{`?`, `for help with ` + CommandName + ` commands`},
		{`g`, `or terminate with semicolon to execute query`},
		{`q`, `to quit`},
	}
	QuitDesc                = `Use \q to quit.`
	UnknownFormatFieldName  = `unknown option: %s`
	FormatFieldInvalid      = `unrecognized value %q for "%s"`
	FormatFieldInvalidValue = `unrecognized value %q for "%s": %s expected`
	FormatFieldNameSetMap   = map[string]string{
		`border`:                   `Border style is %d.`,
		`columns`:                  `Target width is %d.`,
		`expanded`:                 `Expanded display is %s.`,
		`expanded_auto`:            `Expanded display is used automatically.`,
		`fieldsep`:                 `Field separator is %q.`,
		`fieldsep_zero`:            `Field separator is zero byte.`,
		`footer`:                   `Default footer is %s.`,
		`format`:                   `Output format is %s.`,
		`linestyle`:                `Line style is %s.`,
		`locale`:                   `Locale is %q.`,
		`null`:                     `Null display is %q.`,
		`numericlocale`:            `Locale-adjusted numeric output is %s.`,
		`pager`:                    `Pager usage is %s.`,
		`pager_min_lines`:          `Pager won't be used for less than %d line(s).`,
		`recordsep`:                `Field separator is %q.`,
		`recordsep_zero`:           `Record separator is zero byte.`,
		`tableattr`:                `Table attributes are %q.`,
		`time`:                     `Time display is %s.`,
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
	TimingSet            = `Timing is %s.`
	TimingDesc           = `Time: %0.3f ms`
	InvalidValue         = `invalid -%s value %q: %s`
	NotSupportedByDriver = `%s not supported by %s driver`
	RelationNotFound     = `Did not find any relation named "%s".`
	InvalidOption        = `invalid option %q`
	NotificationReceived = `Asynchronous notification %q %sreceived from server process with PID %d.`
	NotificationPayload  = `with payload %q `
	UnknownShortAlias    = `(unk)`
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

// Logo is the logo.
var Logo image.Image

// LogoPng is the embedded logo.
//
//go:embed logo.png
var LogoPng []byte

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

func init() {
	var err error
	if Logo, err = png.Decode(bytes.NewReader(LogoPng)); err != nil {
		panic(err)
	}
}
