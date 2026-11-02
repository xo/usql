package highlight

import (
	"io/fs"

	"github.com/xo/transit"
	"github.com/xo/transit/grammars/clickhouse"
	"github.com/xo/transit/grammars/cql"
	"github.com/xo/transit/grammars/duckdb"
	"github.com/xo/transit/grammars/mysql"
	"github.com/xo/transit/grammars/neo4j"
	"github.com/xo/transit/grammars/oracle"
	"github.com/xo/transit/grammars/postgres/postgres"
	"github.com/xo/transit/grammars/sql"
	"github.com/xo/transit/grammars/sqlite3"
	"github.com/xo/transit/grammars/sqlserver"
	"github.com/xo/transit/grammars/surrealdb"
	"github.com/xo/transit/grammars/trino"
	"github.com/xo/transit/grammars/ydb"
)

// grammar is a SQL grammar of transit that parses the statements of a dialect.
type grammar struct {
	language func() *transit.Language
	queries  fs.FS
}

// The SQL grammars that usql uses. transit does not know the dialects of
// dbmeta, so usql keeps the match of a dialect to a grammar (transit's
// docs/USQL.md).
var (
	genericGrammar    = grammar{sql.Language, sql.Queries}
	postgresGrammar   = grammar{postgres.Language, postgres.Queries}
	mysqlGrammar      = grammar{mysql.Language, mysql.Queries}
	sqlserverGrammar  = grammar{sqlserver.Language, sqlserver.Queries}
	oracleGrammar     = grammar{oracle.Language, oracle.Queries}
	cqlGrammar        = grammar{cql.Language, cql.Queries}
	sqlite3Grammar    = grammar{sqlite3.Language, sqlite3.Queries}
	duckdbGrammar     = grammar{duckdb.Language, duckdb.Queries}
	clickhouseGrammar = grammar{clickhouse.Language, clickhouse.Queries}
	trinoGrammar      = grammar{trino.Language, trino.Queries}
	neo4jGrammar      = grammar{neo4j.Language, neo4j.Queries}
	surrealdbGrammar  = grammar{surrealdb.Language, surrealdb.Queries}
	ydbGrammar        = grammar{ydb.Language, ydb.Queries}
)

// grammarFor returns the SQL grammar for a dialect of dbmeta. A dialect with
// no grammar of its own gets the generic SQL grammar. Before a connection the
// dialect is empty, and the grammar is the one of PostgreSQL, as in psql.
func grammarFor(dialect string) grammar {
	switch dialect {
	case "", "postgres", "cockroachdb", "cratedb", "redshift":
		return postgresGrammar
	case "mysql", "tidb", "memsql", "vitess":
		return mysqlGrammar
	case "sqlserver":
		return sqlserverGrammar
	case "oracle":
		return oracleGrammar
	case "cassandra":
		return cqlGrammar
	case "sqlite3", "libsql", "rqlite":
		return sqlite3Grammar
	case "duckdb", "gizmosql":
		return duckdbGrammar
	case "clickhouse":
		return clickhouseGrammar
	case "trino", "presto", "athena":
		return trinoGrammar
	case "neo4j":
		return neo4jGrammar
	case "surrealdb":
		return surrealdbGrammar
	case "ydb":
		return ydbGrammar
	}
	return genericGrammar
}

// query returns the text of a query file of fsys, such as "highlights.scm", or
// the empty string when the grammar has no such file.
func query(fsys fs.FS, name string) string {
	b, err := fs.ReadFile(fsys, "queries/"+name)
	if err != nil {
		return ""
	}
	return string(b)
}
