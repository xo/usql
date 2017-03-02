package main

const (
	cliDesc = `usql is a universal command-line interface for SQL databases.`

	helpDesc = `You are using usql, the universal command-line interface for SQL databases.
Type:
	\c <url>  connect to url
	\q        quit
`
)

// Args are the command line arguments.
type Args struct {
	DisablePretty     bool     `arg:"--disable-pretty,-p,help:disable pretty formatting"`
	HistoryFile       string   `arg:"--history-file,env:USQL_HISTFILE,help:history file"`
	UserHistoryPrefix string   `arg:"--user-history-prefix,env:USQL_USERHISTPREFIX,help:user history prefix to use"`
	Commands          []string `arg:"--command,-c,help:command to execute"`
	DSN               string   `arg:"positional,help:database url"`
}

// Description provides the go-arg description.
func (a *Args) Description() string {
	return cliDesc + "\n"
}
