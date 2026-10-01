# D4. usql reads all metadata from dbmeta, at once

Status: Decided.

Ken decided on 2026-09-30 how usql moves onto dbmeta, which W21 holds. He
made three choices.

## All at once, not driver by driver

`docs/DBMETA.md` planned the move one driver at a time, keeping usql's own
readers for a database that dbmeta did not model. Ken chose to remove them all
at once, because D3 and his rule that database knowledge belongs in dbmeta
point there.

So `drivers/metadata` holds no query of its own. It picks dbmeta's model by
`URL.Dialect`, runs dbmeta's queries, and prints the answers. The readers that
went are the information schema, PostgreSQL, Oracle, MySQL and Impala ones in
`drivers/metadata`, and the SQLite, SQL Server, ClickHouse, Trino and DuckDB
ones in their driver packages. `drivers.Driver` loses `NewMetadataReader`,
`NewMetadataWriter`, `NewCompleter`, `User` and `ChangePassword`.

The version, the current user and the statement that changes a password come
from dbmeta too, for every database it models. A driver keeps its own
`Version` only for a database that dbmeta has no model for.

The cost is that a database with no dbmeta model answers none of the describe
commands. `noModel` in `metadata_test.go` lists each one and why, and
`TestEveryDriverReadsThroughDbmeta` fails when a driver is missing from it or
has gained a model.

## A local go.work, not go.mod

dbmeta has no release tag. usql builds against a dbmeta checkout beside it
through `go.work`, as `docs/DBMETA.md` decided, and `go.work` is not
committed. So the change is not committed until dbmeta is tagged. Then
`go.mod` requires that tag and the workspace goes.

## Follow dbmeta and psql

usql prints dbmeta's answers the way psql prints the same command, rather
than keeping its old output byte for byte. The recorded output in
`drivers/testdata` and `testdata/cli` was recorded again from the new code.
