# usql backlog

This file records planned work for usql and for the sibling repositories that
usql depends on. Each item names the files, workflows or issue numbers it
touches, so that the work can start without rediscovering the context.

Items 1 to 9 are Ken's work items, in his order. Items 10 to 17 were added from
work that followed. The last sections hold the GitHub working list, drivers that
could be added, and items that have been raised but have no priority yet.

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

## 5. Decide whether Dameng comes back

The driver was removed on 2026-09-23. It is not documented well enough to
review, so it was taken out rather than left in while that is worked out. What
went: `drivers/dameng`, `contrib/dameng`, `internal/dameng.go`, the ADR, the
README row, and the `github.com/godoes/gorm-dameng` dependency.

The `dm`, `dm8` and `dameng` schemes stay registered in dburl, which was left
alone deliberately. usql no longer has a driver for them, so a `dameng://` URL
now reports that the driver is not available.

Before it returns, the driver needs reading rather than trusting. It also
imports Go's `plugin` package through `dm8/security`, which was 44 MB of the
`most` build on its own; see item 11.

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

## 10. Distribute through winget

Publish usql to the Windows Package Manager, so that Windows users install it
the way they install anything else:

    winget install usql

Windows is the platform with no first class route today. The release builds a
zip, and the README points Windows users at the release page or at `go install`.
Scoop was the previous answer and has been removed.

A winget submission is a manifest in `microsoft/winget-pkgs`, holding the
release URL, the SHA256 of the archive and the installer type. The release
workflow already produces both the archive and its checksum, so the manifest
can be generated and submitted from the same job that drafts the release.

This became more worthwhile once duckdb started building on Windows, because
the Windows binary now carries the same driver set as the others.


## 11. Report the three size defects upstream

Measured on 2026-09-22 on linux/amd64, stripped with `-trimpath -ldflags "-s
-w"`. `docs/` holds no copy of the data; the numbers below are what the builds
reported.

### go-ora declares its character set tables as `int`

`github.com/sijms/go-ora/v3/converters/string_conversion_new.go` is 11 MB of
generated Go in 78,264 lines. It holds one `NewStringConverter` with 236 case
arms, and each arm builds a converter from a `dBuffer []int` and an
`eBuffer map[int]int`. That is 789,555 numbers in the slices and 403,054 pairs
in the maps, 1,595,663 numbers in total.

The element type is `int`, which is 8 bytes, so the tables occupy 12.17 MB. The
binary reports 12.11 MB for them, so the two agree. No value in any table
exceeds 65535. As `uint16` the tables would be 3.04 MB, which is 9.1 MB smaller,
and usql carries Oracle in every default build, so that is 9.1 MB off every
binary usql ships.

The `eBuffer` maps cost again at run time. Go compiles a large map literal into
two static arrays and code that inserts every entry, so building one converter
allocates a map of up to tens of thousands of entries.

Report both to sijms, together with the separate `0x80000000` overflow that is
why `drivers/oracle/oracle.go` carries `Build: !(linux && arm)`.

### duckdb-go-bindings passes -rdynamic

`github.com/duckdb/duckdb-go-bindings/lib/linux-amd64` passes `-rdynamic` in its
cgo LDFLAGS, which makes the linker export every symbol into the dynamic symbol
table. That table survives `-s -w`. duckdb costs 54.1 MB of a `most` build, and
almost all of it is that table.

### gorm-dameng imports the plugin package

No longer usql's problem, recorded because it decides whether the driver can
come back. `github.com/godoes/gorm-dameng/dm8/security` was the only importer
of Go's `plugin` package anywhere in usql's dependency graph, and importing it
has the same effect as `-rdynamic`. It cost 44.1 MB of a `most` build. The
driver was removed on 2026-09-23 for unrelated reasons; see item 5.

With both that and duckdb gone, a `most` build went from 258.5 MB to 159 MB,
and `.dynsym` plus `.dynstr` from 37.1 MB to zero. Only duckdb still carries
this cost.


## 12. `\chart file=NAME` should not need terminal graphics

`doExecChart` in `handler/handler.go` returns `text.ErrGraphicsNotSupported`
before it reads its arguments. When `file` is set the chart is written to a file
as SVG and nothing is drawn in the terminal, so the check rejects a request it
could have served. Reproduce it by piping a statement into a build made with
`-tags charts`:

    printf 'select 1 as a, 2 as b \\chart type bar file=/tmp/out.svg\n' | usql sqlite3://:memory:

