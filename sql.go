package main

import (
	"errors"

	// sql drivers
	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// ErrOracleDriverNotAvailable is the oracle driver not available error.
	ErrOracleDriverNotAvailable = errors.New("oracle driver not available")

	// ErrNotConnected is the not connected error.
	ErrNotConnected = errors.New("not connected")
)

// drivers is the available sql drivers.
var drivers = map[string]bool{
	"mssql":    true, // github.com/denisenkom/go-mssqldb
	"mysql":    true, // github.com/go-sql-driver/mysql
	"postgres": true, // github.com/lib/pq
	"sqlite3":  true, // github.com/mattn/go-sqlite3
}
