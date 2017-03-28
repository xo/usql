# About usql

usql is a universal command-line interface for working with SQL databases.

usql provides a universal command-line interface for the following databases:
PostgreSQL, MySQL, Oracle, SQLite, Microsoft SQL Server, Microsoft ADODB
(Windows only), and others.

The goal is to eventually have usql be a drop in replacement for PostgreSQL's
`psql` command, with all the bells/whistles, but with the added benefit of
working with multiple databases.

#### [Releases](https://github.com/knq/usql/releases)

## Installing

Install in the usual Go way:

```sh
# install usql
$ go get -u github.com/knq/usql

# install with oracle support
$ go get -u -tags oracle github.com/knq/usql

# install with oracle + adodb support (windows only)
$ go get -u -tags 'oracle adodb' github.com/knq/usql
```

Alternatively, you can download a binary release for your platform from the
[GitHub releases page](https://github.com/knq/usql/releases).

## Using

`usql` makes use of the [`dburl`](https://github.com/knq/dburl) package for
opening URLs. Almost every database recognized by `dburl` can be opened by
`usql`.  Some example ways to connect to a database:

```sh
# display command line arguments
$ usql --help

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

# connect to a sqlite file
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

## Example Output

The following is an example of connecting to [xo's booktest](https://github.com/knq/xo)
example Oracle database, performing a query, and then connecting to the
PostgreSQL, MySQL, Microsoft SQL Server, and SQLite3 databases and executing
various queries.

<p align="center">
  <a href="https://asciinema.org/a/73gxbg62ny2fx9ppxu0kd8c48" target="_blank">
    <img src="https://asciinema.org/a/73gxbg62ny2fx9ppxu0kd8c48.png" width="654"/>
  </a>
</p>

## Related Projects

* [dburl](https://github.com/knq/dburl) - a Go package providing a standard, URL style mechanism for parsing and opening database connection URLs
* [xo](https://github.com/knq/xo) - a command-line tool to generate Go code from a database schema

## TODO

A list of planned / in progress work:

### General
1. Show remote server version on connect
2. Fix meta command parsing when passed a quoted string ie, \echo "   foo
   bar  " should have all whitespace included in the parameter
3. fix table output
4. Transaction wrapping / starts/commits / "-1" one transaction stuff
5. pager + pipe / gexec/gset support
6. .usqlpass file (same as .psqlpass)
7. SQL variables + environment
8. Proper table formatting + \pset
9. .usqlrc
10. More command line options
11. add support for managing multiple database connections simultaneously
    (@conn syntax, and a ~/.usqlconnections file, and ~/.usqlconfig) (maybe not
    needed, if variable support works "as expected"?)
15. SQL completion (WIP)
16. syntax highlighting (WIP)
17. \encoding and environment/command line options to set encoding of input (to
    convert to utf-8 before feeding to SQL driver)

#### Not important / "Nice to haves":
1. correct operation of interweaved -f/-c commands, ie: -f 1 -c 1 -c 2 -f 2 -f 3 -c 3 runs in the specified order

### Command Processing + `psql` compatibility
* PAGER + EDITOR support (WIP)
* variable support / interpolation + \prompt, \set, \unset
* the \j* commands (WIP)
* \watch
* \errverbose
* formatting settings (\pset, \a, etc)
* all \\d* commands from `psql` (WIP, need to finish work extracting introspection code from `xo`)
* remaining `psql` cli parameters

### Releases

Need to write scripts for packaging and build binaries for:

* Debian/Ubuntu (.deb)
* MacOS X (.pkg)
* Windows (.msi)
* CentOS/RHEL (.rpm)

Additional:
* Submit upstream to Debian unstable (WIP)

### Testing

* full test suite for databases, doing a minimal set of SELECT, INSERT, UPDATE, DELETE

### Future Database Support

Notes / thoughts / comments on adding support for various "databases":

* Google Spanner
* Cassandra
* Atlassian JIRA JQL (why not? lol)
