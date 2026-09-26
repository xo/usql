# Adding a database driver to usql

This document is written for a coding agent. It is the whole procedure, in
order, including the parts that are not code.

Read all of it before you start. Several steps happen in a different repository,
and one of them must happen first.

## The one thing to understand before anything else

In usql, a broken driver and a working driver look the same at build time.

Almost every field of `drivers.Driver` is optional. `drivers.Register` checks
only that the name is not already taken. A driver that cannot parse its own
connection string, or that panics on `\l`, compiles cleanly, registers cleanly,
and passes `go test ./...`.

So compilation is not evidence. Registration is not evidence. The only evidence
is a live connection, and the verification step below says what to run.

## Stop conditions

Stop and report to the person who asked, rather than continuing, if any of
these is true:

1. The scheme does not exist in the version of `dburl` pinned in `go.mod`.
2. The upstream driver fails any pre-flight check in the next section.
3. The driver needs cgo, a vendored native library, or a client install.
4. You cannot run the database locally, in a container or in an embedded mode.

None of these means the driver must be refused. They mean the decision is not
yours. Report what you found and wait.

## Step 1. The pre-flight gate

Answer four questions before you create any package. Record the answers in the
driver's package comment, so that a later reader can see what was checked and
when.

This gate exists because of what the history shows. Several drivers were added
and later removed, and the cause was never the SQL. It was an upstream that
stopped being maintained, or a driver that reached into global process state.

### Is it maintained?

Record the date of the last tagged release and the date of the last commit.
Record whether open issues are being answered. Write dates, not impressions.

A driver whose upstream stops being maintained is dropped from usql. See
"When a driver is broken" in [CONTRIBUTING.md](../CONTRIBUTING.md).

### Does it touch global process state?

This is the question that pays. Read the upstream driver's `init()` and
everything it calls. A driver may register itself with `database/sql` and do
nothing else.

Search the upstream source for these:

    slog.SetDefault      log.SetOutput       log.SetFlags
    log.SetPrefix        flag.Parse          os.Setenv
    http.DefaultClient   http.DefaultTransport
    signal.Notify        rand.Seed           os.Exit

A hit inside package initialization is a stop condition.

The RamSQL driver is why this check exists. Its `engine/log` package called
`slog.SetDefault` from `init` with a handler writing to `os.Stdout`. A default
`slog` logger also redirects the standard `log` package, so importing that one
driver changed the output of the whole binary.

### What does it link?

Record whether the driver is pure Go. A driver that needs cgo, ships a
prebuilt native library, or requires a client install cannot go in the default
build. It also costs far more binary size than a pure Go driver.

The all-driver binary is already about 215 MB. Two known causes are worth
recognizing, because both survive `-s -w` stripping: `-rdynamic` in cgo
LDFLAGS, and any use of Go's `plugin` package. Both force dynamic symbol
export.

### Can it be tested?

Record whether a maintained container image exists, or whether the database
runs embedded. A driver that nobody can run is a driver that nobody will
notice breaking.

Prefer an image the vendor publishes. An image last built years ago is not a
maintained image.

### Is the driver worth the wrapper it needs?

If making the driver work needs a large validation wrapper, the answer is to
not support it. The Dameng driver needed about 130 lines of DSN validation
against a median of about 20 for other drivers, compensating for a parser that
`dburl` cannot import or test. Dameng was later removed.

Emit the connection string and let the driver reject what it rejects.

## Step 2. The scheme, which lives in dburl

