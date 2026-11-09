# The next major release

This file lists what changed in usql since v0.21.6, and what must be checked
by hand before the release. It is the largest set of changes that usql has
had. Automated tests cover much of it, and Ken checks the rest by hand.

Each check names the databases it covers. Tick a check off in a copy of this
file, not here. When a check finds a fault, record it in
[BACKLOG.md](BACKLOG.md) or fix it, and add a test where one can be written.

## Known gaps

Read these first. Some of them can stop the release.

- `%R` in a prompt shows only `=`, `@` or `-`, because rline draws one prompt
  for an entry. rline's backlog item 33 holds a prompt for each row.
- rline writes an escape code to a pipe when color is on. usql turns color off
  unless both input and output are a terminal.
- ODBC no longer reads `usql_trim`, csvq's `SHOW` prints nothing, and Athena
  and Cosmos no longer strip a trailing semicolon. W28 holds them.
- usql requires dbfwd v0.3.1, dbmeta v0.39.0, dburl v0.58.0, tblfmt v0.23.0, transit
  v0.7.5 and rline v0.8.0, all in `go.mod`. It builds with no `go.work`.
- `\copy` still writes its own SQL: the INSERT, and the query for the columns of
  the target. The placeholder of each database comes from dbmeta. W29 moves the
  rest, and is due before the release.

## What changed

### One build, with every driver

usql has one build. A `go build` or `go install` with no build tags includes
every driver and the chart renderer. A release build is that build, with the
features of SQLite that `build.sh` adds.

- `no_<tag>` leaves one driver out, as before. `no_charts` leaves out the chart
  renderer.
- `no_drivers` leaves out every driver and the chart renderer. A driver or
  `charts` named by its tag goes back in, so
  `go build -tags 'no_drivers sqlite3'` builds usql with only the sqlite3
  driver.
- The groups `base`, `most`, `all` and `bad` are gone, with the build tags
  `no_base`, `most`, `all` and `bad`, and the `Group:` line of a driver package.
  A build that used `-tags most` or `-tags all` now uses no tags. A build that
  used `-tags 'no_base x y'` now uses `-tags 'no_drivers x y'`.
- The charts renderer was only in a build with `-tags charts` or `-tags all`.
  It is now in every build, which adds about 12 MiB to the binary.
- The h2 and gizmosql drivers are removed. They were the `bad` group. dburl
  v0.56.0 drops their schemes, with the aliases `gz` and `gizmo`.
- A driver with no hooks has no package under `drivers`. `gen.go` writes its
  import and its registration into `internal/`, from `plainDrivers` in
  `gen.go`, and `\drivers` lists the same drivers as before (W52). Their
  packages are gone, such as `drivers/oracle`, `drivers/duckdb` and
  `drivers/odbc`.
- On Linux, the SQLite ODBC driver crashes in a build that links go-sqlite3
  and DuckDB. The build is not changed for it (W40). The README says to build
  with the tag `libsqlite3`, or with `-tags 'no_drivers odbc libsqlite3'`.
- When a URL names a driver that the build left out, usql prints
  `go install github.com/xo/usql@<version>`, with no tags. It printed
  `-tags 'most <driver>'`, and the name of the branch was `master`, where it is
  `main`. A scheme that usql has no driver for prints no hint.

### Input: rline and transit's usql grammar

usql reads its input through `github.com/xo/rline` and the usql grammar of
`github.com/xo/transit`. They replace gohxs/readline, usql's own `rline`
package and the `stmt` package (W36).

- The package `feed` parses input with the grammar. It holds psql's query
  buffer, says when an entry is finished, and feeds each statement and meta
  command to the handler.
- An entry can span several rows in the editor. Enter starts a new row until
  the entry ends with a `;` or a meta command.
- A variable is replaced when its statement runs, by its range in the
  input. A defined `:'x'` no longer opens a string, and the second colon of
  `::` is no longer a variable.
