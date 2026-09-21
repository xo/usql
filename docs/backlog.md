# usql backlog

This file records planned work for usql and for the sibling repositories that
usql depends on. Each item names the files, workflows or issue numbers it
touches, so that the work can start without rediscovering the context.

Items 1 to 8 are Ken's work items, in his order. The resvg item at the end is
lowest priority. The last section holds items that have been raised but have no
priority yet.

## 1. Bring CI and CD up to date

The repository has three workflows: `announce.yml`, `release.yml` and
`test.yml`.

1. Remove the AUR publishing steps from `.github/workflows/announce.yml`. That
   publishing is automated elsewhere now.
2. Update every action to its current major version. The workflows use
   `actions/checkout@v4`, `actions/setup-go@v5`, `actions/upload-artifact@v4`,
   `actions/download-artifact@v4`, `softprops/action-gh-release@v2`,
   `crazy-max/ghaction-virustotal@v4` and `shimataro/ssh-key-action@v2`. Two
   Dependabot branches are open for this: `actions/checkout-7` and
   `softprops/action-gh-release-3`.
3. Generate better commit logs in the draft release notes.
4. Clean up `build.sh`. Add long options and a help output. Other improvements
   are open for discussion.
5. Run the tests on Ubuntu amd64, Ubuntu arm64, macos-latest and
   windows-latest. `test.yml` runs only on `ubuntu-latest` today. `release.yml`
   already builds on `macos-latest` and `windows-latest`, so the runners exist.

Item 5 depends on item 2 below.

## 2. Fix the unit tests

Make the unit tests work on every system, and against the databases that can
run in GitHub workflows. PostgreSQL, SQLite3 and MySQL are the minimum.
Microsoft SQL Server and Oracle are worth adding if the runners can host them.

The suite passes as of 2026-09-21, serially, against podman. What follows is
what is left.

The container tests cannot run in parallel. `go test ./...` starts three SQL
Server containers at once, and they exhaust the memory of a normal machine.
Every run needs `-p 1`. CI must either use `-p 1` or give each package its own
job.

A failing test leaks its containers. `TestMain` calls `pool.Close` after the
tests return, and `log.Fatalf` exits before that, so every failure leaves its
containers running. The leaked containers then starve the next run. Cleanup
must happen even on a fatal error.

`ramsql` breaks logging for the whole binary. Its `engine/log` package calls
`slog.SetDefault` from `init` at warning level, and that also redirects the
standard `log` package, so every `log.Print` and `log.Fatalf` in usql is
dropped. `drivers/drivers_test.go` now calls `log.SetOutput(os.Stderr)` to work
around it, but usql itself has the same problem and nothing protects it.
Decide whether to drop ramsql, patch it upstream, or restore the logger after
driver registration.

Golden files record upstream drift rather than usql behavior. Four expected
files were regenerated because jOOQ changed the sakila column types and because
tblfmt changed where it puts blank lines. `sqlserver.listIndexes` also depended
on a table that `TestCopy` leaves behind. Consider generating these from a
pinned schema instead of the live one.

`drivers/clickhouse` skips one case. clickhouse-go truncates a parameter list
with the regular expression `\sVALUES\s.*$`, which is case sensitive and needs
whitespace after `VALUES`, so a lowercase `values(` reaches the server with its
placeholders and is rejected. Report it upstream and remove the skip when it is
fixed.

## 3. Close the gap against psql

Compare usql against the current psql command line tool, feature by feature,
and close the gaps that matter.

The psql-compatible variables in `env/vars.go` are part of this. The prompt
substitution escapes are the visible gap. `handler/handler.go` line 498 lists
the ones that are not implemented: `%M`, `%m`, `%>`, `%n`, `%/`, `%~`, `%#`,
`%p`, `%R` and `%x`. Issue 467 is the user-facing report for `%/`, which shows
the database named in the connection URL instead of the database in use after
a `\c` command or a `USE` statement.

## 4. Work through GitHub

Fix the most important outstanding issues. There are 96 open. Move the
repository from issues to discussions. Close the pull requests where closing
makes sense.

The statement parsing bugs belong here. Issue 587 reports that a
backtick-quoted identifier containing a single quote flips the lexer quote
state in the MySQL and MariaDB dialect. The next string literal is then
unescaped a second time, and usql exits with status 0 after sending different
SQL than the user wrote. The same report names two silent no-ops. A statement
passed through `-f` with a trailing semicolon never runs, and any `-f` file
whose last statement has no trailing semicolon is discarded at end of file. The
code is in `stmt/parse.go`.

## 5. Fix the Dameng issues

Issue 589 reports `dameng://` failing with `driver not available` on version
0.21.5, which contains the driver and dburl v0.25.1. All three schemes connect
correctly in local testing on both branches, so the cause is not known yet.

## 6. Create a dbtest package

Move the contrib podman scripts into their own repository, written in Go, that
starts databases as containers through podman. The material is in `contrib/`:
`podman-run.sh`, `podman-stop.sh`, `usql-test.sh`, and one directory for each
database.

Add an MCP tool to that repository, so that a caller can start a database for
testing with a standard configuration.