usql does no connection string parsing. All of it is in
[github.com/xo/dburl](https://github.com/xo/dburl). The scheme must exist there
before usql can use it.

Read `docs/SCHEME.md` in dburl. It is ten numbered steps written for this
task. What follows is only the part that affects choices you make on the usql
side.

### The scheme and the directory name may differ

usql's README table has a column headed `Scheme / Tag`. It holds the build
tag, which is the directory name under `drivers/`, and not the scheme.

They are allowed to differ and sometimes do. DynamoDB is `godynamo` in dburl's
registry and `dynamodb` in usql's table. Netezza is `nzgo` against `netezza`.
Both work, because the tag is also registered as an alias.

You are choosing two names. Do not assume they must match.

### Choosing a generator

A `Gen*` function is a helper. Reusing one says nothing about wire
compatibility, and reuse is normal. Prefer the first of these that fits:

1. `GenFromURL("scheme://localhost:PORT/")` when the connection string is a
   URL that needs a default port.
2. `GenScheme("name")` when it is a URL that needs no default port.
3. A hand written `GenXxx` when the connection string is not a URL, such as a
   `key=value` form, or when the generator has to reject something.

What reuse does claim is that every driver behind that generator accepts the
same connection string. Check that claim against the new driver. Trino and
Presto shared a generator for six years, and it was split when the Presto v2
driver started rejecting what the shared function produced.

### File backed databases need a file type as well as a scheme

A database that reads a file rather than connecting to a host needs a
`RegisterFileType` call in dburl's `init`:

    RegisterFileType("duckdb", isDuckdbHeader, `(?i)\.duckdb$`)

The three arguments are the driver name, a header function, and an extension
pattern. Register the scheme with `Opaque: true`.

The header and the extension are consulted in different situations, and this
is the part that costs debugging time:

1. When the file exists, dburl reads the first 64 bytes and asks each header
   function. The extension is never looked at.
2. When the file does not exist, no header can be read, so the extension
   decides.

That is why `usql test.duckdb` once worked on a path that did not exist and
failed on the same path after the file was created. Only the second case
reached the faulty matcher. Anyone debugging that without knowing the rule
looks at extension registration, where nothing is wrong.

Header functions compare bytes at a fixed offset. They never use a regular
expression.

### Aliases are cheap to add and expensive to remove

usql publishes every alias in its README table, so an alias is advertised as
supported from the moment it exists. Removing one is a change to the front
page.

Presto is the warning. Five aliases were published. Two of them were `s`
suffixed variants meaning TLS, which the v2 driver cannot express, because it
selects TLS from options rather than from the scheme. Removing them was a
break.

Do not add an alias for a spelling you have not seen in use.

### Read the parser, not the documentation

Read the connection string parser in the exact driver version that usql pins
in `go.mod`. Not the driver's documentation, and not memory. Presto broke
because nobody did.

## Step 3. The driver package

Create `drivers/<tag>/<tag>.go`. The package needs an `init()` that calls
`drivers.Register`.

A minimal driver is short, because nearly every hook is optional:

    package mydb

    import (
        _ "github.com/example/mydb-go-driver" // DRIVER
        "github.com/xo/usql/drivers"
    )

    func init() {
        drivers.Register("mydb", drivers.Driver{
            AllowMultilineComments: true,
            Process:                drivers.StripTrailingSemicolon,
        })
    }

Keep the `// DRIVER` comment on the driver import. Other tools in the family
find the driver package by searching for it.

### The package comment carries machine readable lines

`gen.go` reads the package comment. Two lines in it are not decoration.

A `Build:` line gives a constraint for drivers that cannot compile everywhere,
and `gen.go` copies it into the generated file under `internal`:

    // Build: !(linux && arm) && !windows

The duckdb driver uses that constraint, because its prebuilt library has no 32
bit arm build and no Windows build. The oracle driver uses `!(linux && arm)`,
because go-ora writes a constant that overflows a 32 bit integer.

A `Group:` line puts the driver in a build group. The groups are `base`, which
is the default build, `most`, `all`, and `bad` for drivers that are known to be
broken.

Always write the `Group:` line. A driver that omits it lands in `most`, which
ships in regular builds, and that is rarely what a new driver should do. Write
`Group: all` unless the maintainer decides otherwise.

Promoting a driver to `most` or `base` is a decision for the maintainer, not
part of adding it.

### Promoting to base needs one more edit

If a driver is ever promoted to the base group, add it to `baseOrder` in
`gen.go`. That map fixes the order of the base rows in the README.

A driver missing from the map sorts as 0 and ties with postgres. `sort.Slice`
is not stable, so the two rows swap on every run and `go generate` churns the
README.

## Step 4. Generate, do not edit

Run the generator:

    GOPATH=$(go env GOPATH) go run gen.go

It needs `GOPATH` to point at a tree that contains a `dburl` checkout at
`src/github.com/xo/dburl`, because it reads dburl's source rather than the
module cache. It panics with a confusing `lstat` error when `GOPATH` is empty.

These files are generated. Do not edit them by hand:

1. `internal/<tag>.go`, which carries the build constraints.
2. The driver tables in `README.md`.
3. The bundled license files.

The alias column and the file marker in the README come from dburl's registry.
The description, the package path and the `Scheme / Tag` column come from
usql's own data in `gen.go`. So regenerating rewrites the alias column and
leaves the rest of the row alone.

Then add the module and tidy:

    go get github.com/example/mydb-go-driver@latest
    go mod tidy

## Step 5. Metadata introspection

This step has an instruction that used to be "implement the readers you can".
That instruction produced the same gap twenty-one times, so it has been
replaced.

Of 51 registered names, 21 have a metadata reader. Every one of the 21 that
fails a command fails on `\l`, on `\dp`, or on both. Nothing fails `\d`, `\dt`
or `\dn`.

`\l` needs `CatalogReader`. `\dp` needs `PrivilegeSummaryReader`.

### What to do instead

Decide each of these two interfaces explicitly, and record the decision:

1. Does the product have a concept of catalogs or databases that a statement
   can read? If yes, implement `CatalogReader`. If no, write one line saying
   which upstream facility is absent and how you checked.
2. Does the product have grants or roles that a statement can read? If yes,
   implement `PrivilegeSummaryReader`. If no, write the same one line.

Omission is allowed. Silent omission is not. "Nobody tried" and "the product
cannot" look identical in the code, and today most of the thirty gaps cannot
be told apart.

### The rest of the readers

`drivers/metadata/metadata.go` declares fourteen leaf reader interfaces,
aggregated by `ExtendedReader`. Seven decide whether a command runs at all:

    TableReader   ColumnReader   FunctionReader   IndexReader
    SchemaReader  CatalogReader  PrivilegeSummaryReader

The other seven decide how much detail is printed. `SequenceReader`,
`IndexColumnReader`, `TriggerReader`, `ConstraintReader` and
`ConstraintColumnReader` add sections to `\d+`. `FunctionColumnReader` adds
detail to `\df`. `ColumnStatReader` serves `\ss`.

### Use the shared reader where you can

`drivers/metadata/informationschema` is a configurable reader for any product
with a standard `information_schema`. Configure it with options rather than
writing a reader from nothing:

    infos.New(
        infos.WithPlaceholder(func(int) string { return "?" }),
        infos.WithSequences(false),
        infos.WithSystemSchemas([]string{"mysql", "information_schema"}),
    )

Do not return a value that satisfies nothing when the handle is unexpected.
The impala reader returns `struct{}{}` when the handle is not a `*sql.DB`,
which means it loses all metadata support without reporting anything. It is
the only driver that does this, and it is not a pattern to copy.

## Step 6. Verification

Compilation proves nothing here. Run all of this, and report the output rather
than a summary of it.

### Without a database

    go run gen.go                       # then inspect git diff
    go build ./...                      # default build
    go build -tags all ./...            # every driver
    go vet ./...
    go test ./...

The `-tags all` build matters on its own, because a duplicate scheme or alias
panics at registration time rather than at compile time.

Most test failures without a container runtime are expected. Five packages
need containers. Everything else must pass.

### With a database

Start the database, then run each of these and keep the output:

    usql <dsn> -c 'select 1'
    usql <dsn> -c '\l'
    usql <dsn> -c '\dt'
    usql <dsn> -c '\d <table>'
    usql <dsn> -c '\dp'

Try every connection string form that dburl accepts for the scheme, including
the short alias. Try one malformed connection string, and make sure that it
produces an error rather than a panic.

For a file backed driver, test both cases separately, because they take
different paths through dburl:

    usql ./new.db        # the file does not exist: the extension decides
    usql ./existing.db   # the file exists: the header decides

### What counts as done

Do not report the work as complete on compilation, on registration, or on
`go test` alone. In this repository all three are true of a driver that does
not work.

Report the pre-flight answers, the generator diff, and the transcript of a
live connection.

## What goes wrong most often

Adding the driver to usql without the scheme existing in dburl. The code
compiles, connecting fails at runtime, and the next move is usually to write
connection string parsing inside the usql driver package. That parsing does not
belong there and will be rejected. If the scheme is missing, stop at Step 2.

The second most common is hand editing a generated file. The change survives
until somebody runs `go generate`, and then it disappears.

## The checks that tell you what you forgot

| Check | What fails it |
| --- | --- |
| `go build -tags all ./...` | A duplicate scheme or alias, which panics at registration |
| `git diff` after `go run gen.go` | A generated file that was edited by hand |
| `\l` against a live database | `CatalogReader` missing or wrong |
| `\dp` against a live database | `PrivilegeSummaryReader` missing or wrong |
| A malformed connection string | A panic where an error was wanted |
| A file path that does not exist | A missing extension pattern in dburl |
| A file path that does exist | A faulty header function in dburl |
| The upstream `init()` | A driver that writes global process state |

Add a row whenever you find something this document did not catch. The table
is read when something is red, which is more often than the rest of the
document is read.

## Related documents

| Task | Document |
| --- | --- |
| Adding a scheme | `docs/SCHEME.md` in dburl |
| A driver that is broken or unmaintained | [CONTRIBUTING.md](../CONTRIBUTING.md) |
| Running the tests | [CONTRIBUTING.md](../CONTRIBUTING.md) |
| Planned work | [BACKLOG.md](BACKLOG.md) |
| Rules for agents working here | [CLAUDE.md](../CLAUDE.md) |
