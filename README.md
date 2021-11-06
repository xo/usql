<p align="center">
  <img src="https://raw.githubusercontent.com/xo/usql-logo/master/usql.png" height="120">
</p>

<p align="center">
  <a href="#installing" title="Installing">Installing</a> |
  <a href="#building" title="Building">Building</a> |
  <a href="#using" title="Using">Using</a> |
  <a href="#database-support" title="Database Support">Database Support</a> |
  <a href="#features-and-compatibility" title="Features and Compatibility">Features and Compatibility</a> |
  <a href="https://github.com/xo/usql/releases" title="Releases">Releases</a> |
  <a href="#contributing" title="Contributing">Contributing</a>
</p>

<br/>

`usql` is a universal command-line interface for PostgreSQL, MySQL, Oracle
Database, SQLite3, Microsoft SQL Server, [and many other databases][Database Support]
including NoSQL and non-relational databases!

`usql` provides a simple way to work with [SQL and NoSQL databases][Database Support]
via a command-line inspired by PostgreSQL's `psql`. `usql` supports most of the
core `psql` features, such as [variables][], [backticks][], and [commands][]
and has additional features that `psql` does not, such as [syntax highlighting][highlighting],
context-based completion, and [multiple database support][Database Support].

Database administrators and developers that would prefer to work with a tool
like `psql` with non-PostgreSQL databases, will find `usql` intuitive,
easy-to-use, and a great replacement for the command-line clients/tools
for other databases.

[![Unit Tests][usql-ci-status]][usql-ci]
[![Go Reference][goref-usql-status]][goref-usql]
[![Discord Discussion][discord-status]][discord]

[usql-ci]: https://github.com/xo/usql/actions/workflows/test.yml (Test CI)
[usql-ci-status]: https://github.com/xo/usql/actions/workflows/test.yml/badge.svg (Test CI)
[goref-usql]: https://pkg.go.dev/github.com/xo/usql (Go Reference)
[goref-usql-status]: https://pkg.go.dev/badge/github.com/xo/usql.svg (Go Reference)
[discord]: https://discord.gg/yJKEzc7prt (Discord Discussion)
[discord-status]: https://img.shields.io/discord/829150509658013727.svg?label=Discord&logo=Discord&colorB=7289da&style=flat-square (Discord Discussion)

[Installing]: #installing (Installing)
[Building]: #building (Building)
[Using]: #using (Using)
[Database Support]: #database-support (Database Support)
[Features and Compatibility]: #features-and-compatibility (Features and Compatibility)
[Releases]: https://github.com/xo/usql/releases (Releases)
[Contributing]: #contributing (Contributing)

## Installing

`usql` can be installed [via Release][], [via Homebrew][], [via Scoop][] or [via Go][]:

[via Release]: #installing-via-release
[via Homebrew]: #installing-via-homebrew-macos-and-linux
[via Scoop]: #installing-via-scoop-windows
[via Go]: #installing-via-go

### Installing via Release

1. [Download a release for your platform][Releases]
2. Extract the `usql` or `usql.exe` file from the `.tar.bz2` or `.zip` file
3. Move the extracted executable to somewhere on your `$PATH` (Linux/macOS) or
`%PATH%` (Windows)

#### macOS Notes

The recommended installation method on macOS is [via `brew` (see below)][via Homebrew].
If the following or similar error is encountered when attempting to run `usql`:

```sh
$ usql
dyld: Library not loaded: /usr/local/opt/icu4c/lib/libicuuc.68.dylib
  Referenced from: /Users/user/.local/bin/usql
  Reason: image not found
Abort trap: 6
```

Then the [ICU lib][] needs to be installed. This can be accomplished using `brew`:

```
$ brew install icu4c
```

[ICU lib]: http://site.icu-project.org

### Installing via Homebrew (macOS and Linux)

`usql` is available in the [`xo/xo` tap][xo-tap], and can be installed in the
usual way with the [`brew` command][homebrew]:

```sh
# install usql with "most" drivers
$ brew install xo/xo/usql
```

Additional support for [ODBC databases][Database Support] can be installed by
passing `--with-odbc` option during install:

```sh
# install usql with odbc support
$ brew install --with-odbc usql
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
# install usql from master branch with basic database support
# includes PostgreSQL, Oracle Database, MySQL, MS SQL, and SQLite3 drivers
$ go install github.com/xo/usql@master
```

## Building

When building `usql` with [Go][go-project], only drivers for PostgreSQL, MySQL,
SQLite3 and Microsoft SQL Server will be enabled by default. Other databases
can be enabled by specifying the build tag for their [database driver][Database Support].
Additionally, the `most` and `all` build tags include most, and all SQL
drivers, respectively:

```sh
# install all drivers
$ go install -tags all github.com/xo/usql@master

# install with most drivers (excludes unsupported drivers)
$ go install -tags most github.com/xo/usql@master

# install with base drivers and additional support for Oracle Database and ODBC
$ go install -tags 'godror odbc' github.com/xo/usql@master
```

For every build tag `<driver>`, there is also a `no_<driver>` build tag
disabling the driver:

```sh
# install all drivers excluding avatica and couchbase
$ go install -tags 'all no_avatica no_couchbase' github.com/xo/usql@master
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
packages. `usql`'s code is also well-documented -- please refer to the
[Go reference][goref-usql] for an overview of the various packages and APIs.

## Database Support

`usql` works with all Go standard library compatible SQL drivers supported by
[`github.com/xo/dburl`][dburl].

The list of drivers that `usql` was built with can be displayed using the
[`\drivers` command][commands]:

```sh
$ cd $GOPATH/src/github.com/xo/usql
$ export GO111MODULE=on
# build excluding the base drivers, and including cassandra and moderncsqlite
$ go build -tags 'no_postgres no_oracle no_sqlserver no_sqlite3 cassandra moderncsqlite'
# show built driver support
$ ./usql -c '\drivers'
Available Drivers:
  cql [ca, scy, scylla, datastax, cassandra]
  memsql (mysql) [me]
  moderncsqlite [mq, sq, file, sqlite, sqlite3, modernsqlite]
  mysql [my, maria, aurora, mariadb, percona]
  tidb (mysql) [ti]
  vitess (mysql) [vt]
