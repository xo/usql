# usql ![Build Status][travis-ci]

A universal command-line interface for PostgreSQL, MySQL, Oracle Database,
SQLite3, Microsoft SQL Server, [and many other databases][Database Support]
including NoSQL and non-relational databases!

[travis-ci]: https://travis-ci.org/xo/usql.svg?branch=master (https://travis-ci.org/xo/usql)

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

`usql` can be installed by [via Release][], [via Homebrew][], or [via Go][]:

[via Release]: #installing-via-release
[via Homebrew]: #installing-via-homebrew-macos
[via Go]: #installing-via-go

### Installing via Release

1. [Download a release for your platform][Releases]
2. Extract the `usql` or `usql.exe` file from the `.tar.bz2` or `.zip` file
3. Move the extracted executable to somewhere on your `$PATH` (Linux/macOS) or
   `%PATH%` (Windows)

### Installing via Homebrew (macOS)

`usql` is available in the [`xo/xo` tap][xo-tap], and can be installed in the
usual way with the [`brew` command][homebrew]:

```sh
# add tap
$ brew tap xo/xo

# install usql with "most" drivers
$ brew install usql
```

Additional support for [Oracle and ODBC databases][Database Support] can be
installed by passing `--with-*` parameters during install:

```sh
# install usql with oracle and odbc support
$ brew install --with-oracle --with-odbc usql
```

