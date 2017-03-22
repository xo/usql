package text

var (
	CommandName = `usql`

	CommandVersion = `0.0.0-dev`

	Banner = CommandName + `, the universal command-line interface for SQL databases`

	NotConnected = `(not connected)`

	HelpPrefix = "help"

	WelcomeDesc = `Type "` + HelpPrefix + `" for help.`

	QueryBufferEmpty = `Query buffer is empty.`

	QueryBufferReset = `Query buffer reset (cleared).`

	InvalidCommand = `Invalid command \%s. Try \? for help.`

	ExtraArgumentIgnored = `\%s: extra argument "%s" ignored`

	Copyright = Banner + ".\n\n" + License

	MissingRequiredArg = "missing required argument"

	HelpDesc = `You are using ` + Banner + `.
Type: \copyright        distribution terms
      \c[onnect] <url>  connect to url
      \q                quit
      \Z                disconnect
	  \buildinfo        display information about which databases usql supports, depending on build
`

	RowCount = "(%d rows)"

	AvailableDrivers = `Available Drivers:`

	ConnInfo = `You are connected with driver %s (%s)`
)
