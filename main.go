package main

//go:generate ./gen-license.sh

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/mattn/go-isatty"

	// sql drivers
	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/knq/usql/handler"
)

var (
	name    = "usql"
	version = "0.0.0-dev"
)

// drivers are the default sql drivers.
var drivers = map[string]string{
	"cockroachdb": "cockroachdb", // github.com/lib/pq
	"memsql":      "memsql",      // github.com/go-sql-driver/mysql
	"mssql":       "mssql",       // github.com/denisenkom/go-mssqldb
	"mysql":       "mysql",       // github.com/go-sql-driver/mysql
	"postgres":    "pq",          // github.com/lib/pq
	"sqlite3":     "sqlite3",     // github.com/mattn/go-sqlite3
	"tidb":        "tidb",        // github.com/go-sql-driver/mysql
	"vitess":      "vitess",      // github.com/go-sql-driver/mysql
}

// allKnownDrivers is the map of all known drivers.
var allKnownDrivers = map[string]string{
	"adodb":       "adodb",       // github.com/mattn/go-adodb
	"avatica":     "avatica",     // github.com/Boostport/avatica
	"clickhouse":  "clickhouse",  // github.com/kshvakov/clickhouse
	"cockroachdb": "cockroachdb", // github.com/lib/pq
	"couchbase":   "n1ql",        // github.com/couchbase/go_n1ql
	"memsql":      "memsql",      // github.com/go-sql-driver/mysql
	"mssql":       "mssql",       // github.com/denisenkom/go-mssqldb
	"mysql":       "mysql",       // github.com/go-sql-driver/mysql
	"odbc":        "odbc",        // github.com/alexbrainman/odbc
	"oleodbc":     "oleodbc",     // github.com/mattn/go-adodb
	"oracle":      "ora",         // gopkg.in/rana/ora.v4
	"postgres":    "postgres",    // github.com/lib/pq
	"ql":          "ql",          // github.com/cznic/ql/driver
	"saphana":     "hdb",         // github.com/SAP/go-hdb/driver
	"sqlite3":     "sqlite3",     // github.com/mattn/go-sqlite3
	"tidb":        "tidb",        // github.com/go-sql-driver/mysql
	"vitess":      "vitess",      // github.com/go-sql-driver/mysql
	"voltdb":      "voltdb",      // github.com/VoltDB/voltdb-client-go/voltdbclient
}

func main() {
	// circumvent all logic to determine if usql was built with support for a specific driver
	if len(os.Args) == 2 &&
		strings.HasPrefix(os.Args[1], "--has-") &&
		strings.HasSuffix(os.Args[1], "-support") {

		n := os.Args[1][6 : len(os.Args[1])-8]
		if v, ok := allKnownDrivers[n]; ok {
			n = v
		}

		var out int
		if _, ok := drivers[n]; ok {
			out = 1
		}
		fmt.Fprintf(os.Stdout, "%d", out)
		return
	}

	// set available drivers
	handler.SetAvailableDrivers(drivers)

	var err error

	// load current user
	cur, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// parse args
	args := &Args{
		Username:    cur.Username,
		HistoryFile: filepath.Join(cur.HomeDir, ".usql_history"),
	}
	arg.MustParse(args)

	// run
	err = run(args, cur.HomeDir)
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		// extra output for when the oracle driver is not available
		if e, ok := err.(*handler.Error); ok && e.Err == handler.ErrDriverNotAvailable && e.Driver == "ora" {
			fmt.Fprint(os.Stderr, "\ntry:\n\n  go get -u -tags oracle github.com/knq/usql\n\n")
		}
	}
}

// run processes args, processing args.Commands if non-empty, or args.File if
// specified, otherwise launch an interactive readline from stdin.
func run(args *Args, homedir string) error {
	var err error

	// get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// determine interactive/cygwin
	interactive := isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
	cygwin := isatty.IsCygwinTerminal(os.Stdout.Fd()) && isatty.IsCygwinTerminal(os.Stdin.Fd())

	// create handler
	h, err := handler.New(args.HistoryFile, homedir, wd, interactive || cygwin, cygwin)
	if err != nil {
		return err
	}

	// open dsn
	err = h.Open(args.DSN)
	if err != nil {
		return err
	}

	// short circuit if commands provided as args
	if len(args.Commands) > 0 {
		return h.RunCommands(args.Commands)
	}

	return h.RunReadline(args.File, args.Out)
}