- The grammar is one language, and its scanner takes the lexical forms of
  the dialect as options (transit's D108): dollar quotes, block comments,
  `//` and `#` comments, and backticks. They come from dbmeta, and follow the
  dialect of the connection, so they change on `\c`.
- The `BEGIN ... END` body of a stored program stays in one statement, where
  usql sent each piece between semicolons as a statement of its own (W65).
  dbmeta v0.22.0 says which dialects have such bodies (its D249), among them
  SQLite triggers, MySQL, SQL Server and Oracle procedures, and PostgreSQL
  `BEGIN ATOMIC` functions. `BEGIN; ...; END;` is a transaction and still
  splits.
- On ClickHouse, a `;` inside a name in backticks, a `//`, `# ` or `#!`
  comment, or a `$$` or `$tag$` string no longer ends the statement (dbmeta
  v0.23.0 and v0.24.0, its D250 and D251).
- On SQL Server, SQLite, libSQL and rqlite, a name between brackets, such as
  `[n° d'entrée]`, is a quoted name (transit v0.7.4, its D140). A quote, a `;`
  or `--` inside the brackets no longer opens a string, ends the statement or
  starts a comment, and `]]` is a `]`. Discussion #600 reported a script
  that ran no statement after such a name.
- A U+FEFF, the byte order mark that editors on Windows write, is white space
  between statements and inside them (transit's D141). Inside a string or a
  quoted name it stays text. A script that held one anywhere but at its start
  sent it to the database.
- A backslash escapes the next character in a string between single quotes
  and in a name between double quotes only where the database reads it so:
  MySQL and its family, ClickHouse, Databend, Impala, Neo4j, Spanner and the
  others that dbmeta v0.32.0 marks with `BackslashEscapes` (its D266, and
  transit's D142). Everywhere else a backslash is text, as in psql, so
  `select 'C:\data\';` on PostgreSQL, SQL Server, Oracle or SQLite ends the
  statement, where usql waited for more input. `E'...'` keeps its escapes.
- A script given with `-f`, `\i`, `init.sql` or a pipe must be valid UTF-8,
  or usql stops with an error that names the file and the line. Before, usql
  read a script as bytes. A byte order mark at the start of the script decides
  its encoding: UTF-8, UTF-16 or UTF-32. Otherwise the new setting `encoding`
  names it, by the names of the WHATWG Encoding Standard, as in `\uset
  encoding windows-1252` or `--uset encoding=shift_jis`. A file takes the
  setting when it is opened, and a pipe when it is first read. usql decodes a
  script before it parses it, so the byte 0x5C inside a Shift JIS, Big5, GBK
  or GB18030 character is not a backslash.
- A line of many statements is parsed once. usql parsed the rest of the line
  again for each statement, so 4,000 statements on one line took 14 s and
  5 GiB, and now take 6 ms.
- Fuzz tests feed scripts through the parser in each dialect of dbmeta, with
  names and strings that hold quotes, separators, Unicode that is hard to see,
  CRLF line ends, a byte order mark, and the encodings of Windows, ISO 8859,
  KOI8-R, EUC-JP, EUC-KR, Shift JIS, Big5, GBK and GB18030. CONTRIBUTING.md
  says how to run them.
- The arguments of a meta command come from the grammar. `\g` reads its
  option list, such as `(format=csv)`, from the grammar.
- `-o FILE` sends the results to the file, as `\o` does. `\echo` still goes
  to the terminal, as in psql.
- The completer is an rline completer.
- rline binds the keys of GNU readline 8.3, and has all of its
  commands, so any binding in an inputrc works. Ctrl-G aborts as in
  readline, Esc no longer clears the line, Up and Down walk the history and
  not the rows of an entry, and Ctrl-J accepts the line as Enter does. F1
  shows nothing.
- A terminal sends ESC [ I and ESC [ O when its window gains or loses focus.
  rline v0.7.1 rings the bell once for each and changes nothing, as readline
  does. It read them as Tab and F3, and the Tab
  opened the completion menu. This can be the cause of issues 93, 122, 483 and
  490, where keys act as if Ctrl were held after an alt-tab. That is not shown
  yet. rline v0.7.0 also reads Ctrl and a letter in the kitty keyboard
  protocol (CSI u) as the control character.
- Ctrl-C leaves the entry on the screen with `^C` after it, and starts the
  next prompt below, as psql does. The entry is not run and not kept in the
  history.
- Tab fills in what the completions share, and then opens a numbered menu of
  them, as before.
- A pasted script runs whole. rline v0.5.0 keeps the input that is not read
  yet when it sets the terminal, where it threw away all that came after the
  first line of a paste. It turns on bracketed paste, as readline 8.1 does, so
  a paste goes into the entry as rows, and one Enter runs it. `set
  enable-bracketed-paste off` in an inputrc file turns it off.
- Tab completes a meta command after a statement too, as in `select 1 \g` or
  `select 1\gx`, as psql does. It completed one only at the start of an
  entry.
- Tab offers every meta command, from the table of commands that `\?` prints,
  where a list written by hand left out many, such as `\sf`, `\sv`, `\ss`,
  `\gexec` and `\crosstabview`. A kind of object, as the `a` of `\dfa`, is
  offered alone, with each combination of `S` and `+`.
- usql reads the user's inputrc, as psql does through readline: `$INPUTRC`,
  or else `~/.inputrc`, or else `/etc/inputrc`. `$if usql` in the file tests
  the name of usql. The file can bind Tab to something else.
- `\p` prints the query buffer in color, and `\sv` and `\sf` print the
  definition in color, when the input is highlighted and the output is the
  terminal. A pager shows the colors only if it passes them through, as
  `less -R` does.
- Syntax highlighting parses the input with the grammars of transit, and
  colors it with the styles of transit, in place of Chroma. These have a
  grammar of their own, from transit v0.7.0: PostgreSQL, MySQL, SQL Server,
  Oracle, CQL, SQLite (with libSQL and rqlite), DuckDB (with GizmoSQL),
  ClickHouse, Trino (with Presto and Athena), Neo4j, SurrealDB and YDB. The
  other dialects use the generic SQL grammar. transit wrote the PostgreSQL,
  Oracle and SQL Server grammars again, and they are smaller (W32 has the
  sizes).
- A line with only `/` is statement text on every database, as in psql. It
  does not end a statement as SQL*Plus ends one (transit's D138). A script
  for Oracle ends each statement with `;`, and writes each `;` inside a
  package, or in the declarations of a unit before its `BEGIN`, as `\;`.
- Inside a `BEGIN ... END` body, a variable such as the `:new` of an Oracle
  trigger no longer ends the body at the next `;` (transit v0.7.2).
- Two dashes start a comment on every database except Neo4j and ArangoDB,
  where `(p)--()` and `RETURN 5--3` are statement text (dbmeta's D261,
  transit's D139).
- An anonymous block of PL/SQL on Oracle, `BEGIN ... END` or `DECLARE ...
  BEGIN ... END`, stays in one statement, and so do Firebird's `EXECUTE BLOCK`,
  SAP HANA's `DO BEGIN` and a `WITH FUNCTION` of Trino (W65).
- On Oracle, usql removes the `;` at the end of a statement, except after the
  `END` of a PL/SQL unit. It now also keeps the `;` after `END name;`, which
  ends a package and a package body. Write each `;` inside a package, and
  inside the declarations of a unit before its `BEGIN`, as `\;`, as in psql.

### Metadata: dbmeta and psql's describe commands

usql reads every database's metadata through `github.com/xo/dbmeta` (W21).
It has no metadata query of its own (D3, D4).

- The readers in usql are gone, for the information schema, PostgreSQL,
  Oracle, MySQL, Impala, SQLite, SQL Server, ClickHouse, Trino and DuckDB.
- The output follows psql, so it differs from the output of v0.21.6.
- psql's other describe commands are added (W31): `\dA` and its forms,
  `\db`, `\dc`, `\dconfig`, `\dC`, `\dd`, `\dD`, `\ddp`, `\dE`, `\des`, `\det`,
  `\deu`, `\dew`, `\df` with `anptw`, `\dF` and its forms, `\dg`, `\dl`,
  `\lo_list`, `\dL`, `\do`, `\dO`, `\dP`, `\drds`, `\drg`, `\dRp`, `\dRs`,
  `\dT`, `\du`, `\dx`, `\dX`, `\dy`, `\sf`, `\sv` and `\z`.
- A name with both S and +, such as `\dtS+`, works. `\d` takes its class
  letters in any order.
- A pattern folds as the database folds a name that is not quoted, and a
  name between double quotes keeps its case. A pattern also shows the
  objects of the system, as in psql.
- The package `metacmd/describe` formats the answers.
- `\ef [FUNCNAME [LINE]]` and `\ev [VIEWNAME [LINE]]` open the definition of
  a function or a view in the editor, as psql does. The edited statement
  waits in the query buffer for a `;` or `\g`. An unchanged file prints `No
  changes` and leaves the buffer empty. The LINE of `\ef` counts from the
  first line of the body, as `\sf+` numbers it. Each needs a name: psql's
  empty template for a new function is PostgreSQL's own syntax.
- `\sv` writes no `;` after the definition, as psql does. A view that does
  not exist is the error `relation "NAME" does not exist`.
- A table in the aligned and wrapped formats is followed by a blank line, as
  in psql. This holds for the describe commands and for a query in the
  wrapped format, which had none.
- `\d NAME` lists the not-null constraints, the definition of a view and
  the constraint of a partition only with `+`, as psql 18 does.
- More of the describe output matches psql 18 on PostgreSQL:
  - A list that finds nothing names its kind, as in `Did not find any
    tables named "x".`, and `\drds` and `\dRp+` print psql's line.
  - `\di` is one more class of `\d`, so `\dti` and the other combinations
    list indexes with their table, as in psql.
  - `\dPt` and `\dPi` have their own titles and no Type column, and the
    columns of `\dP` follow psql's letters.
  - `\du` and `\dg` print no count of rows, and `\dx+` prints one.
  - The numbers of a sequence align to the right.
  - "Referenced by" lists the foreign keys of every schema, ordered by name,
    and names a table outside the search path with its schema.
  - `\dAp` prints the name of a function alone, `\dC+` prints `no` for a
    cast with no function, and `\dRs+` has psql's columns with Conninfo.
  - A list with no rows keeps psql's columns.
  - `\dA` writes the type of an access method with a capital, as `Index`,
    and `\da` writes `*` for an aggregate with no arguments, as `count(*)`.
  - A list with no schema in its pattern shows only the objects on the
    search path, as psql does (dbmeta v0.31.0, its D262).
  - The lists keep the order of dbmeta's queries, which is psql's order in
    the collation of the database. usql sorted them again in byte order.
  - `\d INDEX` names the column of the index and its type, as `lower` and
    `text` for `lower(name)`. `\d+ INDEX` adds Storage and Stats target, a
    partitioned index is titled `Partitioned index` and lists its
    partitions, and a partition of an index names its parent.
  - The sequence of an identity column says `Sequence for identity column:`.
  - `\dd` lists the comments on constraints, operator classes and families,
    rules and triggers, as psql does, and not those on relations.
  - `\dconfig` with no pattern lists the settings that are not at their
    default or boot value, and `\dLS`, `\drg` and `\dFp+` follow psql.
  - The relation lists have no Rows column, and `\df+` has no Source code
    column, as in psql 18. `\sf` prints the source of a function.
- The editor gets the line before the file, as `+LINE FILE`, as psql passes
  it.
- The version in the banner, the current user, and the statement that
  changes a password come from dbmeta. InfluxDB shows its release from
  GET /ping.

### Settings: `\uset` and `\set`

usql's own settings left `\set` for the new `\uset`, which works as `\pset` does
(W61). `\set` is only for script variables, and its `:name` interpolation.

- The settings are echo_hidden, editor, on_error_stop, pager, prompt1, prompt2,
  quiet, show_host_information, sslmode, syntax_hl, syntax_hl_format,
  syntax_hl_override_bg, syntax_hl_style and term_graphics.
- prompt2 is the prompt of the rows after the first in an entry. Its default
  is the same as prompt1's, and its `%R` shows `-`, so a continued statement
  shows `pg:user@host/db-> ` as psql does.
- `\uset` shows every setting as a table, `\uset NAME` shows one, and `\uset NAME
  VALUE` sets it and says so. A value is read as an argument of any meta command.
- `\set pager less` and `\unset quiet` are errors that say to use `\uset`, so a
  script or an `init.sql` that set a setting with `\set` must change. `:pager` is no
  longer a variable.
- The command line has `--uset NAME=VALUE`, or `-U NAME=VALUE`, and `-q` is
  `--uset quiet=on`. `--set` and `-v` give script variables only. psql's `-U`
  is the user name, and usql's is not: usql reads the user from the URL.

### The pager

The pager moved out of tblfmt into the new package `pager` of usql (W64).
tblfmt v0.21.0 and later have no pager. All
the output of an entry goes through it: the results of a query, the describe
commands, `\?`, `\drivers`, `\echo`, and the listings of `\set`, `\uset`,
`\pset` and `\cset`.

- usql holds the output of an entry until it is as tall as the terminal or
  wider than it. Then it starts the pager, and gives it the held output and the
  rest. Output that fits goes to the terminal when the entry ends. Every format
  pages, CSV and JSON as well.
- The pager writes to the terminal, so less pages, takes its keys and draws its
  colors. Before, it wrote through the line editor and acted like cat.
- The shell runs the pager, so `PAGER="less -S"` works.
- `\pset pager off`, `on` and `always` and `\pset pager_min_lines` apply to all
  output. pager_min_lines is the number of lines the output must have before the
  pager starts, as in psql.
- The pager runs only when standard output is a terminal. `\o`, `\g FILE`,
  `\g |command` and `\watch` do not page.
- `\timing` prints the time after the pager exits, as psql does.
- The pager runs inside rline's `Session.Release`, and usql measures the
  terminal with `Session.Size`. Ctrl-C in the pager goes to the pager, and
  usql keeps running. Ctrl-Z stops both. The editor of `\e` and the shell of
  `\!` run the same way, so Ctrl-C in them no longer ends usql.
- An error goes after the output of its entry, not into the middle of it.

### The fields of a driver

What describes a database moved from `drivers.Driver` to dbmeta (D5):

- the lexical forms
- the old password that SQL Server needs
- CQL's batches, which the grammar of usql now keeps as one statement
  (`BEGIN BATCH`, `BEGIN UNLOGGED BATCH` and `BEGIN COUNTER BATCH`). A batch that
  is not finished when `\g` is typed goes to the server as it is, and fails there.
- what happens to a trailing semicolon
- the case of column names
- the placeholder of the INSERT of `\copy`, whether a write needs autocommit,
  whether every statement is a query, and the name of a product with no
  statement for its version
- scanning a column into the type that the driver reports (the MySQL family and
  Databend)

What describes the DSN moved to dburl: the defaults of the MySQL family
(`parseTime`, `loc` and `sql_mode`), SQLite (`loc`), Cassandra (`timeout`) and
Couchbase (`txtimeout`), which a URL can now override. A MySQL or SQLite URL that
writes `loc=UTC` keeps it.

`Process` is gone. The drivers that used it only to run every statement as a
query now come from dbmeta (`EveryStatementIsAQuery`). `BatchAsTransaction` and
`LexerName` are gone.

### Drivers

- postgres:// opens through pgx. lib/pq is the libpq driver. A build with
  libpq and no pgx gives postgres:// and its aliases to libpq, as a build
  with moderncsqlite and no sqlite3 does for sqlite3. The README lists pgx and
  postgres in one row, named PostgreSQL, and the package `drivers/postgres`
  is gone. The build tags `postgres` and `no_postgres` select pgx.
- `no_pgx` and `no_postgres` leave pgx out of the binary. Redshift,
  CockroachDB, CrateDB and QuestDB then use libpq, through the new package
  `drivers/pgwire`.
- A driver that is built on another driver goes when that driver goes:
  `no_mysql` leaves out memsql, tidb and vitess, and `no_influxdb` leaves out
  influxql. `gen.go` reads this from
  the imports of the driver package.
- PostgreSQL prints more as psql does, with both drivers (W63). float4 has
  the digits of a float4, oid, xid, cid and xid8 are aligned to the right,
  and libpq prints time and timetz as psql does and not as a timestamp.
- An aligned table measures a cell by grapheme cluster (tblfmt v0.19.2), so
  an emoji with a skin tone, a flag, a keycap or U+FE0F keeps its column
  aligned (W63).
- Every format writes what psql 18 writes (tblfmt v0.21.0). An aligned table
  has no space after its last column, a tab becomes spaces, a result with
  columns and no rows has its header, and expanded, html and asciidoc output
  match psql. `\pset format` takes wrapped, latex, latex-longtable and
  troff-ms.
- A binary value prints as psql prints a bytea, on every database: `\x` and
  lower case hex, and `\x` alone for an empty value. dbmeta v0.27.0 says
  which column types are binary (its D258), and tblfmt v0.21.1 writes the
  text in every format. usql printed a binary value with Go escapes before,
  as in `"\x15\x1C/\x00"`. An Oracle
  `RAW(16)` whose column name holds `id`, in any case, such as `staff_id`
  or `GUID`, is how Oracle stores a GUID, and prints as a lower case UUID,
  as a PostgreSQL uuid and a SQL Server uniqueidentifier print (dbmeta
  v0.30.1, its D263, with the column length from xo/oracle v0.3.0). Any
  other `RAW`, such as a hash, prints as `\x` and hex.
- A decimal that a driver returns with database/sql's Decompose method
  prints as psql prints a numeric (tblfmt v0.21.2): aligned to the right,
  with no exponent, and as a number in JSON. SAP HANA's DECIMAL failed to
  print before. The decimals of the dbimp drivers, such as ClickHouse,
  Trino and BigQuery, printed as left aligned text with an exponent, as in
  `1.5E+10`.
- A `float8`, and any 64-bit float, prints as psql prints it (tblfmt
  v0.22.1): with no exponent when its decimal exponent is from -4 to 14, so
  `1450817.032` where usql printed `1.450817032e+06`, and `1e+15` and
  `1e-05` otherwise. `numericlocale` groups a float as psql does, and
  expanded output follows `numericlocale`, `time` and `timezone`.
- `\pset format markdown`, or `-M` and `--markdown` on the command line as
  `-J` is for JSON, writes a table in markdown, with each column
  padded and its alignment in the delimiter row (tblfmt v0.22.0). It writes
  no title and no count of rows, and a blank line after each table, so that
  two tables stay apart.
- `\pset header_align` sets where the names of the columns go in the
  aligned, wrapped and markdown formats: `center` as psql (the default),
  `left`, `right`, or `column`, where each name goes to the side of the
  values of its column, as clickhouse-client does. psql has no such
  variable.
- JSON output is compact, one result set on each line, where it was indented
  before. `\pset json_indent N` indents it by N spaces, `on` is two spaces,
  `off` and 0 are compact, and `\pset json_indent` alone switches between
  compact and two spaces. `-P json_indent=2` sets it on the command line.
- A column is aligned by its database type, as psql aligns it (dbmeta
  v0.18.0, its D245). Numbers, money and numeric identifiers such as oid go
  to the right, and every other type goes to the left. A money column now
  goes to the right in PostgreSQL. dbmeta records the types of 34 dialects.
  For a type that dbmeta does not record, usql aligns the column by its
  values, as before. Athena, Cosmos, Databricks, MaxCompute, Redshift,
  Snowflake, Spanner, Tablestore, Impala and InfluxQL record no types yet.
- Cassandra moves to `github.com/xo/cassandra`, and its scheme is
  `cassandra`, with `cql` as an alias (W45).
- Couchbase and Databend move to `github.com/xo/dbimp`.
- BigQuery, Cosmos DB, Athena, Databricks, Spanner, Snowflake and Exasol move to
  `github.com/xo/dbimp` (W49, W50, W56). Athena's scheme is `athena`, with
  `awsathena`, `s3` and `aws` as aliases. The URLs of Athena, Cosmos DB,
  Databricks, Snowflake and Exasol changed: `athena://key:secret@host/db`,
  `cosmos://user:key@account/db/container`, `databricks://token:pat@workspace/warehouse`,
  `snowflake://user:key@host/database/schema?role=r` (the password is the private
  key as base64url PKCS8 DER) and `exasol://user:password@host:8563/schema`.
  Every earlier URL must change. Cosmos DB reads only. They ran through usql on
  the hosted services, or their emulators, on 2026-10-11 (W55); check each again
  by hand before the release: a query, a wrong credential, `\dt`, `\d` and `\copy`.
- New drivers:
  - CockroachDB, CrateDB, Neo4j, InfluxDB and InfluxQL
  - SurrealDB and ArangoDB
  - Pinot, rqlite and libSQL
  - Drill, Solr, Elasticsearch, OpenSearch and Druid, all read only
  - DynamoDB, which is dbimp's driver in place of godynamo, so the scheme
    `godynamo` goes and the URL needs a key and a secret (W43)
- Removed drivers:
  - mymysql, adodb, sapase and ignite
  - ql, Netezza, RamSQL and Dameng
  - godror, which needed cgo and Oracle's client library. Use the oracle
    driver, which is pure Go.
  - Db2's ODBC setup in `contrib`
- flightsql moves to arrow-go v18.
- Hive and Impala move to dbimp's hive driver, pure Go, which replaces gohive
  and impala-go. The URL is
  `hive://user:password@host:port/database?httpPath=cliservice`, and
  `impala://` takes the same form. The driver speaks only the HTTP transport
  of HiveServer2. Hive needs `hive.server2.transport.mode=http`. The default
  port is 10001 for Hive and 28000 for Impala. The binary ports, 10000 on
  Hive and 21050 on Impala, no longer work, so every earlier URL must change.
  `auth=`, `service=` and `transport=` in a URL now fail, and a URL with no
  database no longer gets `default`. The driver has no transactions, and
  writes each argument into the text of the statement. Impala does not
  report how many rows a statement changed.
- Oracle moves from sijms/go-ora to `github.com/xo/oracle` v0.3.0, pure Go.
  A wrong user or password now prints ORA-01017, where go-ora printed
  "message code error", and usql prompts for the password. An error no longer
  ends with the line " error occur at position: N". The driver
  builds on 32 bit arm, so the oracle driver is no longer left out there.
  `ORACLE_SID` and `ORASID` no longer give the service name of a URL that has
  none. Write the service name in the path of the URL. usql registers the driver with no hooks,
  and the prompt for a refused password comes from the method
  `AuthenticationRefused` of the driver's errors (W66).
- Cassandra moves to xo/cassandra v0.2.2, which reports a refused credential
  itself (W66), so usql registers it with no hooks. usql now also prompts for
  a missing password on Cassandra, which it did not before.
- ODBC moves to xo/odbc v0.3.0, which reports a refused credential itself
  (W66). usql now also prompts for the password when PostgreSQL refuses it
  through psqlODBC, which reports it as state 08001 and not in class 28. A
  locked or expired login no longer prompts on SQL Server and MySQL.
- MaxCompute and Tablestore move to `github.com/xo/dbimp` (dbimp v0.23.0,
  dburl v0.54.0), and dbmeta v0.16.0 reads their metadata. The URLs are `maxcompute://id:key@host/project` and
  `tablestore://id:secret@host/instance`, and every earlier URL must change.
  A host that is not `service.<region>.maxcompute.aliyun.com` needs the
  option `region`. A Tablestore host of the form `instance.region` gets the
  suffix `.ots.aliyuncs.com`. The scheme and the build tag of Tablestore are
  `tablestore`, with `ots` and `ot` as aliases, and the tag `ots` is gone.
  Neither driver has parameters or transactions. MaxCompute writes each
  argument into the statement, and the server cuts a result at 10,000 rows
  with no error. Tablestore sends one statement in each request and holds the
  whole result in memory. Neither ran against the hosted service through usql.
  Check each by hand before the release: a query and a wrong credential.
- Avatica moves to `github.com/xo/dbimp` (W39). Its URL is
  `avatica://user:password@host:port` and not `http://`. Check it against
  the standalone server and Phoenix.
- ClickHouse moves to `github.com/xo/dbimp` (W42). It speaks HTTP only, on
  port 8123 or, with `tls=true`, 8443. The native protocol and port 9000 are
  gone, and so are the `+http` and `+https` transports and the options of
  `clickhouse-go`, such as `secure=true`.
- Trino and Presto move to `github.com/xo/dbimp/trino` (W40). A URL with no
  user connects as `user`. `trs://` and `trinos://` are gone. Presto has
  `\copy` now.
- ODBC moves to `github.com/xo/odbc` (W40), which needs no cgo. Its URL is
  `odbc+<driver>://user:pass@host:port/db` with the settings of the ODBC
  driver in the query. dburl no longer adds a port, a host or a ServiceName,
  and a URL with no driver is an error. A URL with no host and a path, such
  as `odbc+sqlite3:/path/to/file.db`, keeps its driver (dburl v0.55.2).
  The path of the URL means what the ODBC driver needs (xo/odbc v0.3.0, its
  D28). SQL Server takes the instance and the database, SQLite and DuckDB
  take the path of the file, and every other driver takes the whole path as
  the name of the database. The driver once kept only the part after the
  last `/`, so `odbc+sqlite://home/ken/a.db` made `a.db` in the current
  directory. That URL has the host `home`, and SQLite and DuckDB now refuse a
  host other than `localhost`, with an error and no file.
- MaxCompute, Tablestore and YDB choose TLS with `tls=true` (W41), and
  `yds://` and `ydbs://` are gone. MaxCompute and Tablestore use TLS by
  default.
- The registry is keyed by the scheme of dburl.

### Where usql keeps its files

usql reads and writes no file of its own in the home directory. It follows
the XDG Base Directory specification on Linux and other Unix systems.

- The history is `$XDG_STATE_HOME/usql/history`, which is
  `~/.local/state/usql/history` when `XDG_STATE_HOME` is not set. It is
  `~/Library/Application Support/usql/history` on macOS and
  `%LocalAppData%\usql\history` on Windows. usql makes the directory.
  `USQL_HISTORY` still names another file.
- `~/.usql_history` is not read, and its entries are not moved. The format
  of rline differs from that of gohxs/readline.
- At a terminal, usql prints a warning at start when it cannot make the
  directory or read the file, and then does not save the history.
- The startup script is `init.sql` in the configuration directory, next to
  `config.yaml`: `$XDG_CONFIG_HOME/usql/init.sql` or
  `~/.config/usql/init.sql`. `USQLRC` still names another file, and `-X`
  still skips it. `~/.usqlrc` is not read.
- `config.yaml` no longer runs an `init:` script. usql prints a warning, with
  the path of `init.sql`, when the file still holds one.
- The `.usqlpass` password file is gone, with no warning. Named connections in
  `config.yaml` replace it. Tab completion no longer offers its entries.
- `_samples/usqlrc` is now `_samples/init.sql`, and `_samples/usqlpass` is
  gone.

### Other

- `USQL_DSN` names the DSN to open when the command line names none. A
  program that holds a password, such as one that forwards a port, can run
  usql with it and keep the password out of the argument list, which every
  user of the machine can read. A DSN on the command line wins, and `\c`
  never reads the variable. usql removes it from its environment after it
  reads it, so that `\!`, the pager and the editor do not get it. There is
  no variable for the password alone, and `USQL_PASSWORD` is still the
  variable of `-W`.
- usql opens a database in Kubernetes from a URL such as `k://pg.prod/app` or
  `k8s+cnpg://pg.prod/app`, through dbfwd v0.3.1 and `kubectl port-forward`.
  dburl v0.57.0 parses the `k8s` scheme, with the aliases `k`, `kc`, `k8`,
  `kube` and `kubernetes`. The type after `+` takes dburl's aliases, so
  `k+pg://` is PostgreSQL. The password comes from the Secret, so `-W` gives
  an error. The banner and `\conninfo` name what the forward reached, and the
  prompt shows the host from the URL. A lost forward stops each statement
  with its reason, and `\c` with no argument starts it again. usql stops the
  forward on `\c`, `\Z` and at exit.
- The describe commands, `\l` and `\lo_list` take psql 18's modifier `x`,
  which shows one result in expanded mode, such as `\dtx` or `\dSx t`. The
  modifiers go in the order `S`, `x`, `+`, so `\dtx+` works and `\dt+x` does
  not. `\dx` is still the extensions, and `\dxx` shows them expanded.
- `\restrict KEY` and `\unrestrict KEY` work as in psql 18. pg_dump 18
  writes them on the first and the last line of a plain dump. Between them,
  statements run and no meta command does, so a meta command in the data of
  a dump cannot run. Before, usql printed "Invalid command" for both lines.
- `\errverbose` shows the error of the last statement that failed, with every
  field that the driver gives, in psql's verbose layout for every database:
  `ERROR:  CODE: MESSAGE`, then `LINE` and a caret where the server gives a
  position, then one `NAME:  value` line for each other field. PostgreSQL,
  through pgx or lib/pq, matches psql. MySQL gives its SQLSTATE and its own
  number, SQL Server its number, level, state and line, and SQLite its result
  codes. A driver library with no package in usql gives its fields through a
  method `ErrVerbose`, which usql finds with `errors.As`: Oracle (xo/oracle
  v0.4.1), Cassandra (xo/cassandra v0.2.4), ODBC (xo/odbc v0.4.1) and the 27
  drivers of dbimp (v0.25.2). Each library moved to psql's layout on
  2026-11-08, and was checked against a live server. Another driver shows the
  error as usql prints it.
- `\gdesc` shows the name and the type of each column of the query in the
  buffer, without running it, as psql does. It works on PostgreSQL and the
  databases on pgx, which prepare the query and name its types with dbmeta,
  and on SQL Server, through `sp_describe_first_result_set`. Any other
  database refuses it, because it can describe a query only by running it.
- `\df PATTERN TYPES...` lists only the functions whose arguments have the
  types, and `\do PATTERN [LEFT] RIGHT` the operators whose operands do, as
  in psql. `-` means no argument. Only PostgreSQL and CockroachDB match the
  types, and another database says so (dbmeta v0.36.0).
- `\list` and `\html` are psql's long names of `\l` and `\H`.
- At a terminal, `help`, `quit` and `exit` typed alone on a line, with or
  without a `;`, show the help, or quit, as psql does and as usql did before
  the input moved to transit. Inside a statement they are part of the
  statement.
- `\history [COUNT]` shows the history of the line editor, with the number of
  each entry, as bash's `history` does. `\s [FILE]` shows it as psql does,
  or writes it to a file.
- `\k` lists the databases that usql can reach in Kubernetes, with the URL
  that `\c` opens. `\k` lists the current namespace, `\k -A` every
  namespace, `-n NAMESPACE` one namespace, and `--context NAME` another
  context. A scope narrows the list, such as `pg*`, `pg*.prod` or `cnpg://*`,
  and its `k://` is optional. `\k+` adds the roles, the ready pods, the
  databases and where the password comes from. No Secret is read. The URLs
  carry `?context=` only when `\k` named a context.
- On Windows, a path on the drive `K:`, such as `K:\data\app.db`, now parses
  as a `k8s` URL and fails. Write it as `file:///K:/data/app.db`.
- `\c` closes the connection that it replaces. Before, the pool of the old
  connection stayed open until usql exited. usql also closes the connection
  at exit, and rolls back a transaction that is still open.
- tblfmt v0.23.0: with `numericlocale`, a number uses the digits, the group
  sizes, the decimal point and the minus sign of the locale. On Windows, the
  html, asciidoc, latex, latex-longtable, troff-ms and vertical formats end
  each line with `\r\n`. The unaligned format ends its last line with a
  newline, not with `recordsep`, as psql does. The markdown format writes a
  backslash as `\\`.
- The line editor keeps 500 entries of history as psql does, where it kept
  200, and does not redraw the line while more keys wait, so a paste
  without bracketed paste is no longer slower with each row (rline v0.6.0).
  usql warns once when the history file cannot be read or written. Keys
  typed or pasted right after the start are no longer lost at a terminal
  that does not answer the query for graphics, such as Emacs or a serial
  console (rasterm v0.1.22). Keys typed while usql asks the terminal whether
  it draws images are no longer lost: rasterm v0.1.23 keeps them, and rline
  v0.8.0 reads them first. Before, `select 1;` typed at once ran as `elect
  1;`. The
  prompt starts after output that did not end its row, as after `\echo -n`,
  and a `[x]` in a prompt stays as it is written.
- An `INSERT`, `UPDATE` or `DELETE` with `RETURNING` prints its rows, as in
  psql, on PostgreSQL, CockroachDB, SQLite, DuckDB, MariaDB and Firebird, and
  so does SQL Server's `OUTPUT` without `INTO`. It printed only its tag.
  dbmeta now decides from the whole statement whether it returns rows (its
  D264), and `drivers/qtype.go`, which looked at the first words only, is
  gone.
- The prompt's `%/` shows the current database, which dbmeta reads after the
  connection and after a `USE` (its D265). On SQL Server with
  `?database=northwind` it was empty, and after `USE master` it showed the
  old database.
- A foreign key to a table outside the search path names its schema, as in
  `REFERENCES hr.departments(id)`, and a name is quoted only when the
  database would not read it back as written, so SQL Server's `Customers`
  has no quotes.
- Input from `-f`, a pipe or `\i` runs a last statement that has no `;`, as
  psql does. Before, it was dropped and usql exited with status 0.
- `\if`, `\elif`, `\else` and `\endif` work as in psql (W36). Before, they
  did nothing, and every branch ran.
  - Blocks nest. The expression is a boolean, read after its variables are
    replaced: true, false, yes, no, on, off, 1 or 0, in any case, or a unique
    prefix of one. Any other value is an error, and the branch does not run.
  - A branch that does not run sends no statement, runs no meta command
    other than these four, and replaces no variable. Its statements do not go into the
    query buffer, and a file that its `\i` names is not read.
  - An `\elif` or `\else` after `\else`, and an `\elif`, `\else` or
    `\endif` with no `\if`, are errors with psql's text.
  - The terminal, the text of `-c`, and each file of `-f` and `\i` have
    their own blocks. When one ends inside a block, usql prints psql's error
    `reached EOF without finding closing \endif(s)`. The exit status does not
    change, as in psql, unless on_error_stop is on.
  - At a terminal, `%R` in the prompt shows `@` in a branch that does not
    run. A statement or a meta command there prints that it is ignored, and
    Ctrl-C leaves the innermost block, as in psql.
  - An unknown meta command prints `Invalid command \foo. Try \? for help.`,
    where it printed `error: unknown command`.
- The `editor` setting runs through the shell, as psql runs it, so it can
  hold arguments, as in `code --wait`. With no editor set, `\e` runs `vi` on
  a new file, where it failed.
- `\uset fetch_count N`, psql's FETCH_COUNT, shows a result in groups of N
  rows, so that a large result does not have to fit in memory. A result of a
  million rows took 627 MB in the aligned format, and takes 71 MB with
  `fetch_count 1000`.
- `\copy` passes each value as its `Value()`, so a NULL reaches a driver
  whose prepared statement takes only plain values, such as csvq's.
- Tab completion folds the typed text as the database folds a name, so `ord`
  and `ORD` both find `orders` on PostgreSQL, and quotes a name that needs it,
  such as `"MyTable"` or `"Order Details"`. After `ALTER` or `DROP` of a
  kind, it offers the names of that kind. After `CREATE`, `ALTER` and `DROP`,
  it offers the words that the grammar of the database allows, so SQLite is
  not offered `DROP DATABASE`.
- A single-quoted argument of a meta command decodes as psql decodes it:
  `'\q'` is `q`, `'\61'` is `1`, `'\x4'` is the byte 4, and `'\u0041'` is
  `u0041`. A variable in a command between backticks is replaced, and
  `:'name'` is quoted for the shell.
- `\d` of a temporary relation is titled `Table`, `View` or `Sequence`, as in
  psql, where usql wrote `Temporary table`.
- `\ss` needs a table. Its help named a QUERY form that did not exist.
- `\pset recordsep` says `Record separator is`, where it said `Field
  separator is`.
- The README says where usql looks for `config.yaml`: `--config`,
  `USQL_CONFIG`, then the configuration directory of the system, which
  follows `XDG_CONFIG_HOME` on Linux and not on macOS. The help names the
  macOS directory with its `/usql`.
- The sample `config.yaml` no longer has a `queries:` section, which nothing
  read.

- A SQL Server `uniqueidentifier` prints in lower case, as a PostgreSQL `uuid`
  does, so that a column of UUIDs prints the same on both. SQL Server's tools
  print it in upper case.
- A SQL Server `real` prints as sqlcmd prints it, `4.9`, where it printed
  `4.900000095367432`. go-mssqldb widens a `real` to a float64, and usql now
  reads it back as a float32.
- A session runs on one connection, as psql runs one (W67). A `SET`, a
  temporary table or a variable of the session holds until the connection
  ends. Before, each statement ran on whichever connection of the pool was
  free, and on SQL Server a `SET` never held, because go-mssqldb resets a
  connection that the pool takes back. When the connection is lost, usql says
  so, opens a new one, and runs the statement again.
- A large statement read from a file, such as an `INSERT` of many rows, no
  longer costs memory with the square of its size (W68). An `INSERT` of 830
  rows ran out of memory at more than 4 GB, and now takes 87 MiB.
- `\conninfo` prints a table titled Connection Information, as psql 18 does:
  the driver, the database, the user, the host, the port, whether a password
  was used, the version of the server, and the URL. pgx adds the backend PID,
  SSL, superuser and hot standby rows of psql. The password shows as xxxxx.
  `\conninfo -password` shows it, and adds the DSN.
- `size.sh` measures what each driver adds to the binary (W32).
- `go run gen.go -examples` writes the examples of the README from what usql
  prints, against SQL Server with dbmeta's fixture `examples`. CONTRIBUTING.md
  says how.
- The tests against databases are back, on servers that dbmeta's `dbrun`
  starts (W37). They started containers of their own, and were removed for
  that. `TestConnect`, `TestDescribe` and `TestCopy` in
  `drivers/dbrun_test.go` read the URL of the database from `USQL_TEST_URL`,
  and skip when it is not set. `TestDescribe` records what the describe
  commands print on dbrun's fixture `examples`, in `drivers/testdata/dbrun`,
  for PostgreSQL 18, SQL Server 2025, MySQL 26.7, SQLite, moderncsqlite and
  DuckDB. `TestCopy` runs `\copy` between the database and SQLite, both ways.
  CONTRIBUTING.md says how to run them.
- CI runs these tests through `dbrun` (W38) on each push, for PostgreSQL 18,
  SQL Server 2025, MySQL 26.7, SQLite and DuckDB. Every other database is
  tested locally.
- `usql --help` no longer shows `(default true)` after each switch, such as
  `--no-align`, which pflag v1.0.10 printed.
- The message that usql connected goes to the terminal and not to the pager,
  so a long version of the server does not open the pager.
- The Releases badge drops `sort=semver`, on which shields.io times out.
- The README ends with the bar of links to the sister projects.
- The backlog gains W28 to W38.
- `internal/cmd/pgcompare` compares the output of every PostgreSQL type with
  psql (W63).
- `gen.go` lists Trino and Presto in a fixed place in the README table of the
  base drivers. Before, the two rows changed places on each run.
- When the command of `\g |command` quits before it reads all the output, usql
  closes the pipe and waits for the command. Before, it left both.

## What to check by hand

Use dbmeta's `dbrun` to start a database: `DBMETA_OWNER_NAME=usql go run
./cmd/dbrun start NAME`, in dbmeta's `test` directory. `dbrun dsn --json
NAME` gives the URL. Stop each database when its checks are done. Start no
more than a few at a time.

### Connecting and the banner

For every driver in the build, check that it connects, and that the banner
names the database and its release. For the new drivers, check one query and
one error, and a wrong password:

- CockroachDB, CrateDB, Neo4j, InfluxDB and InfluxQL
- SurrealDB and ArangoDB
- Pinot, rqlite and libSQL
- Druid, with `\dtS`, `\d` and a wrong password
- Drill, Solr, Elasticsearch, OpenSearch and DynamoDB, with a statement that
  ends in `;`, and `\dt` on Drill, Solr and Elasticsearch
- Redshift, as a user who is not a superuser: `\dt`, `\d` and `\dv` work, and
  `\dt+` shows no size and rows
- ClickHouse on its HTTP port, and Trino and Presto, with and without a user
- ODBC with a real driver, such as SQL Server's, PostgreSQL's and SQLite's,
  after the crash is fixed
- Couchbase and Databend on dbimp
- Hive and Impala on dbimp's hive driver, over HTTP: `select version()`,
  `\conninfo`, and `\copy` into a table, which inserts each
  row with no transaction

### Databases in Kubernetes

Start a cluster with `DBMETA_OWNER_NAME=usql go run ./cmd/dbrun start
k8s-1.37`, and set `KUBECONFIG` to the path that `dbrun dsn --json k8s-1.37`
prints. Put a database and its Secret in a namespace of your own.

- Connect with `k://NAME.NAMESPACE/` and with a type, such as
  `k8s+postgres://`, and run a query.
- Check the banner and the prompt at a terminal, and `\conninfo`, which shows
  the forward, the URL, and the local URL with the password hidden.
- Check that a name that does not exist lists each step that was tried.
- Scale the database to zero for more than two minutes, run a statement, and
  check that it reports the lost forward. Scale it up, and check that `\c`
  connects again.
- Check that no `kubectl` process is left after `\q`.
- Run `\k`, `\k+`, `\k -A`, `\k -n NAMESPACE`, `\k pg://*` and
  `\k --context NAME`, and open a listed URL with `\c`.

### Changing passwords

Run `\password` as an ordinary user and as an administrator on each database
that dbmeta can change a password on: cassandra, clickhouse, cockroachdb, cratedb,
databend, exasol, memsql, mysql, oracle, postgres, redshift, snowflake,
sqlserver, tidb and vertica.

- Check that the old password stops working and the new one works.
- On SQL Server, check that usql asks for the old password.
- Use a password that holds a quote, a backslash and a space.

### Input, for each set of lexical forms

The dialects of dbmeta use six sets of lexical forms. Check each set with one
of its databases:

| Forms like | Databases |
| --- | --- |
| postgres | cockroachdb, cratedb, duckdb, postgres, questdb, redshift, snowflake, vertica |
| cassandra | cassandra |
| mysql | memsql, mysql, surrealdb, tidb, vitess |
| sqlite | arangodb, impala, libsql, neo4j, rqlite, sqlite3, ydb |
| standard | clickhouse, couchbase, exasol, firebirdsql, hdb, influxql, oracle, presto, sqlserver, trino |
| plain | databend, hive, influxdb |

For each, check:

- a `;` inside a string, a quoted name, a comment, and, where the dialect
  has them, a dollar quote and a backtick
- a statement over several rows, at a terminal and in a script
- `select 1 \g`, `\g` on an empty buffer, `\gx`, `\gset`, `\gexec`, and
  `\g (format=csv) FILE`
- variables: `:x`, `:'x'`, `:"x"`, `:{?x}`, a value that holds a `;` or a
  quote, and a `::` cast
- `\c` to a database with other forms, then a statement that the new forms
  read differently
- CQL: `BEGIN BATCH` … `APPLY BATCH`
- Oracle: a statement with a trailing `;`, and a PL/SQL block that ends with
  `END;`
- Trino and Presto: a statement with a trailing `;`

### Syntax highlighting

Check the color of keywords, strings, comments, variables and meta commands
for each set above, in two styles, at a light and a dark background. Check
that `\uset syntax_hl off` removes every color, that `\uset syntax_hl_format
terminal` draws 16 colors, and that `\uset syntax_hl_style nope` is an error.
Check a long statement of many rows, where each key parses the text again, for
a delay that you can feel. Check that `NO_COLOR=1` starts with highlighting off
(W58). Check that `\p`, `\sv` and `\sf` print in color, through the pager too,
and that they print plain with `\o FILE` and when the output is piped.

### The describe commands

For each database that dbmeta has a model for, run `\d`, `\dt`, `\dv`,
`\di`, `\ds`, `\df`, `\dn`, `\l`, `\dp`, `\du`, `\d NAME` on a table with an
index, a check, a foreign key and a trigger, `\sv`, `\sf`, `\ev`, `\ef`,
and `\ss`. Check
upper case and quoted names on Oracle, and on a database that folds to
lower case. Check `\uset echo_hidden on` and `noexec`. On a database with no
model, check that each command says it is not supported.

### The line editor

At a terminal:

- an entry over several rows, and moving between its rows with the keys of
  readline
- Ctrl-C on an open entry of one row and of several rows, which stays on the
  screen with `^C` after it, Ctrl-D, and Ctrl-D on an empty entry
- the history across a restart, in `$XDG_STATE_HOME/usql/history` and in
  `~/.local/state/usql/history`, and the warning at start when the file is a
  directory
- tab completion of a meta command, a table, a column, a file, and a
  variable, and the numbered menu of completions
- a `~/.inputrc` with a binding under `$if usql`, and one with `set
  editing-mode vi`
- `\prompt`, `\prompt -password`, `\password` and a password prompt at
  connect
- `\e` and `\e FILE`, with editor set, and Ctrl-C in `\! sleep 10`
- help, quit and exit as the first word of an entry
- a terminal that is resized while an entry is open
- an alt-tab away from the terminal and back, while an entry is open, on
  Windows Terminal and on one terminal of Linux or macOS. Typing after it
  inserts the keys, no completion menu opens, and the bell rings at most once
  for each change of focus.

### The pager

At a terminal, with PAGER unset, with `PAGER=less`, and with
`PAGER="less -S"`:

- a query that fits, one that is taller than the terminal, and one with a row
  that is wider than it
- `\?`, `\drivers`, `\d` on a large schema, and `\pset` with no arguments
- `\pset pager off`, `\pset pager always`, and `\pset pager_min_lines 100`
- q in the pager before the output ends, on a query of a million rows
- Ctrl-C in the pager, and Ctrl-C during a long query, which cancels it
- `\uset pager nosuchpager`, which prints the error of the shell
- CSV and JSON output that is taller than the terminal
- `\timing on` with a query that pages
- on Windows, with PAGER unset, which uses more

### Restricted mode

- Make a plain dump with pg_dump 18, add a line such as `\! echo ran` after
  its `\restrict` line, and run the dump with `usql -f`. Check that usql
  refuses the line, and that `\unrestrict` at the end of the dump passes.

### Scripts, pipes and the command line

- `-c` with a statement, with a meta command, and with several of each
- `-f FILE`, `\i`, `\ir`, a file that ends without a `;`, and Windows line
  endings
- `\if`, `\elif`, `\else` and `\endif` in a file of `-f`, in a file of `\i`
  inside a block, and at the terminal, where the prompt shows `@` in a branch
  that does not run and Ctrl-C leaves the block. Compare the output with
  psql's.
- input from a pipe, output to a pipe, and `-o FILE`
- a script in Windows-1252 and one in Shift JIS with `\uset encoding`, a
  script in UTF-16 that Notepad or PowerShell saved, and the error for a
  script that is not valid UTF-8, with `-f`, `\i` and a pipe
- `-1` with a file, and on_error_stop on and off
- `--set`, `--variable`, `init.sql`, `USQLRC`, and `config.yaml` with named
  connections, and the warning for a `config.yaml` that holds `init:`
- `\copy` between two databases, and into a database with no transactions,
  such as rqlite

### The meta commands

Run each meta command in `\?` once, with its options. Check the commands
that read their arguments in a new way most: `\set`, `\echo`, `\o`, `\g`,
`\copy`, `\c`, `\cset`, `\prompt`, `\!` and `\setenv`. Run `\conninfo` and
`\conninfo -password` on pgx, on libpq and on SQLite, and inside a transaction.

### Column alignment

On each base database, select a column of each numeric type and a column of
text, and make sure that the numbers go to the right and the text to the left.
Use a column whose values are all null too. On PostgreSQL, with pgx and with
libpq, a money column goes to the right, and boolean and interval go to the
left, as in psql. This was checked on PostgreSQL 18, Oracle and SQLite on
2026-10-31.

### Builds and packages

- `go build` with no tags, `-tags no_drivers`, `-tags 'no_drivers sqlite3'`
  and `-tags no_sqlite3`, and `\drivers` of each: every driver of the README
  table, none, only sqlite3, and moderncsqlite in place of sqlite3
- `\chart` in a build with no tags, and the error of `\chart` in a build with
  `-tags no_charts`
- `./size.sh`, compared with W32. The chart renderer is now in the build.
- on Linux, an ODBC query through the SQLite ODBC driver in a build with
  `-tags libsqlite3`, and through the PostgreSQL Unicode driver in a build with
  no tags
- the 32 bit arm build of `build.sh`, which leaves out ChaiSQL and DuckDB
- the static builds of `build.sh`, for amd64 and arm64
- macOS, Windows and Linux
- the release archives, the Homebrew formula, the AUR packages and the
  Docker image
