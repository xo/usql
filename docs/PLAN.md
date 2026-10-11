# Plan

`usql` is a universal command line interface for SQL databases. It links many
`database/sql` drivers into one binary and gives each the same psql style
commands.

This file holds the direction of the project and the questions that are open
for Ken. The work items are in [BACKLOG.md](BACKLOG.md), each with a `W`
number. The decisions are in [decisions/](decisions/README.md), each with a
`D` number.

## Direction

- Connection string parsing lives in `dburl`, and the README driver tables are
  generated from its registry by `gen.go`.
- Database metadata moves out of `drivers/metadata` into `dbmeta`, and no
  metadata reader is written here (D3). W7 and W21 track it, and
  [DBMETA.md](DBMETA.md) holds the steps.
- Drivers for databases whose upstream Go driver is missing, dead or heavy come
  from `dbimp`, which is now most of them: it is pure Go, and a driver of it is a
  bare registration, because its errors, row counts and refused credentials are
  generic here. W27 records the review that retired, replaced and regrouped the
  drivers.
- What describes a database (the lexical forms, the placeholder of `\copy`, the
  statement that every one of its statements is a query, the product string) is
  in dbmeta, and the defaults of a DSN are in dburl. A driver holds only Go code.
- The describe commands follow psql (W46), PostgreSQL first and the other
  databases one at a time.
- The containers that usql is tested against come from `dbmeta`'s `dbrun`.
  CONTRIBUTING.md shows how to use it.

## Open questions

For Ken:

- How the release handles the ODBC crash that go-sqlite3 and DuckDB cause
  together: build with `-tags libsqlite3`, leave ODBC out, or make `moderncsqlite`
  the SQLite of a release (W40).
- Whether Cosmos DB keeps no writes, since dbimp's driver reads only.
- Whether DuckDB gets a build group of its own, since it is 37% of a `most`
  binary (W32).
- The history file of gohxs/readline, which rline cannot read (see
  [PROGRESS.md](PROGRESS.md)).
- Whether the drivers that are only a registration lose their packages, with
  the build tags written into `internal/` by `gen.go` (W52).
