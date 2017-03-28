package drivers

import (
	"database/sql"
	"strings"

	// mssql driver
	_ "github.com/denisenkom/go-mssqldb"

	// mysql driver
	"github.com/go-sql-driver/mysql"

	// postgres driver
	"github.com/lib/pq"

	// sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
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

var pwErr = map[string]func(error) bool{
	"mssql": func(err error) bool {
		return strings.Contains(err.Error(), "Login failed for")
	},
	"mysql": func(err error) bool {
		if e, ok := err.(*mysql.MySQLError); ok {
			return e.Number == 1045
		}
		return false
	},
	"postgres": func(err error) bool {
		if e, ok := err.(*pq.Error); ok {
			return e.Code.Name() == "invalid_password"
		}
		return false
	},
}

// IsPasswordErr returns true when the passed err is a authentication /
// password error for the driver.
func IsPasswordErr(name string, err error) bool {
	if f, ok := pwErr[name]; ok {
		return f(err)
	}

	return false
}

// GetDatabaseInfo returns information about the database.
func GetDatabaseInfo(name string, db *sql.DB) (product string, ver string, err error) {
	return "", "", nil
}
