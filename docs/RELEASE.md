# The next major release

This file lists what changed in usql since v0.21.6, and what must be checked
by hand before the release. It is the largest set of changes that usql has
had. Automated tests cover much of it, and Ken checks the rest by hand.

Each check names the databases it covers. Tick a check off in a copy of this
file, not here. When a check finds a fault, record it in
[BACKLOG.md](BACKLOG.md) or fix it, and add a test where one can be written.

## Known gaps

Read these first. Some of them can stop the release.

- The history is not saved for a user who has a `~/.usql_history` from
  gohxs/readline. rline cannot read its backslashes, and then refuses to
  save. It reports nothing. Ken decides the fix (see
  [PROGRESS.md](PROGRESS.md)).
- `\if`, `\elif`, `\else` and `\endif` still do nothing, and every branch
  runs. W36 gives them psql's meaning, and that work is not done.
- `%R` in a prompt shows only `=` or `-`, because rline draws one prompt for
  an entry. rline's backlog item 33 holds a prompt for each row.
- The keys typed in the first moments after start are lost. At start, usql
  asks the terminal whether it draws images, through rasterm. rasterm sends
  ESC[0c and reads up to 1 KB from standard input in one read, so keys typed
  before or with the answer are read and dropped. On a pty, a query typed up
  to 3 s after start was lost, and one typed after 4 s ran.
- rline writes an escape code to a pipe when color is on. usql turns color off
  unless both input and output are a terminal.
- ODBC no longer reads `usql_trim`, csvq's `SHOW` prints nothing, and Athena
  and Cosmos no longer strip a trailing semicolon. W28 holds them.
- usql requires dbmeta v0.27.0, tblfmt v0.21.1, transit v0.7.1 and rline
  v0.5.0, all in `go.mod`. It builds with no `go.work`.
- The tests against databases and their recordings were removed, because
  they started their own containers. W37 brings them back on `dbrun`, and
  W38 runs CI on every database. Both are due before the release.
- **ODBC crashes in the default build.** With go-sqlite3 and DuckDB both
  linked, which every `most` build does, the first ODBC query crashes the
  process, because DuckDB's bindings export the `sqlite3_*` symbols of
  go-sqlite3 into the executable. Building with `-tags libsqlite3` fixes it,
  and that changes the SQLite of the release builds. Ken decides how (W40).
  This stops the release until he does.
- `\copy` still writes its own SQL: the INSERT, and the query for the columns of
  the target. The placeholder of each database comes from dbmeta. W29 moves the
  rest, and is due before the release.

## What changed

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
- On Oracle, a line with only `/` ends a statement, as in SQL*Plus, and is not
  sent with it: after a PL/SQL block, after a `CREATE PROCEDURE ... END;`, and
  after a statement with no `;`. Inside parentheses it stays a division
  (transit v0.7.1, W65).
- An anonymous block of PL/SQL on Oracle, `BEGIN ... END` or `DECLARE ...
  BEGIN ... END`, stays in one statement, and so do Firebird's `EXECUTE BLOCK`,
  SAP HANA's `DO BEGIN` and a `WITH FUNCTION` of Trino (W65).

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
  script or a `.usqlrc` that set a setting with `\set` must change. `:pager` is no
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
  `no_mysql` leaves out memsql, tidb and vitess, `no_influxdb` leaves out
  influxql, and `no_flightsql` leaves out gizmosql. `gen.go` reads this from
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
  `RAW(16)` from `SYS_GUID()` now prints as `\x` and 32 hex digits. A SQL
  Server uniqueidentifier is not binary, and prints in lower case as a
  PostgreSQL uuid does.
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
- flightsql moves to arrow-go v18, and hive to gohive v2.
- Oracle moves from sijms/go-ora to `github.com/xo/oracle` v0.2.0, pure Go.
  A wrong user or password now prints ORA-01017, where go-ora printed
  "message code error", and usql prompts for the password. An error no longer
  ends with the line " error occur at position: N". The driver
  builds on 32 bit arm, so the oracle driver is no longer left out there.
  `ORACLE_SID` and `ORASID` no longer give the service name of a URL that has
  none. Write the service name in the path of the URL. usql registers the driver with no hooks,
  and the prompt for a refused password comes from the method
  `AuthenticationRefused` of the driver's errors (W66).
- Cassandra moves to xo/cassandra v0.2.1, which reports a refused credential
  itself (W66), so usql registers it with no hooks. usql now also prompts for
  a missing password on Cassandra, which it did not before.
- ODBC moves to xo/odbc v0.2.0, which reports a refused credential itself
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
- ODBC moves to `github.com/xo/odbc` (W40), which needs no cgo, and so it is
  in `most` and no longer in `all`. Its URL is
  `odbc+<driver>://user:pass@host:port/db` with the settings of the ODBC
  driver in the query. dburl no longer adds a port, a host or a ServiceName,
  and a URL with no driver is an error.
- MaxCompute, Tablestore and YDB choose TLS with `tls=true` (W41), and
  `yds://` and `ydbs://` are gone. MaxCompute and Tablestore use TLS by
  default.
- The registry is keyed by the scheme of dburl.
- A driver goes in the group `most` when its scheme names a dialect and it
  needs no cgo. Pinot, rqlite, libSQL, Drill, Solr, Elasticsearch, OpenSearch,
  Druid and DynamoDB are in `most`.

### Other

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
index, a check, a foreign key and a trigger, `\sv`, `\sf`, and `\ss`. Check
upper case and quoted names on Oracle, and on a database that folds to
lower case. Check `\uset echo_hidden on` and `noexec`. On a database with no
model, check that each command says it is not supported.

### The line editor

At a terminal:

- an entry over several rows, and moving between its rows with the keys of
  readline
- Ctrl-C on an open entry of one row and of several rows, which stays on the
  screen with `^C` after it, Ctrl-D, and Ctrl-D on an empty entry
- the history across a restart, with and without an old history file
- tab completion of a meta command, a table, a column, a file, and a
  variable, and the numbered menu of completions
- a `~/.inputrc` with a binding under `$if usql`, and one with `set
  editing-mode vi`
- `\prompt`, `\prompt -password`, `\password` and a password prompt at
  connect
- `\e` and `\e FILE`, with editor set, and Ctrl-C in `\! sleep 10`
- help, quit and exit as the first word of an entry
- a terminal that is resized while an entry is open

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

### Scripts, pipes and the command line

- `-c` with a statement, with a meta command, and with several of each
- `-f FILE`, `\i`, `\ir`, a file that ends without a `;`, and Windows line
  endings
- input from a pipe, output to a pipe, and `-o FILE`
- `-1` with a file, and on_error_stop on and off
- `--set`, `--variable`, the RC file and `config.yaml` with named
  connections
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

- `go build -tags all`, the default build, and `-tags no_base`
- `./size.sh -g most`, compared with W32
- the static builds of `build.sh`, for amd64 and arm64
- macOS, Windows and Linux
- the release archives, the Homebrew formula, the AUR packages and the
  Docker image
