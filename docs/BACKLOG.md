# usql backlog

This file records planned work for usql and for the sibling repositories that
usql depends on. Each item names the files, workflows or issue numbers it
touches, so that the work can start without rediscovering the context.

This file is the source of truth for the roadmap. The issue tracker is for
defects and for work that is ready for someone outside the project to pick up.
Anything recorded in both places drifts, and the roadmap in the tracker was
also hiding the real bug reports underneath it.

## How to refer to an item

Every item has an identifier of the form `W` followed by a number. Three rules
govern it:

1. Identifiers are append only. A new item takes the next unused number.
2. An identifier is never reused. W1 is finished, and no later item becomes W1.
3. An identifier is never renumbered. Removing W6 does not turn W7 into W6.

An item that is finished or abandoned keeps its heading and gains a status in
that heading. It does not disappear. A citation that silently starts pointing
at different work is worse than one that points at nothing, because nothing
tells the reader it moved.

A heading with no status is open. The statuses in use are `Done`, `Dropped`
and `Superseded by Wn`.

These rules exist because the items are cited. They are cited in this file, in
commit messages, and in conversation. Before these rules the numbers were
positional, so removing an item renumbered every item after it and quietly
invalidated every reference to them.

Use a separate series for anything that is not a work item. The sibling
repositories `dburl` and `dbmeta` number their design decisions `D1`, `D2` and
so on, and the two series must not be confused where both appear in one commit
message.

## Where the items came from

W2 to W9 are Ken's work items, in his order. W10 to W15 were added from work
that followed. W16 is the roadmap that used to live in the GitHub issue
tracker, and W17 came from it. W18 and after were added later, and each says
where it came from.

W1, bringing CI and CD up to date, is finished.

The last sections hold the GitHub working list, drivers that could be added,
and items that have been raised but have no priority yet.

## W2. Fix the unit tests

Make the unit tests work on every system, and against the databases that can
run in GitHub workflows. PostgreSQL, SQLite3 and MySQL are the minimum.
Microsoft SQL Server and Oracle are worth adding if the runners can host them.

The suite passes as of 2026-09-21, serially, against podman. What follows is
what is left.

The container tests cannot run in parallel. `go test ./...` starts three SQL
Server containers at once, and they exhaust the memory of a normal machine.
Every run needs `-p 1`, which CI does.

Both of the other two are fixed. A failing test no longer leaks its containers:
`TestMain` is `os.Exit(run(m))` and `run` holds the deferred purge, so every
exit path reaches it.

The logging problem is fixed by removing the RamSQL driver. Its `engine/log`
called `slog.SetDefault` from `init` with a handler on `os.Stdout`, and a
default slog logger also redirects the standard `log` package, so every
`log.Print` and `slog` call vanished in any build that carried it. Writing to
stdout was the worse half, because stdout carries the query results.

A workaround in `package main` was written first and then reverted. It fixed
the binary but not usql's packages when imported elsewhere, nor any test binary
that does not link `main`. Gemini and DeepSeek were both asked whether the
driver was worth keeping and both said drop it, independently and without
hedging: mutating the process-wide logger from `init` is disqualifying for a
CLI whose stdout is the data channel, and `sqlite3://:memory:` and the pure-Go
`moderncsqlite` already cover every use for it.

`drivers/drivers_test.go` keeps its `log.SetOutput(os.Stderr)`. It is one line
and nothing stops the next driver doing the same thing.

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

## W3. Close the gap against psql

Compare usql against the current psql command line tool, feature by feature,
and close the gaps that matter.

`\prompt` must work in a non-interactive script. Issue 53, which Ken accepted
in the thread after first closing it: psql allows `\prompt` in a
non-interactive terminal, so usql should too, and "I will make it a point to
have this fixed in the next major release" is the commitment. He also named the
obstacle, which is that the readline package makes it hard to enable on
scripts, so this is entangled with W8. Without it a `-f` script cannot ask
for a value, and `--set NAME=VALUE` is the only way to pass one in.

The psql-compatible variables in `env/vars.go` are part of this. The prompt
substitution escapes are the visible gap. `handler/handler.go` line 498 lists
the ones that are not implemented: `%M`, `%m`, `%>`, `%n`, `%/`, `%~`, `%#`,
`%p`, `%R` and `%x`. Issue 467 is the user-facing report for `%/`, which shows
the database named in the connection URL instead of the database in use after
a `\c` command or a `USE` statement.

## W4. Work through GitHub

93 issues and 24 pull requests are open, reviewed on 2026-09-23. Gemini and
DeepSeek were consulted on which are real; where they disagreed it is said so.

### W4a. The nine scanner reports: close them

Issues 543, 556, 567, 568, 574, 575, 579, 580 and 581 are all dependency
scanner output from nine different users. `govulncheck` against the full build
reports that nothing reachable is affected: 0 vulnerabilities in usql's code,
0 in imported packages, and 3 in required modules that are never called, being
GO-2026-5932, GO-2022-0646 and GO-2022-0635.

