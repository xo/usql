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

On 2026-09-27 the container configuration left `contrib/` as well. dbmeta's
`container` package and `dbrun` command replace `podman-run.sh`,
`podman-stop.sh`, `usql-test.sh` and every `podman-config` and `usql-config`
file. dbmeta records the products it does not model yet, with their images,
ports and setup steps, in its `docs/EVALUATION.md`. The sample `config.yaml`,
`usqlrc` and `usqlpass` files moved to `_samples/`.

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


### Two replacement pull requests are closed

Both are closed rather than merged, because neither library is the
destination. 418 proposed `reeflective/readline` and 535 proposed
`chzyer/readline`, which is the upstream `gohxs` forked eight years ago.

535 was genuinely drop-in: chzyer's `AutoCompleter` is identical in shape to
the fork's. Taking it would have removed the unmaintained dependency today.
The cost is churn in a layer that is being replaced, which is the same
reasoning that froze `drivers/metadata`.

418 was mostly decoupling rather than a swap, and most of that work is already
done. `rline/rline.go:20` already declares an `IO` interface. The only
readline type that leaks outside `rline/` is `readline.AutoCompleter`, in
three places: `drivers/drivers.go:110`, `drivers/drivers.go:494` and
`drivers/completer/completer.go:113`. See W25.

usql stays on `gohxs/readline` until rline is ready. It has had no commits
since 2017.

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


## W18. Two missing interfaces account for every metadata gap (Superseded by W21)

Source: measured on 2026-09-26 while checking dbmeta's `docs/USQL.md`.

Superseded on 2026-09-27 by D3: no metadata reader is written in usql, so the
two interfaces are not added here. The measurement below stays, because it
says what W21 has to answer in `dbmeta`.

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

`drivers.Version` falls back to `SELECT version();` for any driver that
registers no version function. Oracle is not one of them: it registers one
through `orshared.Register`, which an earlier draft of this item missed.
Oracle's defect is different and is W22.

### Moving Oracle costs `\l` until dbmeta answers it

dbmeta's Oracle model registers no `Databases` query. Oracle is one of two
dialects there that cannot answer `\l`, recorded on the grounds that Oracle has
one database per instance and no list to read.

usql answers `\l` for Oracle today, for administrators only, through
`v$parameter`. So moving Oracle as things stand trades one command for
correctness everywhere else, going from 11 of 11 to 10 of 11 on that axis.

dbmeta will not write the query until somebody decides what `\l` means on a
multitenant Oracle server. `SYS_CONTEXT('USERENV', 'DB_NAME')` returns the
container name rather than the pluggable database name, and a listing that
names the container when the connection is to a pluggable database is worse
than no listing. If the answer is the pluggable databases visible to this
connection, there is a view for it. If the answer is the one database you are
attached to, it is a single row.

That decision is a dependency on dbmeta, not usql work. Pull request 570 pairs
`SYS_CONTEXT` with `all_db_links`, which lists links to other databases rather
than databases, which suggests usql's Oracle catalogs query is answering two
questions at once.

### The metadata layer is frozen

Decided on 2026-09-26. `drivers/metadata` takes no more changes before the
migration.

Small fixes are still worth having in principle, and both external reviews
argued for merging them, on the grounds that the migration has no date and
main ships to users today. Ken decided the other way: the layer is being
replaced, and churn in it costs merge pain for code that is going away.

So pull requests against the metadata readers are held rather than merged, and
the defects they report move to the migration instead. Held on this basis:

    524   oracle version query needs privileges an ordinary user lacks
    570   oracle catalogs query reads administrator-only views
    583   postgres catalog access privileges crash on NULL

A held pull request is not a closed one. Each defect has to survive the move,
which means W21 inherits them and the new readers have to answer for them.

### Implemented, waiting for a dbmeta tag

Implemented on 2026-09-30 and staged for review, following D4. usql reads all
of its metadata from dbmeta: the describe commands, tab completion, the
version, the current user and the change of a password. Every usql reader is
gone. `metacmd/describe` holds the commands, which print dbmeta's answers the
way psql prints them, and a small source of names for the completer.

It builds against a dbmeta checkout through a local `go.work`, and is not
committed until dbmeta has a tag for `go.mod` to require. The steps of
`docs/DBMETA.md` it covers are 0 to 2, 4 to 6, and the count of step 4.
Step 3, the information schema model for a database with no model of its own,
is dbmeta's to register, and Ken is deciding which databases get one.

### Open decision

dbmeta has no release tag and its API is still moving. Whether usql depends on
an unreleased module, and whose cadence wins, is Ken's call. The per-driver
shape of the move limits the exposure without removing it.

When usql calls dbmeta, choose the dialect from `u.Dialect`, which dburl
v0.32.0 sets, and never from `u.Driver`. Every PostgreSQL URL now has Driver
`pgx`, which is not a dbmeta dialect. `u.Dialect` is `postgres` for all of them,
`sqlite3` for moderncsqlite, `oracle` for godror, and `mysql` for TiDB,
SingleStore and Vitess. dbmeta's `docs/COMMANDS.md` says the same under "Wiring
usql up".

## W22. Metadata and version queries assume a privileged user (Superseded by W21)

Source: found on 2026-09-26 while triaging pull requests 524 and 570. Answered
the same day.

usql's Oracle queries cannot be run by an ordinary user.
`drivers/oracle/orshared/orshared.go:38` reads the version from `v$instance`,
and the catalogs query reads `v$parameter` and `dba_db_links`. All three are
administrator objects. Pull requests 524 and 570 report this independently.

Nobody had asked usql's queries as anyone other than the administrator, because
the tests connect as a privileged user, so a query an ordinary user cannot run
passes every test here.

### Why this is superseded rather than open

dbmeta already solves it, measured rather than asserted. Its Oracle model
touches exactly one `v$` view in the whole model, which is the version query,
and uses `all_` views everywhere else. On Oracle 26ai as a user granted only
CREATE SESSION:

    SELECT version FROM v$instance                 ORA-00942, no such view
    SELECT banner FROM v$version WHERE ROWNUM = 1  answers

So usql prints a version for an administrator and nothing for anybody else, and
dbmeta prints one for both.

It is also not a one-off audit there. dbmeta runs `TestPrivilegeParity` in CI
on every push, asking every query as the administrator and as each lesser
principal the product has, with the results checked in. A query that starts
depending on who is asking fails their build.

The metadata layer here is frozen, so these queries move to dbmeta rather than
being fixed in place. The defect does not survive the migration. Pull requests
524 and 570 stay held.

### One caution that outlived the item

`SELECT banner FROM v$version` returns five rows on Oracle 11.2, one each for
the database, PL/SQL, CORE, TNS and NLSRTL. It became a single row in 18c.
dbmeta's query carries `WHERE ROWNUM = 1` and has since it was written. Pull
request 524 has no predicate and is nondeterministic on any server before 18c.

## W23. Variables need a scrub, and `\copy` does not resolve them

Source: issue 495 and pull request 542, triaged 2026-09-26.

`\copy` goes straight to `dburl.Parse` on its source and destination, so a
connection variable set with `\cset` is not resolved. Issue 495 reports it:

    \cset MSSQL mssql://USER:PASS@HOST:PORT/DATABASE
    \copy MSSQL csvq://. 'YYY'
    error: invalid database scheme

`\c MSSQL` works, and `\copy` with the URL spelled out works. Only the
combination fails.

Pull request 542 is closed rather than merged. It resolved the variable but
then hand-built a `dburl.URL` from a driver and a DSN instead of going through
`dburl.Parse`, which leaves `Host`, `User`, `Path`, `RawQuery` and `GoDriver`
empty. `drivers.Open` passes the whole URL to any driver with its own `Open`,
and six have one: cassandra, moderncsqlite, pgx, postgres, sqlserver and
vertica. So it would have broken `\copy` for postgres and pgx, which are the
drivers most likely to be used with it.

