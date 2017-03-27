package text

import "strings"

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

	ConnInfo = `You are connected with driver %s (%s)`

	EnterPassword = `Enter Password`

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