Move the check past the point where `cfg` is known, and apply it only when
`cfg.File` is empty. This is older than the charts build tag and is not caused
by it.


## 13. Shrink the charts stack

Charts are behind the `charts` build tag as of 2026-09-22, so this only affects
a build made with `-tags charts` or `-tags all`. Measured on linux/amd64 by
diffing a `-tags charts` build against a `-tags none` build, both stripped with
`-trimpath -ldflags "-s -w"`. The tag costs 12.4 MB, which splits as:

| part | cost |
|---|---:|
| `libresvg.a`, the Rust static library | 3.68 MB |
| `github.com/dop251/goja` | 1.69 MB |
| `golang.org/x/text` | 1.27 MB |
| `echarts.min.js`, embedded | 1.20 MB |
| Go standard library and runtime pulled in | 1.21 MB |
| `.gopclntab`, `.go.type`, `.eh_frame`, `.data.rel.ro` | 2.9 MB |
| `go:func` | 0.17 MB |

Two of those are usql's own modules and can move.

`echartsgoja.go:288` carries `//go:embed *.js`, which embeds `echarts.min.js`
verbatim at 0.98 MB. Minified JavaScript compresses about four to one, so
storing it gzipped and inflating it on first render would return roughly
0.75 MB. `compress/gzip` is already linked into most usql builds.

`libresvg.a` is the largest single piece. The xo/resvg session is designing this
and has confirmed the scope: keep the `text` and `system-fonts` Cargo features,
which are load-bearing, and drop `raster-images` and `svgz` for a lean variant
while leaving them on by default for other consumers. Do not scope work around
removing font shaping. A chart carries about ten `<text>` elements at
`font-family:sans-serif`, covering the axis labels, the title, the subtitle and
the legend, and `xo/resvg` sets `loadSystemFonts: true` at `resvg.go:89`.

`raster-images` is safe to drop for usql specifically, and that is a guarantee
rather than a sample: `metacmd/charts/charts.go` builds the option document from
a closed struct with five fields, Title, Legend, XAxis, YAxis and Series, none
of which can carry an image, a URL or a data URI. usql cannot emit `<image>`.

The 1.27 MB of `golang.org/x/text` is the collation tables that goja links for
locale-aware string comparison. It arrives through echartsgoja's dependency on
goja and is only avoidable by changing engines, which is not worth it.

That session has also split libresvg into one Go module per platform, released
in xo/resvg v0.9.1, with each `libresvg/<platform>` tagged v0.48.1 to match the
vendored resvg version. usql is on v0.10.0, which adds an embedded font for
that project's own tests and a `WithSansSerifFamily` option, and is the same
size. Measured with an empty module cache, a
linux/amd64 build now downloads 11 MiB of archives rather than the 184 MiB of
the monolith, and each platform links only its own submodule. Note that
`go mod tidy` still records all six as indirect requires, so `go mod download`
with no arguments fetches all of them; a build does not.

That is a download win, not a binary size win: the charts build is 58.2 MB
either way, because it is the same archive. It does not resolve issue 494
either. See the resvg item under Lowest priority for what 494 actually needs.


## 14. NULL scan failures

Partly fixed on 2026-09-23. `drivers.NullSafeColumnType` replaces
`tblfmt.WithUseColumnTypes` for the drivers that set `UseColumnTypes`, which are
mysql, mymysql and databend.

The cause was that usql scanned straight into the Go type the driver names for
a column. MySQL describes 37 of the 56 columns of `SHOW REPLICA STATUS` as not
nullable, four of them as `uint32`, and then sends NULL for
`SQL_Remaining_Delay`, so the scan failed with:

    sql: Scan error on column index 43, name "SQL_Remaining_Delay":
    converting NULL to uint32 is unsupported

That is issues 307, 476 and 539, reported against MySQL 5.7 and 8.0.
`drivers/columns_test.go` reproduces it with a driver that reports the same
column shapes and returns NULL for all of them.

The same change fixes a second fault that was not reported. A nullable
`BIGINT UNSIGNED` arrives as `sql.Null[uint64]`, which nothing unwraps on the
way out, so usql printed the JSON of the struct:

    { "V": 0, "Valid": false }