```

The above shows that `usql` was built with only the `mysql`, `cassandra` (ie,
`cql`), and `moderncsqlite` drivers. The output above reflects information
about the drivers available to `usql`, specifically the internal driver name,
its primary URL scheme, the driver's available scheme aliases (shown in
`[...]`), and the real/underlying driver (shown in `(...)`) for wire compatible
drivers.

#### Supported Database Schemes and Aliases

The following are the [Go SQL drivers][go-sql] that `usql` supports, the
associated database, scheme / build tag, and scheme aliases:

<!-- DRIVER DETAILS START -->
| Database             | Scheme / Tag    | Scheme Aliases                                  | Driver Package / Notes                                          |
|----------------------|-----------------|-------------------------------------------------|-----------------------------------------------------------------|
| Microsoft SQL Server | `sqlserver`     | `ms`, `mssql`                                   | [github.com/denisenkom/go-mssqldb][d-sqlserver]                 |
| MySQL                | `mysql`         | `my`, `maria`, `aurora`, `mariadb`, `percona`   | [github.com/go-sql-driver/mysql][d-mysql]                       |
| Oracle Database      | `oracle`        | `or`, `ora`, `oci`, `oci8`, `odpi`, `odpi-c`    | [github.com/sijms/go-ora/v2][d-oracle]                          |
| PostgreSQL           | `postgres`      | `pg`, `pgsql`, `postgresql`                     | [github.com/lib/pq][d-postgres]                                 |
| SQLite3              | `sqlite3`       | `sq`, `file`, `sqlite`                          | [github.com/mattn/go-sqlite3][d-sqlite3]<sup>[†][f-cgo]</sup>   |
|                      |                 |                                                 |                                                                 |
| Alibaba MaxCompute   | `maxcompute`    | `mc`                                            | [sqlflow.org/gomaxcompute][d-maxcompute]                        |
| Apache Avatica       | `avatica`       | `av`, `phoenix`                                 | [github.com/apache/calcite-avatica-go/v5][d-avatica]            |
| Apache H2            | `h2`            |                                                 | [github.com/jmrobles/h2go][d-h2]                                |
| Apache Ignite        | `ignite`        | `ig`, `gridgain`                                | [github.com/amsokol/ignite-go-client/sql][d-ignite]             |
| AWS Athena           | `athena`        | `s3`, `aws`                                     | [github.com/uber/athenadriver/go][d-athena]                     |
| Cassandra            | `cassandra`     | `ca`, `scy`, `scylla`, `datastax`, `cql`        | [github.com/MichaelS11/go-cql-driver][d-cassandra]              |
| ClickHouse           | `clickhouse`    | `ch`                                            | [github.com/ClickHouse/clickhouse-go][d-clickhouse]             |
| Couchbase            | `couchbase`     | `n1`, `n1ql`                                    | [github.com/couchbase/go_n1ql][d-couchbase]                     |
| CSVQ                 | `csvq`          | `cs`, `csv`, `tsv`, `json`                      | [github.com/mithrandie/csvq-driver][d-csvq]                     |
| Cznic QL             | `ql`            | `cznic`, `cznicql`                              | [modernc.org/ql][d-ql]                                          |
| Exasol               | `exasol`        | `ex`, `exa`                                     | [github.com/exasol/exasol-driver-go][d-exasol]                  |
| Firebird             | `firebird`      | `fb`, `firebirdsql`                             | [github.com/nakagami/firebirdsql][d-firebird]                   |
| Genji                | `genji`         | `gj`                                            | [github.com/genjidb/genji/driver][d-genji]                      |
| Google BigQuery      | `bigquery`      | `bq`                                            | [gorm.io/driver/bigquery/driver][d-bigquery]                    |
| Google Spanner       | `spanner`       | `sp`                                            | [github.com/cloudspannerecosystem/go-sql-spanner][d-spanner]    |
| Microsoft ADODB      | `adodb`         | `ad`, `ado`                                     | [github.com/mattn/go-adodb][d-adodb]                            |
| ModernC SQLite3      | `moderncsqlite` | `mq`, `modernsqlite`                            | [modernc.org/sqlite][d-moderncsqlite]                           |
| MySQL MyMySQL        | `mymysql`       | `zm`, `mymy`                                    | [github.com/ziutek/mymysql/godrv][d-mymysql]                    |
| Netezza              | `netezza`       | `nz`, `nzgo`                                    | [github.com/IBM/nzgo][d-netezza]                                |
| PostgreSQL PGX       | `pgx`           | `px`                                            | [github.com/jackc/pgx/v4/stdlib][d-pgx]                         |
| Presto               | `presto`        | `pr`, `prs`, `prestos`, `prestodb`, `prestodbs` | [github.com/prestodb/presto-go-client/presto][d-presto]         |
| SAP ASE              | `sapase`        | `ax`, `ase`, `tds`                              | [github.com/thda/tds][d-sapase]                                 |
| SAP HANA             | `saphana`       | `sa`, `sap`, `hana`, `hdb`                      | [github.com/SAP/go-hdb/driver][d-saphana]                       |
| Trino                | `trino`         | `tr`, `trs`, `trinos`                           | [github.com/trinodb/trino-go-client/trino][d-trino]             |
| Vertica              | `vertica`       | `ve`                                            | [github.com/vertica/vertica-sql-go][d-vertica]                  |
| VoltDB               | `voltdb`        | `vo`, `vdb`, `volt`                             | [github.com/VoltDB/voltdb-client-go/voltdbclient][d-voltdb]     |
|                      |                 |                                                 |                                                                 |
| Apache Hive          | `hive`          | `hi`                                            | [sqlflow.org/gohive][d-hive]                                    |
| Apache Impala        | `impala`        | `im`                                            | [github.com/bippio/go-impala][d-impala]                         |
| Azure CosmosDB       | `cosmos`        | `cm`                                            | [github.com/btnguyen2k/gocosmos][d-cosmos]                      |
| GO DRiver for ORacle | `godror`        | `gr`                                            | [github.com/godror/godror][d-godror]<sup>[†][f-cgo]</sup>       |
| ODBC                 | `odbc`          | `od`                                            | [github.com/alexbrainman/odbc][d-odbc]<sup>[†][f-cgo]</sup>     |
| Snowflake            | `snowflake`     | `sf`                                            | [github.com/snowflakedb/gosnowflake][d-snowflake]               |
|                      |                 |                                                 |                                                                 |
| Amazon Redshift      | `postgres`      | `rs`, `redshift`                                | [github.com/lib/pq][d-postgres]<sup>[‡][f-wire]</sup>           |
| CockroachDB          | `postgres`      | `cr`, `cdb`, `crdb`, `cockroach`, `cockroachdb` | [github.com/lib/pq][d-postgres]<sup>[‡][f-wire]</sup>           |
| OLE ODBC             | `adodb`         | `oo`, `ole`, `oleodbc`                          | [github.com/mattn/go-adodb][d-adodb]<sup>[‡][f-wire]</sup>      |
| SingleStore MemSQL   | `mysql`         | `me`, `memsql`                                  | [github.com/go-sql-driver/mysql][d-mysql]<sup>[‡][f-wire]</sup> |
| TiDB                 | `mysql`         | `ti`, `tidb`                                    | [github.com/go-sql-driver/mysql][d-mysql]<sup>[‡][f-wire]</sup> |
| Vitess Database      | `mysql`         | `vt`, `vitess`                                  | [github.com/go-sql-driver/mysql][d-mysql]<sup>[‡][f-wire]</sup> |
|                      |                 |                                                 |                                                                 |
| **NO DRIVERS**       | `no_base`       |                                                 | _no base drivers (useful for development)_                      |
| **MOST DRIVERS**     | `most`          |                                                 | _all stable drivers_                                            |
| **ALL DRIVERS**      | `all`           |                                                 | _all drivers_                                                   |
| **NO &lt;TAG&gt;**   | `no_<tag>`      |                                                 | _exclude driver with `<tag>`_                                   |

[d-adodb]: https://github.com/mattn/go-adodb
[d-athena]: https://github.com/uber/athenadriver
[d-avatica]: https://github.com/apache/calcite-avatica-go
[d-bigquery]: https://github.com/go-gorm/gorm
[d-cassandra]: https://github.com/MichaelS11/go-cql-driver
[d-clickhouse]: https://github.com/ClickHouse/clickhouse-go
[d-cosmos]: https://github.com/btnguyen2k/gocosmos
[d-couchbase]: https://github.com/couchbase/go_n1ql
[d-csvq]: https://github.com/mithrandie/csvq-driver
[d-exasol]: https://github.com/exasol/exasol-driver-go
[d-firebird]: https://github.com/nakagami/firebirdsql
[d-genji]: https://github.com/genjidb/genji
[d-godror]: https://github.com/godror/godror
[d-h2]: https://github.com/jmrobles/h2go
[d-hive]: https://github.com/sql-machine-learning/gohive
[d-ignite]: https://github.com/amsokol/ignite-go-client
[d-impala]: https://github.com/bippio/go-impala
[d-maxcompute]: https://github.com/sql-machine-learning/gomaxcompute
[d-moderncsqlite]: https://gitlab.com/cznic/sqlite
[d-mymysql]: https://github.com/ziutek/mymysql
[d-mysql]: https://github.com/go-sql-driver/mysql
[d-netezza]: https://github.com/IBM/nzgo
[d-odbc]: https://github.com/alexbrainman/odbc
[d-oracle]: https://github.com/sijms/go-ora
[d-pgx]: https://github.com/jackc/pgx
[d-postgres]: https://github.com/lib/pq
[d-presto]: https://github.com/prestodb/presto-go-client
[d-ql]: https://gitlab.com/cznic/ql
[d-sapase]: https://github.com/thda/tds
[d-saphana]: https://github.com/SAP/go-hdb
[d-snowflake]: https://github.com/snowflakedb/gosnowflake
[d-spanner]: https://github.com/cloudspannerecosystem/go-sql-spanner
[d-sqlite3]: https://github.com/mattn/go-sqlite3
[d-sqlserver]: https://github.com/denisenkom/go-mssqldb
[d-trino]: https://github.com/trinodb/trino-go-client
[d-vertica]: https://github.com/vertica/vertica-sql-go
[d-voltdb]: https://github.com/VoltDB/voltdb-client-go
<!-- DRIVER DETAILS END -->

[f-cgo]: #f-cgo (Requires CGO)
[f-wire]: #f-wire (Wire compatible)

<p>
  <i>
    <a id="f-cgo"><sup>†</sup>Requires CGO</a><br>
    <a id="f-wire"><sup>‡</sup>Wire compatible (see respective driver)</a>
  </i>
</p>

Any of the protocol schemes/aliases shown above can be used in conjunction when
connecting to a database via the command-line or with the [`\connect` command][commands]:

```sh
# connect to a vitess database:
$ usql vt://user:pass@host:3306/mydatabase

