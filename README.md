# About usql

usql is a universal command-line interface for working with SQL databases.

usql provides a universal command line interface for the following databases:
PostgreSQL, MySQL, Oracle, SQLite, and Microsoft SQL Server.

The goal is to eventually have usql be a drop in replacement for PostgreSQL's
`psql` command, with all the bells/whistles, but with the added benefit of
working with more than one database.

## Installing

Install in the usual Go way:

```sh
# install usql
$ go get -u github.com/knq/usql

# install with oracle support
$ go get -u -tags oracle github.com/knq/usql
```

## Using

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

# connect to a oracle database
$ usql or://user:pass@localhost/dbname
$ usql oracle://user:pass@localhost:port/dbname

# connect to a sqlite file
$ usql dbname.sqlite3
$ usql sq://path/to/dbname.sqlite3
$ usql sqlite3://path/to/dbname.sqlite3
$ usql file:/path/to/dbname.sqlite3
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

# TODO
* Better handling of local files (such as unix domain sockets) for Handler.Open
* All the various \\d* commands from `psql`
* SQL completion
