// +build !no_base

package internal

//go:generate ./gen.sh

// KnownBuildTags returns a map of known driver names to its respective build
// tags.
func KnownBuildTags() map[string]string {
	return map[string]string{
		"adodb":       "adodb",      // github.com/mattn/go-adodb
		"avatica":     "avatica",    // github.com/Boostport/avatica
		"clickhouse":  "clickhouse", // github.com/kshvakov/clickhouse
		"n1ql":        "couchbase",  // github.com/couchbase/go_n1ql
		"firebirdsql": "firebird",   // github.com/nakagami/firebirdsql
		"mssql":       "mssql",      // github.com/denisenkom/go-mssqldb
		"mymysql":     "mymysql",    // github.com/ziutek/mymysql/godrv
		"mysql":       "mysql",      // github.com/go-sql-driver/mysql
		"odbc":        "odbc",       // github.com/alexbrainman/odbc
		"ora":         "oracle",     // gopkg.in/rana/ora.v4
		"pgx":         "pgx",        // github.com/jackc/pgx/stdlib
		"postgres":    "postgres",   // github.com/lib/pq
		"ql":          "ql",         // github.com/cznic/ql/driver
		"hdb":         "saphana",    // github.com/SAP/go-hdb/driver
		"sqlite3":     "sqlite3",    // github.com/mattn/go-sqlite3
		"voltdb":      "voltdb",     // github.com/VoltDB/voltdb-client-go/voltdbclient
	}
}