$ usql
(not connected)=> \c vitess://user:pass@host:3306/mydatabase
```

See [the section below on connecting to databases](#connecting-to-databases)
for further details building DSNs/URLs for use with `usql`.

## Using

After [installing][Installing], `usql` can be used similarly to the following:

```sh
# connect to a postgres database
$ usql postgres://booktest@localhost/booktest

# connect to an oracle database
$ usql oracle://user:pass@host/oracle.sid

# connect to a postgres database and run the commands contained in script.sql
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
  -F, --field-separator=FIELD-SEPARATOR ...
                               field separator for unaligned output (default, "|")
  -R, --record-separator=RECORD-SEPARATOR ...
                               record separator for unaligned output (default, \n)
  -T, --table-attr=TABLE-ATTR ...
                               set HTML table tag attributes (e.g., width, border)
  -A, --no-align               unaligned table output mode
  -H, --html                   HTML table output mode
  -t, --tuples-only            print rows only
  -x, --expanded               turn on expanded table output
  -z, --field-separator-zero   set field separator for unaligned output to zero byte
  -0, --record-separator-zero  set record separator for unaligned output to zero byte
  -J, --json                   JSON output mode
  -C, --csv                    CSV output mode
  -G, --vertical               vertical output mode
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

| Component                      | Description                                                                          |
|--------------------------------|--------------------------------------------------------------------------------------|
| `driver`                       | driver scheme name or scheme alias                                                   |
| `transport`                    | `tcp`, `udp`, `unix` or driver name <i>(for ODBC and ADODB)</i>                      |
| `user`                         | username                                                                             |
| `pass`                         | password                                                                             |
| `host`                         | hostname                                                                             |
| `dbname`<sup>[±][f-path]</sup> | database name, instance, or service name/ID                                          |
| `?opt1=a&...`                  | additional database driver options (see respective SQL driver for available options) |
| `/path/to/file`                | a path on disk                                                                       |

