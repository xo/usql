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

> Tip: check out closed PRs for examples, and/or search the codebase
> for names of databases you're familiar with.

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