They are not defects in usql and should be closed. Both models agreed that a
shipped CLI is still operationally different from a library: enterprise
scanners read `go.mod` from the binary's build info and ignore reachability, so
a compliance gate can block the binary whatever `govulncheck` says. The reports
will therefore keep arriving, and closing them without changing anything is
only half an answer.

They disagreed on whether to keep one umbrella issue. Gemini said do not, it
attracts noise. DeepSeek said keep one for tracking. No umbrella issue is the
better call for a project this size.

### W4b. Stop the scanner reports recurring

Both models converged on the same four, in rough order of effect:

1. Run `govulncheck` in CI and publish its output, so the reachability answer
   is a build artifact rather than a claim in a comment.
2. Publish a VEX document with each release marking the unreachable findings
   `not_affected`. Modern scanners consume it and suppress the false alarm.
   This is the only mechanism that actually stops the reports at source.
3. Add `SECURITY.md` stating that usql tracks vulnerabilities through
   `govulncheck`, and that an unreachable transitive finding is ordinary
   dependency maintenance rather than a security incident.
4. Add an issue template for vulnerability reports that asks for `govulncheck`
   output showing a call path.

Keep taking the dependency bumps regardless. They are cheap and they clear most
scanners without any argument.

### W4c. Statement parsing

Issue 587 reports that a backtick-quoted identifier containing a single quote
flips the lexer quote state in the MySQL and MariaDB dialect. The next string
literal is unescaped a second time, and usql exits 0 after sending different
SQL than the user wrote. The same report names two silent no-ops: a statement
passed through `-f` with a trailing semicolon never runs, and any `-f` file
whose last statement lacks a trailing semicolon is discarded at end of file.
The code is `stmt/parse.go`. Pull request 591 fixes the backtick half only and
is rebased onto main locally as `pr-591`.

Issues 165, 166 and 505 are the same area: the `\\` separator, quoted string
processing, and a regression in quoted variable replacement. Fix them together
with W3 rather than piecemeal.

### W4d. Stability, which outranks everything else here

Issue 546, a busy loop at 100% CPU, and 464, a crash after executing a command.
Both models put these first and they are right: a CLI that spins or dies is
worse than one missing a feature. 546 is most likely the interactive loop
polling on EOF or an unexpectedly closed pipe.

Issue 371, high memory usage, has no heap profile attached. Ask for one and
close it if none arrives.

### W4e. `\copy`, which is the data path

Issues 254, 322, 397, 427, 458, 462 and 495. Gemini called this the top
priority after stability, on the grounds that a tool which mangles data or runs
out of memory on a large load gets dropped immediately. 397, 254 and 462 are
probably one fault, the NULL scan in `\copy`; see W14. 322 is a 500 MB file
failing, which suggests the copy reads a whole payload rather than streaming.

### W4f. Output faults, cheap and visible

Issue 448, numbers shown as `1.450817032e+`, is a formatting default and should
be a small fix. Issue 509, control characters in a text field breaking the
output, needs the cell sanitised. Issue 504, `\dt` listing SQLite system
tables, is a missing `sqlite_%` filter. All three are quick and user facing.

Issue 516, csv disabling the pager, Gemini read as intended behaviour, since
csv is meant for redirection. Honour an explicitly set `\pset pager on` and
leave the default as it is.

### W4g. Driver-specific metadata: accept patches, do not write them

Issues 375, athena has no `\d`, and 440, no foreign keys on SQL Server. Gemini
was blunt and correct: one maintainer cannot write bespoke introspection for
fifty databases, most of which are not run locally. Label these for help and
merge contributions with tests. Do not spend core time on them.

### W4h. The terminal cluster belongs to W8

Issues 93, 122, 236, 483, 490, 508, 552 and 472 are all readline behaviour:
garbled input after alt-tab on Windows, DEL deleting a whole word, typing
switching to Ctrl, vi key bindings, the pager and prompt alternating. These do
not get fixed one at a time in usql. They close with the move to `xo/rline`.

### W4i. Questions, blocked on Discussions being enabled

Issues 263, 469 and 485 are questions, and pull request 360 is titled as one.
None can move, because Discussions is switched off for the repository.
Enabling it is a settings change and is Ken's call. Until then they stay open,
because closing a question the author cannot re-ask elsewhere is worse than
leaving it.

475 was not an environment problem. gosnowflake printed a DBUS warning during
`--version`, which is a driver writing uninvited output. It no longer
reproduces: `--version` gives clean stdout and empty stderr with
`DBUS_SESSION_BUS_ADDRESS` unset, and gosnowflake is still linked. Closed as
fixed.

