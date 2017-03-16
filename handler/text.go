package handler

import (
	"fmt"

	"github.com/chzyer/readline"
)

const (
	notConnected = "(not connected)"

	tag = `usql, the universal command-line interface for SQL databases.`

	welcomeDesc = "Type \"help\" for help.\n\n"

	queryBufferEmpty = "Query buffer is empty."

	queryBufferReset = "Query buffer reset (cleared).\n"

	invalidCommand = "Invalid command \\%s. Try \\? for help.\n"

	extraArgumentIgnored = "\\%s: extra argument \"%s\" ignored\n"

	copyright = tag + "\n\n" + license + "\n\n"

	missingRequiredArg = "missing required argument"

	helpDesc = `You are using ` + tag + `
Type: \copyright        distribution terms
      \c[onnect] <url>  connect to url
      \q                quit
      \Z                disconnect
`
)

// cmdErr is a util func to simply write a "\cmd: msg" style error.
func cmdErr(l *readline.Instance, cmd, msg string) (int, error) {
	return fmt.Fprintf(l.Stderr(), "\\%s: %s\n", cmd, msg)
}
