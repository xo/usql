package metacmd

// Code generated by gen.go. DO NOT EDIT.

import (
	"github.com/xo/usql/text"
)

// sections are the command description sections.
var sections = []string{
	"General",
	"Help",
	"Connection",
	"Query Execute",
	"Query View",
	"Query Buffer",
	"Informational",
	"Variables",
	"Input/Output",
	"Control/Conditional",
	"Transaction",
	"Operating System/Environment",
}

// descs are the command descriptions.
var descs [][]desc

// cmds are the command lookup map.
var cmds map[string]func(*Params) error

func init() {
	descs = [][]desc{
		// General
		{
			{Quit, `q`, ``, `quit ` + text.CommandName + ``, false, false},
			{Quit, `quit`, ``, `alias for \q`, true, false},
			{Copyright, `copyright`, ``, `show usage and distribution terms for ` + text.CommandName + ``, false, false},
			{Drivers, `drivers`, ``, `show database drivers available to ` + text.CommandName + ``, false, false},
		},
		// Help
		{
			{Help, `?`, `[commands]`, `show help on ` + text.CommandName + `'s meta (backslash) commands`, false, false},
			{Help, `?`, `options`, `show help on ` + text.CommandName + ` command-line options`, false, false},
			{Help, `?`, `variables`, `show help on special ` + text.CommandName + ` variables`, false, false},
		},
		// Connection
		{
			{Connect, `c`, `DSN or \c NAME`, `connect to dsn or named database connection`, false, false},
			{Connect, `c`, `DRIVER PARAMS...`, `connect to database with driver and parameters`, false, false},
			{Connect, `connect`, ``, `alias for \c`, true, false},
			{Disconnect, `Z`, ``, `close (disconnect) database connection`, false, false},
			{Disconnect, `disconnect`, ``, `alias for \Z`, true, false},
			{Password, `password`, `[USER]`, `change password for user`, false, false},
			{Password, `passwd`, ``, `alias for \password`, true, false},
			{ConnectionInfo, `conninfo`, ``, `display information about the current database connection`, false, false},
		},
		// Query Execute
		{
			{Execute, `g`, `[(OPTIONS)] [FILE] or ;`, `execute query (and send results to file or |pipe)`, false, false},
			{Execute, `go`, ``, `alias for \g`, true, false},
			{Execute, `G`, `[(OPTIONS)] [FILE]`, `as \g, but forces vertical output mode`, false, false},
			{Execute, `ego`, ``, `alias for \G`, true, false},
			{Execute, `gx`, `[(OPTIONS)] [FILE]`, `as \g, but forces expanded output mode`, false, false},
			{Execute, `gexec`, ``, `execute query and execute each value of the result`, false, false},
			{Execute, `gset`, `[PREFIX]`, `execute query and store results in ` + text.CommandName + ` variables`, false, false},
			{Bind, `bind`, `[PARAM]...`, `set query parameters`, false, false},
			{Timing, `timing`, `[on|off]`, `toggle timing of commands`, false, false},
		},
		// Query View
		{
			{Crosstab, `crosstab`, `[(OPTIONS)] [COLUMNS]`, `execute query and display results in crosstab`, false, false},
			{Crosstab, `crosstabview`, ``, `alias for \crosstab`, true, false},
			{Crosstab, `xtab`, ``, `alias for \crosstab`, true, false},
			{Chart, `chart`, `CHART [(OPTIONS)]`, `execute query and display results as a chart`, false, false},
			{Watch, `watch`, `[(OPTIONS)] [INTERVAL]`, `execute query every specified interval`, false, false},
		},
		// Query Buffer
		{
			{Edit, `e`, `[-raw|-exec] [FILE] [LINE]`, `edit the query buffer, raw (non-interpolated) buffer, the exec buffer, or a file with external editor`, false, false},
			{Edit, `edit`, ``, `alias for \e`, true, false},
			{Print, `p`, `[-raw|-exec]`, `show the contents of the query buffer, the raw (non-interpolated) buffer or the exec buffer`, false, false},
			{Print, `print`, ``, `alias for \p`, true, false},
			{Print, `raw`, ``, `alias for \p`, true, false},
			{Print, `exec`, ``, `alias for \p`, true, false},
			{Write, `w`, `[-raw|-exec] FILE`, `write the contents of the query buffer, raw (non-interpolated) buffer, or exec buffer to file`, false, false},
			{Write, `write`, ``, `alias for \w`, true, false},
			{Reset, `r`, ``, `reset (clear) the query buffer`, false, false},
			{Reset, `reset`, ``, `alias for \r`, true, false},
		},
		// Informational
		{
			{Describe, `d[S+]`, `[NAME]`, `list tables, views, and sequences or describe table, view, sequence, or index`, false, false},
			{Describe, `da[S+]`, `[PATTERN]`, `list aggregates`, false, false},
			{Describe, `df[S+]`, `[PATTERN]`, `list functions`, false, false},
			{Describe, `di[S+]`, `[PATTERN]`, `list indexes`, false, false},
			{Describe, `dm[S+]`, `[PATTERN]`, `list materialized views`, false, false},
			{Describe, `dn[S+]`, `[PATTERN]`, `list schemas`, false, false},
			{Describe, `dp[S]`, `[PATTERN]`, `list table, view, and sequence access privileges`, false, false},
			{Describe, `ds[S+]`, `[PATTERN]`, `list sequences`, false, false},
			{Describe, `dt[S+]`, `[PATTERN]`, `list tables`, false, false},
			{Describe, `dv[S+]`, `[PATTERN]`, `list views`, false, false},
			{Describe, `l[+]`, ``, `list databases`, false, false},
			{Stats, `ss[+]`, `[TABLE|QUERY] [k]`, `show stats for a table or a query`, false, false},
		},
		// Variables
		{
			{Set, `set`, `[NAME [VALUE]]`, `set ` + text.CommandName + ` application variable, or show all ` + text.CommandName + ` application variables if no parameters`, false, false},
			{Unset, `unset`, `NAME`, `unset (delete) ` + text.CommandName + ` application variable`, false, false},
			{SetPrint, `pset`, `[NAME [VALUE]]`, `set table print formatting option, or show all print formatting options if no parameters`, false, false},
			{SetPrint, `a`, ``, `toggle between unaligned and aligned output mode`, false, true},
			{SetPrint, `C`, `[TITLE]`, `set table title, or unset if none`, false, true},
			{SetPrint, `f`, `[SEPARATOR]`, `show or set field separator for unaligned query output`, false, true},
			{SetPrint, `H`, ``, `toggle HTML output mode`, false, true},
			{SetPrint, `T`, `[ATTRIBUTES]`, `set HTML <table> tag attributes, or unset if none`, false, true},
			{SetPrint, `t`, `[on|off]`, `show only rows`, false, true},
			{SetPrint, `x`, `[on|off|auto]`, `toggle expanded output`, false, true},
			{SetConn, `cset`, `[NAME [URL]]`, `set named connection, or show all named connections if no parameters`, false, false},
			{SetConn, `cset`, `NAME DRIVER PARAMS...`, `set named connection for driver and parameters`, false, false},
			{Prompt, `prompt`, `[-TYPE] VAR [PROMPT]`, `prompt user to set application variable`, false, false},
		},
		// Input/Output
		{
			{Echo, `echo`, `[-n] [MESSAGE]...`, `write message to standard output (-n for no newline)`, false, false},
			{Echo, `qecho`, `[-n] [MESSAGE]...`, `write message to \o output stream (-n for no newline)`, false, false},
			{Echo, `warn`, `[-n] [MESSAGE]...`, `write message to standard error (-n for no newline)`, false, false},
			{Out, `o`, `[FILE]`, `send all query results to file or |pipe`, false, false},
			{Out, `out`, ``, `alias for \o`, true, false},
			{Copy, `copy`, `SRC DST QUERY TABLE`, `copy results of query from source database into table on destination database`, false, false},
			{Copy, `copy`, `SRC DST QUERY TABLE(A,...)`, `copy results of query from source database into table's columns on destination database`, false, false},
		},
		// Control/Conditional
		{
			{Include, `i`, `FILE`, `execute commands from file`, false, false},
			{Include, `include`, ``, `alias for \i`, true, false},
			{Include, `ir`, `FILE`, `as \i, but relative to location of current script`, false, false},
			{Include, `include_relative`, ``, `alias for \ir`, true, false},
			{Conditional, `if`, `EXPR`, `begin conditional block`, false, false},
			{Conditional, `elif`, `EXPR`, `alternative within current conditional block`, false, false},
			{Conditional, `else`, ``, `final alternative within current conditional block`, false, false},
			{Conditional, `endif`, ``, `end conditional block`, false, false},
		},
		// Transaction
		{
			{Transact, `begin`, `[-read-only [ISOLATION]]`, `begin transaction, with optional isolation level`, false, false},
			{Transact, `commit`, ``, `commit current transaction`, false, false},
			{Transact, `rollback`, ``, `rollback (abort) current transaction`, false, false},
			{Transact, `abort`, ``, `alias for \rollback`, true, false},
		},
		// Operating System/Environment
		{
			{Shell, `!`, `[COMMAND]`, `execute command in shell or start interactive shell`, false, false},
			{Chdir, `cd`, `[DIR]`, `change the current working directory`, false, false},
			{Getenv, `getenv`, `VARNAME ENVVAR`, `fetch environment variable`, false, false},
			{Setenv, `setenv`, `NAME [VALUE]`, `set or unset environment variable`, false, false},
		},
	}
	cmds = make(map[string]func(*Params) error)
	for i := range sections {
		for _, desc := range descs[i] {
			for _, n := range desc.Names() {
				cmds[n] = desc.Func
			}
		}
	}
}