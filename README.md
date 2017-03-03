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

# connect to a sqlite file
$ usql file:dbname.sqlite3
```

## Example Output

The following is an example of connecting to [xo's booktest](https://github.com/knq/xo)
example Oracle, performing a query, and then connecting to the PostgreSQL,
MySQL, Microsoft SQL Server, and SQLite3 databases and performing the same
query.

<p align="center">
  <a href="https://asciinema.org/a/73gxbg62ny2fx9ppxu0kd8c48" target="_blank">
    <img src="https://asciinema.org/a/73gxbg62ny2fx9ppxu0kd8c48.png" width="654"/>
  </a>
</p>

# TODO
* Fix --command/-c execution
* All the various \\d* commands from `psql`
* SQL completion
