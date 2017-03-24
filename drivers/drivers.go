package drivers

import (
	"runtime"

	// mssql driver
	_ "github.com/denisenkom/go-mssqldb"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	// postgres driver
	_ "github.com/lib/pq"

	// sqlite3 driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/knq/dburl"
)

// Drivers are the supported SQL drivers.
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
	"adodb":      "adodb",       // github.com/mattn/go-adodb
	"avatica":    "avatica",     // github.com/Boostport/avatica
	"clickhouse": "clickhouse",  // github.com/kshvakov/clickhouse
	"couchbase":  "n1ql",        // github.com/couchbase/go_n1ql
	"firebird":   "firebirdsql", // github.com/nakagami/firebirdsql
	"mymysql":    "mymysql",     // github.com/ziutek/mymysql/godrv
	"odbc":       "odbc",        // github.com/alexbrainman/odbc
	"oracle":     "ora",         // gopkg.in/rana/ora.v4
	"pgx":        "pgx",         // github.com/jackc/pgx
	"ql":         "ql",          // github.com/cznic/ql/driver
	"saphana":    "hdb",         // github.com/SAP/go-hdb/driver
	"voltdb":     "voltdb",      // github.com/VoltDB/voltdb-client-go/voltdbclient
	"yql":        "yql",         // github.com/mattn/go-yql
}

func init() {
	if runtime.GOOS == "windows" {
		// if no odbc driver, but we have adodb, add 'odbc' as alias to oleodbc
		// driver.
		_, haveODBC := Drivers["odbc"]
		_, haveADODB := Drivers["adodb"]
		if haveADODB && !haveODBC {
			dburl.Unregister("odbc")
			dburl.RegisterAlias("oleodbc", "odbc")
		}
	}
}