The correct shape is to resolve the variable to a connection string and pass
that string to `dburl.Parse`, so that one code path handles both cases.

### This is part of a larger change

Variables in general are getting a clean up, and this fix belongs inside it
rather than ahead of it. Doing it piecemeal means touching `\copy` twice.

## W24. Read a command or a script from standard input

Source: pull request 479, triaged 2026-09-26. Planned for the next major
release.

`usql -c -` and `usql -f -` should read from standard input, which is what
psql does.

Pull request 479 is closed rather than merged. It called `scanner.Scan()`
exactly once on a `bufio.Scanner` with the default line splitter, so it read
only the first line and silently discarded the rest of a piped script. It also
deleted two imports that are still used, and it predates two years of changes
to `run.go`.

The feature is wanted. Read all of standard input rather than one line, and
decide what `-f -` does when it is given more than once.

## W25. Stop leaking `readline.AutoCompleter` out of the rline package

Source: found on 2026-09-26 while triaging pull requests 418 and 535.

`rline/rline.go:20` already declares an `IO` interface, so usql is mostly
decoupled from its line editor already. One type escapes:

    drivers/drivers.go:110              NewCompleter field on drivers.Driver
    drivers/drivers.go:494              NewCompleter func
    drivers/completer/completer.go:113  NewDefaultCompleter func

`readline.AutoCompleter` is in the public `drivers.Driver` struct, so the line
editing library reaches into the driver API. Every driver that supplies its own
completer names the type.

Declare the interface in usql and change the three signatures. The interface
has one method, so this is not inventing an abstraction, it is naming one that
already exists under a different package.

This is deliberately not shaped for any particular library. It is the part of
pull request 418 that is worth having regardless of which editor wins, and it
means W8 does not have to change the driver API.

## W26. Prompt formatting: PROMPT2, PROMPT3 and the documentation

Source: pull requests 577 and 578, closed 2026-09-26. To be done differently.

usql has `PROMPT1` and nothing else. `handler/handler.go:561` already documents
`%w` as "Whitespace of the same width as the most recent output of PROMPT1.
This can be used as a PROMPT2 setting", so the documentation already describes
a variable that does not exist.

Pull request 577 added `PROMPT2` with a default of two spaces and implemented
`%w` by tracking the last rendered prompt width on the handler. Pull request
578 added a prompt formatting section to the README and stated that `PROMPT2`
and `PROMPT3` use the same escape sequences. Neither adds `PROMPT3`, so 578
documented two variables that do not exist and one that another pull request
was adding.

Both are closed because this is being done differently and soon.

### What has to be true when it is done

`PROMPT2` is the continuation prompt for a statement spanning several lines.
`PROMPT3` is what psql prints while reading `COPY` input, so it needs the copy
input path rather than only a variable. Adding the variable without that path
makes it inert.

If `%w` is implemented, its width has to be measured in terminal columns rather
than bytes or runes, because a prompt usually carries colour escapes and may
carry wide characters.

Do not document a prompt variable before it works. That is how the current
state arose.

## W27. Retire, replace and regroup drivers

Source: a review of every driver on 2026-09-27, with Gemini. DeepSeek was asked
five times and returned nothing each time. Ken made the decisions below the
same day. Module dates come from the Go proxy.

Remove these four. Each also drops its scheme from dburl, which has been told:

- mymysql, `github.com/ziutek/mymysql`, last released 2015-01-09.
  `go-sql-driver/mysql` covers the same database.
- adodb, `github.com/mattn/go-adodb`, last released 2018-05-15. odbc covers
  it. The `oleodbc` override goes with it.
- sapase, `github.com/thda/tds`, last released 2019-09-27. `SAP/go-ase` last
  changed on 2025-05-22 and is not a replacement.
- ignite, `github.com/amsokol/ignite-go-client`, last released 2019-01-04.

Switch these three to a maintained module:

- flightsql moves to `github.com/apache/arrow-go/v18`, v18.8.0. That module is
  already an indirect dependency, so this removes a second Arrow tree.
- maxcompute moves to `github.com/aliyun/aliyun-odps-go-sdk/sqldriver`,
  v0.4.26. It registers as `odps` and takes a different DSN, so dburl's
  generator changes with it. The SDK carries its own Arrow package, so check
  the dependency weight before merging.
- postgres moves from `github.com/lib/pq` to `github.com/jackc/pgx/v5/stdlib`.
  DSN handling, error text and type mapping change, so re-record the
  `cli_test.go` goldens and run it against a live database before release.

Regroup these two:

- csvq moves from `base` to `most`. It is finished rather than broken.
- h2 moves to `bad`, as triage before removal. `github.com/jmrobles/h2go` last
  released on 2020-11-14.

Keep these two, on forks that Ken is starting:

- cassandra. `github.com/MichaelS11/go-cql-driver` last released on
  2020-09-20 and wraps `gocql/gocql`, which has since moved to
  `github.com/apache/cassandra-gocql-driver/v2`.
- couchbase. `github.com/couchbase/go_n1ql` last changed on 2022-03-03.

Both forks bring the code up to modern Go. usql and dburl switch to them when
they exist.

As of 2026-09-27 the removals, both regroupings, and the flightsql and
maxcompute switches are done, on dburl v0.31.0. usql registers the odps driver
under `maxcompute` as well, because that is the name dburl gives the scheme.

The postgres switch is done too, on dburl v0.32.0, following dburl's D22.
`postgres://`, `cockroachdb://` and `redshift://` reach pgx, which is now in
`base` and also answers to the `postgres` build tag. lib/pq stays supported
as `drivers/libpq`, in `most`, and dburl sends `pq://` and `libpq://` to it.
It still registers as `postgres`, which is lib/pq's name in `database/sql`.

The Cassandra fork is done. The cassandra driver uses `github.com/xo/cql`
v0.1.0, which wraps `github.com/apache/cassandra-gocql-driver/v2`. It returns
the real connection error, so usql no longer captures the gocql log to find
it. A NULL now shows as empty rather than as a zero value.

The Couchbase driver is replaced by `github.com/xo/dbimp/couchbase` v0.1.0,
written in dbimp rather than forked from `go_n1ql`. It registers as
`couchbase` and takes only a `couchbase://user:pass@host:port/` URL. It
landed with dburl v0.33.0, which renamed the scheme from `n1ql` and keeps
`n1ql` and `n1` as aliases, following dburl's D25. usql adds
`txtimeout=30m` to a URL that has none, because the server ends a transaction
after 15 seconds by default. It does not set `durability_level`, and a
single-node server needs `durability_level=none` to commit a transaction.

Netezza was removed on 2026-09-29, and its scheme goes from dburl. Ken
decided it: the driver was never tested, nobody is known to use it, and no
copy of Netezza is available to test with.

Most driver packages under `drivers/` have no `Group:` line and fall to
`most` by default, which breaks rule 10 in `AGENTS.md`. Add the line to each
as they are touched.

## W28. Statement handling for ODBC, csvq, Athena and Cosmos

Source: Ken, on 2026-09-30, while dbmeta took over the grammar of each
database as its D143.

Four drivers changed a statement before usql ran it, through the `Process`
field of `drivers.Driver`. Ken removed those hooks. What a database does with a
statement belongs in dbmeta's syntax options, applied for every driver of it,
and that needs a larger design than D143. Each database is taken on its own.

What each hook did, and what is lost until this item is done:

- ODBC stripped a trailing semicolon, unless the statement ended with `END;`,
  when the URL set `usql_trim`. The key is now ignored.
  `dburl.OdbcIgnoreQueryPrefixes` still keeps `usql_` keys out of the DSN.
- csvq ran a `SHOW` statement as exec and pointed csvq's output at standard
  output, because csvq prints the answer to `SHOW` rather than returning rows.
  `SHOW` now prints nothing.
- Athena and Cosmos stripped a trailing semicolon.