Both the value and the NULL are printed correctly now.

Still open:

Pull requests 524, 526, 570 and 583 fix NULL scans in the metadata readers,
which is a different path from the one fixed here: those build their own scan
destinations rather than going through tblfmt. Review and merge them, then
check whether a shared helper would serve them too.

`drivers.go` has the same fault in the `\copy` path, at the
`reflect.New(columnTypes[i].ScanType())` near line 595. A NULL in a source
column fails the copy. It needs the same treatment, but the destination is
handed to `ExecContext` rather than printed, so the mapping is not identical.

The tblfmt fix raises that package's go directive to 1.27.1. main already
declares `go 1.27.1` so it costs nothing there, but release-21 is `go 1.26.1`
and release-20 is `go 1.25`, and both pin tblfmt v0.18.3. The fix therefore
cannot be backported to either release branch without raising its Go floor,
which is not a thing to do on a release branch. It reaches users through the
next minor instead.

The `sql.Null[T]` half of this is fixed in tblfmt v0.19.0, which usql is on.
Verified by building usql with `WithUseColumnTypes(true)` in place of
`NullSafeColumnType` and running against live MariaDB: tblfmt's own path now
produces identical output for every format.

`NullSafeColumnType` still stays, because the crash is not something tblfmt can
fix. `WithUseColumnTypes` still builds `reflect.New(ct.ScanType())`, so
`database/sql` refuses the NULL before the formatter ever sees the row.

tblfmt has its own session. Route changes and questions there rather than
editing the package.


## 15. `\pset numericlocale` corrupts the csv and json formats

Fixed in tblfmt on 2026-09-23, not yet released. No usql change is needed and
no issue was filed.

A locale formatted number was built with `newValue`, which marks a value Raw
and unquoted, so it skipped escaping. So `\pset numericlocale on` with
`\pset format json` emitted `[{"n":1,234,567}]`, which does not parse, and csv
emitted a bare `1,234,567`, which a reader takes as three fields against a
one-column header.

The outcome differs by format, which is correct. csv applies the locale and
quotes only the field that needs it, matching psql 18.6 byte for byte: `999`
bare and `"1,000"` quoted. json ignores the locale entirely and numbers stay
numbers, because csv has no type system and JSON does, so a column's JSON type
must not follow a display option.

Verified through usql against live MariaDB, including that output with
numericlocale off is byte identical to before the fix.

Untested: a locale whose grouping separator is not a comma, or whose decimal
separator is a comma.


## 16. The null string is not aligned the way psql aligns it

Fixed in tblfmt on 2026-09-23, not yet released. No usql change is needed.

psql aligns the null string to the column it lands in. tblfmt always aligned it
left. Tested against psql 18.6 and a real Postgres:

    psql -P null='(null)' -c "select 12345678901234567890::numeric as n,
                                     'txt'::text as t
                              union all select null, null;"

              n           |   t
    ----------------------+--------
     12345678901234567890 | txt
                   (null) | (null)

Right-aligned under the numeric column, left-aligned under the text column.
usql printed it left-aligned in both.

The cause was one shared `empty` Value per encoder, built with a zero Align,
with no knowledge of the column it was printed in. A null now follows its
column's alignment in the table and template encoders, and a column of mixed
type or of all nulls keeps the left default. tblfmt confirmed psql's `-H`
output right-aligns an html null cell in a numeric column too.

This was part of item 3, closing the gap against psql.


## 17. Three user-visible changes from tblfmt v0.19.0

usql is on v0.19.0. All three were verified against live MariaDB and Postgres.

`NaN`, `+Inf` and `-Inf` now print as `NaN`, `Infinity` and `-Infinity` in
every format, and as strings in json. This changes usql's default aligned
output and is right: it is byte identical to psql 18.6 on the same query, and
`to_jsonb` gives the same three as strings. The old spellings were Go's `%v`.

A uint64 is always a JSON string, so a MySQL `BIGINT UNSIGNED` column gives
`{"n":"42"}` as well as `{"n":"18446744073709551615"}`. usql argued against the
earlier form of this, which quoted only above 2^63 and so changed a column's
JSON type partway down. That is fixed: the column is now one type throughout,
which is the property that matters. The remaining difference from PostgreSQL,
where `to_jsonb(42::numeric)` is bare, is a judgement call rather than a
defect, and the objection is withdrawn.