## 7. Create a dbmeta package

Move the database metadata out of usql into its own repository, so that the
dbtpl project can share it. The code is in `drivers/metadata/`.

Expand the metadata to cover more databases. Then use dbtpl to generate the
model code for the shared metadata.

## 8. Switch to xo/rline

Replace the readline implementation with `github.com/xo/rline`. usql currently
depends on `github.com/gohxs/readline`, which has had no commits since 2017.
Issue 528 asks for this change and proposes `chzyer/readline`, which is the
upstream that gohxs forked.

Several open issues point at the readline layer and can close with this work.
Issue 546 reports a busy loop at 100 percent CPU in static builds. Issue 508
reports input alternating between the pager and the prompt. Issue 552 asks for
vim mode. Issue 490 reports the delete key removing a whole word. Issue 483
reports the terminal behaving as though Control is held down. Issue 472 reports
that a tab character cannot be sent as input.

## 9. Clean up the documentation

Create a friendly website and documentation site at usql.app.

`README.md` carries almost all of the documentation today, at roughly 64 KB in
a single file. Part of it is generated: `gen.go` rebuilds the driver table and
the link definitions from the build tags and from the dburl scheme registry, so
any new site must either keep that generation step or replace it.

## Tier 2: GitHub issues and pull requests

Items 1 to 9 are the programme. This section is the working list drawn from the
96 open issues and 25 open pull requests, reviewed on 2026-09-21.

### Add the official Oracle driver

Add `github.com/oracle/go-oracledb`, Oracle's own driver for Go, as an
additional Oracle driver. This does not replace `github.com/sijms/go-ora/v3`,
which stays as it is.

### Released defects, fix first

Issue 590 breaks `go install`, and issue 589 broke the Dameng driver in the
same release. Both are recorded under "Not yet scheduled" below with their
causes and fixes. 589 is already fixed in dburl v0.25.2.

### Pull requests to merge

These are one-file fixes from repeat contributors, most of them for NULL scan
errors in driver metadata: 524 and 570 for Oracle, 526 and 583 for catalog and
function metadata, 572 for Snowflake `$$` strings, 573 for Impala `SET`, 577
for PROMPT2, 578 for prompt documentation, which closes issue 374, 521 for a
manual page, 538 for config file locations, and 479 for stdin support.

### Pull requests to review carefully

591 fixes the backtick lexer bug in issue 587, but only one of the three
failures that issue reports, and the branch is rebased onto main locally as
`pr-591`. 565 rewrites metacmd tokenization for psql compatibility in 431 lines
of one file, and its description reads as a specification handed to a model
rather than a report of work done, so read the diff before trusting it. 501
adds bulk load for MySQL and 542 adds connection variables to `\copy`.

### Pull requests to close

418 and 535 both replace the readline layer and conflict with item 8. 571, 582
and 584 duplicate the dependency and action work in item 1. 585 duplicates the
Dameng driver that is already merged. 360 belongs in Discussions.

### Issues to close rather than fix

Item 8 subsumes the terminal cluster, so 122, 137, 236, 483, 490, 528 and 552
close with it. Item 1 subsumes the ten dependency scanner reports, which are
543, 556, 567, 568, 574, 575, 579, 580 and 581. 470 is already fixed on main.
589 is fixed in dburl v0.25.2. 263, 469 and 485 are questions and move to
Discussions. 391 and 475 are environment problems rather than usql defects.

### The theme being underweighted

NULL scan failures in driver metadata. Issues 476 and 539, and pull requests
524, 526, 570 and 583, are all the same shape: a `\d` or catalog query panics
or errors on a NULL column, across Oracle, SQL Server and MySQL. Four of the
open pull requests already fix parts of it.

## Tier 3: drivers that could be added

Verified on 2026-09-21. Every module below resolves on the proxy and calls
`sql.Register` in non-test code. Gemini and DeepSeek proposed others that do
not, and those were discarded.

| Database            | Module                                             | Latest                | Released   |
| ------------------- | -------------------------------------------------- | --------------------- | ---------- |
| IBM Db2             | `github.com/ibmdb/go_ibm_db`                       | v0.5.4                | 2025-10-07 |
| libSQL and Turso    | `github.com/tursodatabase/libsql-client-go/libsql` | v0.0.0-20260528064733 | 2026-05-28 |
| TDengine            | `github.com/taosdata/driver-go/v3`                 | v3.8.2                | 2026-07-09 |
| Dolt                | `github.com/dolthub/driver`                        | v1.88.1               | 2026-05-07 |
| MonetDB             | `github.com/MonetDB/MonetDB-Go/v2`                 | v2.0.3                | 2025-08-24 |
| rqlite              | `github.com/rqlite/gorqlite/stdlib`                | v0.0.0-20260504155303 | 2026-05-04 |
| openGauss and MogDB | `gitee.com/opengauss/openGauss-connector-go-pq`    | v1.0.8                | 2025-08-20 |

Db2 is the strongest candidate, because `contrib/db2/` already ships a
container definition and no Db2 driver exists anywhere in the tree. It needs
cgo and a download of the IBM CLI driver.