The other `Process` hooks, in arangodb, influxdb, neo4j and surrealdb, did
nothing but run every statement as a query. That is now the `AlwaysQuery`
field, and `Process` is no longer a field of `drivers.Driver`.

### Open for discussion

Ken may remove ODBC, Athena and Cosmos, because none of them can be tested
against a server that CI can start. He would keep csvq, and wants to discuss
it with others first. Nothing is removed until that discussion.

### What has to be true when it is done

- dbmeta holds the rule for each of the four that is kept, and usql applies it
  with no code for that database.
- No code in usql changes a statement's text for one database.

## W29. Move the SQL of \copy into dbmeta

Source: Ken, on 2026-09-30. After the metadata readers went to dbmeta (D4),
the `Copy` hook of `drivers.Driver` holds the last SQL that usql writes for
one database or another. It needs a careful design, like W28. Ken set it due
before the next release, on 2026-10-02.

What `\copy` builds today:

- `drivers.CopyWithInsert` and `drivers.FlexibleCopyWithInsert`, in
  `drivers/drivers.go`, write `INSERT INTO table VALUES (...)`. Each driver
  passes the placeholder of its database: `$1` by default, `?` for ClickHouse,
  csvq, DuckDB, Exasol, MySQL, SQLite and Trino, `:1` for Oracle, and SQL
  Server's own. Impala runs the inserts outside a transaction.
- The same code, and the pgx and libpq drivers, find the columns of the target
  with `SELECT * FROM table WHERE 1=0`, or `SELECT cols FROM table WHERE 1=0`
  when the target names its columns.
- pgx and libpq then use the COPY protocol of their Go driver, through
  `CopyFrom` and `pq.CopyIn`. That is a feature of the driver, and D5 keeps it
  in the driver.

The placeholder, the insert statement and the query for the columns of a
table all describe the database, so by D5 they belong in dbmeta. The design
has to decide how a client asks dbmeta for them, and what happens for a
database that dbmeta has no model for.

### Other SQL that is still in usql

A search of every Go file outside the tests found no other statement that
reads metadata, runs DDL or changes a password. These remain:

- `Version` hooks for seven databases that dbmeta has no model for: Athena,
  csvq, ArangoDB, Neo4j and YDB run a query, and SurrealDB and InfluxDB call
  their client library.
- Two fallbacks in `drivers/drivers.go` for a database with no model:
  `SELECT version();` in `Version` and `SELECT current_user` in `User`.
- Text in `metacmd/describe` that is printed and never run: the
  `FOREIGN KEY ... REFERENCES` line built for SQLite, the `CHECK (...)` around
  an expression from information_schema, and the `CREATE OR REPLACE VIEW`
  line that `\sv` puts above a definition.
- The keyword tables in `drivers/qtype.go`, which decide whether a statement
  returns rows. They hold statements of PostgreSQL, MySQL and CQL.

Ken decides which of these move to dbmeta, and in which item.

## W30. Research an ODBC driver written on purego: Done

Done on 2026-10-07. `github.com/xo/odbc` v0.1.0 is that driver, pure Go on
purego, with no cgo, and usql uses it. The crash that it has in usql's default
build, and the question of its metadata, are in W40.

Source: Ken, on 2026-09-30.

Research a new ODBC driver for `database/sql`, written from the start on
`github.com/ebitengine/purego`. purego calls C functions from Go without cgo,
by loading the shared library at run time.

usql's `odbc` driver uses `github.com/alexbrainman/odbc`, and the README marks
it as needing cgo. So it is left out of a build without cgo, and CI installs
`unixodbc-dev` to build it. A driver on purego would load the ODBC driver
manager, such as `libodbc` from unixODBC, iODBC or `odbc32.dll`, when a
connection opens.

Questions for the research:

- Which platforms purego supports, and whether it can call every ODBC
  function the driver needs, including those that take callbacks or pass
  structs by value.
- How the driver finds the driver manager on each platform, and what error a
  user sees when there is none.
- How the data types of ODBC map to Go, and how that compares with
  `alexbrainman/odbc`.
- How to test it in CI against a real database through a real ODBC driver.
- Whether it lives in the xo family, and under which name.

This bears on W28, which asks whether ODBC stays in usql at all. A driver
without cgo could join the `most` group, and it could be tested.

## W31. The rest of psql's describe commands

Source: Ken, on 2026-09-30, once W21 put all of usql's metadata on dbmeta.

Status: Implemented, waiting for W21 to be committed.

usql now answers every command in dbmeta's `docs/COMMANDS.md` that dbmeta has
a query for. `metacmd/describe/lists.go` holds one spec for each: the query,
psql's title, and psql's columns. `metacmd/cmds.go` routes the commands, and
`TestEveryListIsACommand` checks that each spec is reachable. The new commands
are `\dA`, `\dAc`, `\dAf`, `\dAo`, `\dAp`, `\db`, `\dc`, `\dconfig`,
`\dC`, `\dd`, `\dD`, `\ddp`, `\dE`, `\des`, `\det`, `\deu`, `\dew`,
`\df` with `anptw`, `\dF`, `\dFd`, `\dFp`, `\dFt`, `\dg`, `\dl`,
`\lo_list`, `\dL`, `\do`, `\dO`, `\dP`, `\drds`, `\drg`, `\dRp`,
`\dRs`, `\dT`, `\du`, `\dx`, `\dX`, `\dy`, `\sf`, `\sv` and `\z`.

Two fixes came with them. `\dS+`, `\dtS+` and every other form with both S
and + were unknown commands, because a command name expanded each letter in
brackets alone. And a pattern now shows the objects of the system, as psql
does, which checks `!showSystem && !pattern`.

### What is still different from psql

dbmeta added the fields these commands lacked as its D147: the whole
definition of a function for `\sf`, the value with its unit for `\dconfig`,
the size for `\dT+`, the storage type for `\dAc`, the applicable types for
`\dAf`, the ICU rules for `\dO`, psql's columns for `\dX`, and the token map
for `\dF+`. usql prints all of them. These differences remain:

- An empty result prints `(0 rows)` with no title and no header. tblfmt does
  this for every result, not only for describe commands.
- The second pattern of `\dAc` and `\dAf` matches the name of the class or
  family, where psql matches the input type. dbmeta takes a name.
- `\dP` takes no `itn` letters.

A describe command now reads a name between double quotes as it is written,
as psql does. Before, usql removed the quotes before the pattern was folded,
so `\d "Foo"` looked up `foo`.

## W32. Review the size that each driver adds to the binary

Source: Ken, on 2026-10-01. Measure again at each of these times:

- Before a release.
- After `update-deps.sh` moves the dependencies.
- When a driver is added or switched to another module.

Run `./size.sh -g most`, which measures the drivers that a release ships,
and replace the table below. CONTRIBUTING.md says how the script works. A
driver whose share grows a lot is a question for Ken: keep it in `most`,
move it to `all`, or find a smaller module.

### The measurement of 2026-10-01

A release build with every driver of `most` was 203.1 MB, stripped, with
go1.27.1 on linux/amd64. The ten largest drivers make up 133.9 MB of it, and
29 of the 51 drivers add less than 1 MB each.

| MB saved | Driver left out | What it brings |
| ---: | --- | --- |
| 54.2 | duckdb | the C++ library of DuckDB, with 4 MB of ICU data |
| 17.4 | spanner | the Google Cloud SDK, gRPC, and the protos of Envoy |
| 13.5 | oracle | the character set tables of go-ora |
| 10.9 | snowflake | the clients of S3 and Azure Blob storage, and Arrow |
| 9.6 | chai | Pebble, the storage engine of CockroachDB |
| 7.6 | athena | version 1 of the AWS SDK |
| 7.1 | ydb | the YDB SDK and its gRPC protos |
| 4.8 | clickhouse | ch-go and clickhouse-go |
| 4.6 | databricks | its SDK, with Arrow v12 |
| 4.2 | bigquery | the Google Cloud SDK, with Arrow v15 |
| 3.6 | moderncsqlite | SQLite translated to Go |
| 3.2 | csvq | the csvq engine |
| 2.7 | dynamodb | the DynamoDB client of AWS SDK v2 |
| 2.5 | sqlserver | go-mssqldb |
| 2.2 | ots | the Tablestore SDK |
| 1.9 | maxcompute | the ODPS SDK |
| 1.7 | sqlite3 | the C library of SQLite |