391 is usql running inside the Emacs shell on Windows, which is not a terminal,
so the prompt and the line editor are off. That is documented behaviour rather
than a defect, but W8 may change what is possible there, so it stays open
until the rline switch lands.

### W4j. Be ruthless with the rest

Both models said the same thing unprompted: for a project this size, close
anything over a year old that has no reproduction, no stack trace and no heap
profile, saying it can be reopened with one. Most will never be reopened. The
alternative is that the real bugs above stay buried under sixty that are not.


## W5. `\echo -n` and `\warn -n` are erased by the line editor

Issue 215, which stays open as an issue because it is a defect rather than a
plan. Tier 1.

`\echo -n` and `\warn -n` suppress the trailing newline correctly, and then the
line editor clears the line before the next prompt is drawn, so the output is
never seen:

    (not connected)=> \echo -n foo
    (not connected)=>

psql leaves it in place and draws the prompt after it:

    postgres=# \echo -n foo
    foopostgres=#

The output is being written and then overwritten, so the fix is in how the
prompt is redrawn after a command that deliberately left the cursor mid-line.
That makes it a question for W8 as much as for the `\echo` implementation,
and it should be checked against `xo/rline` before being fixed in the current
in-tree editor.


## W6. Create a dbtest package (Dropped)

Dropped on 2026-09-26. Do not start this work and do not reinstate it without
reading what follows.

The plan was to move the podman scripts in `contrib/` into a separate Go
repository that starts databases as containers, and to give that repository an
MCP tool so a caller could start a database for testing with a standard
configuration.

Most of what the package was for has been built in `dbmeta` instead. dbmeta
runs databases in containers across several versions and flavors of each
product, because it has to answer what each one supports, and that requires the
same container handling this item described. Building a second thing that
starts the same containers would leave two of them to keep working.

What this item wanted that dbmeta does not already provide is the MCP tool.
That is a small addition to dbmeta if it is still wanted, not a repository.

The scripts in `contrib/` stay where they are. They work, and nothing here
depends on moving them.

## W7. Create a dbmeta package (In progress, outside this repository)

Move the database metadata out of usql into its own repository, so that the
dbtpl project can share it. The code here is in `drivers/metadata/`.

The repository exists and is being built. As of 2026-09-26 it has models for
nine products and is working on SAP HANA. It records its design decisions in
`docs/PLAN.md` and its per product notes in `docs/DIALECT.md`, and it keeps a
`docs/USQL.md` describing how usql reads metadata today and what it would take
to read it from dbmeta instead.

Nothing has been removed from usql yet, and this item stays open until it is.
The migration plan is [DBMETA.md](DBMETA.md), and the work is W21.

Two facts about usql that the work needs, measured on 2026-09-26 against
`-tags all`:

Of 51 registered names, 21 have a metadata reader and 30 have none. Of the 21,
the gaps are only ever `\l` and `\dp`. See W18.

usql declares 19 interfaces in `drivers/metadata/metadata.go`. Fourteen are
leaf readers, aggregated by `ExtendedReader`. Seven of the fourteen decide
whether a command runs at all. The other seven decide how much `\d+` prints.

## W8. Switch to xo/rline

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


### The readline label is the shared queue

Anything tagged `readline` on the tracker is rline's to plan for, and that
session has been told so. Thirteen issues carry it: 72, 93, 122, 215, 236, 320,
414, 472, 483, 490, 508, 528 and 552.

    https://github.com/xo/usql/issues?q=is%3Aissue+is%3Aopen+label%3Areadline

Deliberately not tagged, so the boundary is on record. 196 and 282 are about
which candidates autocomplete offers, which the metadata reader produces, not
how the editor behaves. 467 is prompt content and 516 is pager configuration.
422 and 342 are about where a password comes from rather than how it is read.

546, a busy loop at 100% CPU against `usql_static`, is tagged but unconfirmed.
The reporter's diagnosis, which they say came from an AI and nobody has
verified, is that the read loop stops blocking in a static build and returns
empty results. If that is right it is rline's; if not it is ours.

### What the rline session found, 2026-09-23

Two mechanisms behind the terminal cluster, worth recording because neither is
guessable from the issue titles.

The four-issue family, 93, 122, 483 and 490, where ordinary keys start acting
as if Ctrl were held. rline's escape decoder maps `CSI I` to Tab and `CSI O` to
F3. Those two sequences are focus-in and focus-out, which is what a terminal
sends on alt-tab, the trigger two of the four reporters named. A focus
notification therefore arrives as a keystroke, a spurious Tab opens the
completion menu, and the keys afterwards do something other than insert
themselves. The mis-decode is measured. That something else leaves focus
reporting enabled, and that this explains all four reports, are separate claims
and not yet shown.

508, the pager, is a different seam and not the other end of the same one.
dradtke reproduced it in twenty lines that import `usql/rline` directly with no
database and no usql, running `less` as a child while calling `Next`, and
confirmed with strace that `less` reads every other keypress. Two readers on
one terminal, with the kernel giving each input to whichever reads first. usql
starts the pager without suspending the line editor.

