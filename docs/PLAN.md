# Plan

`usql` is a universal command line interface for SQL databases. It links many
`database/sql` drivers into one binary and gives each the same psql style
commands.

This file holds the direction of the project and the questions that are open
for the maintainers. The work items are in [BACKLOG.md](BACKLOG.md), each with
a `W` number. The decisions are in [decisions/](decisions/README.md), each with
a `D` number.

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
- usql has one build, with every driver. `no_<tag>` leaves one driver out, and
  `no_drivers` leaves out every driver that the build does not name by its tag.
  A driver with no hooks has no package, and `gen.go` writes its import and its
  registration into `internal/` (W52).
- What describes a database (the lexical forms, the placeholder of `\copy`, the
  statement that every one of its statements is a query, the product string) is
  in dbmeta, and the defaults of a DSN are in dburl. A driver holds only Go code.
- The describe commands follow psql (W46), PostgreSQL first and the other
  databases one at a time.
- The containers that usql is tested against come from `dbmeta`'s `dbrun`.
  CONTRIBUTING.md shows how to use it.

## Open questions

1. W80, the configuration model: does an alias take arguments, and how?
   For example, the words after `\NAME` could be added to the end of its
   text, or be read as `:1` and `:2`. The maintainer settled the other
   questions of W80 on 2026-11-09, and W80 records the answers. W80 is built.
   Until this question is answered, an alias that is given an argument fails
   with "an alias takes no arguments".
2. W82, a T-SQL procedure with no `BEGIN ... END`: where does its body end?
   usql cuts it at the first `;`, and the rest of its statements run as
   queries when a dump loads. Either transit's grammar reads the body of a
   `CREATE PROCEDURE`, `FUNCTION` or `TRIGGER` of SQL Server to the end of the
   script or to a marker, or the dump writes each `;` in such a body as `\;`.
   Found on 2026-11-10.