Three things to read the table by:

- A number is what leaving out that one driver saves. Two drivers that share
  a module each save only what the other does not use. So pgx and mysql show
  almost nothing, because postgres, cockroachdb, memsql and the other aliases
  import them.
- The build holds three versions of Apache Arrow: arrow-go v18 for flightsql,
  v15 for bigquery and v12 for databricks. A driver that moved to v18 would
  share it.
- `bad` drivers are not measured, because no release builds them.

## W33. Make the README a landing page

Source: Ken, on 2026-10-01, after a review with Gemini and DeepSeek.

README.md is 1662 lines. 714 of them are Features and Compatibility, and 184
are pasted output of `--help` and `\?`. Both reviewers proposed that the
README answers a visitor's first questions, and that docs/ holds the
reference. Their merged proposal is about 230 lines:

| Section | Lines |
| --- | ---: |
| Header, pitch and badges | 10 |
| Install commands | 25 |
| First five minutes: install, connect, query | 40 |
| Supported databases, a compact generated list | 25 |
| Common connection examples | 30 |
| Backslash commands, an overview | 25 |
| Features, each linked into docs/ | 50 |
| Documentation index | 15 |
| Contributing | 10 |

The reference moves under docs/, with the listings generated by gen.go: the
full driver table, the output of `--help`, the `\?` listing, and the
variables. The configuration, copying, terminal graphics, time formats and
troubleshooting become pages of their own. The two reviewers agreed that the
database names stay in the README, because a visitor and GitHub's search look
for them there.

Open for Ken: the names of the files under docs/, and whether one generator
writes the README sections, docs/ and the man pages of W34 from one source.

## W34. Man pages for usql

Source: Ken, on 2026-10-01, after a review with Gemini and DeepSeek. The man
pages are generated from code, a person can edit their prose, and every
release package installs them.

The pages:

- usql(1): the command line, the environment and the files.
- usql-commands(1): the backslash commands and the variables. Both reviewers
  first said section 7, and then agreed that section 7 is for conventions,
  not for the use of a program.
- usqlrc(5): the formats of config.yaml, .usqlpass and the RC file.

How they are made:

- gen.go fills templates from the data that the code already holds: the cobra
  flags in run.go, the descriptions in metacmd/descs.go, and env's
  variables. A person writes the prose of each page in the template.
- Both reviewers said that go-md2man cannot write the definition lists that
  120 commands need. That is wrong. go-md2man v2 turns a Markdown definition
  list into roff `.TP`, which is the list a man page uses. So the prose can be
  Markdown. Writing man(7) directly is the other choice, and it is harder for
  a contributor to edit.
- A backslash in the text is escaped for roff. Unescaped, `\d` is a roff
  command that moves the text down.
- The generated pages are committed under man/, and a test fails when they
  differ from what the generator writes.
- The date in `.TH` is set when a release is tagged, with SOURCE_DATE_EPOCH
  as an override for packagers. A date from the last commit would change the
  pages, and fail the test, on every commit.

How they are installed:

- The tarball and the zip of build.sh hold man/ next to the binary.
- The Homebrew formula in xo/homebrew-xo calls man1.install and man5.install.
- The AUR packages install them with `install -Dm644` under
  /usr/share/man.
- A hidden flag prints a page, for Docker and Windows, which have no man(1).
  It is a flag and not a subcommand, because a bare word on the command line
  is a DSN, such as a named connection. usql prints its completion scripts
  through flags too, such as `--completion-script-bash`.

Open for Ken: Markdown or man(7) for the prose, the name of the flag, and how
the version in `.TH` gets into a committed page.

## W35. Match usql to transit's usql grammar

Source: Ken decided the design of transit's usql grammar on 2026-10-01, as
transit's D101 and D102. The grammar parses usql's input: the meta commands,
the variables, and the SQL that it hands to a SQL grammar for the dialect.
Nothing is due until usql uses the grammar. Two points touch usql:

- Outside a string, `\\` is psql's separator meta command. It ends the
  arguments of the meta command before it, and the rest of the line is SQL.
  usql's splitter reads `\\` as an escaped backslash today (stmt/stmt.go), so
  the two differ there until usql follows psql.
- The grammar splits a meta command's name into its base and its modifiers,
  such as `dtS+`, from usql's command list in metacmd/descs.go. So a new or
  renamed meta command in usql needs a change to the grammar too.

The rest of the design is transit's. The grammar has one variant for each
family of dialects that usql's lexer options make: postgres, mysql, sqlite,
block comments only, cql, and no options. Variables reach the SQL grammar as
placeholders of the same length. Each meta command is one node, with a name
and typed arguments that usql maps to the kinds it completes. `\if` and its
branches are flat commands, as in psql.

## W36. Read input through rline and the usql grammar

Source: Ken, on 2026-10-01. The usql grammar of transit replaces the stmt
package, and rline replaces gohxs/readline and usql's rline package. usql has
no API to keep, and it takes the API of rline and transit as they are. This
design came from transit, rline, Gemini and DeepSeek. Where they disagreed,
psql decides.

### What the grammar gives

The root of the tree is a list of statements and meta commands. A statement
ends at a `;` outside parentheses, at a meta command, or at the end of the
input. An open string, quoted identifier, dollar quote or block comment is one
token to the end of the input, never an ERROR node. A variable is a node with
a byte range. Each family of dialects is its own language, so a change of
family on `\c` parses the unfinished text again, with no old tree. The
grammar joins nothing across a `;`, so a PL/SQL or T-SQL block, a CQL batch,
Oracle's `/` and T-SQL's GO are usql's to handle, as today.

### The design

- The query buffer stays, with psql's meaning. It is the raw text of the
  statements that are not sent yet.
  - A statement that ends in `;` joins the buffer, the buffer runs, and it
    becomes the last query.
  - A statement with no `;` stays in the buffer.
  - `\g` sends the buffer once, or the last query when the buffer is
    empty. `\p`, `\r`, `\e` and `\w` act on the buffer, as in psql.
  - A first reviewer proposed to replace the buffer with the last statement.
    That runs `select 1 \g` twice, so it is not taken.
- An entry is finished when its last top-level node is a statement that ends
  in `;`, or a meta command. An open string, comment or parenthesis keeps it
  open. An entry of only white space or a comment is finished when the
  buffer is empty. rline's WithContinue asks this.
- The parse is kept with the text it came from, and is shared by
  WithContinue, the highlighter and the completer. A parse runs only when the
  text differs, after an edit of the old tree. rline calls the three in no
  promised order, and does not always give them the same text.
- With no terminal, rline reads plain lines and calls none of the three.
  usql collects the lines itself. It parses from the first byte that has not
  run, runs every complete node except the last, and counts the lines that
  ran, for the line numbers of errors. This covers pipes, `-f` files, `\i`
  and the RC file. A `-c` string is parsed whole.
- A variable is replaced when its node runs, by its byte range, so the value
  is never scanned again. That fixes the splicing of `:'x'` and the colon of
  `::`. `:'x'` and `:"x"` double the quote in the value, as psql does.
- `\if`, `\elif`, `\else` and `\endif` get psql's meaning: a stack, and
  an inactive branch runs nothing and expands nothing.
- help, quit and exit are recognized when an entry is only that word.
- CQL batches keep the joining that usql does today, from dbmeta's Batches.
- The stmt package goes after a differential test passes. The test runs the
  old splitter and the new one over a corpus: the stmt tests, testdata and
  the grammar's corpus. The two known bugs of stmt are listed as expected
  differences.