Please note that Oracle support requires using the [`xo/xo` tap's][xo-tap]
`instantclient-sdk` formula. Any other `instantclient-sdk` formulae or older
versions of the Oracle Instant Client SDK [should be uninstalled][xo-tap-notes]
prior to attempting the above:

```sh
# uninstall the instantclient-sdk formula
$ brew uninstall InstantClientTap/instantclient/instantclient-sdk

# remove conflicting tap
$ brew untap InstantClientTap/instantclient
```

### Installing via Go

`usql` can be installed in the usual Go fashion:

```sh
# install usql with basic database support (includes PosgreSQL, MySQL, SQLite3, and MS SQL drivers)
$ go get -u github.com/xo/usql
```

Support for additional databases can be specified with [build tags][Database Support]:

```sh
# install usql with most drivers (excludes drivers requiring CGO)
$ go get -u -tags most github.com/xo/usql

# install usql with all drivers (includes drivers requiring CGO, namely Oracle and ODBC drivers)
$ go get -u -tags all github.com/xo/usql
```

## Building

When building `usql` with [Go][go-project], only drivers for PostgreSQL, MySQL,
SQLite3 and Microsoft SQL Server will be enabled by default. Other databases
can be enabled by specifying the build tag for their [database driver][Database Support].
Additionally, the `most` and `all` build tags include most, and all SQL
drivers, respectively:

```sh
# install all drivers
$ go get -u -tags all github.com/xo/usql

# install with most drivers (same as all but excludes Oracle/ODBC)
$ go get -u -tags most github.com/xo/usql

# install with base drivers and Oracle/ODBC support
$ go get -u -tags 'oracle odbc' github.com/xo/usql
```

For every build tag `<driver>`, there is also the `no_<driver>` build tag
disabling the driver:

```sh
# install all drivers excluding avatica and couchbase
$ go get -u -tags 'all no_avatica no_couchbase' github.com/xo/usql
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

The databases supported, the respective build tag, and the driver used by `usql` are:

| Driver               | Build Tag  | Driver Used                                                                      |
|----------------------|------------|----------------------------------------------------------------------------------|
| Microsoft SQL Server | mssql      | [github.com/denisenkom/go-mssqldb][d-mssql]                                      |
| MySQL                | mysql      | [github.com/go-sql-driver/mysql][d-mysql]                                        |
| PostgreSQL           | postgres   | [github.com/lib/pq][d-postgres]                                                  |
| SQLite3              | sqlite3    | [github.com/mattn/go-sqlite3][d-sqlite3]                                         |
| Oracle               | oracle     | [gopkg.in/rana/ora.v4][d-oracle]                                                 |
|                      |            |                                                                                  |
| MySQL                | mymysql    | [github.com/ziutek/mymysql/godrv][d-mymysql]                                     |
| PostgreSQL           | pgx        | [github.com/jackc/pgx/stdlib][d-pgx]                                             |
|                      |            |                                                                                  |
| Apache Avatica       | avatica    | [github.com/Boostport/avatica][d-avatica]                                        |
| Cassandra            | cassandra  | [github.com/MichaelS11/go-cql-driver][d-cassandra]                               |
| ClickHouse           | clickhouse | [github.com/kshvakov/clickhouse][d-clickhouse]                                   |
| Couchbase            | couchbase  | [github.com/couchbase/go_n1ql][d-couchbase]                                      |
| Cznic QL             | ql         | [github.com/cznic/ql][d-ql]                                                      |
| Firebird SQL         | firebird   | [github.com/nakagami/firebirdsql][d-firebird]                                    |
| Microsoft ADODB      | adodb      | [github.com/mattn/go-adodb][d-adodb]                                             |
| ODBC                 | odbc       | [github.com/alexbrainman/odbc][d-odbc]                                           |
| Presto               | presto     | [github.com/prestodb/presto-go-client/presto][d-presto]                          |
| SAP HANA             | hdb        | [github.com/SAP/go-hdb/driver][d-hdb]                                            |
| VoltDB               | voltdb     | [github.com/VoltDB/voltdb-client-go/voltdbclient][d-voltdb]                      |
|                      |            |                                                                                  |
| Google Spanner       | spanner    | github.com/xo/spanner (not yet public)                                           |
|                      |            |                                                                                  |
| **MOST DRIVERS**     | most       | all drivers excluding Oracle and ODBC (requires CGO and additional dependencies) |
| **ALL DRIVERS**      | all        | all drivers                                                                      |


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
usql, the universal command-line interface for SQL databases.

usql 0.7.0
Usage: usql [--command COMMAND] [--file FILE] [--output OUTPUT] [--username USERNAME] [--password] [--no-password] [--no-rc] [--single-transaction] [--set SET] DSN

Positional arguments:
  DSN                    database url

Options:
  --command COMMAND, -c COMMAND
                         run only single command (SQL or internal) and exit
  --file FILE, -f FILE   execute commands from file and exit
  --output OUTPUT, -o OUTPUT
                         output file
  --username USERNAME, -U USERNAME
                         database user name [default: ken]
  --password, -W         force password prompt (should happen automatically)
  --no-password, -w      never prompt for password
  --no-rc, -X            do not read start up file
  --single-transaction, -1
                         execute as a single transaction (if non-interactive)
  --set SET, -v SET      set variable NAME=VALUE
  --help, -h             display this help and exit
  --version              display version and exit
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

| Component          | Description                                                               |
|--------------------|---------------------------------------------------------------------------|
| driver             | driver name or alias                                                      |
| transport          | `tcp`, `udp`, `unix` or driver name <i>(for ODBC and ADODB)</i>           |
| user               | username                                                                  |
| pass               | password                                                                  |
| host               | hostname                                                                  |
| dbname<sup>*</sup> | database name, instance, or service name/ID                               |
| ?opt1=a&...        | database driver options (see respective SQL driver for available options) |
| /path/to/file      | a path on disk                                                            |

<i><sup><b>*</b></sup> for Microsoft SQL Server, the syntax to supply an
instance and database name is `/instance/dbname`, where `/instance` is
optional. For Oracle databases, `/dbname` is the unique database ID (SID).</i>

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
  \q                    quit usql
  \copyright            show usql usage and distribution terms
  \drivers              display information about available database drivers
  \g [FILE] or ;        execute query (and send results to file or |pipe)
  \gexec                execute query and execute each value of the result
  \gset [PREFIX]        execute query and store results in usql variables

Help
  \? [commands]         show help on backslash commands
  \? options            show help on usql command-line options
  \? variables          show help on special variables

Query Buffer
  \e [FILE] [LINE]      edit the query buffer (or file) with external editor
  \p                    show the contents of the query buffer
  \raw                  show the raw (non-interpolated) contents of the query buffer
  \r                    reset (clear) the query buffer
  \w FILE               write query buffer to file

Input/Output
  \echo [STRING]        write string to standard output
  \i FILE               execute commands from file
  \ir FILE              as \i, but relative to location of current script

Transaction
  \begin                begin a transaction
  \commit               commit current transaction
  \rollback             rollback (abort) current transaction

Connection
  \c URL                connect to database with url
  \c DRIVER PARAMS...   connect to database with SQL driver and parameters
  \Z                    close database connection
  \password [USERNAME]  change the password for a user
  \conninfo             display information about the current database connection

Operating System
  \cd [DIR]             change the current working directory
  \setenv NAME [VALUE]  set or unset environment variable
  \! [COMMAND]          execute command in shell or start interactive shell

Variables
  \prompt [TEXT] NAME   prompt user to set internal variable
  \set [NAME [VALUE]]   set internal variable, or list all if no parameters
  \unset NAME           unset (delete) internal variable
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

```
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
$ ./usql pg://
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

## TODO

`usql` aims to eventually provide a drop-in replacement for PostgreSQL's `psql`
command. This is on-going -- an attempt has been made in good-faith to provide
support for the most frequently used aspects/features of `psql`. Compatability
(where possible) with `psql`, takes general development priority.

##### General

0.  updated asciinema demo
1.  support more prompt configuration, colored prompt by default
2.  add window title / status output
2.  change `drivers.Convert*` to drivers.Marshal style interfaces
3.  allow configuration for JSON encoding/decoding output
4.  return single 'driver' type handling marshaling / scanning of types / columns
5.  implement a table writer that follows "optional func" parameter style, is streaming /
    handles marshalers, can handle the different configuration options for `\pset`
6.  implement "extended" display for queries (for `\gx` / formatting)
7.  implement better environment variable handling
8.  implement proper readline
9.  tab-completion of queries
10. show hidden (client) queries (`\set SHOW_HIDDEN`)
11. fix multiline behavior to mimic `psql` properly (on arrow up/down through history)
12. proper `PAGER` support
13. `\qecho` + `\o` support
14. context-based completion (WIP)
15. full `\if` `\elif` `\else` `\endif` support
16. fix `WITH ... DELETE` queries (postgresql)
17. better `--help` support/output cli, man pages
18. translations
16. `\encoding` and environment/command line options to set encoding of input (to
    convert to UTF-8 before feeding to SQL driver) (how important is this ... ?)

##### Command Processing + `psql` compatibility

1. formatting settings (`\pset`, `\a`, etc)
2. all `\d*` commands from `psql` (WIP, need to finish work extracting introspection code from `xo`)
3. `\ef` and `\ev` commands from `psql` (WIP, need to finish work extracting stored procs / funcs / views for all the major databases)
3. `\watch`
4. `\errverbose` (show verbose info for last error)
5. remaining `psql` cli parameters
6. `\j*` commands (WIP)
7. `\copy` (add support for copying between two different databases ...?)

###### Low priority compatibity fixes:

1. correct operation of interweaved `-f`/`-c` commands, ie: `usql -f 1 -c 1 -c 2 -f 2 -f 3 -c 3` runs in the specified order

##### Testing

1. test suite for databases, doing minimal of `SELECT`, `INSERT`, `UPDATE`, `DELETE` for every database

##### Future Database Support

1. Redis CLI
2. Native Oracle
3. InfluxDB
4. CSV via SQLite3 vtable
5. Google Spanner
6. Google Sheets via SQLite3 vtable
7. [Charlatan][d-charlatan]
8. InfluxDB IQL
9. Aerospike AQL
10. ArrangoDB AQL
11. OrientDB SQL
12. Cypher / SparQL
13. Atlassian JIRA JQL

## Related Projects

* [dburl][dburl] - Go package providing a standard, URL-style mechanism for parsing and opening database connection URLs
* [xo][xo] - Go command-line tool to generate Go code from a database schema

[dburl]: https://github.com/xo/dburl
[dburl-schemes]: https://github.com/xo/dburl#protocol-schemes-and-aliases
[godoc]: https://godoc.org/github.com/xo/usql
[go-project]: https://golang.org/project
[go-time]: https://golang.org/pkg/time/#pkg-constants
[homebrew]: https://brew.sh/
[xo]: https://github.com/xo/xo
[xo-tap]: https://github.com/xo/homebrew-xo
[xo-tap-notes]: https://github.com/xo/homebrew-xo#oracle-notes
[chroma]: https://github.com/alecthomas/chroma
[chroma-formatter]: https://github.com/alecthomas/chroma#formatters
[chroma-style]: https://xyproto.github.io/splash/docs/all.html

[commands]: #backslash-commands (Commands)
[backticks]: #backtick-d-parameters (Backtick Parameters)
[highlighting]: #syntax-highlighting (Syntax Highlighting)
[variables]: #variables-and-interpolation (Variable Interpolation)

[d-mssql]: https://github.com/denisenkom/go-mssqldb
[d-mysql]: https://github.com/go-sql-driver/mysql
[d-postgres]: https://github.com/lib/pq
[d-sqlite3]: https://github.com/mattn/go-sqlite3
[d-oracle]: https://gopkg.in/rana/ora.v4
[d-mymysql]: https://github.com/ziutek/mymysql
[d-pgx]: https://github.com/jackc/pgx
[d-avatica]: https://github.com/Boostport/avatica
[d-cassandra]: https://github.com/MichaelS11/go-cql-driver
[d-clickhouse]: https://github.com/kshvakov/clickhouse
[d-couchbase]: https://github.com/couchbase/go_n1ql
[d-ql]: https://github.com/cznic/ql
[d-firebird]: https://github.com/nakagami/firebirdsql
[d-adodb]: https://github.com/mattn/go-adodb
[d-odbc]: https://github.com/alexbrainman/odbc
[d-presto]: https://github.com/prestodb/presto-go-client
[d-hdb]: https://github.com/SAP/go-hdb
[d-sqlago]: https://github.com/a-palchikov/sqlago
[d-voltdb]: https://github.com/VoltDB/voltdb-client-go
[d-spanner]: https://github.com/xo/spanner
[d-charlatan]: github.com/BatchLabs/charlatan
