# About usql

`usql` is a universal command-line interface for PostgreSQL, MySQL, Oracle,
SQLite3, Microsoft SQL Server, [and other databases](#database_support).

#### [Quickstart][] | [Demo][] | [Database Support][] | [Connection Strings][] | [Commands][] | [Building][] | [Compatibility][] | [Related][] | [Releases][]

[Quickstart]: #quickstart (Quickstart)
[Demo]: #interactive-demo (Demonstration)
[Database Support]: #database-support (Database Support)
[Connection Strings]: #database-connection-strings (Database Connection Strings)
[Commands]: #backslash--commands (Commands)
[Building]: #build-install (Build/Install from Source)
[Compatibility]: #compatibility-and-todo (Compatibility with PostgreSQL psql and TODO)
[Related][]: #related-projects (Related Projects)
[Releases]: https://github.com/knq/usql/releases (Project Releases)

## Quickstart

1. [Download a release for your platform](https://github.com/knq/usql/releases)
2. Extract the `.zip` (Windows), or `.tar.bz2` (OS X/Linux) file and place the
   `usql` executable somewhere on your `%PATH%` (Windows), or your `$PATH` (OS X/Linux)
3. Connect to a database using `usql driver://user:pass@host/dbname`, and
   execute a SQL query, or [command](#commands):

```sh
$ usql postgres://booktest@localhost/booktest
error: pq: 28P01: password authentication failed for user "booktest"
Password:
You are connected with driver postgres (PostgreSQL 9.6.2)
Type "help" for help.

pg:booktest@localhost/booktest=> select * from books;
  book_id | author_id | isbn | booktype |      title       | year |            available            |      tags
+---------+-----------+------+----------+------------------+------+---------------------------------+-----------------+
        1 |         1 |    1 | FICTION  | asotenhuastonehu | 2016 | 2017-03-19T02:48:36.27928+07:00 | {}
        2 |         1 |    2 | FICTION  | asotenhuastonehu | 2016 | 2017-03-19T02:48:36.27928+07:00 | {cool,disastor}
        3 |         1 |    3 | FICTION  | asotenhuastonehu | 2001 | 2017-03-19T02:48:36.27928+07:00 | {cool}
(3 rows)

pg:booktest@localhost/booktest=> \p
select * from books;
pg:booktest@localhost/booktest=> \g
  book_id | author_id | isbn | booktype |      title       | year |            available            |      tags
+---------+-----------+------+----------+------------------+------+---------------------------------+-----------------+
        1 |         1 |    1 | FICTION  | asotenhuastonehu | 2016 | 2017-03-19T02:48:36.27928+07:00 | {}
        2 |         1 |    2 | FICTION  | asotenhuastonehu | 2016 | 2017-03-19T02:48:36.27928+07:00 | {cool,disastor}
        3 |         1 |    3 | FICTION  | asotenhuastonehu | 2001 | 2017-03-19T02:48:36.27928+07:00 | {cool}
(3 rows)

pg:booktest@localhost/booktest=> \q
```

Alternatively, if you already have a [working Go build environment](https://golang.org/doc/install),
you may [install directly](#build_install) in the usual Go fashion:

```sh
# install usql with most SQL drivers
$ go get -u -tags most github.com/knq/usql
```

## Interactive Demo

The below is a demonstration using `usql` with [xo's booktest](https://github.com/knq/xo)
simple test database, showcasing a early version of `usql`. In the
demonstration, `usql` connects to an Oracle database instance, performs some
basic queries, and subsequently connects to PostgreSQL, MySQL, Microsoft SQL
Server, and SQLite3 databases, and similarly executes the same basic queries
against each database.

<p align="center">
  <a href="https://asciinema.org/a/73gxbg62ny2fx9ppxu0kd8c48" target="_blank">
    <img src="https://asciinema.org/a/73gxbg62ny2fx9ppxu0kd8c48.png" width="654"/>
  </a>
</p>

Please note the above was recorded using the an early release version of
`usql`, and since then, `usql` has become significantly more robust and
feature-compatible with PostgreSQL's `psql`.

## Database Support

`usql` aims to provide support for all Go standard library compatible SQL
drivers -- with an emphasis on supporting the drivers that sister project,
[`dburl`](https://github.com/knq/dburl), provides "out-of-the-box" URL support
for.

The databases currently supported by `usql` (and related build tag name) are
summarized below:

| Drivers              | Build Tag            | Driver Package                                                                                                                 |
|----------------------|----------------------|--------------------------------------------------------------------------------------------------------------------------------|
| Microsoft SQL Server | mssql<sup>*</sup>    | [github.com/denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb)                                                   |
| MySQL                | mysql<sup>*</sup>    | [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)                                                       |
| PostgreSQL           | postgres<sup>*</sup> | [github.com/lib/pq](https://github.com/lib/pq)                                                                                 |
| SQLite3              | sqlite3<sup>*</sup>  | [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)                                                             |
| Oracle               | oracle               | [gopkg.in/rana/ora.v4](https://gopkg.in/rana/ora.v4)                                                                           |
| MySQL                | mymysql              | [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql)                                                           |
| PostgreSQL           | pgx                  | [github.com/jackc/pgx/stdlib](https://github.com/jackc/pgx)                                                                    |
|                      |                      |                                                                                                                                |
| Apache Avatica       | avatica              | [github.com/Boostport/avatica](https://github.com/Boostport/avatica)                                                           |
| ClickHouse           | clickhouse           | [github.com/kshvakov/clickhouse](https://github.com/kshvakov/clickhouse)                                                       |
| Couchbase            | couchbase            | [github.com/couchbase/go_n1ql](https://github.com/couchbase/go_n1ql)                                                           |
| Cznic QL             | ql                   | [github.com/cznic/ql](https://github.com/cznic/ql)                                                                             |
| Firebird SQL         | firebird             | [github.com/nakagami/firebirdsql](https://github.com/nakagami/firebirdsql)                                                     |
| Microsoft ADODB      | adodb                | [github.com/mattn/go-adodb](https://github.com/mattn/go-adodb)                                                                 |
| ODBC                 | odbc                 | [github.com/alexbrainman/odbc](https://github.com/alexbrainman/odbc)                                                           |
| SAP HANA             | hdb                  | [github.com/SAP/go-hdb/driver](https://github.com/SAP/go-hdb)                                                                  |
| Sybase SQL Anywhere  | sqlany               | [github.com/a-palchikov/sqlago](https://github.com/a-palchikov/sqlago)                                                         |
| VoltDB               | voltdb               | [github.com/VoltDB/voltdb-client-go/voltdbclient](github.com/VoltDB/voltdb-client-go])                                         |
| YQL                  | yql                  | [github.com/mattn/go-yql](https://github.com/mattn/go-yql)                                                                     |
|                      |                      |                                                                                                                                |
| Google Spanner       | spanner              | github.com/knq/spanner (not yet public)                                                                                        |
|                      |                      |                                                                                                                                |
| **MOST DRIVERS**     | most                 | (all drivers listed above, excluding the drivers for Oracle and ODBC, which require third-party dependencies to build/install) |
| **ALL DRIVERS**      | all                  | (all drivers listed above)                                                                                                     |

<i><sup>*</sup>included by default when building</i>

## Database Connection Strings

Database connection strings, or "data source name" (aka DSNs), used with `usql`
have the same parsing rules as a normal URL, and have the following two forms:

```
   protocol+transport://user:pass@host/dbname?opt1=a&opt2=b
   protocol:/path/to/file
```

Where:

| Component          | Description                                                                          |
|--------------------|--------------------------------------------------------------------------------------|
| protocol           | driver name or alias (see below)                                                     |
| transport          | "tcp", "udp", "unix" or driver name <i>(for ODBC connections)</i>                    |
| user               | username                                                                             |
| pass               | password                                                                             |
| host               | host                                                                                 |
| dbname<sup>*</sup> | database, instance, or service name/ID to connect to                                 |
| ?opt1=...          | additional database driver options (see respective SQL driver for available options) |

<i><sup><b>*</b></sup> for Microsoft SQL Server, the syntax to supply an
instance and database name is `/instance/dbname`, where `/instance` is
optional. For Oracle databases, `/dbname` is the unique database ID (SID).</i>

Additionally, if `usql` is passed a URL without a leading `scheme://`, `usql` will
attempt to locate the path on disk, and if it exists will open it accordingly.
Specifically, if `usql` finds a Unix Domain Socket, it will attempt to open it
using the `mysql` driver, or when a directory is found, `usql` will attempt to
open the path using the `postgres` driver; last, if it the path is a regular
file, `usql` will attempt to open the file using the `sqlite3` driver.

`usql` recognizes the same drivers and scheme aliases from the [`dburl`](https://github.com/knq/dburl)
package. Please see the `dburl` documentation for more in-depth information on
how DSNs are built from standard URLs. Additionally, all of the above formats
can be used in conjuction with the `\c` (or `\connect`) backslash meta command.

### Example Connection Strings

The following are example connection strings (DSNs) and some additional ways to
connect to databases with `usql`:

```sh
# connect to a postgres database
$ usql pg://user:pass@localhost/dbname
$ usql pgsql://user:pass@localhost/dbname
$ usql postgres://user:pass@localhost:port/dbname

# connect to a mysql database
$ usql my://user:pass@localhost/dbname
$ usql mysql://user:pass@localhost:port/dbname
$ usql /var/run/mysqld/mysqld.sock

# connect to a mssql (Microsoft SQL) database
$ usql ms://user:pass@localhost/dbname
$ usql mssql://user:pass@localhost:port/dbname

# connect using Windows domain authentication to a mssql (Microsoft SQL)
# database
$ runas /user:ACME\wiley /netonly "usql mssql://host/dbname/"

# connect to a oracle database
$ usql or://user:pass@localhost/dbname
$ usql oracle://user:pass@localhost:port/dbname

# connect to a pre-existing sqlite database
$ usql dbname.sqlite3

# note: when not using a "<scheme>://" or "<scheme>:" prefix, the file must already
# exist; if it doesn't, please prefix with file:, sq:, sqlite3: or any other
# scheme alias recognized by the dburl package for sqlite databases, and sqlite
# will create a new database, like the following:
$ usql sq://path/to/dbname.sqlite3
$ usql sqlite3://path/to/dbname.sqlite3
$ usql file:/path/to/dbname.sqlite3

# connect to a adodb ole resource (windows only)
$ usql adodb://Microsoft.Jet.OLEDB.4.0/myfile.mdb
$ usql "adodb://Microsoft.ACE.OLEDB.12.0/?Extended+Properties=\"Text;HDR=NO;FMT=Delimited\""
```

## Backslash (`\`) Commands

The following are the currently supported backslash (`\ `) meta commands
available to interactive `usql` sessions or to included (ie, `\i`) scripts:

```txt
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
  \r                    reset (clear) the query buffer
  \w FILE               write query buffer to file

Input/Output
  \echo [STRING]        write string to standard output
  \i FILE               execute commands from file
  \ir FILE              as \i, but relative to location of current script

Connection
  \c URL                connect to database with url
  \c DRIVER PARAMS...   connect to database with SQL driver and parameters
  \Z                    close database connection
  \conninfo             display information about the current database connection

Operating System
  \cd [DIR]             change the current working directory
  \setenv NAME [VALUE]  set or unset environment variable
```

More coming soon!

In time, the aim of the `usql` project is to provide out-of-the-box support for
most of `psql`'s backslash meta commands. As of now, this is a continuing, and
on-going effort -- and a substantial, good-faith attempt has been made to
provide support for the most frequently used `psql` meta commands -- note,
however that `usql` is not close to a 100% replacement/drop-in compatible with
`psql`. ***CAVEAT USER***.

## Build/Install from Source

You can build or install `usql` from source in the usual Go fashion:

```sh
# install usql (includes support for PosgreSQL, MySQL, SQLite3, and MS SQL)
$ go get -u github.com/knq/usql
```

Please note that default calls to `go get`, `go build`, or `go install` will
only include drivers for PostgreSQL, MySQL, SQLite3 and Microsoft SQL Server.

If you need additional support for a database driver (or wish to disable a
specific driver), you may use additional build tags with `go get`, `go build`,
or `go install`. Please refer to the table in the [Database Support][] section
above for the names of the various build tags.

Note that for every build tag `<name>`, there is an additional tag `no_<name>`,
that disables the respective driver(s). Additionally, there are the build tags
`most` and `all`, that include most, and all SQL drivers, respectively.

As such, you can easily (and quickly) recompile `usql` with by combining any
number of build tags to enable/disable specific drivers as needed:

```sh
# install all drivers
$ go get -u -tags all github.com/knq/usql

# install with "most" drivers (same as "all" but excludes oracle/odbc)
$ go get -u -tags most github.com/knq/usql

# install with base drivers and oracle / odbc support
$ go get -u -tags 'oracle odbc' github.com/knq/usql

# install all drivers but exclude avatica, and couchbase drivers
$ go get -u -tags 'all no_avatica no_couchbase'
```

For reference, [`usql` releases](https://github.com/knq/usql/releases) are
built with the `most` tag, and with [additional SQLite3 specific build tags](contrib/build-release.sh).

### Using `usql` as a library

Significant effort has gone into making `usql`'s codebase modular, and reusable
by other developers wishing to leverage the existing features of `usql`. As
such, if you would like to build your own SQL command-line interface (e.g, for
use with a SQL-like project, or otherwise as an "official" client), it is
relatively straight-forward and easy to do so.

Please refer to the [main command-line entry point](main.go) to see how `usql`
uses its constituent packages to create a interactive command-line
handler/interpreter. Additionally, `usql`'s code is fairly well-documented --
please refer to the [GoDoc listing](https://godoc.org/github.com/knq/usql) to
see how it's all put together.

## Compatibility and TODO

The goal of the `usql` project is to eventually provide a drop-in replacement
for the amazing PostgreSQL's `psql` command -- including all bells/whistles --
but with the added benefit of working with practically any database.

Eventually, `usql` developers hope to leverage the power of Go and have plans for
more features than the base `psql` command provides.

#### TODO

A list of planned / in progress work:

##### General
1. .usqlpass file (ie .psqlpass)
2. Fix meta command parsing when passed a quoted string ie, \echo "   foo
   bar  " should have all whitespace included in the parameter
3. fix table output / formatting
4. Transaction wrapping / starts/commits / "-1" one transaction stuff
5. pager + pipe / gexec/gset support
6. SQL variables + environment
7. add support for managing multiple database connections simultaneously
    (@conn syntax, and a ~/.usqlconnections file, and ~/.usqlconfig) (maybe not
    needed, if variable support works "as expected"?)
8. SQL completion (WIP)
9. syntax highlighting (WIP)
10. \encoding and environment/command line options to set encoding of input (to
    convert to utf-8 before feeding to SQL driver) (how important is this ... ?)
11. better --help support/output cli, man pages

##### Not important / "Nice to haves":
1. correct operation of interweaved -f/-c commands, ie: -f 1 -c 1 -c 2 -f 2 -f 3 -c 3 runs in the specified order

##### Command Processing + `psql` compatibility
* PAGER + EDITOR support (WIP)
* variable support / interpolation + \prompt, \set, \unset
* the \j* commands (WIP)
* \watch
* \errverbose
* formatting settings (\pset, \a, etc)
* all \\d* commands from `psql` (WIP, need to finish work extracting introspection code from `xo`)
* remaining `psql` cli parameters

##### Releases

Need to write scripts for packaging and build binaries for:

* Debian/Ubuntu (.deb)
* MacOS X (.pkg)
* Windows (.msi)
* CentOS/RHEL (.rpm)

Additional:
* Submit upstream to Debian unstable (WIP)

##### Testing

* full test suite for databases, doing a minimal set of SELECT, INSERT, UPDATE, DELETE

##### Future Database Support

Notes / thoughts / comments on adding support for various "databases":

* Google Spanner
* Cassandra
* Atlassian JIRA JQL (why not? lol)

## Related Projects

* [dburl](https://github.com/knq/dburl) - a Go package providing a standard, URL style mechanism for parsing and opening database connection URLs
* [xo](https://github.com/knq/xo) - a command-line tool to generate Go code from a database schema