json output is indented rather than compact. Nothing in usql depends on the old
shape and no golden covers the json format, but anyone byte-comparing usql's
json between versions will see every line move.

Also: tblfmt now documents its last-column padding difference from psql as
deliberate. Measuring it showed psql pads its header to full width and trims
its data rows, so psql is inconsistent with itself. usql should not chase that.


## Tier 2: GitHub issues and pull requests

Items 1 to 17 are the programme. This section is the working list drawn from the
96 open issues and 25 open pull requests, reviewed on 2026-09-21.

### Add the official Oracle driver

Add `github.com/oracle/go-oracledb`, Oracle's own driver for Go, as an
additional Oracle driver. This does not replace `github.com/sijms/go-ora/v3`,
which stays as it is.

### Released defects, fix first

Issue 590 broke `go install` and issue 589 broke the Dameng driver, both in
release 0.21.5. 590 is fixed, by replacing the `exclude` directive with an
ordinary pin of `github.com/uber-go/tally v3.5.10+incompatible`, since
`go install module@version` refuses any `exclude` or `replace`. 589 was fixed
by dburl v0.25.2, which gave the dameng scheme an `Override` so that
`u.Driver` is `dm` again, but usql no longer ships a Dameng driver, so it
should be closed as no longer applicable rather than as fixed. See item 5.

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
and 584 duplicate the dependency and action work in item 1. 585 adds a Dameng
driver, which usql removed on 2026-09-23; close it with item 5's reasoning. 360
belongs in Discussions.

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

### Confirm that the non-Windows, non-macOS, non-Linux builds now work

Charts are behind the `charts` build tag as of this change, so resvg is linked
only by `go build -tags charts` and `go build -tags all`. A default build and a
`most` build no longer reference it at all.

`github.com/xo/resvg` ships prebuilt static Rust artifacts instead of building
from source. Version 0.8.0 contains six copies of `libresvg.a`, for
`darwin_amd64`, `darwin_arm64`, `linux_amd64`, `linux_arm`, `linux_arm64` and
`windows_amd64`. The cgo preamble in `resvg.go` names those same six pairs in
its `#cgo <goos>,<goarch> LDFLAGS` lines. The file carries no build constraint,
so no other platform can compile the package. On any other platform the C code
compiles, because `CFLAGS` still points at `libresvg/resvg.h`, but no `LDFLAGS`
line matches. The linker therefore receives no `-L` and no `-lresvg`, and every
`resvg_` symbol is undefined. Issue 494 is this failure on the FreeBSD port for
arm64, armv7 and i386.

That failure should now be reachable only with the `charts` tag, so a default
build and a `most` build should work on those platforms. Build usql on the
freebsd-rline, netbsd-rline and omnios-rline machines to confirm it. Note that
resvg may not be the only package those platforms cannot build, so a failure
there is not by itself a sign that this gating is wrong.

Gating does not close issue 494. The xo/resvg session established that a
per-platform module split does not close it either, because the `#cgo CFLAGS`
line in `resvg.go` carries no build constraint: the C still compiles on an
unsupported platform and the link still fails on undefined `resvg_*` symbols,
from a different set of files. Closing 494 needs an explicit build-tag-gated
stub in xo/resvg that fails to build with a message naming the platform, or
that compiles to a renderer which reports that it is unavailable. That work is
recorded in xo/resvg's `libresvg/README.md`.

Fixing charts on those platforms is a separate job and costs more. resvg can
learn to link a system `libresvg` when the platform provides one, which is what
the FreeBSD port needs. The Rust crate can also be cross-built for more targets
and the artifacts vendored, which grows the module and needs a Rust toolchain
for each platform.

## Not yet scheduled

Apply the `update-deps.sh` report fix to `release-21`. The `REMAINING:` report
on that branch uses the module list from before the update. `go list -m -u`
then fails on a module that `go mod tidy` removed, and the report claims that
nothing is outdated. The branch also still carries an obsolete
`github.com/opencontainers/runc` entry in `SKIP`. The fix is in a stash on that
branch named `release-21: update-deps.sh REMAINING report fix`.