### What waits for rline

These are rline's backlog items 33, 34 and 35, which are for Ken:

- A prompt for each row of an entry, for `%R`. Until then, a fixed
  continuation marker.
- Text to start ReadLine with, for `\e`. Until then, the edited text goes
  into the query buffer and is printed.
- A way to change WithContinue for `\prompt`. Until then, the function
  knows when usql reads for `\prompt`. The answer is also taken out of the
  history.

### Decided

Ken decided on 2026-10-01 that the chroma lexer goes, with no temporary
highlighter in its place. Highlighting comes back from transit's grammars,
whose tag is due soon. Until then, input has no color.

Ken also decided that usql does not add T-SQL's GO. Taken as defaults, for
Ken to change: Oracle's `/` stays out too, as in psql, and the package that
replaces stmt is named `input`.

## W37. Bring back the tests against databases, and their recordings

Source: Ken, on 2026-10-02. Due before the next release.

The tests of `drivers` and `drivers/clickhouse` started their own containers
through `ory/dockertest`, and loaded jOOQ's Sakila schema into PostgreSQL,
MySQL, SQL Server, Trino and ClickHouse. Ken had them removed on 2026-10-02,
with their recordings, because dbmeta owns every container setup. These went:

- TestWriter, which compared the output of the describe commands with the
  recordings in `drivers/testdata`
- the TestCopy of `drivers` and of `drivers/clickhouse`, which copied rows
  between databases with `\copy`
- the image definition, `gen-golden.sh`, and the fixtures in
  `drivers/testdata` and `drivers/clickhouse/testdata`
- `github.com/ory/dockertest/v4`, and the job of CI that ran them

Much of the output of the describe commands still needs fixing, and the rest
of usql's plumbing is updated first. Then:

- Write the tests again against servers that dbmeta's `dbrun` starts. A test
  takes the URL of its server from the environment, and skips when it is not
  set.
- Load a fixture of the test's own over the connection, rather than build an
  image.
- Record the output of the describe commands again, and the sessions of
  TestCLI in `testdata/cli`.
- Cover `\copy` between two databases again.

## W38. Run CI on every database, through dbmeta and dbrun

Source: Ken, on 2026-10-02. Due before the next release.

Overhaul usql's workflows, so that they test usql against every database
that can run on GitHub's runners, through dbmeta's `dbrun`. usql writes no
container setup of its own.

dbmeta's `test.yml` shows the shape. One job builds `dbrun` and prints the
list of releases with `dbrun list --json --names tested`, and `nightly` for
the full list. A matrix runs one job for each release. Each job starts its
server with `dbrun`, takes the URL from `dbrun dsn`, and runs the tests. The
workflow names no database, so the list lives only in dbmeta.

- Choose which of dbmeta's selectors usql runs on each push, and which every
  night.
- Run W37's tests, TestCLI and the manual checks of `docs/RELEASE.md` that can
  be automated, such as changing passwords and the describe commands.
- Keep the jobs that build every driver, the static builds and the release
  archives.

## W39. Move Avatica to dbimp's driver: Done

Source: dbimp, on 2026-10-07. dbimp released an Avatica driver in v0.10.0
(its W24), and asks usql to use it in place of
`github.com/apache/calcite-avatica-go/v5`. That drops protobuf, the Kerberos
library and the digest library from `go.mod`.

Done on 2026-10-07, on dburl v0.41.0 and dbimp v0.10.1. dburl's D47 made the
`avatica` scheme name `github.com/xo/dbimp/avatica`, with the DSN
`avatica://user:password@host:port`, no port added and no path, and the keys
`tls` and `auth`. phoenix stays an alias. The old DSN was `http://`, which
the new driver refuses. A driver registered under a scheme that dburl does
not know fails `TestRegistryIsKeyedByScheme`.

In `drivers/avatica/avatica.go`:

- It imports `github.com/xo/dbimp/avatica`. The Apache driver and the modules
  that only it needed, and some left over from earlier removals, are out of
  `go.mod`.
- `Err` reads `Code` and `Message` of `*avatica.Error`.
- `IsPasswordErr` takes HTTP 401, and also the error that the server gives a
  wrong password: HTTP 500 with the exception `SQLInvalidAuthorizationSpec`.
- `Version` runs `VALUES (DATABASE_VERSION())`, which works on the standalone
  server and returns the release of HSQLDB, not of Avatica. The Phoenix Query
  Server has no such statement, and the driver does not send
  `databaseProperties`, so there the banner says the release is unknown. A
  dbmeta model for Avatica would be the place for it.

Checked on 2026-10-07 with `dbrun`, on avatica-1.29.0 and phoenix-2.0-5.0:
the banner, a table of several types with NULLs, a syntax error, a missing
table, a duplicate key, a rollback and a commit, and a wrong password.
avatica-1.28.0 was not run.

## W40. Move ODBC, Trino and Presto to the drivers that dburl now names

Progress on 2026-10-07, on dburl v0.44.0, dbimp v0.13.0 and xo/odbc v0.1.0.
Trino and Presto are moved and checked. ODBC is moved, and it crashes in the
default build. See the end of this item.

- Trino and Presto: both import `github.com/xo/dbimp/trino`, and each has its
  own hooks, so that `no_trino` leaves Presto alone. dburl writes the user
  `user` and `flavor=presto`, so `trino://host` and `presto://host` connect.
  `RowsAffected` is `drivers.RowsAffectedOrZero`, and `\copy` inserts without
  a transaction, because Trino's connectors take a write only with
  autocommit. Presto had no `\copy` before, and has one now. Checked on
  trino-483 and presto-0.299 with dbrun: the banners name the product and
  its release, types with NULLs, row counts, errors, `\copy` of three rows,
  and `\dt` and `\d`. The wrong password was not checked, because those
  servers take any user.
- `drivers.RowsAffectedOrZero` is new, for a server that gives no count. It
  replaces the copies in Databend and ClickHouse. The shared `\copy` code
  counts one row for an insert that gives no count, and still stops on any
  other error.

Source: dburl, on 2026-10-07. Ken decided three scheme moves in dburl that
day (its D49 and D50). They are staged there and not released. Do nothing in
usql until dburl sends the tag.

- ODBC moves to `github.com/xo/odbc`, which needs no cgo. The registered name
  stays `odbc`, and the connection string is passed on as it is. This bears
  on W30, which researched an ODBC driver on purego, and on W28, which asked
  whether to keep ODBC at all. Take both to Ken.

  xo/odbc v0.1.0 was released on 2026-10-07, so dburl's release no longer
  waits for it. It is a purego driver for Linux, macOS and Windows, with no
  cgo. Its data source name is a URL, `odbc+<driver>://user:pass@host:port/db`,
  or an ODBC connection string. A setting of the ODBC driver goes in the
  query, such as `INITSTMT` for ANSI SQL mode on MariaDB and MySQL. Its
  README has an FAQ for that. usql's old driver read the key `usql_trim`,
  which W28 removed.

  Its `sql.Conn.Raw` gives a connection with `GetInfoString`,
  `GetInfoUint16`, `GetInfoUint32`, `Tables`, `Columns` and `PrimaryKeys`,
  which read the metadata of any database that has an ODBC driver.
  `odbc.Drivers` and `odbc.DataSources` list what the driver manager knows.
  dbmeta decided on 2026-10-07 to build no model on them, and the sender says
  that the fallback for a database that dbmeta does not know belongs to the
  client. That would be a metadata reader in usql, which D3 forbids. It is
  Ken's decision whether ODBC is the exception, or whether its describe
  commands stay unsupported.
