package handler

import (
	"fmt"
	"strings"

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

	helpPrefix = "help"

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

// writeErr writes an error to stderr when err is not nil.
func writeErr(l *readline.Instance, err error, prefixes ...string) {
	if err != nil {
		fmt.Fprintf(l.Stderr(), "error: %s%v\n", strings.Join(prefixes, ""), err)
	}
}

// notImpl is a simple helper for not yet implemented commands.
func notImpl(l *readline.Instance, cmd string) {
	fmt.Fprintf(l.Stderr(), "COMMAND `\\%s` IS NOT YET IMPLEMENTED.\n", cmd)
}