`github.com/ncruces/go-sqlite3/driver` v0.35.5 also qualifies, but usql already
links two SQLite drivers, so it is the weakest of these.

### Replace the retired Arrow module

`github.com/apache/arrow/go/v17` is frozen at v17.0.0 from 2024-07-11 and is
the only version that line ever published. The flightsql driver moved to
`github.com/apache/arrow-go/v18`, now v18.8.0 from 2026-09-04, which is already
an indirect dependency, so usql carries two Arrow trees today.

### Drivers whose upstream is dead

None of these has a maintained replacement that implements `database/sql`, so
removing one means dropping that database.

| Driver     | Module                                              | Latest          | Released   |
| ---------- | --------------------------------------------------- | --------------- | ---------- |
| mymysql    | `github.com/ziutek/mymysql`                         | v1.5.4          | 2015-01-09 |
| adodb      | `github.com/mattn/go-adodb`                         | v0.0.1          | 2018-05-15 |
| sapase     | `github.com/thda/tds`                               | v0.1.7          | 2019-09-27 |
| ignite     | `github.com/amsokol/ignite-go-client`               | v0.12.2         | 2019-01-04 |
| h2         | `github.com/jmrobles/h2go`                          | v0.5.0          | 2020-11-14 |
| cassandra  | `github.com/MichaelS11/go-cql-driver`               | v0.1.1          | 2020-09-20 |
| maxcompute | `sqlflow.org/gomaxcompute`                          | v0.0.0-20210805 | 2021-08-05 |
| couchbase  | `github.com/couchbase/go_n1ql`                      | v0.0.0-20220303 | 2022-03-03 |
| ots        | `github.com/aliyun/aliyun-tablestore-go-sql-driver` | v0.0.0-20220418 | 2022-04-18 |

mymysql is the clearest drop, because `go-sql-driver/mysql` covers the same
database and shipped v1.10.1 on 2026-09-02. adodb is next, because `internal`
already aliases odbc to adodb only on Windows and odbc covers it.

### The Vertica container is gone

`contrib/vertica/podman-config` points at `docker.io/vertica/vertica-ce`, which
no longer exists. OpenText moved to the `opentext/` namespace and published no
`vertica-ce` there. `opentext/vertica-k8s` exists but ignores the
`APP_DB_USER` and `APP_DB_PASSWORD` values the config sets, so substituting it
would produce a config that silently does not work. Either build the image from
an RPM as the upstream README now says, or drop the entry.

## Lowest priority

Add a build tag that compiles out resvg and the chart commands on platforms
other than Windows, macOS and Linux.

`github.com/xo/resvg` ships prebuilt static Rust artifacts instead of building
from source. Version 0.8.0 contains six copies of `libresvg.a`, for
`darwin_amd64`, `darwin_arm64`, `linux_amd64`, `linux_arm`, `linux_arm64` and
`windows_amd64`. The cgo preamble in `resvg.go` names those same six pairs in
its `#cgo <goos>,<goarch> LDFLAGS` lines. The file carries no build constraint,
so no other platform can compile the package.

On any other platform the C code compiles, because `CFLAGS` still points at
`libresvg/resvg.h`, but no `LDFLAGS` line matches. The linker therefore
receives no `-L` and no `-lresvg`, and every `resvg_` symbol is undefined.
Issue 494 is this failure on the FreeBSD port for arm64, armv7 and i386.

usql links resvg on every build today. `handler/handler.go` imports
`github.com/xo/resvg` at line 36, in an ordinary file with no build tag, and
there is no `no_resvg` tag to turn it off. This blocks NetBSD, FreeBSD and
illumos builds whatever Go toolchain the host has.

Two other approaches exist and cost more. resvg can learn to link a system
`libresvg` when the platform provides one, which is what the FreeBSD port needs
and which fixes issue 494 at the source. The Rust crate can also be cross-built
for more targets and the artifacts vendored, which grows the module and needs a
Rust toolchain for each platform.

## Not yet scheduled

The `exclude` directive in `go.mod` breaks `go install`. Version 0.21.5 cannot
be installed with `go install github.com/xo/usql@latest`, because Go treats the
named module as the main module and refuses any `exclude` or `replace`
directive. Issue 590 reports it, and issue 394 was the same failure in 2023.
The line is `exclude github.com/uber-go/tally v5.0.0+incompatible`, and it is
present on `main` and on `release-21`. The fix is to pin
`github.com/uber-go/tally v3.5.10+incompatible` as an ordinary requirement, and
to add that module to the `SKIP` list in `update-deps.sh` so that `go get -u`
does not raise it again. Nothing in the module graph needs version 5, because
athenadriver requires version 3.3.17.

Apply the `update-deps.sh` report fix to `release-21`. The `REMAINING:` report
on that branch uses the module list from before the update. `go list -m -u`
then fails on a module that `go mod tidy` removed, and the report claims that
nothing is outdated. The branch also still carries an obsolete
`github.com/opencontainers/runc` entry in `SKIP`. The fix is in a stash on that
branch named `release-21: update-deps.sh REMAINING report fix`.