[f-path]: #f-path (URL Paths for Databases)

<p>
  <i>
    <a id="f-path">
      <sup>±</sup>Some databases, such as Microsoft SQL Server, or Oracle
      Database support a path component (ie, <code>/dbname</code>) in the form
      of <code>/instance/dbname</code>, where <code>/instance</code> is the
      optional service identifier (aka "SID") or database instance
    </a>
  </i>
</p>

#### Driver Aliases

`usql` supports the same driver names and aliases from the [`dburl`][dburl]
package. Most databases have at least one or more alias - please refer to the
[`dburl` documentation][dburl-schemes] for all supported aliases.

##### Short Aliases

All database drivers have a two character short form that is usually the first
two letters of the database driver. For example, `pg` for `postgres`, `my` for
`mysql`, `ms` for `sqlserver` (formerly known as `mssql`), `or` for `oracle`,
or `sq` for `sqlite3`.

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

# connect to a sqlserver database
$ usql sqlserver://user:pass@host/instancename/dbname
$ usql ms://user:pass@host/dbname
$ usql ms://user:pass@host/instancename/dbname
$ usql mssql://user:pass@host:port/dbname
$ usql ms://

# connect to a sqlserver database using Windows domain authentication
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
  \q                                   quit usql
  \copyright                           show usql usage and distribution terms
  \drivers                             display information about available database drivers

Query Execute
  \g [(OPTIONS)] [FILE] or ;           execute query (and send results to file or |pipe)
  \crosstabview [(OPTIONS)] [COLUMNS]  execute query and display results in crosstab
  \G [(OPTIONS)] [FILE]                as \g, but forces vertical output mode
  \gexec                               execute query and execute each value of the result
  \gset [PREFIX]                       execute query and store results in usql variables
  \gx [(OPTIONS)] [FILE]               as \g, but forces expanded output mode
  \watch [(OPTIONS)] [DURATION]        execute query every specified interval

