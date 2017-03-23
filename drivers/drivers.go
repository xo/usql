package drivers

import (
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Drivers are the default sql drivers.
var Drivers = map[string]string{
	"cockroachdb": "cockroachdb", // github.com/lib/pq
	"memsql":      "memsql",      // github.com/go-sql-driver/mysql
	"mssql":       "mssql",       // github.com/denisenkom/go-mssqldb
	"mysql":       "mysql",       // github.com/go-sql-driver/mysql
	"postgres":    "pq",          // github.com/lib/pq
	"sqlite3":     "sqlite3",     // github.com/mattn/go-sqlite3
	"tidb":        "tidb",        // github.com/go-sql-driver/mysql
	"vitess":      "vitess",      // github.com/go-sql-driver/mysql
}

// KnownDrivers is the map of known drivers.
//
// build tag -> driver name
var KnownDrivers = map[string]string{
	"adodb":      "adodb",      // github.com/mattn/go-adodb
	"avatica":    "avatica",    // github.com/Boostport/avatica
	"clickhouse": "clickhouse", // github.com/kshvakov/clickhouse
	"couchbase":  "n1ql",       // github.com/couchbase/go_n1ql
	"odbc":       "odbc",       // github.com/alexbrainman/odbc
	"oracle":     "ora",        // gopkg.in/rana/ora.v4
	"ql":         "ql",         // github.com/cznic/ql/driver
	"saphana":    "hdb",        // github.com/SAP/go-hdb/driver
	"voltdb":     "voltdb",     // github.com/VoltDB/voltdb-client-go/voltdbclient
}