The rline session then reproduced it against its own port, with a control:
without a child holding the terminal the editor read 10 of 10 keys, and with
one it read 0 of 10. The child there was `cat` rather than `less`, and the
difference is the useful part. `cat` reads greedily in a loop and takes
everything; `less` reads one key and waits, which is why the report describes a
clean alternation rather than total starvation. **The split is a property of
the other reader, not of the bug.** So a fix that tunes the sharing rather than
ending it would behave differently against every program somebody sets `PAGER`
to.

rline needs an explicit contract for handing the terminal to a child and taking
it back. Two constraints on that design. The first pager invocation after
startup behaves and later ones do not, which the two-readers account does not
explain on its own, since two readers should race the first time as readily as
the tenth; either the first invocation differs, or it races and wins. And the
fix has to end the sharing rather than tune it, for the reason above.

472, sending a literal tab, is a feature request against rline rather than a
usql regression: `key.CtrlV` exists as a code with nothing implementing
literal-next, and bracketed paste is absent from the port and was absent from
isocline before it.

## W9. Clean up the documentation

Create a friendly website and documentation site at usql.app.

`README.md` carries almost all of the documentation today, at roughly 64 KB in
a single file. Part of it is generated: `gen.go` rebuilds the driver table and
the link definitions from the build tags and from the dburl scheme registry, so
any new site must either keep that generation step or replace it.

## W10. Distribute through winget

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


## W11. Report the three size defects upstream

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

### dburl v0.26.0 removes the Dameng schemes

Recorded because it changes what a user sees and arrives through a dependency
bump rather than through anything here. usql removed the Dameng driver on
2026-09-23 and left dburl alone, so `dameng://` currently reports that the
driver is not available. dburl is now removing the `dm` scheme and its `dm8`
and `dameng` aliases as well, in v0.26.0, after the same question was put to
that project separately. Once usql is on it the message becomes unknown
database scheme, which is more accurate.

usql is on v0.26.0, which removes them. The plan had been to pin to v0.25.4 and
keep the release boring, and the evidence went the other way. On v0.25.4 a
`dameng://` URL reports that the `dm` driver is not available and then prints
usql's rebuild hint:

    error: dm: driver not available

    try:

      go install -tags 'most dm' github.com/xo/usql@master

That hint is wrong. The `dm` build tag was deleted along with the driver, so
following it produces a binary that still cannot connect. v0.26.0 reports
`unknown database scheme` and prints nothing further, which is accurate.

**Put this in usql's release notes.** The error a user sees for a `dameng://`
URL changes, and it changes through a dependency bump rather than through
anything in usql's own history, so someone searching usql's changelog for it
would find nothing. All three forms are affected: `dameng://`, `dm://` and
`dm8://`. usql is clean either way, with no reference to `GenDameng` or to the
schemes outside this file.

### The duckdb promotion depends on dburl v0.25.4

Blocking, and not obvious from the diff. dburl's duckdb header matcher was
`regexp.MustCompile("^.{8}DUCK.{8}")`, and Go's `.` matches a rune rather than
a byte, so a checksum containing a newline or a valid multi-byte UTF-8 sequence
misaligns the magic. duckdb writes `DUCK` at offset 8 after an 8-byte checksum
that covers the fixed header block, so the value is constant per storage
version: detection is all-or-none per duckdb build, not per file.

Three constants have been seen. `67274d4c681e71c3` and `2eb427ecd613b61e` pass.
`06d76f27c2dab3b9` fails, and that is the one duckdb-go/v2 writes today, so
every database created by the version usql links is undetectable. Every form
fails once the file exists: `file:x.duckdb`, `x.duckdb`, `./x.duckdb` and an
absolute path, with three different error messages between them because `Parse`
reported any `SchemeType` failure as `ErrUnknownFileExtension`.

Fixed in dburl v0.25.4, which usql is on. Verified against the released module
with a file carrying the failing constant: all four forms, `file:x.duckdb`,
`x.duckdb`, `./x.duckdb` and an absolute path, open it and return the row.

dburl's own suite never caught it because `testdata/test.duckdb` has been
byte-identical since it was added in `95e9c5f`, and its checksum is one of the
passing ones. A real file was testing the matcher for its entire life and could
not vary.

duckdb is a base driver as of 2026-09-23, which makes the cost below apply to
the default build rather than only to `most`. `go build` with no tags went from
45.7 MB to 99.9 MB. Issue 446 asked for the promotion and had been declined
because duckdb was not available on Windows out of the box; that stopped being
true when the MinGW toolchain went into the release workflow. linux/arm still
excludes it through the `Build:` constraint, so the 32-bit arm build is
unaffected.

### duckdb-go-bindings passes -rdynamic

