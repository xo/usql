<p align="center">
  <img src="https://raw.githubusercontent.com/xo/usql-logo/master/usql.png" height="120">
</p>

<p align="center">
  <a href="#installing" title="Installing">Installing</a> |
  <a href="#building" title="Building">Building</a> |
  <a href="#database-support" title="Database Support">Database Support</a> |
  <a href="#using" title="Using">Using</a> |
  <a href="#features-and-compatibility" title="Features and Compatibility">Features and Compatibility</a> |
  <a href="https://github.com/xo/usql/releases" title="Releases">Releases</a> |
  <a href="#contributing" title="Contributing">Contributing</a>
</p>

<br/>

`usql` is a universal command-line interface for PostgreSQL, MySQL, Oracle
Database, SQLite3, Microsoft SQL Server, [and many other databases][databases]
including NoSQL and non-relational databases!

`usql` provides a simple way to work with [SQL and NoSQL databases][databases]
via a command-line inspired by PostgreSQL's `psql`. `usql` supports most of the
core `psql` features, such as [variables][variables], [backticks][backticks],
[backslash commands][commands] and has additional features that `psql` does
not, such as [multiple database support][databases], [copying between databases][copying],
[syntax highlighting][highlighting], and [context-based completion][completion].

Database administrators and developers that would prefer to work with a tool
like `psql` with non-PostgreSQL databases, will find `usql` intuitive,
easy-to-use, and a great replacement for the command-line clients/tools
for other databases.

[![Unit Tests][usql-ci-status]][usql-ci]
[![Go Reference][goref-usql-status]][goref-usql]
[![Releases][release-status]][Releases]
[![Discord Discussion][discord-status]][discord]

[usql-ci]: https://github.com/xo/usql/actions/workflows/test.yml (Test CI)
[usql-ci-status]: https://github.com/xo/usql/actions/workflows/test.yml/badge.svg (Test CI)
[goref-usql]: https://pkg.go.dev/github.com/xo/usql (Go Reference)
[goref-usql-status]: https://pkg.go.dev/badge/github.com/xo/usql.svg (Go Reference)
[release-status]: https://img.shields.io/github/v/release/xo/usql?display_name=tag&sort=semver (Latest Release)
[discord]: https://discord.gg/yJKEzc7prt (Discord Discussion)
[discord-status]: https://img.shields.io/discord/829150509658013727.svg?label=Discord&logo=Discord&colorB=7289da&style=flat-square (Discord Discussion)

[installing]: #installing (Installing)
[databases]: #database-support (Database Support)
[releases]: https://github.com/xo/usql/releases (Releases)

## Installing

`usql` can be installed [via Release][], [via Homebrew][], [via AUR][], [via
Scoop][] or [via Go][]:

[via Release]: #installing-via-release
[via Homebrew]: #installing-via-homebrew-macos-and-linux
[via AUR]: #installing-via-aur-arch-linux
[via Scoop]: #installing-via-scoop-windows
[via Go]: #installing-via-go

### Installing via Release

1. [Download a release for your platform][releases]
2. Extract the `usql` or `usql.exe` file from the `.tar.bz2` or `.zip` file
3. Move the extracted executable to somewhere on your `$PATH` (Linux/macOS) or
  `%PATH%` (Windows)

### Installing via Homebrew (macOS and Linux)

Install `usql` from the [`xo/xo` tap][xo-tap] in the usual way with the [`brew`
command][homebrew]:

```sh
# install usql with most drivers
$ brew install xo/xo/usql
```

Support for [ODBC databases][databases] is available through the `--with-odbc`
install flag:

```sh
# add xo tap
$ brew tap xo/xo

# install usql with odbc support
$ brew install --with-odbc usql
```

### Installing via AUR (Arch Linux)

Install `usql` from the [Arch Linux AUR][aur] in the usual way with the [`yay`
command][yay]:

```sh
# install usql with most drivers
$ yay -S usql
```

Alternately, build and [install using `makepkg`][arch-makepkg]:

```sh
$ git clone https://aur.archlinux.org/usql.git && cd usql
$ makepkg -si
==> Making package: usql 0.12.10-1 (Fri 26 Aug 2022 05:56:09 AM WIB)
==> Checking runtime dependencies...
==> Checking buildtime dependencies...
==> Retrieving sources...
  -> Downloading usql-0.12.10.tar.gz...
...
```

