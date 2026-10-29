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
- At a pseudo-terminal, the first keys typed after start can be lost, and
  rline writes an escape code to a pipe when color is on. usql turns color
  off unless both input and output are a terminal.
- ODBC no longer reads `usql_trim`, csvq's `SHOW` prints nothing, and Athena
  and Cosmos no longer strip a trailing semicolon. W28 holds them.
- usql requires dbmeta v0.15.2, transit v0.5.0 and rline v0.1.0, all in
  `go.mod`. It builds with no `go.work`.
- The tests against databases and their recordings were removed, because
  they started their own containers. W37 brings them back on `dbrun`, and
  W38 runs CI on every database. Both are due before the release.
- **ODBC crashes in the default build.** With go-sqlite3 and DuckDB both
  linked, which every `most` build does, the first ODBC query crashes the
  process, because DuckDB's bindings export the `sqlite3_*` symbols of
  go-sqlite3 into the executable. Building with `-tags libsqlite3` fixes it,
  and that changes the SQLite of the release builds. Ken decides how (W40).
  This stops the release until he does.
- Ctrl-C ends usql. It cancels the query, and then rline raises the signal again
  and the process ends. The same happens with Ctrl-C in the pager. rline v0.1.0
  watches SIGINT for the life of the session, and the rline session was told on
  2026-10-29. This stops the release until rline fixes it.
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
- The arguments of a meta command come from the grammar. `\g` reads its
  option list, such as `(format=csv)`, from the grammar.
- `-o FILE` sends the results to the file, as `\o` does. `\echo` still goes
  to the terminal, as in psql.
- The completer is an rline completer.

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
- `\uset` shows every setting as a table, `\uset NAME` shows one, and `\uset NAME
  VALUE` sets it and says so. A value is read as an argument of any meta command.
- `\set pager less` and `\unset quiet` are errors that say to use `\uset`, so a
  script or a `.usqlrc` that set a setting with `\set` must change. `:pager` is no
  longer a variable.
- The command line has `--uset NAME=VALUE`, and `-q` is `--uset quiet=on`. `--set`
  and `-v` give script variables only.

### The pager

The pager moved out of tblfmt into the new package `pager` of usql (W64).
usql requires tblfmt v0.20.0, which has no pager. All
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
- MaxCompute and Tablestore move to `github.com/xo/dbimp` (dbimp v0.21.1,
  dburl v0.54.0). The URLs are `maxcompute://id:key@host/project` and
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

- `size.sh` measures what each driver adds to the binary (W32).
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
(W58).

### The describe commands

For each database that dbmeta has a model for, run `\d`, `\dt`, `\dv`,
`\di`, `\ds`, `\df`, `\dn`, `\l`, `\dp`, `\du`, `\d NAME` on a table with an
index, a check, a foreign key and a trigger, `\sv`, `\sf`, and `\ss`. Check
upper case and quoted names on Oracle, and on a database that folds to
lower case. Check `\uset echo_hidden on` and `noexec`. On a database with no
model, check that each command says it is not supported.

### The line editor

At a terminal:

- an entry over several rows, and moving between its rows
- Ctrl-C on an open entry, Ctrl-D, and Ctrl-D on an empty entry
- the history across a restart, with and without an old history file
- tab completion of a meta command, a table, a column, a file, and a
  variable
- `\prompt`, `\prompt -password`, `\password` and a password prompt at
  connect
- `\e` and `\e FILE`, with editor set
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
`\copy`, `\c`, `\cset`, `\prompt`, `\!` and `\setenv`.

### Builds and packages

- `go build -tags all`, the default build, and `-tags no_base`
- `./size.sh -g most`, compared with W32
- the static builds of `build.sh`, for amd64 and arm64
- macOS, Windows and Linux
- the release archives, the Homebrew formula, the AUR packages and the
  Docker image