- Trino and Presto both move to `github.com/xo/dbimp/trino`, which registers
  `trino`. For `presto`, `URL.SchemeName` is `presto` and `URL.Driver` is
  `trino`, with the dialect `presto`. usql needs a registry entry keyed
  `presto` that opens the `trino` driver. It replaces `trino-go-client` and
  `presto-go-client`. The DSN
  `trino://user:password@host:port/catalog/schema` is a guess until dbimp
  tags the driver.
- The old ports 8080 and 8443, the forced user `user` and the default catalog
  are gone from both generators. Check what usql's Trino and Presto drivers
  assumed from them.

Checked on 2026-10-07: dbimp v0.12.0 has `github.com/xo/dbimp/trino`, for
Trino and Presto, with the key `flavor` (empty asks the server through
GET /v1/info). With dburl v0.43.0, a DSN with a user parses: `trino://u@host`,
`trino://u:p@host:8443/cat/sch` and `presto://u@host/cat`. A URL with no
user does not: `trino://host` and `presto://host` give `trino://host`, and
the driver refuses it with "no user, and both servers need one". dburl
v0.42.0 forced the user `user`.

Ken decided on 2026-10-07 that dburl writes the default user `user` for a
trino or presto URL that has none, and that the driver keeps its error.
dburl staged it in its D50, and also made the scheme set `flavor=trino` or
`flavor=presto` unless the URL names it, so `presto://u@host/cat` becomes
`trino://u@host/cat?flavor=presto` and the driver does not ask the server.
It is not released, because `xo/odbc` has no tag and D49 is in the same
release. So usql waits for that release, which comes after v0.43.0. Then
it takes that release, moves the drivers, and tests, in one commit. `Version`
reads `SELECT version()` on Trino, and /v1/info or nothing on Presto, which
has no `version()`.

ODBC, on 2026-10-07: `drivers/odbc` imports `github.com/xo/odbc` v0.1.0,
and `alexbrainman/odbc` is out of `go.mod`. It needs no cgo, so by the rule
for groups it moved from `all` to `most`. `IsPasswordErr` reads class 28 of
SQLSTATE, the failed authorization of every database. The metadata question
above is still open.

**Blocker, with its cause.** In a build that links both go-sqlite3 and DuckDB,
which is the default build and every `most` build, the first ODBC query
crashes the process with a SIGSEGV. The odbc project found the cause on
2026-10-07, and it is in how usql links, not in `xo/odbc`:

- DuckDB's bindings link with `-rdynamic`, which exports every global symbol of
  the usql executable. That includes the 268 `sqlite3_*` functions of the
  SQLite that go-sqlite3 carries. The SQLite ODBC driver is a shared library
  that needs `sqlite3_*` functions too, and the dynamic linker gives it the
  copies in the usql executable for most of them, and the system's
  `libsqlite3` for the few that go-sqlite3 leaves out. Two builds of SQLite
  then answer one connection. This holds for any ODBC driver that carries or
  loads SQLite, and only when both go-sqlite3 and DuckDB are linked.
- Checked on 2026-10-07: the default build exports 268 `sqlite3_*` symbols and
  crashes 6 of 6 times. The tag `libsqlite3`, which builds go-sqlite3 against
  the system SQLite, exports 0 and worked 6 of 6 times with DuckDB still
  linked. On that build ODBC gave data with NULLs, errors with their SQLSTATE,
  and a rollback that undid its insert.

The ways out are Ken's decision, because each changes a release:

1. Build with `-tags libsqlite3`. go-sqlite3 then uses the SQLite of the
   system, so the tags of `build.sh` (`sqlite_fts5`, `sqlite_json1` and the
   others) stop applying, the features of SQLite depend on the distribution,
   and the static builds need `libsqlite3.a`. Every installer who builds with
   `go install` needs the tag as well.
2. Leave ODBC out of the builds that link both, which makes it an extra
   again, with a README note.
3. Make `moderncsqlite`, which needs no cgo, the `sqlite3` driver of the
   release builds, and keep go-sqlite3 behind a tag.
4. Ask the DuckDB bindings to stop exporting symbols.

Not decided. Do not release until one is.

Two limits of the drivers, not usql's: the SQLite ODBC driver fails on
`values (1)` with HY010, because it returns no data from the execute, and
`select * from (values (1))` works. And `xo/odbc` v0.1.0 fails on a statement
that returns no result set when it goes through `QueryContext`, such as
`create table`, with "fetching: return code -1". usql runs `create table`
through `ExecContext`, so it does not meet that today. The odbc project has the
fix in its working tree, and Ken decides about a v0.1.1.

The Homebrew formula in `xo/homebrew-xo` and the README paragraph about
`--with-odbc` need a look at the same time, because ODBC is no longer an extra.

## W41. Follow dburl's removal of TLS by scheme: Done

Done on 2026-10-07 on dburl v0.44.0. The MaxCompute example of the README now
uses `tls=false`, and the generated table lost `yds`, `ydbs`, `trs` and
`trinos`.

