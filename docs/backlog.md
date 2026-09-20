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
