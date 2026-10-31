package readme

// Examples are the examples of the README, in the order they appear.
var Examples = []Example{
	{
		Name: "drivers",
		Tags: "no_base cassandra moderncsqlite mysql",
		Lines: []string{
			"$ cd $GOPATH/src/github.com/xo/usql",
			"",
			"# build excluding the base drivers, and including cassandra, moderncsqlite and mysql",
			"$ go build -tags 'no_base cassandra moderncsqlite mysql'",
			"",
			"# show built driver support",
		},
		Prog: "./usql",
		Args: []string{"-c", `\drivers`},
	},
	{
		Name: "help",
		Args: []string{"--help"},
	},
	{
		Name:    "queries",
		DB:      true,
		Command: true,
		Args:    []string{"{url}"},
		Input: []string{
			`select * from authors;`,
			`select title from books`,
			`where author_id = 2`,
			`\p`,
			`\g`,
		},
	},
	{
		Name:    "commands",
		Command: true,
		Input:   []string{`\?`},
	},
	{
		Name:  "set-unset",
		Input: []string{`\unset FOO`, `\set FOO bar`},
	},
	{
		Name:  "set-show",
		Setup: []string{`\set FOO bar`},
		Input: []string{`\set`},
	},
	{
		Name:  "interpolate",
		DB:    true,
		Args:  []string{"{url}"},
		Input: []string{`\set FOO 'Stephen King'`, `select * from authors where name = :'FOO';`},
	},
	{
		Name:  "interpolate-quoted",
		DB:    true,
		Args:  []string{"{url}"},
		Input: interpolateQuoted,
	},
	{
		Name:  "interpolate-print",
		DB:    true,
		Args:  []string{"{url}"},
		Setup: interpolateQuoted,
		Input: []string{`\p`, `\raw`},
	},
	{
		Name:  "not-interpolated",
		DB:    true,
		Args:  []string{"{url}"},
		Setup: []string{`\set FOO bar`},
		Input: []string{`select ':FOO' as foo;`, `\p`},
	},
	{
		Name:  "cset-connect",
		DB:    true,
		Input: []string{`\cset my_conn {url}`, `\c my_conn`},
	},
	{
		Name:  "cset-show",
		Setup: []string{`\cset my_conn postgres://user:pass@localhost`},
		Input: []string{`\cset`},
	},
	{
		Name:  "pset-set",
		Input: []string{`\pset time Kitchen`, `\a`},
	},
	{
		Name:  "pset-show",
		Setup: []string{`\pset time Kitchen`},
		Input: []string{`\pset`},
	},
	{
		Name:  "uset",
		Input: []string{`\uset pager "less -R"`, `\uset on_error_stop on`, `\uset`},
	},
	{
		Name:  "backticks-echo",
		Input: []string{"\\echo Welcome `echo $USER` -- 'running on:' \"(\" `uname -s` \")\""},
	},
	{
		Name:  "backticks-set",
		Input: []string{"\\set MYVAR `uname -s`", `\set`, `\echo :MYVAR`},
	},
	{
		Name: "copy-csv",
		Files: [][2]string{
			{"authors.csv", "author_id,name\n1,Isaac Asimov\n2,Stephen King\n"},
			{"books.csv", "book_id,author_id,title\n1,1,I Robot\n2,2,Carrie\n3,2,Cujo\n"},
		},
		Command: true,
		Input: []string{
			`-- setting variables to make connections easier`,
			`\set SOURCE_DSN csvq://.`,
			`\set DESTINATION_DSN sqlite3:booktest.db`,
			`-- connecting to the destination and creating the schema`,
			`\c :DESTINATION_DSN`,
			`create table authors (author_id integer, name text);`,
			`create table books (book_id integer not null primary key autoincrement, author_id integer, title text);`,
			`-- adding an extra row to books prior to copying`,
			`insert into books (author_id, title) values (1, 'Foundation');`,
			`-- disconnecting to demonstrate that \copy opens new database connections`,
			`\disconnect`,
			`-- copying data from SOURCE -> DESTINATION`,
			`\copy :SOURCE_DSN :DESTINATION_DSN 'select * from authors' authors`,
			`\copy :SOURCE_DSN :DESTINATION_DSN 'select author_id, title from books' 'books(author_id, title)'`,
			`\c :DESTINATION_DSN`,
			`select * from authors;`,
			`select * from books;`,
		},
	},
	{
		Name:    "syntax-highlighting",
		Command: true,
		Input:   []string{`\uset syntax_hl_style dracula`, `\uset syntax_hl off`},
	},
	{
		Name:    "completion-select",
		DB:      true,
		Command: true,
		Args:    []string{"{url}"},
		Input:   []string{"select * f\t"},
	},
	{
		Name:    "completion-meta",
		DB:      true,
		Command: true,
		Args:    []string{"{url}"},
		Input:   []string{"\\g\t"},
	},
	{
		Name:    "time",
		DB:      true,
		Command: true,
		Args:    []string{"{url}"},
		Input:   timeInput("select timestamp '2026-10-31 09:41:27.123456' as stamp;"),
		Products: map[string][]string{
			"sqlserver": timeInput("select cast('2026-10-31 09:41:27.123456' as datetime2) as stamp;"),
			"oracle":    timeInput("select timestamp '2026-10-31 09:41:27.123456' as stamp from dual;"),
		},
	},
	{
		Name:    "host-information-env",
		DB:      true,
		Command: true,
		Lines:   []string{"$ export USQL_SHOW_HOST_INFORMATION=false"},
		Env:     []string{"USQL_SHOW_HOST_INFORMATION=false"},
		Args:    []string{"{url}"},
	},
	{
		Name:    "host-information-uset",
		DB:      true,
		Command: true,
		Args:    []string{"--uset", "show_host_information=false", "{url}"},
		Input:   []string{`\uset show_host_information true`, `\connect {url}`},
	},
}

// interpolateQuoted sets the variables of a query that names its table and
// column with them, and leaves the query in the buffer.
var interpolateQuoted = []string{
	`\set TBLNAME authors`,
	`\set COLNAME name`,
	`\set FOO 'Isaac Asimov'`,
	`select * from :TBLNAME where :"COLNAME" = :'FOO'`,
}

// timeInput runs query, which selects a timestamp, before and after \pset
// time changes its format.
func timeInput(query string) []string {
	return []string{`\pset time`, query, `\pset time Kitchen`, query}
}
