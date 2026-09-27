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
- Database metadata moves out of `drivers/metadata` into `dbmeta`. W7 and W21
  track it, and [DBMETA.md](DBMETA.md) holds the steps.
- New drivers for databases whose upstream Go driver is missing or dead come
  from `dbimp`, as Couchbase and SurrealDB did. W27 records the review that
  retired, replaced and regrouped the drivers.
- The containers that usql is tested against come from `dbmeta`'s `dbrun`.
  CONTRIBUTING.md shows how to use it.

## Open questions

1. Hard rule 3 in `AGENTS.md` says that `gen.go` reads dburl's source tree,
   so `GOPATH` must point at one. `gen.go` now reads the pinned dburl module,
   names no `GOPATH`, and runs with `GOPATH` empty. Should the rule be deleted?
2. The SurrealDB driver has no metadata reader, so `\d` and the other describe
   commands report that they are not supported. SurrealDB lists its objects
   through `INFO FOR`, which returns one object rather than rows. Should usql
   write a reader for it, or wait for a SurrealDB model in `dbmeta`?
