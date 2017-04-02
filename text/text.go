package text

import (
	"regexp"
	"strings"
)

var (
	CommandName = `usql`

	CommandVersion = `0.0.0-dev`

	Banner = CommandName + `, the universal command-line interface for SQL databases`

	NotConnected = `(not connected)`

	HelpPrefix = `help`

	WelcomeDesc = `Type "` + HelpPrefix + `" for help.`

	QueryBufferEmpty = `Query buffer is empty.`

	QueryBufferReset = `Query buffer reset (cleared).`

	InvalidCommand = `Invalid command \%s. Try \? for help.`

	ExtraArgumentIgnored = `\%s: extra argument "%s" ignored`

	MissingRequiredArg = `\%s: missing required argument`

	Copyright = Banner + ".\n\n" + License

	RowCount = `(%d rows)`

	AvailableDrivers = `Available Drivers:`

	ConnInfo = `Connected with driver %s (%s)`

	BadPassFile = `could not open "%s", not a file`

	BadPassFileMode = `password file "%s" has group or world access`

	BadPassFileLine = `line %d of password file incorrectly formatted`

	BadPassFileFieldEmpty = `line %d field %d of password file cannot be blank`

	BadPassFileUsername = `username in line %d of password file cannot contain *`

	EnterPassword = `Enter password: `

	EnterPreviousPassword = `Enter previous password: `

	PasswordsDoNotMatch = `Passwords do not match, trying again ...`

	NewPassword = `Enter new password: `

	ConfirmPassword = `Confirm password: `

	PasswordChangeFailed = `\password for "%s" failed: %v`

	CouldNotSetVariable = `could not set variable "%s"`

	//PasswordChangeSucceeded = `\password succeeded for "%s"`

	HelpDesc string

	HelpBanner = `You are using ` + Banner + `.`

	HelpCommandPrefix = `Type:  `

	HelpCommands = [][]string{
		[]string{`copyright`, `for distribution terms`},
		[]string{`h`, `for help with SQL commands`},
		[]string{`?`, `for help with ` + CommandName + ` commands`},
		[]string{`g`, `or terminate with semicolon to execute query`},
		[]string{`q`, `to quit`},
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
func Command() string {
	return spaceRE.ReplaceAllString(CommandName, "")
}

// CommandLower returns the lower case command name without spaces.
func CommandLower() string {
	return strings.ToLower(Command())
}

// CommandUpper returns the upper case command name without spaces.
func CommandUpper() string {
	return strings.ToUpper(Command())
}
