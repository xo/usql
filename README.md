# usql [![Build Status][travis-ci]](https://travis-ci.org/xo/usql)

A universal command-line interface for PostgreSQL, MySQL, Oracle Database,
SQLite3, Microsoft SQL Server, [and many other databases][Database Support]
including NoSQL and non-relational databases!

[travis-ci]: https://travis-ci.org/xo/usql.svg?branch=master "Travis CI"

[Installing][] | [Building][] | [Using][] | [Database Support][] | [Features and Compatibility][] | [Releases][]

[Installing]: #installing (Installing)
[Building]: #building (Building)
[Using]: #using (Using)
[Database Support]: #database-support (Database Support)
[Features and Compatibility]: #features-and-compatibility (Features and Compatibility)
[Releases]: https://github.com/xo/usql/releases (Releases)

## Overview

`usql` provides a simple way to work with [SQL and NoSQL databases][Database Support]
via a command-line inspired by PostgreSQL's `psql`. `usql` supports most of the
core `psql` features, such as [variables][], [backticks][], and [commands][]
and has additional features that `psql` does not, such as [syntax highlighting][highlighting],
context-based completion, and [multiple database support][Database Support].

Database administrators and developers that would prefer to work with a tool
like `psql` with non-PostgreSQL databases, will find `usql` intuitive,
easy-to-use, and a great replacement for the command-line clients/tools
for other databases.

## Installing

`usql` can be installed [via Release][], [via Homebrew][], [via Scoop][] or [via Go][]:

[via Release]: #installing-via-release
[via Homebrew]: #installing-via-homebrew-macos
[via Scoop]: #installing-via-scoop-windows
[via Go]: #installing-via-go

### Installing via Release

1. [Download a release for your platform][Releases]
2. Extract the `usql` or `usql.exe` file from the `.tar.bz2` or `.zip` file
3. Move the extracted executable to somewhere on your `$PATH` (Linux/macOS) or
`%PATH%` (Windows)

### Installing via Homebrew (macOS and Linux)

`usql` is available in the [`xo/xo` tap][xo-tap], and can be installed in the
usual way with the [`brew` command][homebrew]:

```sh
# install usql with "most" drivers
$ brew install xo/xo/usql
```

Additional support for [Oracle (godror) and ODBC databases][Database Support] can be
installed by passing `--with-*` parameters during install:

```sh
# install usql with oracle (godror) and odbc support
$ brew install --with-godror --with-odbc usql
```

