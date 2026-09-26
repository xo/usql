# Rules for agents working on usql

This file holds rules that an agent can break without noticing. It is not a
design document. The reasoning behind the project lives in the documents listed
at the bottom.

## Which document to read

| Task | Document |
| --- | --- |
| Adding a database driver | [docs/DRIVER.md](docs/DRIVER.md) |
| Adding a connection scheme | `docs/SCHEME.md` in [dburl](https://github.com/xo/dburl) |
| A driver that is broken or unmaintained | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Putting a feature behind a build tag | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Running the tests | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Planned work, and how to cite it | [docs/BACKLOG.md](docs/BACKLOG.md) |
| What usql does, for a user | [README.md](README.md) |

A document that is not in this table does not exist. If you find one, either
add a row for it or delete it.

## Hard rules

1. Do not edit generated files. `internal/*.go`, the driver tables in
   `README.md`, and the bundled license files are all produced by `gen.go`.
   Change `gen.go` or the source it reads, then regenerate.

2. Do not write connection string parsing in this repository. All of it lives
   in `github.com/xo/dburl`. A scheme must exist there before usql can use it.

3. Run `GOPATH=$(go env GOPATH) go run gen.go` to regenerate. It reads dburl's
   source tree rather than the module cache, so `GOPATH` must point at a tree
   containing `src/github.com/xo/dburl`. With `GOPATH` empty it panics with a
   confusing `lstat` error.

4. Build with `-tags all` before claiming that a driver change works. A
   duplicate scheme or alias panics at registration time, which the default
   build does not reach.

5. Compilation is not evidence that a driver works. Nearly every field of
   `drivers.Driver` is optional, so a driver that cannot connect still
   compiles, registers, and passes `go test ./...`.

6. Five test packages need a container runtime: `drivers`,
   `drivers/clickhouse`, `drivers/sqlserver`,
   `drivers/metadata/informationschema` and `drivers/metadata/postgres`.
   Their failure without podman or docker is expected. Any other failure is
   not.

7. Container tests cannot run in parallel. Use `go test -p 1`, which is what
   CI does. Without it the run starts three SQL Server containers at once and
   exhausts memory.

8. Backlog identifiers are append only. Never reuse a `W` number and never
   renumber one. An item that is finished or abandoned keeps its heading and
   gains a status. See the rules at the top of `docs/BACKLOG.md`.

9. A driver may register itself with `database/sql` and do nothing else. A
   driver that writes global process state from `init()` belongs in the `bad`
   group or out of the tree. `slog.SetDefault` is the case that has already
   happened.

10. Always write a `Group:` line in a driver's package comment. A driver that
    omits it defaults to `most`, which ships in regular builds.

11. The repository root holds `README.md`, `CONTRIBUTING.md` and this file.
    Every other document goes under `docs/`.

12. Do not add an alias for a connection scheme spelling you have not seen in
    use. usql publishes every alias in its README table, so an alias is
    advertised from the moment it exists, and removing one is a break.

## Writing

Documentation and comments in this project are written in plain English.
Short sentences, active voice, no semicolons, and no em dashes. State the fact
rather than its importance.

Where a number in prose can be generated from something the code already
knows, generate it. `gen.go` builds the README driver tables from dburl's
registry for this reason: a generated table cannot go stale, because there is
nothing for it to drift from.

Where a number cannot be generated, a test should check it. Put the unit into
the sentence the test searches, so that changing the unit breaks the test.
Every wrong figure found in this project so far was a right count of the wrong
thing.

## Sibling repositories

usql depends on several repositories in the same family. Changes often cross
between them.

| Repository | What it holds |
| --- | --- |
| `xo/dburl` | Connection string parsing, schemes, aliases, file types |
| `xo/dbmeta` | Database metadata, being moved out of `drivers/metadata` |
| `xo/tblfmt` | Result set formatting |
| `xo/rline` | Line editing, intended to replace `gohxs/readline` |
| `xo/dbtpl` | Code generation from database schemas |

`dburl` and `dbmeta` record design decisions as numbered entries, `D1`, `D2`
and so on. That series is separate from this project's `W` work items. Do not
mix them in one commit message.

When a driver is dropped from usql, its scheme is dropped from dburl.