Query Buffer
  \e [FILE] [LINE]                     edit the query buffer (or file) with external editor
  \p                                   show the contents of the query buffer
  \raw                                 show the raw (non-interpolated) contents of the query buffer
  \r                                   reset (clear) the query buffer
  \w FILE                              write query buffer to file

Help
  \? [commands]                        show help on backslash commands
  \? options                           show help on usql command-line options
  \? variables                         show help on special variables

Input/Output
  \echo [-n] [STRING]                  write string to standard output (-n for no newline)
  \qecho [-n] [STRING]                 write string to \o output stream (-n for no newline)
  \warn [-n] [STRING]                  write string to standard error (-n for no newline)
  \o [FILE]                            send all query results to file or |pipe
  \i FILE                              execute commands from file
  \ir FILE                             as \i, but relative to location of current script

Informational
  \d[S+] [NAME]                        list tables, views, and sequences or describe table, view, sequence, or index
  \da[S+] [PATTERN]                    list aggregates
  \df[S+] [PATTERN]                    list functions
  \di[S+] [PATTERN]                    list indexes
  \dm[S+] [PATTERN]                    list materialized views
  \dn[S+] [PATTERN]                    list schemas
  \ds[S+] [PATTERN]                    list sequences
  \dt[S+] [PATTERN]                    list tables
  \dv[S+] [PATTERN]                    list views
  \l[+]                                list databases
  \ss[+] [TABLE|QUERY] [k]             show stats for a table or a query

Formatting
  \pset [NAME [VALUE]]                 set table output option
  \a                                   toggle between unaligned and aligned output mode
  \C [STRING]                          set table title, or unset if none
  \f [STRING]                          show or set field separator for unaligned query output
  \H                                   toggle HTML output mode
  \T [STRING]                          set HTML <table> tag attributes, or unset if none
  \t [on|off]                          show only rows
  \x [on|off|auto]                     toggle expanded output

Transaction
  \begin                               begin a transaction
  \commit                              commit current transaction
  \rollback                            rollback (abort) current transaction

Connection
  \c URL                               connect to database with url
  \c DRIVER PARAMS...                  connect to database with SQL driver and parameters
  \Z                                   close database connection
  \password [USERNAME]                 change the password for a user
  \conninfo                            display information about the current database connection

Operating System
  \cd [DIR]                            change the current working directory
  \setenv NAME [VALUE]                 set or unset environment variable
  \! [COMMAND]                         execute command in shell or start interactive shell
  \timing [on|off]                     toggle timing of commands

Variables
  \prompt [-TYPE] <VAR> [PROMPT]       prompt user to set variable
  \set [NAME [VALUE]]                  set internal variable, or list all if no parameters
  \unset NAME                          unset (delete) internal variable
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
set using `\pset time <FORMAT>`:

```sh
$ usql pg://
Connected with driver postgres (PostgreSQL 13.2 (Debian 13.2-1.pgdg100+1))
Type "help" for help.

pg:postgres@=> \pset
time                     RFC3339Nano
pg:postgres@=> select now();
             now
-----------------------------
 2021-05-01T22:21:44.710385Z
(1 row)

pg:postgres@=> \pset time Kitchen
Time display is "Kitchen" ("3:04PM").
pg:postgres@=> select now();
   now
---------
 10:22PM
(1 row)

pg:postgres@=>
```

Any [Go supported time format][go-time] or the standard Go const name (for example,
`Kitchen`, in the above).

##### Constants

| Constant Name | Value                                 |
|---------------|---------------------------------------|
| ANSIC         | `Mon Jan _2 15:04:05 2006`            |
| UnixDate      | `Mon Jan _2 15:04:05 MST 2006`        |
| RubyDate      | `Mon Jan 02 15:04:05 -0700 2006`      |
| RFC822        | `02 Jan 06 15:04 MST`                 |
| RFC822Z       | `02 Jan 06 15:04 -0700`               |
| RFC850        | `Monday, 02-Jan-06 15:04:05 MST`      |
| RFC1123       | `Mon, 02 Jan 2006 15:04:05 MST`       |
| RFC1123Z      | `Mon, 02 Jan 2006 15:04:05 -0700`     |
| RFC3339       | `2006-01-02T15:04:05Z07:00`           |
| RFC3339Nano   | `2006-01-02T15:04:05.999999999Z07:00` |
| Kitchen       | `3:04PM`                              |
| Stamp         | `Jan _2 15:04:05`                     |
| StampMilli    | `Jan _2 15:04:05.000`                 |
| StampMicro    | `Jan _2 15:04:05.000000`              |
| StampNano     | `Jan _2 15:04:05.000000000`           |