`github.com/duckdb/duckdb-go-bindings/lib/linux-amd64` passes `-rdynamic` in its
cgo LDFLAGS, which makes the linker export every symbol into the dynamic symbol
table. That table survives `-s -w`. duckdb costs 54.1 MB of a `most` build, and
almost all of it is that table.

### gorm-dameng imported the plugin package

Resolved by removing the driver on 2026-09-23, not by a report. It was the only
importer of Go's `plugin` package anywhere in the graph, and importing it has
the same effect as `-rdynamic`. Removing it took a build with every driver from
260.2 MB to 215.7 MB. duckdb still carries the same cost through `-rdynamic`.


## W12. `\chart file=NAME` should not need terminal graphics

`doExecChart` in `handler/handler.go` returns `text.ErrGraphicsNotSupported`
before it reads its arguments. When `file` is set the chart is written to a file
as SVG and nothing is drawn in the terminal, so the check rejects a request it
could have served. Reproduce it by piping a statement into a build made with
`-tags charts`:

    printf 'select 1 as a, 2 as b \\chart type bar file=/tmp/out.svg\n' | usql sqlite3://:memory:

Move the check past the point where `cfg` is known, and apply it only when
`cfg.File` is empty. This is older than the charts build tag and is not caused
by it.


## W13. Shrink the charts stack

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


## W14. NULL scan failures that remain

The reported crash is fixed and released into main, and issues 307, 476 and 539
are closed. `drivers.NullSafeColumnType` chooses a destination that tolerates a
NULL while keeping the type information the alignment and time format need.
What is left is the same fault in two other places.

Pull requests 524, 526, 570 and 583 fix NULL scans in the metadata readers,
which build their own scan destinations rather than going through tblfmt.
Review and merge them, then decide whether a shared helper would serve them.

`\copy` has the identical fault at the `reflect.New(columnTypes[i].ScanType())`
near `drivers/drivers.go:595`. A NULL in a source column fails the copy. That
is issue 397, and issues 254 and 462 are probably the same fault. The
destination there is handed to `ExecContext` rather than printed, so the
mapping is not identical to `NullSafeColumnType`.


## W15. Output changes that came in with tblfmt v0.19.0

Landed, recorded because they change what users see and may draw reports.

`NaN`, `+Inf` and `-Inf` now print as `NaN`, `Infinity` and `-Infinity`, which
is byte identical to psql 18.6. A uint64 is always a JSON string, so a MySQL
`BIGINT UNSIGNED` column is one JSON type for every row. json output is
indented rather than compact.

usql still passes `drivers.NullSafeColumnType` rather than
`tblfmt.WithUseColumnTypes`, and must keep doing so: the latter still builds
`reflect.New(ct.ScanType())`, which is the crash, and that is upstream of
anything tblfmt can fix.


## W16. Roadmap, moved here from GitHub issues

Eighteen issues that Ken filed, mostly in January 2021, as his own roadmap in
issue form. They were never stale, only in the wrong place: a roadmap kept in
an issue tracker and a roadmap kept in this file are two roadmaps, and the
issue count then hides the real bug reports underneath them. Gemini and
DeepSeek were both asked where they belonged, and both said the duplication was
the problem; Gemini's answer, that this file is the source of truth and the
tracker is for defects and contributor-ready work, is the one taken.

They are closed on GitHub with a comment pointing here. Issue 215 stayed open
as a defect and is W5.

Anything below that is ready for someone else to pick up should be re-filed as
a narrow issue with acceptance criteria when that is true, rather than left
here with `help wanted` on it. That was the one real cost of the move, and it
is worth paying attention to: `is:issue label:"help wanted"` is how outside
contributors and aggregators find work, and this file is not indexed by any of
them.

### Already owned by another item

Four were closed because the work is described elsewhere in this file, not
because it was dropped.

137, a wrapper for the C readline library, wanted a standalone package with the
same interface as `rline`, selectable by build tag. W8 supersedes it.