Please note that Oracle Database support requires using the [`xo/xo`
tap's][xo-tap] `instantclient-sdk` formula. Any other `instantclient-sdk`
formulae or older versions of the Oracle Instant Client SDK [should be
uninstalled][xo-tap-notes] prior to attempting the above:

```sh
# uninstall the instantclient-sdk formula
$ brew uninstall InstantClientTap/instantclient/instantclient-sdk

# remove conflicting tap
$ brew untap InstantClientTap/instantclient
```

### Installing via Scoop (Windows)

`usql` can be installed using [Scoop](https://scoop.sh):

```powershell
# install scoop if not already installed
iex (new-object net.webclient).downloadstring('https://get.scoop.sh')

scoop install usql
```

### Installing via Go

`usql` can be installed in the usual Go fashion:

```sh
# install usql with basic database support (includes PostgreSQL, Oracle Database, MySQL, MS SQL, and SQLite3 drivers)
$ GO111MODULE=on go get github.com/xo/usql
```

Support for additional databases can be specified with [build tags][Database Support]:

```sh
# install usql with most drivers (excludes drivers requiring CGO)
$ GO111MODULE=on go get -tags most github.com/xo/usql

# install usql with all drivers (includes drivers requiring CGO, namely Oracle and ODBC drivers)
$ GO111MODULE=on go get -tags all github.com/xo/usql
```

## Building

When building `usql` with [Go][go-project], only drivers for PostgreSQL, MySQL,
SQLite3 and Microsoft SQL Server will be enabled by default. Other databases
can be enabled by specifying the build tag for their [database driver][Database Support].
Additionally, the `most` and `all` build tags include most, and all SQL
drivers, respectively:

```sh
# install all drivers
$ GO111MODULE=on go get -tags all github.com/xo/usql

# install with most drivers (same as all but excludes Oracle/ODBC)
$ GO111MODULE=on go get -tags most github.com/xo/usql

# install with base drivers and Oracle Database (OCI)/ODBC support
$ GO111MODULE=on go get -tags 'godror odbc' github.com/xo/usql
```

For every build tag `<driver>`, there is also the `no_<driver>` build tag
disabling the driver:

```sh
# install all drivers excluding avatica and couchbase
$ GO111MODULE=on go get -tags 'all no_avatica no_couchbase' github.com/xo/usql
```

### Release Builds

[Release builds][Releases] are built with the `most` build tag. Additional
[SQLite3 build tags](build-release.sh) are also specified for releases.

### Embedding

An effort has been made to keep `usql`'s packages modular, and reusable by
other developers wishing to leverage the `usql` code base. As such, it is
possible to embed or create a SQL command-line interface (e.g, for use by some
other project as an "official" client) using the core `usql` source tree.

Please refer to [main.go](main.go) to see how `usql` puts together its
packages. `usql`'s code is also well-documented -- please refer to the [GoDoc
listing][godoc] for an overview of the various packages and APIs.

## Database Support

`usql` works with all Go standard library compatible SQL drivers supported by
[`github.com/xo/dburl`][dburl].

The list of drivers that `usql` was built with can be displayed using the
[`\drivers` command][commands]:

```sh
$ cd $GOPATH/src/github.com/xo/usql
$ export GO111MODULE=on
$ go build -tags 'no_most postgres mysql cql sqlite3' && ./usql
Type "help" for help.

(not connected)=> \drivers
Available Drivers:
  cockroachdb (postgres) [cr, cdb, crdb, cockroach]
  memsql (mysql) [me]
  mssql [ms, sqlserver]
  mysql [my, maria, aurora, mariadb, percona]
  postgres [pg, pgsql, postgresql]
  redshift (postgres) [rs]
  sqlite3 [sq, file, sqlite]
  tidb (mysql) [ti]
  vitess (mysql) [vt]
(not connected)=>
```

The above shows that `usql` was built with only the `postgres`, `mysql`, `cql`,
and `sqlite3` drivers. The output above reflects information about the drivers
available to `usql`, specifically the available driver and its primary URL
scheme, the driver's available aliases (shown in `[...]`), and the
real/underlying driver (shown in `(...)`) for the database.

Any of the protocol schemes or aliases shown above can be used in conjunction
with the [`\connect` command][commands] when connecting to a database.

#### Supported Database Schemes and Aliases

The following is a table of all drivers, schemes, and aliases that `usql`
supports:

<!-- START SCHEME TABLE -->
| Database (scheme/driver)        | Protocol Aliases [real driver]           |
|---------------------------------|------------------------------------------|
| Microsoft SQL Server (mssql)    | ms, sqlserver                            |
| MySQL (mysql)                   | my, mariadb, maria, percona, aurora      |
| Oracle Database (oracle)        | or, ora, oracle, oci, oci8, odpi, odpi-c |
| PostgreSQL (postgres)           | pg, postgresql, pgsql                    |
| SQLite3 (sqlite3)               | sq, sqlite, file                         |
|                                 |                                          |
| Amazon Redshift (redshift)      | rs [postgres]                            |
| CockroachDB (cockroachdb)       | cr, cockroach, crdb, cdb [postgres]      |
| MemSQL (memsql)                 | me [mysql]                               |
| TiDB (tidb)                     | ti [mysql]                               |
| Vitess (vitess)                 | vt [mysql]                               |
|                                 |                                          |
| MySQL (mymysql)                 | zm, mymy                                 |
| Oracle Database (godror)        | gr                                       |
| PostgreSQL (pgx)                | px                                       |
|                                 |                                          |
| Alibaba MaxCompute (maxcompute) | mc                                       |
| Apache Avatica (avatica)        | av, phoenix                              |
| Apache H2 (h2)                  | h2                                       |
| Apache Hive (hive)              | hi                                       |
| Apache Ignite (ignite)          | ig, gridgain                             |
| Apache Impala (impala)          | im                                       |
| AWS Athena (athena)             | s3                                       |
| Azure Cosmos (cosmos)           | cm                                       |
| Cassandra (cql)                 | ca, cassandra, datastax, scy, scylla     |
| ClickHouse (clickhouse)         | ch                                       |
| Couchbase (n1ql)                | n1, couchbase                            |
| Cznic QL (ql)                   | ql, cznic, cznicql                       |
| Firebird SQL (firebirdsql)      | fb, firebird                             |
| Genji (genji)                   | gj                                       |
| Google BigQuery (bigquery)      | bq                                       |
| Google Spanner (spanner)        | sp                                       |
| Microsoft ADODB (adodb)         | ad, ado                                  |
| ModernC SQLite (moderncsqlite)  | mq, modernsqlite                         |
| ODBC (odbc)                     | od                                       |
| OLE ODBC (oleodbc)              | oo, ole, oleodbc [adodb]                 |
| Presto (presto)                 | pr, prestodb, prestos, prs, prestodbs    |
| SAP ASE (tds)                   | ax, ase, sapase                          |
| SAP HANA (hdb)                  | sa, saphana, sap, hana                   |
| Snowflake (snowflake)           | sf                                       |
| Trino (trino)                   | tr, trino, trinos, trs                   |
| Vertica (vertica)               | ve                                       |
| VoltDB (voltdb)                 | vo, volt, vdb                            |
<!-- END SCHEME TABLE -->

#### Go Drivers and Build Tags

The following are the [Go SQL drivers][go-sql] that `usql` supports, and the
associated Go build tag:

| Driver               | Build Tag     | Driver Used                                                           |
|----------------------|---------------|-----------------------------------------------------------------------|
| Microsoft SQL Server | mssql         | [github.com/denisenkom/go-mssqldb][d-mssql]                           |
| MySQL                | mysql         | [github.com/go-sql-driver/mysql][d-mysql]                             |
| Oracle Database      | oracle        | [github.com/sijms/go-ora][d-oracle]                                   |
| PostgreSQL           | postgres      | [github.com/lib/pq][d-postgres]                                       |
| SQLite3              | sqlite3       | [github.com/mattn/go-sqlite3][d-sqlite3]                              |
|                      |               |                                                                       |
| MySQL                | mymysql       | [github.com/ziutek/mymysql/godrv][d-mymysql]                          |
| Oracle Database      | godror        | [github.com/godror/godror][d-godror]                                  |
| PostgreSQL           | pgx           | [github.com/jackc/pgx/stdlib][d-pgx]                                  |
|                      |               |                                                                       |
| Alibaba MaxCompute   | maxcompute    | [sqlflow.org/gomaxcompute][d-maxcompute]                              |
| Apache Avatica       | avatica       | [github.com/Boostport/avatica][d-avatica]                             |
| Apache H2            | h2            | [github.com/jmrobles/h2go][d-avatica]                                 |
| Apache Hive          | hive          | [sqlflow.org/gohive][d-hive]                                          |
| Apache Ignite        | ignite        | [github.com/amsokol/ignite-go-client][d-ignite]                       |
| Apache Impala        | impala        | [github.com/bippio/go-impala][d-impala]                               |
| AWS Athena           | athena        | [github.com/uber/athenadriver/go][d-athena]                           |
| Azure Cosmos         | cosmos        | [github.com/btnguyen2k/gocosmos][d-cosmos]                            |
| Cassandra            | cassandra     | [github.com/MichaelS11/go-cql-driver][d-cassandra]                    |
| ClickHouse           | clickhouse    | [github.com/kshvakov/clickhouse][d-clickhouse]                        |
| Couchbase            | couchbase     | [github.com/couchbase/go_n1ql][d-couchbase]                           |
| Cznic QL             | ql            | [modernc.org/ql][d-ql]                                                |
| Firebird SQL         | firebird      | [github.com/nakagami/firebirdsql][d-firebird]                         |
| Genji                | genji         | [github.com/genjidb/genji/sql/driver][d-genji]                        |
| Google BigQuery      | bigquery      | [gorm.io/driver/bigquery/driver][d-bigquery]                          |
| Google Spanner       | spanner       | [github.com/rakyll/go-sql-driver-spanner][d-spanner]                  |
| Microsoft ADODB      | adodb         | [github.com/mattn/go-adodb][d-adodb]                                  |
| ModernC SQLite       | moderncsqlite | [modernc.org/sqlite][d-moderncsqlite]                                 |
| ODBC                 | odbc          | [github.com/alexbrainman/odbc][d-odbc]                                |
| Presto               | presto        | [github.com/prestodb/presto-go-client/presto][d-presto]               |
| SAP ASE              | tds           | [github.com/thda/tds][d-tds]                                          |
| SAP HANA             | hdb           | [github.com/SAP/go-hdb/driver][d-hdb]                                 |
| Snowflake            | snowflake     | [github.com/snowflakedb/gosnowflake][d-snowflake]                     |
| Trino                | trino         | [github.com/trinodb/trino-go-client/trino][d-trino]               |
| Vertica              | vertica       | [github.com/vertica/vertica-sql-go][d-vertica]                        |
| VoltDB               | voltdb        | [github.com/VoltDB/voltdb-client-go/voltdbclient][d-voltdb]           |
|                      |               |                                                                       |
| **MOST DRIVERS**     | most          | all drivers excluding ODBC (requires CGO and additional dependencies) |
| **ALL DRIVERS**      | all           | all drivers                                                           |


## Using

After [installing][Installing], `usql` can be used similarly to the following:

```sh
# connect to a postgres database
$ usql postgres://booktest@localhost/booktest

# connect to an oracle database
$ usql oracle://user:pass@host/oracle.sid

# connect to a postgres database and run script.sql
$ usql pg://localhost/ -f script.sql
```

#### Command-line Options

Supported command-line options:

```sh
$ usql --help
usql, the universal command-line interface for SQL databases

Usage:
  usql [OPTIONS]... [DSN]

Arguments:
  DSN                            database url

Options:
  -c, --command=COMMAND ...    run only single command (SQL or internal) and exit
  -f, --file=FILE ...          execute commands from file and exit
  -w, --no-password            never prompt for password
  -X, --no-rc                  do not read start up file
  -o, --out=OUT                output file
  -W, --password               force password prompt (should happen automatically)
  -1, --single-transaction     execute as a single transaction (if non-interactive)
  -v, --set=, --variable=NAME=VALUE ...
                               set variable NAME to VALUE
  -P, --pset=VAR[=ARG] ...     set printing option VAR to ARG (see \pset command)
  -A, --no-align               unaligned table output mode
  -F, --field-separator=TEXT   field separator for unaligned output (default, "|")
  -H, --html                   HTML table output mode
  -R, --record-separator=TEXT  record separator for unaligned output (default, \n)
  -t, --tuples-only            print rows only
  -T, --table-attr=TEXT        set HTML table tag attributes (e.g., width, border)
  -x, --expanded               turn on expanded table output
  -z, --field-separator-zero   set field separator for unaligned output to zero byte
  -0, --record-separator-zero  set record separator for unaligned output to zero byte
  -J, --json                   JSON output mode
  -C, --csv                    CSV output mode
  -V, --version                display version and exit
```

### Connecting to Databases

`usql` opens a database connection by [parsing a URL][dburl] and passing the
resulting connection string to [a database driver][Database Support]. Database
connection strings (aka "data source name" or DSNs) have the same parsing rules
as URLs, and can be passed to `usql` via command-line, or to the `\connect` or
`\c` [commands][].

Connection strings look like the following:

```txt
   driver+transport://user:pass@host/dbname?opt1=a&opt2=b
   driver:/path/to/file
   /path/to/file
```

Where the above are:

| Component          | Description                                                                          |
|--------------------|--------------------------------------------------------------------------------------|
| driver             | driver name or alias                                                                 |
| transport          | `tcp`, `udp`, `unix` or driver name <i>(for ODBC and ADODB)</i>                      |
| user               | username                                                                             |
| pass               | password                                                                             |
| host               | hostname                                                                             |
| dbname<sup>*</sup> | database name, instance, or service name/ID                                          |
| ?opt1=a&...        | additional database driver options (see respective SQL driver for available options) |
| /path/to/file      | a path on disk                                                                       |

<i><sup><b>*</b></sup> for Microsoft SQL Server, `/dbname` can be
`/instance/dbname`, where `/instance` is optional. For Oracle Database,
`/dbname` is of the form `/service/dbname` where `/service` is the service name
or SID, and `/dbname` is optional. Please see below for examples.</i>

#### Driver Aliases

`usql` supports the same driver names and aliases from the [`dburl`][dburl]
package. Most databases have at least one or more alias - please refer to the
[`dburl` documentation][dburl-schemes] for all supported aliases.

##### Short Aliases

All database drivers have a two character short form that is usually the first
two letters of the database driver. For example, `pg` for `postgres`, `my` for
`mysql`, `ms` for `mssql`, `or` for `oracle`, or `sq` for `sqlite3`.

#### Passing Driver Options

Driver options are specified as standard URL query options in the form of
`?opt1=a&obt2=b`. Please refer to the [relevant database driver's
documentation][Database Support] for available options.

#### Paths on Disk

If a URL does not have a `driver:` scheme, `usql` will check if it is a path on
disk. If the path exists, `usql` will attempt to use an appropriate database
driver to open the path.

If the specified path is a Unix Domain Socket, `usql` will attempt to open it
using the MySQL driver. If the path is a directory, `usql` will attempt to open
it using the PostgreSQL driver. If the path is a regular file, `usql` will
attempt to open the file using the SQLite3 driver.

#### Driver Defaults

As with URLs, most components in the URL are optional and many components can
be left out. `usql` will attempt connecting using defaults where possible:

```sh
# connect to postgres using the local $USER and the unix domain socket in /var/run/postgresql
$ usql pg://
```

Please see documentation for [the database driver][Database Support] you are
connecting with for more information.

### Connection Examples

The following are example connection strings and additional ways to connect to
databases using `usql`:

```sh
# connect to a postgres database
$ usql pg://user:pass@host/dbname
$ usql pgsql://user:pass@host/dbname
$ usql postgres://user:pass@host:port/dbname
$ usql pg://
$ usql /var/run/postgresql
$ usql pg://user:pass@host/dbname?sslmode=disable # Connect without SSL

# connect to a mysql database
$ usql my://user:pass@host/dbname
$ usql mysql://user:pass@host:port/dbname
$ usql my://
$ usql /var/run/mysqld/mysqld.sock

# connect to a mssql (Microsoft SQL) database
$ usql ms://user:pass@host/dbname
$ usql ms://user:pass@host/instancename/dbname
$ usql mssql://user:pass@host:port/dbname
$ usql ms://

# connect to a mssql (Microsoft SQL) database using Windows domain authentication
$ runas /user:ACME\wiley /netonly "usql mssql://host/dbname/"

# connect to a oracle database
$ usql or://user:pass@host/sid
$ usql oracle://user:pass@host:port/sid
$ usql or://

# connect to a cassandra database
$ usql ca://user:pass@host/keyspace
$ usql cassandra://host/keyspace
$ usql cql://host/
$ usql ca://

# connect to a sqlite database that exists on disk
$ usql dbname.sqlite3

# NOTE: when connecting to a SQLite database, if the "<driver>://" or
# "<driver>:" scheme/alias is omitted, the file must already exist on disk.
#
# if the file does not yet exist, the URL must incorporate file:, sq:, sqlite3:,
# or any other recognized sqlite3 driver alias to force usql to create a new,
# empty database at the specified path:
$ usql sq://path/to/dbname.sqlite3
$ usql sqlite3://path/to/dbname.sqlite3
$ usql file:/path/to/dbname.sqlite3

# connect to a adodb ole resource (windows only)
$ usql adodb://Microsoft.Jet.OLEDB.4.0/myfile.mdb
$ usql "adodb://Microsoft.ACE.OLEDB.12.0/?Extended+Properties=\"Text;HDR=NO;FMT=Delimited\""

# connect with ODBC driver (requires building with odbc tag)
$ cat /etc/odbcinst.ini
[DB2]
Description=DB2 driver
Driver=/opt/db2/clidriver/lib/libdb2.so
FileUsage = 1
DontDLClose = 1

[PostgreSQL ANSI]
Description=PostgreSQL ODBC driver (ANSI version)
Driver=psqlodbca.so
Setup=libodbcpsqlS.so
Debug=0
CommLog=1
UsageCount=1
# connect to db2, postgres databases using ODBC
$ usql odbc+DB2://user:pass@localhost/dbname
$ usql odbc+PostgreSQL+ANSI://user:pass@localhost/dbname?TraceFile=/path/to/trace.log
```

### Executing Queries and Commands

The interactive intrepreter reads queries and [meta (`\ `) commands][commands],
sending the query to the connected database:

```sh
$ usql sqlite://example.sqlite3
Connected with driver sqlite3 (SQLite3 3.17.0)
Type "help" for help.

sq:example.sqlite3=> create table test (test_id int, name string);
CREATE TABLE
sq:example.sqlite3=> insert into test (test_id, name) values (1, 'hello');
INSERT 1
sq:example.sqlite3=> select * from test;
  test_id | name
+---------+-------+
        1 | hello
(1 rows)

sq:example.sqlite3=> select * from test
sq:example.sqlite3-> \p
select * from test
sq:example.sqlite3-> \g
  test_id | name
+---------+-------+
        1 | hello
(1 rows)

sq:example.sqlite3=> \c postgres://booktest@localhost
error: pq: 28P01: password authentication failed for user "booktest"
Enter password:
Connected with driver postgres (PostgreSQL 9.6.6)
pg:booktest@localhost=> select * from authors;
  author_id |      name
+-----------+----------------+
          1 | Unknown Master
          2 | blah
          3 | aoeu
(3 rows)

pg:booktest@localhost=>
```

Commands may accept one or more parameter, and can be quoted using either `'`
or `"`. Command parameters may also be [backtick'd][backticks].

### Backslash Commands

Currently available commands:

```sh
$ usql
Type "help" for help.

(not connected)=> \?
General
  \q                              quit usql
  \copyright                      show usql usage and distribution terms
  \drivers                        display information about available database drivers
  \g [FILE] or ;                  execute query (and send results to file or |pipe)
  \gexec                          execute query and execute each value of the result
  \gset [PREFIX]                  execute query and store results in usql variables
  \gx                             as \g, but forces expanded output mode

Help
  \? [commands]                   show help on backslash commands
  \? options                      show help on usql command-line options
  \? variables                    show help on special variables

Query Buffer
  \e [FILE] [LINE]                edit the query buffer (or file) with external editor
  \p                              show the contents of the query buffer
  \raw                            show the raw (non-interpolated) contents of the query buffer
  \r                              reset (clear) the query buffer
  \w FILE                         write query buffer to file

Input/Output
  \echo [STRING]                  write string to standard output
  \i FILE                         execute commands from file
  \ir FILE                        as \i, but relative to location of current script

Formatting
  \pset [NAME [VALUE]]            set table output option
  \a                              toggle between unaligned and aligned output mode
  \C [STRING]                     set table title, or unset if none
  \f [STRING]                     show or set field separator for unaligned query output
  \H                              toggle HTML output mode
  \T [STRING]                     set HTML <table> tag attributes, or unset if none
  \t [on|off]                     show only rows
  \x [on|off|auto]                toggle expanded output

Transaction
  \begin                          begin a transaction
  \commit                         commit current transaction
  \rollback                       rollback (abort) current transaction

Connection
  \c URL                          connect to database with url
  \c DRIVER PARAMS...             connect to database with SQL driver and parameters
  \Z                              close database connection
  \password [USERNAME]            change the password for a user
  \conninfo                       display information about the current database connection

Operating System
  \cd [DIR]                       change the current working directory
  \setenv NAME [VALUE]            set or unset environment variable
  \! [COMMAND]                    execute command in shell or start interactive shell
  \timing [on|off]                toggle timing of commands

Variables
  \prompt [-TYPE] [PROMPT] <VAR>  prompt user to set variable
  \set [NAME [VALUE]]             set internal variable, or list all if no parameters
  \unset NAME                     unset (delete) internal variable
```

## Features and Compatibility

The `usql` project's goal is to support all standard `psql` commands and
features. Pull Requests are always appreciated!

#### Variables and Interpolation

`usql` supports client-side interpolation of variables that can be `\set` and
`\unset`:

```sh
$ usql
(not connected)=> \set
(not connected)=> \set FOO bar
(not connected)=> \set
FOO = 'bar'
(not connected)=> \unset FOO
(not connected)=> \set
(not connected)=>
```

A `\set` variable, `NAME`,  will be directly interpolated (by string
substitution) into the query when prefixed with `:` and optionally surrounded
by quotation marks (`'` or `"`):

```sh
pg:booktest@localhost=> \set FOO bar
pg:booktest@localhost=> select * from authors where name = :'FOO';
  author_id | name
+-----------+------+
          7 | bar
(1 rows)
```

The three forms, `:NAME`, `:'NAME'`, and `:"NAME"`, are used to interpolate a
variable in parts of a query that may require quoting, such as for a column
name, or when doing concatenation in a query:

```sh
pg:booktest@localhost=> \set TBLNAME authors
pg:booktest@localhost=> \set COLNAME name
pg:booktest@localhost=> \set FOO bar
pg:booktest@localhost=> select * from :TBLNAME where :"COLNAME" = :'FOO'
pg:booktest@localhost-> \p
select * from authors where "name" = 'bar'
pg:booktest@localhost-> \raw
select * from :TBLNAME where :"COLNAME" = :'FOO'
pg:booktest@localhost-> \g
  author_id | name
+-----------+------+
          7 | bar
(1 rows)

pg:booktest@localhost=>
```

**Note**: variables contained within other strings will **NOT** be
interpolated:

```sh
pg:booktest@localhost=> select ':FOO';
  ?column?
+----------+
  :FOO
(1 rows)

pg:booktest@localhost=> \p
select ':FOO';
pg:booktest@localhost=>
```

#### Backtick'd parameters

[Meta (`\ `) commands][commands] support backticks on parameters:

```sh
(not connected)=> \echo Welcome `echo $USER` -- 'currently:' "(" `date` ")"
Welcome ken -- currently: ( Wed Jun 13 12:10:27 WIB 2018 )
(not connected)=>
```

Backtick'd parameters will be passed to the user's `SHELL`, exactly as written,
and can be combined with `\set`:

```sh
pg:booktest@localhost=> \set MYVAR `date`
pg:booktest@localhost=> \set
MYVAR = 'Wed Jun 13 12:17:11 WIB 2018'
pg:booktest@localhost=> \echo :MYVAR
Wed Jun 13 12:17:11 WIB 2018
pg:booktest@localhost=>
```

#### Passwords

`usql` supports reading passwords for databases from a `.usqlpass` file
contained in the user's `HOME` directory at startup:

```sh
$ cat $HOME/.usqlpass
# format is:
# protocol:host:port:dbname:user:pass
postgres:*:*:*:booktest:booktest
$ usql pg://
Connected with driver postgres (PostgreSQL 9.6.9)
Type "help" for help.

pg:booktest@=>
```

Note: the `.usqlpass` file cannot be readable by other users. Please set the
permissions accordingly:

```sh
$ chmod 0600 ~/.usqlpass
```

#### Runtime Configuration (RC) File

`usql` supports executing a `.usqlrc` contained in the user's `HOME` directory:

```sh
$ cat $HOME/.usqlrc
\echo WELCOME TO THE JUNGLE `date`
\set SYNTAX_HL_STYLE paraiso-dark
$ usql
WELCOME TO THE JUNGLE Thu Jun 14 02:36:53 WIB 2018
Type "help" for help.

(not connected)=> \set
SYNTAX_HL_STYLE = 'paraiso-dark'
(not connected)=>
```

The `.usqlrc` file is read by `usql` at startup in the same way as a file
passed on the command-line with `-f` / `--file`. It is commonly used to set
startup environment variables and settings.

You can temporarily disable the RC-file by passing `-X` or `--no-rc` on the
command-line:

```sh
$ usql --no-rc pg://
```

#### Host Connection Information

By default, `usql` displays connection information when connecting to a
database. This might cause problems with some databases or connections. This
can be disabled by setting the system environment variable `USQL_SHOW_HOST_INFORMATION`
to `false`:

```sh
$ export USQL_SHOW_HOST_INFORMATION=false
$ usql pg://booktest@localhost
Type "help" for help.

pg:booktest@=>
```

`SHOW_HOST_INFORMATION` is a standard [`usql` variable][variables],
and can be `\set` or `\unset`. Additionally, it can be passed via the
command-line using `-v` or `--set`:

```sh
$ usql --set SHOW_HOST_INFORMATION=false pg://
Type "help" for help.

pg:booktest@=> \set SHOW_HOST_INFORMATION true
pg:booktest@=> \connect pg://
Connected with driver postgres (PostgreSQL 9.6.9)
pg:booktest@=>
```

#### Syntax Highlighting

Interactive queries will be syntax highlighted by default, using
[Chroma][chroma]. There are a number of [variables][] that control syntax
highlighting:

| Variable                | Default                         | Values            | Description                                                  |
|-------------------------|---------------------------------|-------------------|--------------------------------------------------------------|
| `SYNTAX_HL`             | `true`                          | `true` or `false` | enables syntax highlighting                                  |
| `SYNTAX_HL_FORMAT`      | _dependent on terminal support_ | formatter name    | [Chroma formatter name][chroma-formatter]                    |
| `SYNTAX_HL_OVERRIDE_BG` | `true`                          | `true` or `false` | enables overriding the background color of the chroma styles |
| `SYNTAX_HL_STYLE`       | `monokai`                       | style name        | [Chroma style name][chroma-style]                            |

#### Time Formatting

Some databases support time/date columns that [support formatting][go-time]. By
default, `usql` formats time/date columns as [RFC3339Nano][go-time], and can be
set using the [`TIME_FORMAT` variable][variables]:

```sh
$ usql pg://
Connected with driver postgres (PostgreSQL 9.6.9)
Type "help" for help.

pg:booktest@=> \set
TIME_FORMAT = 'RFC3339Nano'
pg:booktest@=> select now();
                now
+----------------------------------+
  2018-06-14T03:24:12.481923+07:00
(1 rows)

pg:booktest@=> \set TIME_FORMAT Kitchen
pg:booktest@=> \g
   now
+--------+
  3:24AM
(1 rows)
```

Any [Go supported time format][go-time] or const name (for example, `Kitchen`,
in the above) can be used for `TIME_FORMAT`.

## Contributing

`usql` is currently a WIP, and is aiming towards a 1.0 release soon.
Well-written PRs are always welcome -- and there is a clear backlog of issues
marked `help wanted` on the GitHub issue tracker!

<p style="text-align: center">
[*Please pick up an issue today, and submit a PR tomorrow!*][help-wanted]
</p>

## Related Projects

* [dburl][dburl] - Go package providing a standard, URL-style mechanism for parsing and opening database connection URLs
* [xo][xo] - Go command-line tool to generate Go code from a database schema

[dburl]: https://github.com/xo/dburl
[dburl-schemes]: https://github.com/xo/dburl#protocol-schemes-and-aliases
[godoc]: https://godoc.org/github.com/xo/usql
[go-project]: https://golang.org/project
[go-time]: https://golang.org/pkg/time/#pkg-constants
[go-sql]: https://golang.org/pkg/database/sql/
[homebrew]: https://brew.sh/
[xo]: https://github.com/xo/xo
[xo-tap]: https://github.com/xo/homebrew-xo
[xo-tap-notes]: https://github.com/xo/homebrew-xo#oracle-notes
[chroma]: https://github.com/alecthomas/chroma
[chroma-formatter]: https://github.com/alecthomas/chroma#formatters
[chroma-style]: https://xyproto.github.io/splash/docs/all.html
[help-wanted]: https://github.com/xo/usql/issues?q=is:open+is:issue+label:%22help+wanted%22

[commands]: #backslash-commands (Commands)
[backticks]: #backtick-d-parameters (Backtick Parameters)
[highlighting]: #syntax-highlighting (Syntax Highlighting)
[variables]: #variables-and-interpolation (Variable Interpolation)

[d-adodb]: https://github.com/mattn/go-adodb
[d-athena]: https://github.com/uber/athenadriver
[d-avatica]: https://github.com/Boostport/avatica
[d-bigquery]: https://gorm.io/driver/bigquery/driver
[d-cassandra]: https://github.com/MichaelS11/go-cql-driver
[d-clickhouse]: https://github.com/kshvakov/clickhouse
[d-cosmos]: https://github.com/btnguyen2k/gocosmos
[d-couchbase]: https://github.com/couchbase/go_n1ql
[d-firebird]: https://github.com/nakagami/firebirdsql
[d-genji]: https://github.com/genjidb/genji
[d-godror]: https://github.com/godror/godror
[d-hdb]: https://github.com/SAP/go-hdb
[d-hive]: https://sqlflow.org/gohive
[d-ignite]: https://github.com/amsokol/ignite-go-client
[d-impala]: https://github.com/bippio/go-impala
[d-maxcompute]: https://sqlflow.org/gomaxcompute
[d-moderncsqlite]: https://modernc.org/sqlite
[d-mssql]: https://github.com/denisenkom/go-mssqldb
[d-mymysql]: https://github.com/ziutek/mymysql
[d-mysql]: https://github.com/go-sql-driver/mysql
[d-odbc]: https://github.com/alexbrainman/odbc
[d-oracle]: https://github.com/sijms/go-ora
[d-pgx]: https://github.com/jackc/pgx
[d-postgres]: https://github.com/lib/pq
[d-presto]: https://github.com/prestodb/presto-go-client
[d-ql]: https://modernc.org/ql
[d-snowflake]: https://github.com/snowflakedb/gosnowflake
[d-spanner]: https://github.com/rakyll/go-sql-driver-spanner
[d-sqlago]: https://github.com/a-palchikov/sqlago
[d-sqlite3]: https://github.com/mattn/go-sqlite3
[d-tds]: https://github.com/thda/tds
[d-trino]: https://github.com/trinodb/trino-go-client
[d-vertica]: https://github.com/vertica/vertica-sql-go
[d-voltdb]: https://github.com/VoltDB/voltdb-client-go