#### Copy

`usql` implements the `\copy` command that reads data from a database connection
and writes it into another one. It requires 4 parameters:
* source connection string
* destination connection string
* source query
* destination table name, optionally with columns

Connection strings support same syntax as in `\connect`. Source query needs to be quoted. Source query must
select same number of columns and in same order as they're defined in the destination table, unless
they're specified for the destination, as `table_name(column1, column2, ...)`. Quote the whole expression,
if it contains spaces. `\copy` does not attempt to perform any data type conversion. Use `CAST` in the source query
to ensure data types compatible with destination table. Some drivers may have limited data type support,
and they might not work at all when combined with other limited drivers.

Unlike `psql`, `\copy` in `usql` cannot read data directly from files. Drivers like `csvq` can help with this,
since they support reading CSV and JSON files.

```sh
$ cat books.csv
book_id,author_id,isbn,title,year,available,tags
3,1,3,one,2018,"2018-06-01 00:00:00",{}
4,2,4,two,2019,"2019-06-01 00:00:00",{}

$ usql -c "\copy csvq://. sqlite3://test.db 'select * from books' 'books'"
Copied 2 rows
```

Note that it might be a better idea to use tools dedicated to the destination database to load data in a robust way.

`\copy` reads data from plain `SELECT` queries. Most drivers that have `\copy` enabled use `INSERT` statements,
except for PostgreSQL ones, which use `COPY TO`. Because data needs to be downloaded from one database and uploaded
into another, don't expect same performance as in `psql`. For loading large amount of data efficiently,
use tools native to the destination database.

You can use `\copy` with variables. Better yet, put those `\set` commands in your runtime configuration file
at `$HOME/.usqlrc` and passwords at `$HOME/.usqlpass`.

```sh
$ usql
Type "help" for help.

(not connected)=> \set pglocal postgres://postgres@localhost:49153?sslmode=disable
(not connected)=> \set oralocal godror://system@localhost:1521/orasid
(not connected)=> \copy :pglocal :oralocal 'select staff_id, first_name from staff' 'staff(staff_id, first_name)'
```

## Contributing

`usql` is currently a WIP, and is aiming towards a 1.0 release soon.
Well-written PRs are always welcome -- and there is a clear backlog of issues
marked `help wanted` on the GitHub issue tracker!

[*Please pick up an issue today, and submit a PR tomorrow!*][help-wanted]

For more technical details, see [CONTRIBUTING.md](https://github.com/xo/usql/blob/master/CONTRIBUTING.md).

## Related Projects

* [dburl][dburl] - Go package providing a standard, URL-style mechanism for parsing and opening database connection URLs
* [xo][xo] - Go command-line tool to generate Go code from a database schema

[dburl]: https://github.com/xo/dburl
[dburl-schemes]: https://github.com/xo/dburl#protocol-schemes-and-aliases
[go-project]: https://golang.org/project
[go-time]: https://golang.org/pkg/time/#pkg-constants
[go-sql]: https://golang.org/pkg/database/sql/
[homebrew]: https://brew.sh/
[xo]: https://github.com/xo/xo
[xo-tap]: https://github.com/xo/homebrew-xo
[chroma]: https://github.com/alecthomas/chroma
[chroma-formatter]: https://github.com/alecthomas/chroma#formatters
[chroma-style]: https://xyproto.github.io/splash/docs/all.html
[help-wanted]: https://github.com/xo/usql/issues?q=is:open+is:issue+label:%22help+wanted%22

[commands]: #backslash-commands (Commands)
[backticks]: #backtick-d-parameters (Backtick Parameters)
[highlighting]: #syntax-highlighting (Syntax Highlighting)
[variables]: #variables-and-interpolation (Variable Interpolation)
