# D5. A driver field that describes the database moves to dbmeta

Status: Decided.

Ken decided on 2026-09-30 which fields of `drivers.Driver` move to dbmeta.
The rule is that a field which says something about the database, and alters
what usql does or how it uses a dialect's queries, belongs in dbmeta, keyed by
the dialect. A field that describes a Go driver, a URL or usql's own interface
stays. dbmeta reviewed every field with Gemini and DeepSeek first.

## What moves to dbmeta

- The grammar of the dialect: `AllowDollar`, `AllowMultilineComments`,
  `AllowCComments`, `AllowHashComments` and `AllowBacktick`.
- `RequirePreviousPassword`, which says the statement that changes a password
  takes the old one, as SQL Server's does.
- `BatchQueryPrefixes`, the statements that begin and end a batch, as CQL's
  `BEGIN BATCH` and `APPLY BATCH`.
- The rule for a trailing semicolon, the part of `Process` that belongs to the
  database, such as Oracle stripping it unless the statement ends with `END;`.
- How the dialect folds an unquoted name, which usql uses for patterns and to
  lower-case column names for display. `LowerColumnNames` is derived from it
  rather than set per driver.
- The handling of ODBC's `usql_trim` key and of csvq's `SHOW` statements.

dbmeta added all but the last as its D143, and usql reads them from there.
The grammar, `RequirePreviousPassword`, `BatchQueryPrefixes` and
`LowerColumnNames` are gone from `drivers.Driver`, and so are the `Process`
hooks of Oracle, Trino and Presto. `LowerColumnNames` is true for a database
that folds to upper case, which adds SAP HANA, Firebird and Snowflake to the
Oracle and Exasol drivers that set it.

A database that dbmeta has no model for allows only block comments. Nearly
every such driver set that and nothing more. A wider default is not safe,
because a hash comment would swallow a SQL Server `#temp` table over ODBC.

The last item waits. Ken decided on 2026-09-30 that the `Process` hooks of
ODBC, csvq, Athena and Cosmos are removed rather than kept in usql. Their
statement handling belongs in dbmeta's syntax options, applied automatically,
and that needs a larger design. W28 holds it, and the possible removal of
ODBC, Athena and Cosmos, which cannot be tested against a server.

With those hooks gone, a `Process` hook did one other thing: run every
statement as a query. The `AlwaysQuery` field replaces the `Process` field
for that. It says how usql runs a statement, and it stays in usql.

## What stays in usql

- `LexerName` becomes a map in usql from the dialect to a Chroma lexer, in
  `drivers/drivers.go`, because the dialect already names the product.
- `Copy`. Its generic `INSERT` takes its placeholder from dbmeta, and a native
  bulk path such as pgx's `CopyFrom` stays in its driver.
- `Name`, `UseColumnTypes`, `ForceParams`, `Open`, `IsPasswordErr`, `Err`,
  `ColumnTypes`, `ConvertBytes`, `ConvertMap`, `ConvertSlice`,
  `ConvertDefault`, `RowsAffected`, and `Version` for a database that dbmeta
  has no model for.

## A rarely used database does not shape usql

`BatchAsTransaction` is removed, and the ql driver with it. Only ql set the
field, and it made usql run a batch in a transaction for that one driver. Ken
decided that a rarely used database must not force logic or architecture into
usql. ql was an unfinished experiment. dburl drops its scheme as its D42.
