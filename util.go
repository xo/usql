package main

const (
	aboutDesc = `usql is the universal command-line interface for SQL databases.
`

	welcomeDesc = `Type "help" for help.

`

	helpDesc = `You are using usql, the universal command-line interface for SQL databases.
Type: \c[onnect] <url>  connect to url
      \q                quit
`
)

// Args are the command line arguments.
type Args struct {
	DSN           string   `arg:"positional,help:database url"`
	Commands      []string `arg:"-c,--command,separate,help:run only single command (SQL or internal) and exit"`
	DisablePretty bool     `arg:"-p,--disable-pretty,help:disable pretty formatting"`
	NoRC          bool     `arg:"-X,--disable-rc,help:do not read start up file"`
	File          string   `arg:"-f,--file,help:execute commands from file then exit"`
	Out           string   `arg:"-o,--output,help:output file"`
	Username      string   `arg:"-U,--username,help:database user name"`
	HistoryFile   string   `arg:"--hist-file,env:USQL_HISTFILE,help:history file"`
}

// Description provides the go-arg description.
func (a *Args) Description() string {
	return aboutDesc
}