### Installing via Scoop (Windows)

Install `usql` using [Scoop](https://scoop.sh):

```powershell
# Optional: Needed to run a remote script the first time
> Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# install scoop if not already installed
> irm get.scoop.sh | iex

# install usql with scoop
> scoop install usql
```

### Installing via Go

Install `usql` in the usual Go fashion:

```sh
# install latest usql version with base drivers
$ go install github.com/xo/usql@latest
```

See [below for information](#building) on `usql` build tags.

## Building

When building `usql` out-of-the-box with `go build` or `go install`, only the
[`base` drivers][databases] for PostgreSQL, MySQL, SQLite3, Microsoft SQL
Server, Oracle, CSVQ will be included in the build:

```sh
# build/install with base drivers (PostgreSQL, MySQL, SQLite3, Microsoft SQL Server,
# Oracle, CSVQ)
$ go install github.com/xo/usql@master
```

Other databases can be enabled by specifying the [build tag for their database
driver][databases].

```sh
# build/install with base, Avatica, and ODBC drivers
$ go install -tags 'avatica odbc' github.com/xo/usql@master
```

For every build tag `<driver>`, there is also a `no_<driver>` build tag
that will disable the driver:

```sh
# build/install most drivers, excluding Avatica, Couchbase, and PostgreSQL
$ go install -tags 'most no_avatica no_couchbase no_postgres' github.com/xo/usql@master
```

By specifying the build tags `most` or `all`, the build will include most, and
all SQL drivers, respectively:

```sh
# build/install with most drivers (excludes CGO drivers and problematic drivers)
$ go install -tags most github.com/xo/usql@master

# build/install all drivers (includes CGO drivers and problematic drivers)
$ go install -tags all github.com/xo/usql@master
```

## Database Support

`usql` works with all Go standard library compatible SQL drivers supported by
[`github.com/xo/dburl`][dburl].

The list of drivers that `usql` was built with can be displayed using the
[`\drivers` command][commands]:

```sh
$ cd $GOPATH/src/github.com/xo/usql

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

### Supported Database Schemes and Aliases

The following are the [Go SQL drivers][go-sql] that `usql` supports, the
associated database, scheme / build tag, and scheme aliases:

<!-- DRIVER DETAILS START -->
| Database             | Scheme / Tag    | Scheme Aliases                                  | Driver Package / Notes                                           |
|----------------------|-----------------|-------------------------------------------------|------------------------------------------------------------------|
| PostgreSQL           | `postgres`      | `pg`, `pgsql`, `postgresql`                     | [github.com/lib/pq][d-postgres]                                  |
| MySQL                | `mysql`         | `my`, `maria`, `aurora`, `mariadb`, `percona`   | [github.com/go-sql-driver/mysql][d-mysql]                        |
| Microsoft SQL Server | `sqlserver`     | `ms`, `mssql`, `azuresql`                       | [github.com/microsoft/go-mssqldb][d-sqlserver]                   |
| Oracle Database      | `oracle`        | `or`, `ora`, `oci`, `oci8`, `odpi`, `odpi-c`    | [github.com/sijms/go-ora/v2][d-oracle]                           |
| SQLite3              | `sqlite3`       | `sq`, `file`, `sqlite`                          | [github.com/mattn/go-sqlite3][d-sqlite3] <sup>[†][f-cgo]</sup>   |
| CSVQ                 | `csvq`          | `cs`, `csv`, `tsv`, `json`                      | [github.com/mithrandie/csvq-driver][d-csvq]                      |
|                      |                 |                                                 |                                                                  |
| Alibaba MaxCompute   | `maxcompute`    | `mc`                                            | [sqlflow.org/gomaxcompute][d-maxcompute]                         |
| Alibaba Tablestore   | `ots`           | `ot`, `tablestore`                              | [github.com/aliyun/aliyun-tablestore-go-sql-driver][d-ots]       |
| Apache Avatica       | `avatica`       | `av`, `phoenix`                                 | [github.com/apache/calcite-avatica-go/v5][d-avatica]             |
| Apache H2            | `h2`            |                                                 | [github.com/jmrobles/h2go][d-h2]                                 |
| Apache Ignite        | `ignite`        | `ig`, `gridgain`                                | [github.com/amsokol/ignite-go-client/sql][d-ignite]              |
| AWS Athena           | `athena`        | `s3`, `aws`, `awsathena`                        | [github.com/uber/athenadriver/go][d-athena]                      |
| Azure CosmosDB       | `cosmos`        | `cm`                                            | [github.com/btnguyen2k/gocosmos][d-cosmos]                       |
| Cassandra            | `cassandra`     | `ca`, `scy`, `scylla`, `datastax`, `cql`        | [github.com/MichaelS11/go-cql-driver][d-cassandra]               |
| ClickHouse           | `clickhouse`    | `ch`                                            | [github.com/ClickHouse/clickhouse-go/v2][d-clickhouse]           |
| Couchbase            | `couchbase`     | `n1`, `n1ql`                                    | [github.com/couchbase/go_n1ql][d-couchbase]                      |
| Cznic QL             | `ql`            | `cznic`, `cznicql`                              | [modernc.org/ql][d-ql]                                           |
| Exasol               | `exasol`        | `ex`, `exa`                                     | [github.com/exasol/exasol-driver-go][d-exasol]                   |
| Firebird             | `firebird`      | `fb`, `firebirdsql`                             | [github.com/nakagami/firebirdsql][d-firebird]                    |
| Genji                | `genji`         | `gj`                                            | [github.com/genjidb/genji/driver][d-genji]                       |
| Google BigQuery      | `bigquery`      | `bq`                                            | [gorm.io/driver/bigquery/driver][d-bigquery]                     |
| Google Spanner       | `spanner`       | `sp`                                            | [github.com/googleapis/go-sql-spanner][d-spanner]                |
| Microsoft ADODB      | `adodb`         | `ad`, `ado`                                     | [github.com/mattn/go-adodb][d-adodb]                             |
| ModernC SQLite3      | `moderncsqlite` | `mq`, `modernsqlite`                            | [modernc.org/sqlite][d-moderncsqlite]                            |
| MySQL MyMySQL        | `mymysql`       | `zm`, `mymy`                                    | [github.com/ziutek/mymysql/godrv][d-mymysql]                     |
| Netezza              | `netezza`       | `nz`, `nzgo`                                    | [github.com/IBM/nzgo/v14][d-netezza]                             |
| PostgreSQL PGX       | `pgx`           | `px`                                            | [github.com/jackc/pgx/v5/stdlib][d-pgx]                          |
| Presto               | `presto`        | `pr`, `prs`, `prestos`, `prestodb`, `prestodbs` | [github.com/prestodb/presto-go-client/presto][d-presto]          |
| SAP ASE              | `sapase`        | `ax`, `ase`, `tds`                              | [github.com/thda/tds][d-sapase]                                  |
| SAP HANA             | `saphana`       | `sa`, `sap`, `hana`, `hdb`                      | [github.com/SAP/go-hdb/driver][d-saphana]                        |
| Snowflake            | `snowflake`     | `sf`                                            | [github.com/snowflakedb/gosnowflake][d-snowflake]                |
| Trino                | `trino`         | `tr`, `trs`, `trinos`                           | [github.com/trinodb/trino-go-client/trino][d-trino]              |
| Vertica              | `vertica`       | `ve`                                            | [github.com/vertica/vertica-sql-go][d-vertica]                   |
| VoltDB               | `voltdb`        | `vo`, `vdb`, `volt`                             | [github.com/VoltDB/voltdb-client-go/voltdbclient][d-voltdb]      |
|                      |                 |                                                 |                                                                  |
| GO DRiver for ORacle | `godror`        | `gr`                                            | [github.com/godror/godror][d-godror] <sup>[†][f-cgo]</sup>       |
| ODBC                 | `odbc`          | `od`                                            | [github.com/alexbrainman/odbc][d-odbc] <sup>[†][f-cgo]</sup>     |
|                      |                 |                                                 |                                                                  |
| Amazon Redshift      | `postgres`      | `rs`, `redshift`                                | [github.com/lib/pq][d-postgres] <sup>[‡][f-wire]</sup>           |
| CockroachDB          | `postgres`      | `cr`, `cdb`, `crdb`, `cockroach`, `cockroachdb` | [github.com/lib/pq][d-postgres] <sup>[‡][f-wire]</sup>           |
| OLE ODBC             | `adodb`         | `oo`, `ole`, `oleodbc`                          | [github.com/mattn/go-adodb][d-adodb] <sup>[‡][f-wire]</sup>      |
| SingleStore MemSQL   | `mysql`         | `me`, `memsql`                                  | [github.com/go-sql-driver/mysql][d-mysql] <sup>[‡][f-wire]</sup> |
| TiDB                 | `mysql`         | `ti`, `tidb`                                    | [github.com/go-sql-driver/mysql][d-mysql] <sup>[‡][f-wire]</sup> |
| Vitess Database      | `mysql`         | `vt`, `vitess`                                  | [github.com/go-sql-driver/mysql][d-mysql] <sup>[‡][f-wire]</sup> |
|                      |                 |                                                 |                                                                  |
| Apache Hive          | `hive`          | `hi`                                            | [sqlflow.org/gohive][d-hive]                                     |
| Apache Impala        | `impala`        | `im`                                            | [github.com/bippio/go-impala][d-impala]                          |
|                      |                 |                                                 |                                                                  |
| **NO DRIVERS**       | `no_base`       |                                                 | _no base drivers (useful for development)_                       |
| **MOST DRIVERS**     | `most`          |                                                 | _all stable drivers_                                             |
| **ALL DRIVERS**      | `all`           |                                                 | _all drivers, excluding bad drivers_                             |
| **BAD DRIVERS**      | `bad`           |                                                 | _bad drivers (broken/non-working drivers)_                       |
| **NO &lt;TAG&gt;**   | `no_<tag>`      |                                                 | _exclude driver with `<tag>`_                                    |

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
[d-ots]: https://github.com/aliyun/aliyun-tablestore-go-sql-driver
[d-pgx]: https://github.com/jackc/pgx
[d-postgres]: https://github.com/lib/pq
[d-presto]: https://github.com/prestodb/presto-go-client
[d-ql]: https://gitlab.com/cznic/ql
[d-sapase]: https://github.com/thda/tds
[d-saphana]: https://github.com/SAP/go-hdb
[d-snowflake]: https://github.com/snowflakedb/gosnowflake
[d-spanner]: https://github.com/googleapis/go-sql-spanner
[d-sqlite3]: https://github.com/mattn/go-sqlite3
[d-sqlserver]: https://github.com/microsoft/go-mssqldb
[d-trino]: https://github.com/trinodb/trino-go-client
[d-vertica]: https://github.com/vertica/vertica-sql-go
[d-voltdb]: https://github.com/VoltDB/voltdb-client-go
<!-- DRIVER DETAILS END -->

[f-cgo]: #f-cgo (Requires CGO)
[f-wire]: #f-wire (Wire compatible)

<p>
  <i>
    <a id="f-cgo"><sup>†</sup> Requires CGO</a><br>
    <a id="f-wire"><sup>‡</sup> Wire compatible (see respective driver)</a>
  </i>
</p>

Any of the protocol schemes/aliases shown above can be used in conjunction when
connecting to a database via the command-line or with the [`\connect` and
`\copy` commands][commands]:

```sh
# connect to a vitess database:
$ usql vt://user:pass@host:3306/mydatabase

$ usql
(not connected)=> \c vitess://user:pass@host:3306/mydatabase

$ usql
(not connected)=> \copy csvq://. pg://localhost/ 'select * ....' 'myTable'
```

See [the section below on connecting to databases][connecting] for further
details building DSNs/URLs for use with `usql`.

## Using

After [installing][], `usql` can be used similarly to the following:

```sh
# connect to a postgres database
$ usql postgres://booktest@localhost/booktest

# connect to an oracle database
$ usql oracle://user:pass@host/oracle.sid

# connect to a postgres database and run the commands contained in script.sql
$ usql pg://localhost/ -f script.sql
```

### Command-line Options

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
resulting connection string to [a database driver][databases]. Database
connection strings (aka "data source name" or DSNs) have the same parsing rules
as URLs, and can be passed to `usql` via command-line, or to the [`\connect`,
`\c`, and `\copy` commands][commands].

Database connection strings look like the following:

```txt
   driver+transport://user:pass@host/dbname?opt1=a&opt2=b
   driver:/path/to/file
   /path/to/file
```

Where the above are:

| Component                       | Description                                                                          |
|---------------------------------|--------------------------------------------------------------------------------------|
| `driver`                        | driver scheme name or scheme alias                                                   |
| `transport`                     | `tcp`, `udp`, `unix` or driver name <i>(for ODBC and ADODB)</i>                      |
| `user`                          | username                                                                             |
| `pass`                          | password                                                                             |
| `host`                          | hostname                                                                             |
| `dbname` <sup>[±][f-path]</sup> | database name, instance, or service name/ID                                          |
| `?opt1=a&...`                   | additional database driver options (see respective SQL driver for available options) |
| `/path/to/file`                 | a path on disk                                                                       |

[f-path]: #f-path (URL Paths for Databases)

<p>
  <i>
    <a id="f-path">
      <sup>±</sup> Some databases, such as Microsoft SQL Server, or Oracle
      Database support a path component (ie, <code>/dbname</code>) in the form
      of <code>/instance/dbname</code>, where <code>/instance</code> is the
      optional service identifier (aka "SID") or database instance
    </a>
  </i>
</p>

#### Driver Aliases

`usql` supports the same driver names and aliases as [the `dburl`
package][dburl]. Databases have at least one or more aliases. See [`dburl`'s
scheme documentation][dburl-schemes] for a list of all supported aliases.

##### Short Aliases

All database drivers have a two character short form that is usually the first
two letters of the database driver. For example, `pg` for `postgres`, `my` for
`mysql`, `ms` for `sqlserver`, `or` for `oracle`, or `sq` for `sqlite3`.

#### Passing Driver Options

Driver options are specified as standard URL query options in the form of
`?opt1=a&obt2=b`. Refer to the [relevant database driver's
documentation][databases] for available options.

#### Paths on Disk

If a URL does not have a `driver:` scheme, `usql` will check if it is a path on
disk. If the path exists, `usql` will attempt to use an appropriate database
driver to open the path.

When the path is a Unix Domain Socket, `usql` will attempt to open it using the
MySQL driver. When the path is a directory, `usql` will attempt to open it
using the PostgreSQL driver. And, lastly, when the path is a regular file,
`usql` will attempt to open the file using the SQLite3 driver.

#### Driver Defaults

As with URLs, most components in the URL are optional and many components can
be left out. `usql` will attempt connecting using defaults where possible:

```sh
# connect to postgres using the local $USER and the unix domain socket in /var/run/postgresql
$ usql pg://
```

See the relevant documentation [on database drivers][databases] for more
information.

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

# Note: when connecting to a SQLite database, if the "<driver>://" or
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

The interactive intrepreter reads queries and [meta (`\`) commands][commands],
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
or `"`. Command parameters [may also be backticked][backticks].

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
  \copy SRC DST QUERY TABLE            copy query from source url to table on destination url
  \copy SRC DST QUERY TABLE(A,...)     copy query from source url to columns of table on destination url
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
  \dp[S] [PATTERN]                     list table, view, and sequence access privileges
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
  \t [on|off]                          show only rows
  \T [STRING]                          set HTML <table> tag attributes, or unset if none
  \x [on|off|auto]                     toggle expanded output

Transaction
  \begin                               begin a transaction
  \begin [-read-only] [ISOLATION]      begin a transaction with isolation level
  \commit                              commit current transaction
  \rollback                            rollback (abort) current transaction

Connection
  \c DSN                               connect to database url
  \c DRIVER PARAMS...                  connect to database with driver and parameters
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

An overview of `usql`'s features, functionality, and compability with `psql`:

* [Variables and Interpolation][variables]
* [Backticks][backticks]
* [Passwords][usqlpass]
* [Runtime Configuration (RC) File][usqlrc]
* [Copying Between Databases][copying]
* [Syntax Highlighting][highlighting]
* [Time Formatting][timefmt]
* [Context Completion][completion]
* [Host Connection Information](#host-connection-information)

The `usql` project's goal is to support as much of `psql`'s core features and
functionality, and aims to be as compatible as possible - [contributions are
always appreciated][contributing]!

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

> **Note**
>
> Variables contained within other strings <b><u>will not</b></u> be interpolated:

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

#### Backticks

[Meta (`\`) commands][commands] support backticks on parameters:

```sh
(not connected)=> \echo Welcome `echo $USER` -- 'currently:' "(" `date` ")"
Welcome ken -- currently: ( Wed Jun 13 12:10:27 WIB 2018 )
(not connected)=>
```

Backticked parameters will be passed to the user's `SHELL`, exactly as written,
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

> **Note**
>
> The `.usqlpass` file cannot be readable by other users, and the permissions
> should be set accordingly:

```sh
chmod 0600 ~/.usqlpass
```

#### Runtime Configuration (RC) File

`usql` supports executing a `.usqlrc` runtime configuration (RC) file contained
in the user's `HOME` directory:

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

The `.usqlrc` file is read at startup in the same way as a file passed on the
command-line with `-f` / `--file`. It is commonly used to set startup
environment variables and settings.

RC-file execution can be temporarily disabled at startup by passing `-X` or
`--no-rc` on the command-line:

```sh
$ usql --no-rc pg://
```

#### Copying Between Databases

`usql` provides a `\copy` command that reads data from a source database DSN
and writes to a destination database DSN:

```sh
$ usql
(not connected)=> \copy :PGDSN :MYDSN 'select book_id, author_id from books' 'books(id, author_id)'
```

> **Note**
>
> `usql`'s `\copy` is distinct from and <b><u>does not</u></b> function like
> `psql`'s `\copy`.

##### Parameters

The `\copy` command has two parameter forms:

```txt
SRC DST QUERY TABLE

SRC DST QUERY TABLE(COL1, COL2, ..., COLN)
```

Where:

* `SRC` - is the [source database URL][connecting] to connect to, and where the
  `QUERY` will be executed
* `DST` - is the [destination database URL][connecting] to connect to, and where
  the destination `TABLE` resides
* `QUERY` - is the query to execute on the `SRC` connection, the results of which
  will be copied to `TABLE`
* `TABLE` - is the destination table name, followed by an optional SQL-like column
  list of the form `(COL1, COL2, ..., COLN)`
* `(COL1, COL2, ..., COLN)` - a list of the destination column names, 1-to-N

The usual rules for [variables, interpolation, and quoting][variables] apply to
`\copy`'s parameters.

###### Quoting

`QUERY` and `TABLE` ***must*** be quoted when containing spaces:

```sh
$ usql
(not connected)=> echo :SOURCE_DSN :DESTINATION_DSN
pg://postgres:P4ssw0rd@localhost/ mysql://localhost
(not connected)=> \copy :SOURCE_DSN :DESTINATION_DSN 'select * from mySourceTable' 'myDestination(colA, colB)'
COPY 2
```

###### Column Counts

The `QUERY` ***must*** return the same number of columns as defined by
the `TABLE` expression:

```sh
$ usql
(not connected)=> \copy csvq:. sq:test.db 'select * from authors' authors
error: failed to prepare insert query: 2 values for 1 columns
(not connected)=> \copy csvq:. sq:test.db 'select name from authors' authors(name)
COPY 2
```

###### Datatype Compatibilty and Casting

The `\copy` command does not attempt to perform any kind of datatype
conversion.

If a `QUERY` returns columns with different datatypes than expected by the
`TABLE`'s column, the `QUERY` can use the source database's conversion/casting
functionality to cast columns to a datatype that will work for `TABLE`'s
columns:

```sh
$ usql
(not connected)=> \copy postgres://user:pass@localhost mysql://user:pass@localhost 'SELECT uuid_column::TEXT FROM myPgTable' myMyTable
COPY 1
```

###### Importing Data from CSV

The `\copy` command is capable of importing data from CSV's (or any other
database!) using the `csvq` driver:

```sh
$ cat authors.csv
author_id,name
1,Isaac Asimov
2,Stephen King
$ cat books.csv
book_id,author_id,title
1,1,I Robot
2,2,Carrie
3,2,Cujo
$ usql
(not connected)=> -- setting variables to make connections easier
(not connected)=> \set SOURCE_DSN csvq://.
(not connected)=> \set DESTINATION_DSN sqlite3:booktest.db
(not connected)=> -- connecting to the destination and creating the schema
(not connected)=> \c :DESTINATION_DSN
Connected with driver sqlite3 (SQLite3 3.38.5)
(sq:booktest.db)=> create table authors (author_id integer, name text);
CREATE TABLE
(sq:booktest.db)=> create table books (book_id integer not null primary key autoincrement, author_id integer, title text);
CREATE TABLE
(sq:booktest.db)=> -- adding an extra row to books prior to copying
(sq:booktest.db)=> insert into books (author_id, title) values (1, 'Foundation');
INSERT 1
(sq:booktest.db)=> -- disconnecting to demonstrate that \copy opens new database connections
(sq:booktest.db)=> \disconnect
(not connected)=> -- copying data from SOURCE -> DESTINATION
(not connected)=> \copy :SOURCE_DSN :DESTINATION_DSN 'select * from authors' authors
COPY 2
(not connected)=> \copy :SOURCE_DSN :DESTINATION_DSN 'select author_id, title from books' 'books(author_id, title)'
COPY 3
(not connected)=> \c :DESTINATION_DSN
Connected with driver sqlite3 (SQLite3 3.38.5)
(sq:booktest.db)=> select * from authors;
 author_id |     name
-----------+--------------
         1 | Isaac Asimov
         2 | Stephen King
(2 rows)

sq:booktest.db=> select * from books;
 book_id | author_id |   title
---------+-----------+------------
       1 |         1 | Foundation
       2 |         1 | I Robot
       3 |         2 | Carrie
       4 |         2 | Cujo
(4 rows)
```

> **Note**
>
> When importing large datasets (> 1GiB) from one database to another, it is
> better to use a database's native clients and tools.

###### Reusing Connections with Copy

The `\copy` command (and all `usql` commands) [works with variables][variables].
When scripting, or when needing to perform multiple `\copy` operations from/to
multiple sources/destinations, the best practice is to `\set` connection
variables either in a script or in [the `$HOME/.usqlrc` RC script][usqlrc].

Similarly, passwords can be stored for easy reuse (and kept out of scripts) by
storing in [the `$HOME/.usqlpass` password file][usqlpass].

For example:

```sh
$ cat $HOME/.usqlpass
postgres:*:*:*:postgres:P4ssw0rd
godror:*:*:*:system:P4ssw0rd
$ usql
Type "help" for help.

(not connected)=> \set pglocal postgres://postgres@localhost:49153?sslmode=disable
(not connected)=> \set orlocal godror://system@localhost:1521/orasid
(not connected)=> \copy :pglocal :orlocal 'select staff_id, first_name from staff' 'staff(staff_id, first_name)'
COPY 18
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

The `SYNTAX_*` variables are regular `usql` variables, and can be `\set` and
`\unset`:

```sh
$ usql
(not connected)=> \set SYNTAX_HL_STYLE dracula
(not connected)=> \unset SYNTAX_HL_OVERRIDE_BG
```

#### Context Completion

When using the interactive shell, context completion is available in `usql` by
hitting the `<Tab>` key. For example, hitting `<Tab>` can complete some parts
of `SELECT` queries on a PostgreSQL databases:

```sh
$ usql
Connected with driver postgres (PostgreSQL 14.4 (Debian 14.4-1.pgdg110+1))
Type "help" for help.

pg:postgres@=> select * f<Tab>
fetch            from             full outer join
```

Or, for example completing [backslash commands][commands] while connected to a
database:

```sh
$ usql my://
Connected with driver mysql (10.8.3-MariaDB-1:10.8.3+maria~jammy)
Type "help" for help.

my:root@=> \g<Tab>
\g     \gexec \gset  \gx
```

Not all commands, contexts, or databases support completion. If you're
interested in helping to make `usql`'s completion better, see [the section
below on contributing][contributing].

Command completion can be canceled with `<Control-C>`.

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

`usql`'s time format supports any [Go supported time format][go-time], or can
be any standard Go const name, such as `Kitchen` above. See below for an
overview of the [available time constants](#time-constants).

##### Time Constants

The following are the time constant names available in `usql`, corresponding
time format value, and example display output:

| Constant    | Format                                | Display <sup>[↓][f-ts]</sup>        |
|-------------|--------------------------------------:|------------------------------------:|
| ANSIC       | `Mon Jan _2 15:04:05 2006`            | `Wed Aug  3 20:12:48 2022`          |
| UnixDate    | `Mon Jan _2 15:04:05 MST 2006`        | `Wed Aug  3 20:12:48 UTC 2022`      |
| RubyDate    | `Mon Jan 02 15:04:05 -0700 2006`      | `Wed Aug 03 20:12:48 +0000 2022`    |
| RFC822      | `02 Jan 06 15:04 MST`                 | `03 Aug 22 20:12 UTC`               |
| RFC822Z     | `02 Jan 06 15:04 -0700`               | `03 Aug 22 20:12 +0000`             |
| RFC850      | `Monday, 02-Jan-06 15:04:05 MST`      | `Wednesday, 03-Aug-22 20:12:48 UTC` |
| RFC1123     | `Mon, 02 Jan 2006 15:04:05 MST`       | `Wed, 03 Aug 2022 20:12:48 UTC`     |
| RFC1123Z    | `Mon, 02 Jan 2006 15:04:05 -0700`     | `Wed, 03 Aug 2022 20:12:48 +0000`   |
| RFC3339     | `2006-01-02T15:04:05Z07:00`           | `2022-08-03T20:12:48Z`              |
| RFC3339Nano | `2006-01-02T15:04:05.999999999Z07:00` | `2022-08-03T20:12:48.693257Z`       |
| Kitchen     | `3:04PM`                              | `8:12PM`                            |
| Stamp       | `Jan _2 15:04:05`                     | `Aug  3 20:12:48`                   |
| StampMilli  | `Jan _2 15:04:05.000`                 | `Aug  3 20:12:48.693`               |
| StampMicro  | `Jan _2 15:04:05.000000`              | `Aug  3 20:12:48.693257`            |
| StampNano   | `Jan _2 15:04:05.000000000`           | `Aug  3 20:12:48.693257000`         |

[f-ts]: #f-ts (Timestamp Value)

<p>
  <i>
    <a id="f-ts"><sup>↓</sup> Generated using timestamp <code>2022-08-03T20:12:48.693257Z</code></a><br>
  </i>
</p>

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

## Additional Notes

The following are additional notes and miscellania related to `usql`:

### Release Builds

[Release builds][releases] are built with the `most` build tag and with
additional [SQLite3 build tags (see: `build-release.sh`)](build-release.sh).

### macOS

The recommended installation method on macOS is [via `brew` (see above)][via Homebrew]
due to the way libary dependencies for the `sqlite3` driver are done on macOS.
If the following (or similar) error is encountered when attempting to run `usql`:

```sh
$ usql
dyld: Library not loaded: /usr/local/opt/icu4c/lib/libicuuc.68.dylib
  Referenced from: /Users/user/.local/bin/usql
  Reason: image not found
Abort trap: 6
```

Then missing library dependency can be fixed by installing
[`icu4c`](http://site.icu-project.org) using `brew`:

```sh
$ brew install icu4c
Running `brew update --auto-update`...
==> Downloading ...
...

$ usql
(not connected)=>
```

## Contributing

`usql` is currently a WIP, and is aiming towards a 1.0 release soon.
Well-written PRs are always welcome -- and there is a clear backlog of issues
marked `help wanted` on the GitHub issue tracker! For [technical details on
contributing, see CONTRIBUTING.md](CONTRIBUTING.md).

[_Pick up an issue today, and submit a PR tomorrow!_][help-wanted]

## Related Projects

* [dburl][dburl] - Go package providing a standard, URL-style mechanism for parsing
  and opening database connection URLs
* [xo][xo] - Go command-line tool to generate Go code from a database schema

[dburl]: https://github.com/xo/dburl
[dburl-schemes]: https://github.com/xo/dburl#protocol-schemes-and-aliases
[go-time]: https://golang.org/pkg/time/#pkg-constants
[go-sql]: https://golang.org/pkg/database/sql/
[homebrew]: https://brew.sh/
[xo]: https://github.com/xo/xo
[xo-tap]: https://github.com/xo/homebrew-xo
[chroma]: https://github.com/alecthomas/chroma
[chroma-formatter]: https://github.com/alecthomas/chroma#formatters
[chroma-style]: https://xyproto.github.io/splash/docs/all.html
[help-wanted]: https://github.com/xo/usql/issues?q=is:open+is:issue+label:%22help+wanted%22
[aur]: https://aur.archlinux.org/packages/usql
[yay]: https://github.com/Jguer/yay
[arch-makepkg]: https://wiki.archlinux.org/title/makepkg

[backticks]: #backticks (Backticks)
[commands]: #backslash-commands (Commands)
[completion]: #context-completion (Context Completion)
[connecting]: #connecting-to-databases (Connecting to Databases)
[contributing]: #contributing (Contributing)
[copying]: #copying-between-databases (Copying Between Databases)
[highlighting]: #syntax-highlighting (Syntax Highlighting)
[timefmt]: #time-formatting (Time Formatting)
[usqlpass]: #passwords (Passwords)
[usqlrc]: #runtime-configuration-rc-file (Runtime Configuration File)
[variables]: #variables-and-interpolation (Variable Interpolation)
