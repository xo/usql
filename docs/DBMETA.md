# Migrating metadata to dbmeta

This is the plan for moving usql off its own metadata `Reader` and `Writer` and
onto [dbmeta](https://github.com/xo/dbmeta). No code has been written.

It is a plan, not a decision. The ordering in it differs from the order the
work was first described, and the reason is in "Why the order changed".

## What is wrong today, measured

`drivers/metadata` is 6083 lines including tests.

`drivers/metadata/metadata.go:120` declares `type Reader interface{}`. The base
reader is the empty interface. Every capability is found by a runtime type
assertion, and there are 52 of them outside tests.

That one line causes the rest. A driver that satisfies nothing compiles,
registers and passes the tests. Of 51 registered names, 21 have a reader. Every
one of the 21 that fails a command fails on `\l` or `\dp`. One driver returns
`struct{}{}` as its reader, which the type system accepts.

`Writer` is eight methods, and each one queries, formats and writes:

    DescribeFunctions(*dburl.URL, string, string, bool, bool) error
    ListTables(*dburl.URL, string, string, bool, bool) error
    ListAllDbs(*dburl.URL, string, bool) error

Positional strings, bare bools, a `*dburl.URL` on every call, and only an error
back. `DefaultWriter` holds the reader, the database handle and an `io.Writer`.
It carries a field commented "custom functions for easier overloading" that
replaces `ListAllDbs` at construction. `NewDefaultWriter` returns a constructor
rather than a `Writer`.

Neither interface takes a `context.Context`. Timeouts are set by an option
stashed on a `LoggingReader`.

## Two facts that constrain the design

These were checked rather than assumed, and each removes a decision.

No project outside usql imports `usql/drivers/metadata`. dbtpl, dbmeta and
tblfmt were all searched. So there is no external compatibility burden, and
every caller of both interfaces is in this repository where the compiler can
find it.

`resultSet` in `metadata.go` already implements `tblfmt.ResultSet` exactly:
`Next`, `Scan`, `Columns`, `Close`, `Err` and `NextResultSet`. `writer.go`
calls into tblfmt eleven times.

The second fact decides what a formatter produces. tblfmt already owns aligned,
unaligned, CSV and the other output modes. A formatter that renders bytes would
duplicate it. A formatter that produces a `tblfmt.ResultSet` does not.

## The boundary

dbmeta reads. tblfmt renders. usql decides what to print.

That division is dbmeta's D5 and it does not change. The writer does not move.
Three options were put to the dbmeta project and this is the one its
`docs/USQL.md` describes.

Two readings were wrong and are recorded so nobody re-derives them:

dbmeta does not own models without queries. A model there is the queries. There
are 12 of them, and a query set for each of 55 object kinds, registered per
dialect. Taking the types without the queries takes the smallest part.

dbmeta does not take the formatting. It answers what the database says.

## Why the order changed

The work was first described as: remove `Reader` and `Writer`, move the readers
to dbmeta, turn `Writer` into `Formatter`, then wire up the commands.

Two things move earlier in this plan.

The formatter work comes before the reader swap, not after. It is entirely
inside this repository, it has no external consumers, and it is where the
psql-matching behavior lives. Doing it first means the reader swap changes one
thing at a time.

A step that produces no visible result comes first of all. It is described in
the next section, and it is the step most likely to be skipped.

## Step 0. Extract the filtering, change nothing else

usql's output matches psql. The matching is done partly by the readers and
partly inside `DefaultWriter`, which queries, filters and renders in one
object.

dbmeta returns what the database says and does not shape a result so that
somebody's output looks right. So if a reader is replaced by a dbmeta query
before that filtering is pulled out, usql's output changes, and it changes
silently, because it appears as extra rows rather than as an error.

Pull the psql-matching filtering out of `DefaultWriter` and into the formatting
layer, with the current readers still in place. Confirm that the output is
unchanged, byte for byte.

This step has no visible result. That is why it is written down first.

## Step 1. Redesign the writer as a formatter

Three layers, not two. Both external reviews reached this independently, and it
is the answer to a question the first design sketch got wrong.

    reader (dbmeta)  ->  command layer (usql)  ->  formatter (usql)  ->  tblfmt

A formatter that takes one result set cannot express `\d`, which composes five
sections: columns, indexes, constraints, triggers and sequences. That
composition is not reading and it is not rendering. It belongs to a middle
layer that the current design does not have, because `DefaultWriter` is doing
all three jobs at once.

So the formatter formats one section. The command layer runs the queries,
writes the section headings, and calls the formatter once per section.

### What the formatter produces

A `tblfmt.ResultSet`, not bytes. tblfmt then renders it in whatever output mode
the session is in.

This makes the formatter a pure function from metadata to a result set, which
makes it testable without a database. That matters here: five test packages
need a container runtime today, and none of the formatting needs one.

### What goes in the signature

An options struct, not functional options and not positional bools. There are
eleven fixed call sites in one switch. Functional options are for open-ended
constructor knobs and would add a closure allocation per `\d`.

Drop `*dburl.URL` from every method. It is resolved once in the command layer.

Take a `context.Context` on the readers, not only on the formatter. The queries
are the slow part. This also deletes the timeout option stashed on
`LoggingReader`, which exists only because there was no context to carry a
deadline.

Drop `Params map[string]string`. It is untyped and forces a formatter to parse
something the command layer already parsed.

Delete the `listAllDbs` override field. It is a hook that exists because the
interface could not express the case.

### No compatibility shim for this part

Both interfaces are internal and no other project imports them. The compiler
finds every caller. A shim here would cost more than the change and would
outlive its purpose.

## Step 2. Capability discovery

This is where the two external reviews disagreed, and the disagreement is worth
recording because the losing answer is the one that looks more obvious.

One proposal was a unified interface with a `Capabilities()` bitmask, and a
`BaseReader` that every driver embeds and that returns a not-supported error
for anything it does not implement.

Do not do this. Embedding a base that answers everything makes every driver
satisfy every interface, so capability can no longer be detected by assertion
at all. A caller has to call and inspect the error. That reintroduces the
current failure in a new form, where nothing can be counted without running it.
A hand-maintained bitmask has the second problem: it can disagree with the code
and nothing notices.

The better answer keeps small optional interfaces, which is how the standard
library does this, and fixes the real defect, which is not the assertions but
that there are 52 of them scattered against `interface{}`. Resolve them once:

    type Capabilities struct {
        Catalog   CatalogReader
        Table     TableReader
        Column    ColumnReader
        Index     IndexReader
        Privilege PrivilegeSummaryReader
        // and the rest
    }

    func CapabilitiesOf(r Reader) Capabilities

Every assertion happens in one function. Callers test a typed field against nil
and then call a statically typed method.

Make the base interface non-empty. A single required method is enough, and it
makes the driver that returns `struct{}{}` stop compiling.

### The sets

Keep the fourteen typed sets at the dbmeta boundary. They carry typed fields
and real tests.

Do not replace them with a generic `Set[T]`. A formatter taking `Set[T]` still
receives `any` at the call site, which reintroduces the type switch that the
generic was meant to remove.

Convert to the single render shape at the formatter boundary instead. That is
one small method per set type, and a formatter then never changes when dbmeta
grows a fifteenth kind.

## Step 3. The information schema adapter, which ships first

This is the earliest visible gain and it is not a native model.

dbmeta has `models/informationschema`, a shared model answering 12 of the 55
object kinds against any database with a standard `information_schema`: tables,
schemas, columns, functions, privileges, constraints, sequences, constraint
columns, routine parameters, views and the current schema.

30 of usql's 51 registered names have no metadata reader at all. Any of them
with a standards-shaped catalog gets those twelve with no per-driver query
written.

Ship this before any native model moves. It is one adapter, it lands coverage
across a set of names rather than one, and it is the cheapest demonstration
that the boundary works.

## Step 4. Move drivers one at a time

The move is per driver and reversible. usql is never blocked on dbmeta for a
product dbmeta does not cover.

### What dbmeta answers today

Measured by dbmeta on 2026-09-26, by asking every registered query whether it
supports the newest release of its dialect. The units are usql's: 11 dispatch
commands, and 5 further sections that `\d NAME` adds.

    PostgreSQL   11/11, 5/5
    MariaDB      11/11, 5/5
    SQL Server   11/11, 5/5
    SAP HANA     11/11, 5/5
    MySQL        11/11, 4/5   no sequence
    ClickHouse   11/11, 2/5   no sequence, trigger or constraint column
    Oracle       10/11, 5/5   no \l
    Firebird     10/11, 5/5   no \dn
    DuckDB       10/11, 3/5   no \dp, no index column or trigger
    SQLite       10/11, 4/5   no \dp, no sequence
    Cassandra    10/11, 4/5   no \l, no sequence
    Trino         8/11, 0/5   no \df, \da, \di, and no sections
    Presto        8/11, 0/5   the same

### Four products where usql answers nothing today

Four of those thirteen map to usql drivers that register no metadata reader at
all. Each was checked against the driver source:

    SAP HANA    drivers/saphana     0/11 today, 11/11 and 5/5 with dbmeta
    Firebird    drivers/firebird    0/11 today, 10/11 and 5/5 with dbmeta
    Cassandra   drivers/cassandra   0/11 today, 10/11 and 4/5 with dbmeta
    Presto      drivers/presto      0/11 today,  8/11 with dbmeta

These are the largest single gains available, and they are gains rather than
refactors. SAP HANA is the largest of all: a product that answers nothing today
and would answer every command and every section.

None of them can come early. SAP HANA has no `information_schema`, so it needs
its native model rather than the shared adapter in Step 3. They belong in this
phase.

### Which native model moves first

MariaDB, but the reason is the sections rather than the commands.

MariaDB is 10 of 11 today and lacks only `CatalogReader`, so the command gain
is one: `\l`. The gain worth having is 5 of 5 sections under `\d NAME`,
against the four the mysql family answers today.

Not PostgreSQL. PostgreSQL already works and moving it would demonstrate
nothing.

Then the rest, in whatever order suits usql.

### Make the remainder countable

A partial migration becomes permanent when nobody can see how partial it is. A
deadline does not fix this. A count does.

Add a test over the driver registry: every registered name either reads through
dbmeta or has an entry saying why not. This is the same exemption map as
[W18](BACKLOG.md), with the same two properties. A contributor cannot pass it
by doing nothing, and one who genuinely cannot move a driver writes one line
saying why.

The list then shows how many entries say "not yet" and how many say "cannot",
which are different states that look identical today.

## Step 5. Version queries

Version queries move to dbmeta, and the data already exists there. Every model
declares a version query, its columns and a parser, and a dialect returns a
version set plus the display line.

Two qualifications.

usql keeps its own fallback for every name dbmeta does not cover. That fallback
is `SELECT version();`, which Oracle does not have, and the oracle driver
registers no version function. So the fallback stays wrong for Oracle until
Oracle moves. dbmeta found this by comparing its version query against usql's,
which is a comparison nobody had been asked to make.

The version is read once when the connection opens, because everything else
resolves against it. That is a change to usql's connection path, not to a
reader.

## Step 6. Change password, which is a bug fix

This is already in dbmeta and it fixes something rather than relocating it.

`Dialect.ChangePassword` takes no database handle and runs nothing. It builds
the statement and returns it as text for usql to execute. dbmeta still executes
only reads, so D5 holds.

usql changes a password in seven drivers today and escapes nothing. Each one
concatenates the password into the statement, so a password containing a quote
or a backslash either breaks the statement or sets something other than what
was asked.

The escaping cannot be done without the server. Whether a backslash escapes
inside a string literal is `sql_mode` on MySQL and MariaDB, and
`standard_conforming_strings` on PostgreSQL. dbmeta reads both.

## Step 7. Decide what the three-valued answer prints

dbmeta can answer Supported, NotSupported, TooOld or NotBuilt for a query.

usql today cannot tell "this database has no such object" from "there are none
of them". Both print an empty table.

Once it can tell them apart, somebody has to decide what each one prints.
TooOld matters most, because it is actionable: the object exists in the product
and this server is older than the release that added it. Firebird reports it
for three commands on 3.0.

This is a user-visible behavior change. Decide it here rather than leaving it
to whoever writes the first adapter.

## Step 8. Decide what a NULL means

dbmeta records whether a NULL means the value is null or the server is too old
to have the column.

usql has no way to express that and will render both as empty. Deciding is
cheap. Forgetting is permanent, because once both render as empty nobody will
know there was a distinction to make.

## Step 9. Share the container list

dbmeta's `container` package is Go data naming every release it tests against.
It deliberately starts nothing and imports no container client, so that a
downstream project can read the list and bring its own podman or docker.

usql should read that list rather than keep a second one. Testing the same
products separately on both sides means twice the containers for one behavior.

This is the same conclusion the [W6](BACKLOG.md) tombstone reached from the
other direction.

## Smaller things that are cheap now and expensive later

dbmeta's arguments carry a catalog filter and `metadata.Filter` does not. Only
Trino uses it, where a table really is catalog, schema and name, and one server
reaches many catalogs at once. It blocks nothing, but it cannot be added later
without touching every call site.

Decide which of the eleven commands fail and which print nothing when a
capability is missing. A sentinel error is needed either way.

Identifier case and quoting differ by product. `\dt myTable` means different
things on PostgreSQL and on Oracle. Pattern normalization belongs in the
per-product reader, not in the command parser.

## One question that is not technical

dbmeta has not been released. There is no tag, the API is still moving, and its
container package changed twice on 2026-09-26.

Whether usql takes a dependency on an unreleased module, and whose release
cadence wins, is Ken's decision. The per-driver shape of the move means usql is
never blocked on dbmeta for a product dbmeta does not cover, which limits the
exposure but does not remove it.

## Two review disagreements, recorded

Capability discovery is in Step 2 above.

The other was streaming. One review argued that materializing a whole set into
memory will exhaust it on a large table. The other argued that buffering is
required regardless, because column widths cannot be computed without seeing
every row.

For metadata the second is right. These queries list tables, columns and
indexes, and the counts are bounded by the schema rather than by the data. Do
not design a streaming interface for this layer. The concern is real for row
results, which this layer does not carry.
