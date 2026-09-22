# Contributing to usql

Any contributions are welcome. If you found a bug, or a missing feature,
take a look at existing [issues](https://github.com/xo/usql/issues)
and create a new one if needed.

You can also open up a [pull request](https://github.com/xo/usql/pulls) (PR)
with code or documentation changes.

# Adding a new driver

1. Add a new schema in [dburl](https://github.com/xo/dburl).
2. Create a new go package in `drivers`. It should have an `init()` function,
   that would call `drivers.Register()`.
3. Regenerate the `internal` package, the driver table in `README.md` and the
   license files by running `go run gen.go`.
4. Add any new required modules using `go get` or by editing `go.mod` manually
   and running `go mod tidy`.
5. Run all tests, build `usql` and see if the new driver works.
6. Update `README.md`.

A driver that cannot compile everywhere says so in its package comment with a
`Build:` line, and `gen.go` adds that constraint to the file it generates in
`internal`. The duckdb driver uses `Build: !(linux && arm) && !windows`, because
its prebuilt library has no 32 bit arm build and no windows build. The oracle
driver uses `Build: !(linux && arm)`, because go-ora writes a constant that
overflows a 32 bit int.

## Building duckdb on Windows

The duckdb driver links prebuilt static archives rather than building DuckDB
from source, and the archives are built by MinGW-Builds GCC 14.2.0 for x86_64,
UCRT, posix-seh. Linking needs a toolchain that matches all three, and two out
of three is not enough:

- An MSVCRT gcc, which is the MSYS2 MINGW64 environment, fails on
  `__stdio_common_vsnprintf_s` and other `__stdio_common_*` symbols. Those are
  UCRT functions, so the runtime is wrong.
- A newer UCRT gcc, which is the MSYS2 UCRT64 environment today, gets past the
  runtime and fails on `__emutls_v._ZSt11__once_call`. The archives were built
  by a GCC whose libstdc++ used emulated TLS for that, and a newer one does
  not. The version is wrong.

Install the matching toolchain and put it first on PATH:

    winget install --id BrechtSanders.WinLibs.POSIX.UCRT.LLVM --exact --silent

It lands under `%LOCALAPPDATA%\Microsoft\WinGet\Packages\` in a
`mingw64\bin` directory. `go build` finds `gcc` there and turns CGO on by
itself. The workflows install the same package for the same reason.

Checking that `go build ./drivers/duckdb/` succeeds is not enough. Building a
package compiles it; the archives are only linked when an executable is
produced, so every failure above appears at the binary link and not before.
Build the binary.

## Adding a platform constraint to a driver

Put the constraint in the driver's package comment rather than in `build.sh`. A constraint in the
generated file also excludes the driver from `go build -tags most`, while one in
the script only applies to a build made through the script.

> Tip: check out closed PRs for examples, and/or search the codebase
> for names of databases you're familiar with.

## Putting a feature behind a build tag

A package that is not a driver can still be a large part of the size of the
binary. The charts renderer is one: it runs Apache ECharts in a JavaScript
engine and rasterizes the result with an SVG renderer, which together add about
12 MiB to every binary they are linked into.

The only way to leave such a package out is a build tag that reaches the import
itself, so these are gated the same way the drivers are. Add an entry to the
`features` list in `gen.go`:

```go
{
    Tag:   "charts",
    Pkg:   "github.com/xo/usql/metacmd/charts/echarts",
    Desc:  "ECharts chart renderer",
    Group: "all",
}
```

`gen.go` then writes `internal/charts.go`, which imports that package under the
constraint the group gives. The group has the same meaning as a driver's, and
`Build:` works the same way, so a feature that cannot compile everywhere states
so there.

A feature is not a driver. It registers no URL scheme, and it does not appear in
`internal.KnownBuildTags` or in the driver table in the README.

The rest of usql must not import the package directly. Give it a registry in a
package that costs nothing, have the gated package register itself from its
`init`, and have the caller ask the registry whether the feature is there. See
`metacmd/charts/renderer.go` for the registry and
`metacmd/charts/echarts/echarts.go` for the implementation. Without that, the
import in the caller links the package back in and the tag does nothing.

Run `go generate` after editing `gen.go`. It needs `GOPATH` set, and
`github.com/xo/dburl` checked out beside usql.

# Running the tests

Some tests need nothing. The rest start a database in a container.

## Tests that need no container

These run anywhere usql builds:

    go test ./stmt/... ./env/... ./handler/... ./drivers/completer/... ./drivers/metadata ./drivers/dameng/...

`drivers/sqlite3/sqshared` also belongs to this group. It builds its sakila
database with the linked SQLite driver and caches the schema under
`drivers/sqlite3/sqshared/testdata`, so it needs the network only on the first
run.

## Tests that need a container runtime

`drivers`, `drivers/clickhouse`, `drivers/metadata/informationschema`,
`drivers/metadata/postgres` and `drivers/sqlserver` each start their own
containers through `ory/dockertest`.

With podman, point the tests at the podman socket first. dockertest speaks the
Docker API, and podman serves it:

    systemctl --user start podman.socket
    export DOCKER_HOST=unix:///run/user/1000/podman/podman.sock

Run the whole suite one package at a time:

    go test -p 1 -tags "$(./build.sh -T)" ./...

Use `-p 1`. Go runs packages in parallel by default, three of these packages
each start a SQL Server container, and three at once exhaust the memory on a
normal workstation. Every one of them then times out.

Keep about 8 GB free for the containers. SQL Server alone takes around 1.5 GB.

## When a test fails

A failing test leaves its containers running. The container cleanup happens
after the tests return, and a fatal error in `TestMain` exits before it. Remove
them before the next run, or the next run fails for lack of memory:

    podman ps -a --format '{{.ID}}\t{{.Image}}' | grep usql- | cut -f1 | xargs -r podman rm -f

## Terminal tests

usql turns off its prompt, line editor, completion and colour when it does not
own a terminal, so piping input into it exercises none of them. `cli_test.go`
records sessions against a real pseudo-terminal and compares them against the
golden transcripts in `testdata/cli`:

    go test -tags "sqlite3 no_base" -run TestCLI .

Every session uses SQLite, so none of them needs a container. Linux and macOS
record. Windows has no pseudo-terminal device file, so the test skips there.

Re-record after a deliberate change to the output:

    go test -tags "sqlite3 no_base" -run TestCLI -update .

Read the diff before you keep a re-recording. A golden here records what usql
does rather than what it ought to do, so accepting one blindly freezes whatever
it does now, including a bug. Before trusting a new golden, break the code it
covers on purpose and make sure that the golden fails. A break that does not
fail it means the golden covers nothing.

## Command line tests

`contrib/` holds a container definition and a connection string for each
database. Start one and query it through usql:

    ./contrib/podman-run.sh postgres
    ./contrib/usql-test.sh

`podman-run.sh` takes the name of any subdirectory of `contrib/` that holds a
`podman-config` file, or `all`, or `test`. `usql-test.sh` builds nothing: it
runs the `usql` binary in the repository root, or the one on your PATH, against
every database that is currently running.

# Enabling metadata introspection for a driver

For `\d*` commands to work, `usql` needs to know how to read the structure of a
database. A driver must provide a metadata reader, by setting the
`NewMetadataReader` property in the `drivers.Driver` structure passed to
`drivers.Register()`. This needs to be a function that given a database and
reader options, returns a reader instance for this particular driver.

If the database has a `information_schema` schema, with standard tables like
`tables` and `columns`, you can use an existing reader from the
`drivers/informationschema` package. Since there are usually minor difference
in objects defined in that schema in different databases, there's a set of
options to configure this reader. Refer to the [package
docs](https://pkg.go.dev/github.com/xo/usql/drivers/metadata/informationschema)
for details.

If you can't use the `informationschema` reader, consider implementing a new
one. It should implement at least one of the following reader interfaces:

- CatalogReader
- SchemaReader
- TableReader
- ColumnReader
- IndexReader
- IndexColumnReader
- FunctionReader
- FunctionColumnReader
- SequenceReader

Every of these interfaces consist of a single function, that takes a `Filter`
structure as an argument, and returns a set of results and an error.

Example drivers using their own readers include:

- `sqlite3`
- `oracle` and `godror` sharing the same reader

If you want to use the `informationschema` reader, but need to override one or
more readers, use the `metadata.NewPluginReader(readers ...Reader)` function.
It returns an object calling reader functions from the last reader passed in
the arguments, that implements it.

Example drivers extending an `informationschema` reader using a plugin reader:

- `postgres`

`\d*` commands are actually implemented by a metadata writer. There's currently
only one, but it too can be replaced and/or extended.

# Enabling autocomplete for a driver

If a driver provides a metadata reader, the default completer will use it. A
driver can provide it's own completer, by setting the `NewCompleter` property
in the `drivers.Driver` structure passed to `drivers.Register()`.