Source: dburl, on 2026-10-07. Ken decided that no scheme chooses TLS by its
scheme or transport any more, because the dbimp drivers read `tls=true`
(dburl's D51). dburl released it in v0.43.0, together with D49 and D50 of
W40.

usql cannot take v0.43.0 yet. Checked on 2026-10-07, with usql's own
drivers: `trino://u:p@host:8443/cat/sch` generated the DSN
`http://u:p@host:8443?catalog=cat&schema=sch` in v0.42.0, which
`trino-go-client` reads, and generates `trino://u:p@host:8443/cat/sch` in
v0.43.0, which it does not. `presto://` changes its driver from `presto` to
`trino` the same way. So a bump breaks both drivers until dbimp's Trino
driver exists, and usql moves them in the same change (W40). usql stays on
dburl v0.42.0 until then, unless Ken asks dburl for a release that holds D51
alone.

- ydb: the aliases `yds` and `ydbs` go, and return an unknown scheme error.
  `ydb://host?tls=true` writes grpcs on port 2135, and `ydb://` stays grpc on
  port 2136. `yd` stays.
- maxcompute and ots (alias `tablestore`): the `+http` and `+https`
  transports go, so `mc+https://` is an invalid transport. TLS is on by
  default, because they are hosted services, and `tls=false` writes http.
- dburl removes the `tls` key from the query, because the drivers do not
  know it.
- clickhouse keeps `+http` and `+https` until dbimp writes its driver.

What usql has to change when it bumps:

- README.md: the MaxCompute example `mc+http://accessid:accesskey@host/api`.
  Write it with `?tls=false` in its place. Check the line above it and the
  note below it.
- The row of YDB in the generated table loses `yds` and `ydbs` when
  `gen.go` runs again. Do not edit it by hand.
- No test, sample or other document of usql uses a scheme that goes. A
  search on 2026-10-07 found only the two places above.

## W42. Decide whether ClickHouse moves to dbimp's driver: Done

Done on 2026-10-07. Ken decided to move (dburl's D53), and usql took it on
dburl v0.44.0 and dbimp v0.13.0. The ClickHouse driver is dbimp's, over HTTP
only, and `clickhouse-go` is out of `go.mod`. Its `RowsAffected` is the shared
`drivers.RowsAffectedOrZero`, its `Err` and `IsPasswordErr` read the codes 194
and 516 of `*clickhouse.Error`, and `\copy` inserts without a transaction,
because the server has none. Checked on clickhouse-26.9 with dbrun: the
banner, types with NULLs, a count, errors, a wrong password, `\copy` of three
rows, and `\dt` and `\d`.

Source: dbimp, on 2026-10-07. dbimp v0.13.0 has `github.com/xo/dbimp/clickhouse`,
for ClickHouse over HTTP. It could replace `ClickHouse/clickhouse-go/v2` in
`drivers/clickhouse`. dbimp asks usql to wait for dburl and for Ken, because
the choice is his. Nothing is done.

What the move changes:

- The driver speaks HTTP only, on port 8123. `clickhouse-go` also speaks the
  native protocol. A user who connects to the native port loses that.
  dburl's `clickhouse` scheme still names `clickhouse-go`, and still has the
  transports `+http` and `+https`, which dburl keeps until dbimp's driver is
  in (its D51).
- Both register the name `clickhouse`, so a program can link one of the two.
  The driver is in the `base` group, so it is in the default build.
- `RowsAffected` goes. usql's hook returns 0 because `clickhouse-go` has no
  count. dbimp's driver returns `dbimp.ErrNotSupported`, and usql would keep
  the hook that turns that into 0, as it does for Databend.
- `Err` reads the `Exception` of `clickhouse-go`. dbimp's driver gives a
  `clickhouse.Error`, and a wrong password has the code 194 or 516. usql's
  `IsPasswordErr` takes only 516 today.
- `SELECT version()` gives the version, which dbmeta already reads.
- The tests that started ClickHouse in a container are gone with W37. They
  come back on `dbrun`.

Size, from W32: `clickhouse-go` adds 4.8 MB to a release build, with ch-go.
dbimp's driver would drop most of that, and the answer depends on what it
imports.

Open for Ken: whether to move at all, and whether to keep the native
protocol. If it moves, dburl moves its scheme first, and usql follows in one
commit, as with Avatica (W39).

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

Completion is per database. It reads names from dbmeta's model of each one,
so a pass for one says nothing about the others.
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
| libSQL and Turso    | `github.com/tursodatabase/libsql-client-go/libsql` | v0.0.0-20260528064733 | 2026-05-28 |
| Dolt                | `github.com/dolthub/driver`                        | v1.88.1               | 2026-05-07 |
| MonetDB             | `github.com/MonetDB/MonetDB-Go/v2`                 | v2.0.3                | 2025-08-24 |
| rqlite              | `github.com/rqlite/gorqlite/stdlib`                | v0.0.0-20260504155303 | 2026-05-04 |
| openGauss and MogDB | `gitee.com/opengauss/openGauss-connector-go-pq`    | v1.0.8                | 2025-08-20 |

IBM Db2 is out of scope. Ken decided on 2026-09-29 that it cannot be tested
here, so usql removed its Db2 client setup from `contrib/` and does not add a
Db2 driver. dbmeta records the decision as its D130.

TDengine is dropped as well. dbimp writes no TDengine driver (its D127), and
dburl v0.38.0 removed the provisional scheme (its D41).

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

Resolved on 2026-09-27. usql's `contrib/vertica/podman-config` pointed at
`docker.io/vertica/vertica-ce`, which no longer exists. That file is gone along
with the rest of the container configuration. dbmeta now models Vertica and
starts releases 7.2, 9.1, 10.1 and 25.1 through `dbrun`. dbmeta's D88 records
how.

### vertica-sql-go splits a statement at every semicolon

`github.com/vertica/vertica-sql-go` splits each statement at its semicolons
before it sends it. It does not know that the `BEGIN ... END` body of a SQL
function contains them. `CREATE FUNCTION ... BEGIN RETURN ...; END` therefore
cannot be sent through usql at all, and the semicolon cannot be left out. A
PL/vSQL procedure works, because its body is dollar quoted and survives the
split. dbmeta found this while building its Vertica model, and its D88 has the
details. Report it upstream.

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

## W43. Take the Drill, Solr, Elasticsearch, OpenSearch and DynamoDB drivers: Done, with one gap

Done on 2026-10-07 on dburl v0.46.0 and dbimp v0.14.0. The five drivers are
dbimp's, pure Go over HTTP, and all are in the group `most`. DynamoDB replaces
`btnguyen2k/godynamo`, so the scheme `godynamo` goes, the aliases are `dy`,
`dyn` and `dynamo`, and the URL is
`dynamodb://key:secret@host:port?region=us-east-1`. `go.mod` loses godynamo and
the AWS modules that only it needed. Each driver reports no count of rows
affected through `drivers.RowsAffectedOrZero`, and none has a metadata model
in dbmeta, so the describe commands report that they are not supported (D3).

Checked through usql on drill-1.22.0, solr-10.0.0, elasticsearch-9.5.3,
opensearch-3.9.0, dynamodb-3.3.1 and alternator-2026.3 from dbrun: a query,
NULLs, an error with its name, and a wrong password (HTTP 401 for Solr,
Elasticsearch and OpenSearch, the login page for Drill, and
`UnrecognizedClientException` for Alternator, which is the one flavor that
checks the signature). Drill also ran `CREATE TABLE AS` and `DROP TABLE`.
Solr needs the collection in the URL path.

The gap, closed on 2026-10-08: Drill, Solr and Elasticsearch refuse a
statement that ends in `;`, and dbmeta v0.4.0 holds a model for each, with
`TerminatorStripped`. usql took v0.4.0, dropped their `noModel` entries and
their Version hooks, and checked a piped `select 1 as a;` and `\dt` on all
three. Drill's banner reads `Apache Drill 1.22.0` from dbmeta, and Solr and
Elasticsearch print only their driver name, because dbmeta reads no version
for them yet. OpenSearch takes `;` already and has no model, and DynamoDB has
no model either, so both keep `noModel` entries.

Not checked: Drill 1.21.2, the other Solr, Elasticsearch, OpenSearch and
DynamoDB releases, and a write to DynamoDB, because DynamoDB Local refuses
`CREATE TABLE` as PartiQL.

`TestRegistryIsKeyedByScheme` fails in the default build with "found 16
registered drivers, expected at least 20". It failed before W43, and it passes
with `-tags most`.

## W44. Take the Druid driver: Done

Done on 2026-10-08 on dburl v0.46.0 and dbimp v0.14.0. The driver is dbimp's,
pure Go over HTTP, read only, in the group `most`. dbmeta has a Druid model, so
the version and the describe commands come from it and usql adds no reader
(D3), and a trailing `;` is stripped by the model's Terminator.

Checked through usql on druid-38.0.0 from dbrun: the banner (`Apache Druid
38.0.0`), a query with NULLs, an error with its code, the refusal of an
`INSERT`, a wrong password (HTTP 401), the ordinary user, and `\dtS`. Not
checked: druid-37.0.0, and `\d` of a table with data, because the server held
none.

## W45. Move Cassandra from xo/cql to xo/cassandra: Done

Waits on dburl and xo/cassandra. Ken said on 2026-10-08 that the Cassandra
driver moves from `github.com/xo/cql` to `github.com/xo/cassandra`, and drops
the CQL name, so that it is consistent with the other drivers. Checked the
same day: the checkout of `xo/cassandra` still has the module path
`github.com/xo/cql` and the tag v0.1.0, and dburl v0.46.0 still names the
scheme `cql`, with the aliases `ca`, `cassandra`, `datastax`, `scy` and
`scylla`. Nothing has been released under the new name.

What usql changes when dburl and the driver release it:

- `go.mod` takes the new module, and `drivers/cassandra` imports it, in
  place of `github.com/xo/cql`, and registers under the scheme name that
  dburl then gives, which is `cassandra` in place of `cql`. The `See:` and
  `// DRIVER` lines follow, and `gen.go` rewrites the README tables, the
  `internal` build tags and the link bar.
- `CQL_DEBUG` and its log line are renamed. Check whether the old name stays
  as a spelling that is still read, which Ken decides.
- The scheme `cql` goes from dburl, or stays as an alias. That is dburl's
  and Ken's decision, and usql adds no alias of its own (rule 12). The
  README example `usql cql://host/` and the tag `cql` change with it.
- dbmeta names its dialect `cql`, and `drivers.Language` and the version and
  metadata reads go through it. Ask dbmeta whether the dialect keeps its
  name, because the language is still CQL.
- `docs/RELEASE.md` names `cql` in the password-change list and the table
  of dialects, and says "Cassandra moves to `github.com/xo/cql`". Fix all
  three. The batches of CQL stay, because that is the name of the language.
- Check on a server from dbrun: the banner, a query, `BEGIN BATCH` to
  `APPLY BATCH`, a wrong password, `\dt` and `\copy`.

Update, 2026-10-08: the cassandra project says the module is now
`github.com/xo/cassandra` and the driver registers as `cassandra`, but it has
no tag, and its repository holds the change uncommitted. dburl has no release
that names it either, and `gen.go` reads dburl's registry from `go.mod`, so
usql cannot move with a `replace` alone: the README tables, the build tags and
`internal/internal.go` would still say `cql`. usql waits for both tags. The
peer's note about `LexerName` does not apply, because the chroma lexer is gone
(W36).

What the cassandra project says changes in the driver, 2026-10-08. `ParseDSN`
and `NewConnector` keep their signatures. `ParseDSN` reads only a
`cassandra://` URL, with one host in the host part and more hosts as repeated
`host` keys, so it must get the URL that dburl writes after dburl moves, and
usql's `Open` hook passes the DSN that usql builds from that URL. Value types
change: decimal is `*apd.Decimal`, varint `*big.Int`, date `dbimp.Date`, time
`dbimp.LocalTime`, duration `dbimp.Interval`, inet `netip.Addr`, list, set and
tuple `[]any`, a map with a text key and a UDT `map[string]any`, and a vector
`dbimp.Vector[T]`. `BeginTx` and named arguments fail with an error that wraps
`dbimp.ErrNotSupported`. The option types become functions, such as
`WithConsistency` and `WithPageSize`.

usql has no code for the types or the options, because `drivers/cassandra`
only opens the connection and reads errors. Check these on a server when it
moves: `\copy` into and out of Cassandra, because the shared copy code starts
a transaction and `BeginTx` now fails (it needs `FlexibleCopyWithInsert` with
no transaction, as for rqlite), a `BEGIN`, and the printing of each value type
above, which `tblfmt` formats.


Update, 2026-10-08, later: usql took dburl v0.47.0 and `github.com/xo/cassandra`
v0.1.0, which pulled dbimp to v0.15.0. `drivers/cassandra` imports the new
module, registers `cassandra`, and reads `CASSANDRA_DEBUG` in place of
`CQL_DEBUG` (Ken may want the old name read too). The README and
`docs/RELEASE.md` follow. Open: dburl's dialect for Cassandra is now
`cassandra`, and dbmeta v0.4.0 and its main branch still name it `cql`, so
dbmeta has no model for the dialect and `TestEveryDriverReadsThroughDbmeta`
fails. dbmeta has been asked to rename it. Do not commit this move before that
release. The `cql` rows of the password and statement-form lists in
`docs/RELEASE.md` are dbmeta's dialect names, and change with it. Not yet
checked on a server: the banner, a query, `BEGIN BATCH`, a wrong password,
`\dt`, `\copy` (the `BeginTx` problem above) and the new value types.

## W46. Bring the describe output back to psql and to v0.21.6, PostgreSQL first

Ken asked on 2026-10-08 for the describe commands to give at least the
information of the last release (v0.21.6), with the columns and names of psql,
and the same column names where it is the same column. Extra columns and extra
commands are fine. psql's style and names win over the old usql's. The
databases follow PostgreSQL one at a time, with v0.21.6 as the baseline where
there is no psql.

The harness is `testdata/describe`. `fixture.sql` builds a database with every
kind of PostgreSQL object, `commands.txt` lists 179 commands in their forms, and
`run.sh` runs them on psql, the old usql and the new one, as csv.
`compare.py` lists the columns that psql has and usql lacks, `values.py` lists
the cells that differ for columns both have, and `sections.py` compares the
sections of `\d+ NAME`. It starts no database, so it is not a test (rule 6), and
W37 will turn its results into goldens.

First run, 2026-10-08, on postgres-18 with dbmeta v0.4.0:

- Three commands fail with a NULL scan error in dbmeta: `\d` of a table with
  an INCLUDE or expression index (`descending`), `\ddp` (`schema`) and `\drds`
  (`database`).
- The relation lists lack Owner, Persistence, Access method and Size (and
  v0.21.6's Rows). `\di` lacks Owner, Persistence and Size, and v0.21.6's
  Primary? and Unique?. `\d+` lacks Collation, Storage, Compression and Stats
  target for columns, and almost every footer section of psql: partitions,
  inheritance, policies, publications, not-null constraints, rules, statistics
  objects, view definitions, options, access method, and the server of a
  foreign table.
- Other lists differ by one or two columns, such as `\dn+`, `\dx`, `\l`, `\dRp`,
  `\dRs+`, `\dFp+`, `\do+`, `\dC+`, `\df+`, and `\dD`.
- Commands that new usql does not know: `\da+`, `\dPt`, `\dPi`, `\dPn`.
- v0.21.6 had far fewer sections than psql. Only Indexes, Check constraints,
  Foreign-key constraints, Referenced by and Triggers match.

The report went to dbmeta on 2026-10-08, which owns the fields. usql's
formatter owns the layout, the names and the order.

Done on 2026-10-08. dbmeta v0.7.0 names the dialect `cassandra`, and usql took it,
so `TestEveryDriverReadsThroughDbmeta` passes again. Checked through usql on
cassandra-5.0 from dbrun: the banner (`Cassandra 5.0.9, CQL 3.4.7, Protocol
v5`), a query with NULLs, and the types decimal, varint, date, time, inet,
list, map and duration, `BEGIN BATCH` with a `$$x;y$$` string, a wrong
password, and `\dt` and `\d`. `\copy` says that the driver does not support it,
as it never did, because the driver has no Copy hook. Open: dbmeta's Cassandra
model ignores the schema and name filters, and the formatter does not narrow
the result, so `\dt ks.*` also lists the system keyspaces. dbmeta has been
asked for one shared helper. Not checked: scylla.


Progress, 2026-10-08, on dbmeta v0.8.0 (D197 and D198). The three scan
failures are gone. The formatter now prints the Owner, Persistence, Access
method, Size and Rows of relations, the Owner, Persistence and Size of
indexes, the Collation, Storage, Compression and Stats target of columns, the
identity and generated defaults, the partial-index WHERE and the index flags
(INVALID, CLUSTER, DEFERRABLE, INITIALLY DEFERRED, REPLICA IDENTITY), the
Leakproof? and Source code of functions, psql's titles for each class, the
`Unlogged` prefix of a title, `\da+` and `\dP[itn]`. A column that no row has a
value for is left out, so databases that give none keep their output. Left, and
asked of dbmeta: the sections of `\d+ NAME` (D199), the fields and text forms
listed in the report of the same day (`\dP` sizes, FDW options, `\dT+`
elements, privileges and policies, `\dn+`, `\dx`, `\l`, `\dRp`, `\dRs+`,
`\dFp+`, `\dAo+`, sequence cache), and the values of `\dD`, `\dFd+` and the
`\dA` family. After that, the other databases follow, with v0.21.6 as the
baseline.

Progress, 2026-10-09, on dbmeta v0.9.0 (D199, D200 and D201). The formatter now
also prints the sections of `\d+ NAME`: the partition key, the partitions and
their bounds, `Partition of`, the inheritance of a table, policies, rules,
statistics objects, publications, not-null constraints, view definitions,
options, access method, and the server and options of a foreign table. The
lines of an index come from the definition of the database, with `UNIQUE
CONSTRAINT` and `EXCLUDE`. Names of related tables are written without a schema
where the database says they resolve without one. `\dn+`, `\dx`, `\l`, `\dRp`,
`\dRs+`, `\do+`, `\dAo+`, `\dP`, the FDW options, `\dT+` elements, and the
column privileges and policies of `\dp` follow psql. The harness shows these
differences left: `\d` of an index lacks the Storage and Stats target of its
columns and the type of an expression, `\dFp+` prints no per-parser table,
`\dC+` gives no Leakproof? (the cast has none), the column options of a
foreign table are empty, and `\dRs+` cannot be compared on a database with no
subscriptions.