165, the special `\\` separator. psql accepts an escaped backslash as a
statement separator, and only the first one: `\x \\ select 1; \\ select * from
foo;` runs the first two and then reports `invalid command \`. W3.

166, quoted string processing and variable interpolation in metacmds. The
issue carries a full side-by-side of psql against usql covering `:{?name}`,
single and double quoted interpolation, standard escape decoding, `E''` style
escaping, and backtick interpolation. `testdata/quotes.sql` is in the tree and
is the reference. W3, and the largest single piece of it.

217, additional variable types, evaluation and interpolation. A long proposal
that deliberately breaks from psql: extended variable types with their own
prefixes for connections and queries, plus shell and ruby style interpolation
and expression evaluation. W3 should settle psql parity first, since this
builds on top of it.

### psql compatibility, still wanted

141 and 142, LaTeX and troff table output, both for psql parity. These belong
to tblfmt rather than usql.

147, `\if`, `\elif`, `\else` and `\end`. Note that psql's condition evaluation
is deliberately simplistic, treating any non-empty string as true, so parity
here is easier than it looks.

158, `\ef` and `\ev`, which need introspection in place first.

160, `\errverbose`, to show the detailed error information a driver can give
beyond the message.

161, the `--echo-*` command line flags.

374, document prompt formatting. The prompt escapes are W3; this is the
documentation half.

### Proposals, not scheduled

144, a shared buffer pool for table output. Recorded as needing a large
overhaul of tblfmt, which now has its own session, so this is a question to put
there rather than work to do here.

146, a `\copy` proposal, which predates the `\copy` that exists. The open
questions in it are still open, and item 4e holds the current `\copy` defects.

154, an expanded test suite covering SELECT, INSERT, UPDATE and DELETE against
each major database and syntax compatibility across all of them, plus the
popular non-major ones. W2 has taken the container half of this; the
per-database statement coverage has not been done.

157, support for databases that have a Go API but no `database/sql` driver:
Redis, InfluxDB and IQL, Aerospike AQL, ArangoDB AQL, OrientDB SQL, Cypher and
SPARQL, JIRA JQL. Needs either a generic adapter or an overhaul of the drivers
package to admit non-SQL backends. Tier 3 holds the drivers that do have a
`database/sql` driver, and is the cheaper list.

162, `\j*` metacommands for processing fields with JavaScript in flight, with
`\jset` to define a function and `\j` to apply it to the current or last
statement buffer, composing with `\copy` for ETL. Note that usql already links
a JavaScript engine through the charts renderer, though only under the `charts`
build tag, so the cost of this is lower than it was in 2021.

216, a `\values` command producing an ephemeral result set from expressions,
for testing, variable manipulation and named queries. Depends on 217's
evaluation syntax.

267, an `\import` proposal.


## W17. A password command hook

Issue 422, accepted by Ken in the thread and moved here. Not started.

Let a user point usql at a command that returns a password, so a password
manager can be the source instead of a plaintext `.usqlpass`. The reporter has
twelve databases and keeps their passwords in KeePass; they explicitly did not
ask for KeePass support, only a hook, so that 1Password, Bitwarden, Vault,
`pass` and anything else can be wired up by the user.

The obvious design does not work. Making `.usqlpass` executable and running it
fails on Windows, which usql ships binaries for and where there is no execute
bit and a program needs an extension. The reporter noticed this themselves in
the thread and suggested naming the command in `.usqlrc` instead. Some setting
that names a command is therefore the shape to take, not a mode bit on the
existing file.

Open questions from the thread. What the command is given: the reporter
suggested the fields already used for a passfile lookup, meaning protocol,
host, port, database and user, with the driver name expanded to its full form
and the usual defaults filled in. And what the contract is: presumably the
password on stdout, a non-zero exit meaning no password rather than an error,
and a timeout so a hung helper does not hang usql.

Prior art worth reading before designing it. git's credential helpers solve the
same problem and have settled on a key-value protocol on stdin and stdout,
which is more extensible than positional arguments. psql has no equivalent, so
there is no compatibility constraint here and no reason to invent a
psql-shaped answer.

usql resolves passwords through `github.com/xo/dburl/passfile` today, called
from `env/env.go`. That package matches entries out of a file and has no notion
of running anything, so the hook belongs in usql above it rather than inside
dburl.


## W18. Two missing interfaces account for every metadata gap

Source: measured on 2026-09-26 while checking dbmeta's `docs/USQL.md`.

Of 51 registered names under `-tags all`, 21 have a metadata reader. Every one
of the 21 that fails a command fails on `\l`, on `\dp`, or on both. Nothing
fails `\d`, `\dt` or `\dn`.

    postgres pgx cockroachdb redshift sqlserver trino duckdb   11/11
    mysql mymysql memsql tidb vitess nzgo databend snowflake   10/11  no \l
    oracle godror                                              10/11  no \dp
    sqlite3 moderncsqlite                                       9/11  no \dp, no \l
    clickhouse                                                  8/11  no \di, \dp, \l
    impala                                                      6/11  no \da, \df, \di, \dp, \l

`\l` needs `CatalogReader`. `\dp` needs `PrivilegeSummaryReader`. So two
interfaces, implemented where the product supports them, take almost every
driver in usql to full coverage. That is a much smaller piece of work than a
ratio column suggests, and a ratio column is what hid it.

### Why the same gap appeared 21 times

The instruction was "implement the readers you can". That cannot be checked,
so it cannot fail, so it produced whatever happened. Naming the two interfaces
in prose is the weaker half of the fix.

The stronger half is a test over the driver registry: every registered name
either implements `CatalogReader` or has an entry in an exemption map giving
the reason it does not, and the same for `PrivilegeSummaryReader`. A
contributor cannot pass it by doing nothing, and a contributor who genuinely
cannot implement it writes one line saying why. The exemptions then become a
readable list of product facts instead of a silence.

Thirty of the 51 will need an entry on the first run. That is not the test
failing, it is the first time the gap has been written down. Expect a good
fraction of those entries to read "nobody tried" rather than "the product
cannot", because today the two are indistinguishable.

This shape is `TestEveryDialectIsMeasuredForParity` in dbmeta, which was
written for the same problem.

### One measurement artifact worth knowing

`drivers/metadata/impala/metadata.go:69` returns `struct{}{}`, commented as a
reader with no capabilities, when the handle is not a `*sql.DB`. It is
deliberate, but impala is the only driver that can lose all metadata support
without reporting anything.

## W19. Guard the numbers that documentation states

Source: an exchange with the dbmeta session on 2026-09-26, in which four
separate figures across the two projects turned out to be wrong.

Prose decays silently. Four wrong numbers were found in one afternoon: four
open pull requests that were two, eight metadata commands that were eleven,
eight gating readers that were seven, and 47 drivers that were 51.

Every one of them was a right count of the wrong thing. So a test that checks
a number without also checking what is being counted catches none of them.

### Generate before testing

usql already solves this for its largest table. `gen.go` builds both README
driver tables from the dburl registry, so the table cannot drift, because there
is nothing for it to drift from. Generation is stronger than a test: a test
reports that prose is stale, generation means it never was.

Where a number can be generated, generate it. The test is for the residue.

### What the test looks like

Count the real thing, then search each document for a number claiming to be
it. Three details are load bearing:

1. Assert that every pattern still matches something. A pattern that stops
   matching is worse than a wrong number, because the test goes green and
   guards nothing.
2. Use one anchored pattern per claim. Two counts a sentence apart get read as
   each other otherwise.
3. Check the file the numbers are about, not only the files that cite it.
   dbmeta checked three files, skipped the one the numbers described, and that
   was the file that rotted.

Write the numbers as digits. A test cannot read "eleven".

Put the unit inside the sentence the test searches, so that changing the unit
breaks the test. Match `(\d+) registered names with a reader` rather than
`(\d+) drivers`.

## W20. Test that no driver writes global process state

Source: proposed by the dbmeta session on 2026-09-26, while writing
[DRIVER.md](DRIVER.md).

`docs/DRIVER.md` tells a contributor to read the upstream driver's `init()`
and to stop if it writes global process state. That is a question a person has
to remember to ask, which is the weaker form of the same check.

Make it a test. Walk the driver packages, resolve each imported upstream
driver module, and fail when package initialization calls any of:

    slog.SetDefault      log.SetOutput         log.SetFlags
    log.SetPrefix        flag.Parse            os.Setenv
    http.DefaultClient   http.DefaultTransport
    signal.Notify        rand.Seed             os.Exit

A driver may register itself with `database/sql` and do nothing else.

### Why this is worth building

It would have caught RamSQL before the merge rather than after. Its
`engine/log` called `slog.SetDefault` from `init` with a handler on
`os.Stdout`, and a default `slog` logger also redirects the standard `log`
package, so importing that one driver changed the output of the whole binary.
It was found by watching the test suite misbehave, which took far longer than
reading an import graph would have.

### The design questions to settle first

This is not a small test, and it is listed as an item rather than written into
the documentation change for that reason.

Deciding what counts as package initialization means following calls out of
`init()` rather than pattern matching a single file. `go/packages` or
`golang.org/x/tools/go/ssa` can do it. A cheap first version could search only
the direct `init()` bodies of the imported driver package, which would still
have caught RamSQL.

A second question is what to do about a legitimate hit. An exemption map with
a reason for each entry is the shape W18 already uses, and it keeps the test
from becoming something people delete.

A third is cost. Resolving every upstream module makes this slow, so it may
belong behind a build tag or in CI rather than in `go test ./...`.

### A cheaper version that is worth doing first

An empty `main` that imports one driver, and asserts that `slog.Default()`,
the standard `log` flags and prefix, `flag.CommandLine` and
`http.DefaultClient` are unchanged after the import. It answers the question
by observation rather than by analysis, and it needs no module resolution.

## W21. Migrate metadata to dbmeta

Source: planned on 2026-09-26 with the dbmeta session, gemini and deepseek.

The plan is [DBMETA.md](DBMETA.md). This item exists so the work can be cited.
W7 is the dbmeta repository itself. W21 is usql's side of the move.

The boundary is that dbmeta reads, tblfmt renders, and usql decides what to
print. The writer does not move.

### The root cause this fixes

`drivers/metadata/metadata.go:120` declares `type Reader interface{}`. The base
reader is the empty interface, so every capability is a runtime type assertion,
and there are 52 of them outside tests. A driver that satisfies nothing
compiles, registers and passes the tests. That single line is why W18 exists.

### Order, which is not the order it was first described in

Two steps move earlier than first planned.

The formatter work comes before the reader swap. It is entirely internal, no
project outside usql imports `drivers/metadata`, and it is where the
psql-matching behavior lives.

Before even that, extract the psql-matching filtering out of `DefaultWriter`
with the current readers still in place, and confirm the output is unchanged.
dbmeta returns what the database says and does not shape results to match
somebody's output. Swapping a reader before that filtering moves changes usql's
output silently, as extra rows rather than as an error.

That step has no visible result, which is why it is the one most likely to be
skipped.

### The early win is not a native model

30 of 51 registered names have no reader. dbmeta's shared information schema
model answers 12 object kinds against any database with a standard
`information_schema`, with no per-driver query written. That adapter lands
coverage across a set of names at once and should ship before any native model
moves.

MariaDB is the first native model to move. The command gain is one, `\l`,
because the mysql family lacks only `CatalogReader`. The gain worth having is
5 of 5 sections under `\d NAME` against the 4 answered today.

The largest gains are four products where usql answers nothing at all, because
their drivers register no reader: SAP HANA at 0 of 11 today against 11 of 11
and 5 of 5 with dbmeta, Firebird at 10 of 11, Cassandra at 10 of 11 and Presto
at 8 of 11. None can come early. SAP HANA has no `information_schema`, so it
needs a native model rather than the shared adapter.

### Two bugs this fixes rather than relocates

usql changes a password in seven drivers and escapes nothing. Each concatenates
the password into the statement, so a password holding a quote or a backslash
breaks it or sets the wrong thing. Correct escaping needs the server, because
the rule is `sql_mode` on MySQL and MariaDB and `standard_conforming_strings`
on PostgreSQL.

`drivers.Version` falls back to `SELECT version();`, which Oracle does not
have, and the oracle driver registers no version function.

### Open decision

dbmeta has no release tag and its API is still moving. Whether usql depends on
an unreleased module, and whose cadence wins, is Ken's call. The per-driver
shape of the move limits the exposure without removing it.

## Tier 2: GitHub issues and pull requests

W2 to W15 are the programme, and W4 holds the issue triage. This section
is the remaining working list, reviewed on 2026-09-21 and updated on
2026-09-23, when 93 issues and 24 pull requests were open.

### Add the official Oracle driver

Add `github.com/oracle/go-oracledb`, Oracle's own driver for Go, as an
additional Oracle driver. This does not replace `github.com/sijms/go-ora/v3`,
which stays as it is.

### Released defects

Both are resolved and both issues are closed. 590 broke `go install` on 0.21.5
and was fixed by replacing the `exclude` directive with an ordinary pin of
`github.com/uber-go/tally v3.5.10+incompatible`, since `go install
module@version` refuses any `exclude` or `replace`. 589 reported the Dameng
driver failing; the driver has since been removed, so it was closed as not
planned.

### Make the completion test comprehensive

`completion_test.go` drives a real tab keypress through a pseudo-terminal and
asserts the table name is offered. It covers duckdb alone. Widen it.

Completion is per driver. Each supplies its own metadata reader, and some
supply their own completer, so a pass for one says nothing about the others.
The duckdb bug this test was written for proves the point: the driver was
registered with MySQL's completer, so every completion ran a MySQL function
against duckdb, and it reached a release because nothing tested completion for
any driver at all.

What a comprehensive version needs. One case per driver that has a metadata
reader, sharing a single recorded session shape rather than one test each. The
containerised drivers need the same fixtures `drivers_test.go` already starts,
so it probably belongs beside that rather than in the root package, and it must
not start a second set of containers. The file-backed drivers, sqlite3,
moderncsqlite, duckdb and csvq, need no container and are the cheap half.

Two things to assert, because they fail differently. That the expected name is
offered, and that the completer logged no metadata error: it swallows a failed
query and carries on with no candidates, so a broken completer and a database
with nothing to complete look identical on screen.

It is slow. Recording a session takes seconds and the duckdb build needs cgo,
so this wants its own CI job rather than a place in the matrix.


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

418 and 535 both replace the readline layer and conflict with W8. 571, 582
and 584 duplicate dependency and action work that is already done. 585 adds a
Dameng driver, which usql removed on 2026-09-23. 360 belongs in Discussions.

### Issues to close rather than fix

Done on 2026-09-23. The nine scanner reports 543, 556, 567, 568, 574, 575, 579,
580 and 581 are closed as not planned, on the govulncheck evidence in item 4a.
470 and 475 are closed as fixed, each reproduced against main first. 589 and
590 were closed earlier, and 476 and 539 with the NULL fix.

Still open and waiting on something. W8 subsumes the terminal cluster, so
122, 137, 236, 483, 490, 528 and 552 close when the rline switch lands, not
before. 263, 469 and 485 need Discussions enabled; see item 4i.

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
