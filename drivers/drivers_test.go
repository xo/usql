// Package drivers_test runs integration tests for drivers package
// on real databases running in containers. During development, to avoid rebuilding
// containers every run, add the `-cleanup=false` flags when calling `go test github.com/xo/usql/drivers`.
package drivers_test

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	dt "github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	_ "github.com/xo/usql/internal"
)

type Database struct {
	BuildArgs  []dc.BuildArg
	RunOptions *dt.RunOptions
	DSN        string
	ReadyDSN   string
	Exec       []string

	DockerPort string
	Resource   *dt.Resource
	URL        *dburl.URL
	DB         *sql.DB
}

const (
	pw = "yourStrong123_Password"
)

var (
	dbs = map[string]*Database{
		"pgsql": {
			BuildArgs: []dc.BuildArg{
				{Name: "BASE_IMAGE", Value: "postgres:13"},
				{Name: "SCHEMA_URL", Value: "https://raw.githubusercontent.com/jOOQ/sakila/main/postgres-sakila-db/postgres-sakila-schema.sql"},
				{Name: "TARGET", Value: "/docker-entrypoint-initdb.d"},
				{Name: "USER", Value: "root"},
			},
			RunOptions: &dt.RunOptions{
				Name: "usql-pgsql",
				Cmd:  []string{"-c", "log_statement=all", "-c", "log_min_duration_statement=0"},
				Env:  []string{"POSTGRES_PASSWORD=pw"},
			},
			DSN:        "postgres://postgres:pw@localhost:%s/postgres?sslmode=disable",
			DockerPort: "5432/tcp",
		},
		"pgx": {
			BuildArgs: []dc.BuildArg{
				{Name: "BASE_IMAGE", Value: "postgres:13"},
				{Name: "SCHEMA_URL", Value: "https://raw.githubusercontent.com/jOOQ/sakila/main/postgres-sakila-db/postgres-sakila-schema.sql"},
				{Name: "TARGET", Value: "/docker-entrypoint-initdb.d"},
				{Name: "USER", Value: "root"},
			},
			RunOptions: &dt.RunOptions{
				Name: "usql-pgsql",
				Cmd:  []string{"-c", "log_statement=all", "-c", "log_min_duration_statement=0"},
				Env:  []string{"POSTGRES_PASSWORD=pw"},
			},
			DSN:        "pgx://postgres:pw@localhost:%s/postgres?sslmode=disable",
			DockerPort: "5432/tcp",
		},
		"mysql": {
			BuildArgs: []dc.BuildArg{
				{Name: "BASE_IMAGE", Value: "mysql:8"},
				{Name: "SCHEMA_URL", Value: "https://raw.githubusercontent.com/jOOQ/sakila/main/mysql-sakila-db/mysql-sakila-schema.sql"},
				{Name: "TARGET", Value: "/docker-entrypoint-initdb.d"},
				{Name: "USER", Value: "root"},
			},
			RunOptions: &dt.RunOptions{
				Name: "usql-mysql",
				Cmd:  []string{"--general-log=1", "--general-log-file=/var/lib/mysql/mysql.log"},
				Env:  []string{"MYSQL_ROOT_PASSWORD=pw"},
			},
			DSN:        "mysql://root:pw@localhost:%s/sakila?parseTime=true",
			DockerPort: "3306/tcp",
		},
		"sqlserver": {
			BuildArgs: []dc.BuildArg{
				{Name: "BASE_IMAGE", Value: "mcr.microsoft.com/mssql/server:2019-latest"},
				{Name: "SCHEMA_URL", Value: "https://raw.githubusercontent.com/jOOQ/sakila/main/sql-server-sakila-db/sql-server-sakila-schema.sql"},
				{Name: "TARGET", Value: "/schema"},
				{Name: "USER", Value: "mssql:0"},
			},
			RunOptions: &dt.RunOptions{
				Name: "usql-sqlserver",
				Env:  []string{"ACCEPT_EULA=Y", "SA_PASSWORD=" + pw},
			},
			DSN:        "sqlserver://sa:" + url.QueryEscape(pw) + "@127.0.0.1:%s?database=sakila",
			ReadyDSN:   "sqlserver://sa:" + url.QueryEscape(pw) + "@127.0.0.1:%s?database=master",
			Exec:       []string{"/opt/mssql-tools/bin/sqlcmd", "-S", "localhost", "-U", "sa", "-P", pw, "-d", "master", "-i", "/schema/sql-server-sakila-schema.sql"},
			DockerPort: "1433/tcp",
		},
		"trino": {
			BuildArgs: []dc.BuildArg{
				{Name: "BASE_IMAGE", Value: "trinodb/trino:359"},
			},
			RunOptions: &dt.RunOptions{
				Name: "usql-trino",
			},
			DSN:        "trino://test@localhost:%s/tpch/sf1",
			DockerPort: "8080/tcp",
		},
	}
	cleanup bool
)

func TestMain(m *testing.M) {
	var only string
	flag.BoolVar(&cleanup, "cleanup", true, "delete containers when finished")
	flag.StringVar(&only, "dbs", "", "comma separated list of dbs to test: pgsql, mysql, sqlserver, trino")
	flag.Parse()

	if only != "" {
		runOnly := map[string]struct{}{}
		for _, dbName := range strings.Split(only, ",") {
			dbName = strings.TrimSpace(dbName)
			runOnly[dbName] = struct{}{}
		}
		for dbName := range dbs {
			if _, ok := runOnly[dbName]; !ok {
				delete(dbs, dbName)
			}
		}
	}

	pool, err := dt.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	for dbName, db := range dbs {
		var ok bool
		db.Resource, ok = pool.ContainerByName(db.RunOptions.Name)
		if !ok {
			buildOpts := &dt.BuildOptions{
				ContextDir: "./testdata/docker",
				BuildArgs:  db.BuildArgs,
			}
			db.Resource, err = pool.BuildAndRunWithBuildOptions(buildOpts, db.RunOptions)
			if err != nil {
				log.Fatalf("Could not start %s: %s", dbName, err)
			}
		}

		hostPort := db.Resource.GetPort(db.DockerPort)
		db.URL, err = dburl.Parse(fmt.Sprintf(db.DSN, hostPort))
		if err != nil {
			log.Fatalf("Failed to parse %s URL %s: %v", dbName, db.DSN, err)
		}

		if len(db.Exec) != 0 {
			if db.ReadyDSN == "" {
				db.ReadyDSN = db.DSN
			}
			readyURL, err := dburl.Parse(fmt.Sprintf(db.ReadyDSN, hostPort))
			if err != nil {
				log.Fatalf("Failed to parse %s ready URL %s: %v", dbName, db.ReadyDSN, err)
			}
			if err := pool.Retry(func() error {
				readyDB, err := drivers.Open(readyURL, nil, nil)
				if err != nil {
					return err
				}
				return readyDB.Ping()
			}); err != nil {
				log.Fatalf("Timed out waiting for %s to be ready: %s", dbName, err)
			}
			// No TTY attached to facilitate debugging with delve
			exitCode, err := db.Resource.Exec(db.Exec, dt.ExecOptions{})
			if err != nil || exitCode != 0 {
				log.Fatalf("Could not load schema for %s: %s", dbName, err)
			}
		}

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		var openErr error
		if retryErr := pool.Retry(func() error {
			db.DB, openErr = drivers.Open(db.URL, nil, nil)
			if openErr != nil {
				return openErr
			}
			return db.DB.Ping()
		}); retryErr != nil {
			log.Fatalf("Timed out waiting for %s:\n%s\n%s", dbName, retryErr, openErr)
		}
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if cleanup {
		for _, db := range dbs {
			if err := pool.Purge(db.Resource); err != nil {
				log.Fatal("Could not purge resource: ", err)
			}
		}
	}

	os.Exit(code)
}

func TestWriter(t *testing.T) {
	type testFunc struct {
		label  string
		f      func(w metadata.Writer, u *dburl.URL) error
		ignore string
	}
	testCases := []struct {
		dbName string
		funcs  []testFunc
	}{
		{
			dbName: "pgsql",
			funcs: []testFunc{
				{
					label: "descTable",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.DescribeTableDetails(u, "film*", true, false)
					},
				},
				{
					label: "listTables",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListTables(u, "tvmsE", "film*", true, false)
					},
				},
				{
					label: "listFuncs",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.DescribeFunctions(u, "", "", false, false)
					},
				},
				{
					label: "listIndexes",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListIndexes(u, "", true, false)
					},
				},
				{
					label: "listSchemas",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListSchemas(u, "", true, false)
					},
				},
			},
		},
		{
			dbName: "mysql",
			funcs: []testFunc{
				{
					label: "descTable",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.DescribeTableDetails(u, "film*", true, false)
					},
				},
				{
					label: "listTables",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListTables(u, "tvmsE", "film*", true, false)
					},
				},
				{
					label: "listFuncs",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.DescribeFunctions(u, "", "", false, false)
					},
				},
				{
					label: "listIndexes",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListIndexes(u, "", true, false)
					},
				},
				{
					label: "listSchemas",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListSchemas(u, "", true, false)
					},
				},
			},
		},
		{
			dbName: "sqlserver",
			funcs: []testFunc{
				{
					label: "descTable",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.DescribeTableDetails(u, "film*", true, false)
					},
					// primary key indices get random names; ignore them
					ignore: "PK__.*__.{16}",
				},
				{
					label: "listTables",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListTables(u, "tvmsE", "film*", true, false)
					},
				},
				{
					label: "listFuncs",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.DescribeFunctions(u, "", "", false, false)
					},
				},
				{
					label: "listIndexes",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListIndexes(u, "", true, false)
					},
					// primary key indices get random names; ignore them
					ignore: "PK__.*__.{16}",
				},
				{
					label: "listSchemas",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListSchemas(u, "", true, false)
					},
				},
			},
		},
		{
			dbName: "trino",
			funcs: []testFunc{
				{
					label: "descTable",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.DescribeTableDetails(u, "order*", true, false)
					},
				},
				{
					label: "listTables",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListTables(u, "tvmsE", "order*", true, false)
					},
				},
				{
					label: "listSchemas",
					f: func(w metadata.Writer, u *dburl.URL) error {
						return w.ListSchemas(u, "", true, false)
					},
				},
			},
		},
	}
	for _, test := range testCases {
		for _, testFunc := range test.funcs {
			actual := fmt.Sprintf("testdata/%s.%s.actual.txt", test.dbName, testFunc.label)
			fo, err := os.Create(actual)
			if err != nil {
				t.Fatalf("Cannot create results file %s: %v", actual, err)
			}

			db, ok := dbs[test.dbName]
			if !ok {
				continue
			}
			w, err := drivers.NewMetadataWriter(context.Background(), db.URL, db.DB, fo)
			if err != nil {
				log.Fatalf("Could not create writer %s %s: %v", test.dbName, testFunc.label, err)
			}

			err = testFunc.f(w, db.URL)
			if err != nil {
				log.Fatalf("Could not write %s %s: %v", test.dbName, testFunc.label, err)
			}
			err = fo.Close()
			if err != nil {
				t.Fatalf("Cannot close results file %s: %v", actual, err)
			}

			expected := fmt.Sprintf("testdata/%s.%s.expected.txt", test.dbName, testFunc.label)
			err = filesEqual(expected, actual, testFunc.ignore)
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func TestCopy(t *testing.T) {
	pg, ok := dbs["pgsql"]
	if !ok {
		t.Skip("Skipping copy tests, as they require PostgreSQL which was not selected for tests")
	}
	// setup test data, ignoring errors, since there'll be duplicates
	_, _ = pg.DB.Exec("ALTER TABLE staff DROP CONSTRAINT staff_address_id_fkey")
	_, _ = pg.DB.Exec("ALTER TABLE staff DROP CONSTRAINT staff_store_id_fkey")
	_, _ = pg.DB.Exec("INSERT INTO staff VALUES (1, 'John', 'Doe', 1, 'john@invalid.com', 1, true, 'jdoe', 'abc', now(), 'abcd')")

	type setupQuery struct {
		query string
		check bool
	}

	testCases := []struct {
		dbName       string
		setupQueries []setupQuery
		src          string
		dest         string
	}{
		{
			dbName: "pgsql",
			setupQueries: []setupQuery{
				{query: "DROP TABLE staff_copy"},
				{query: "CREATE TABLE staff_copy AS SELECT * FROM staff WHERE 0=1", check: true},
			},
			src:  "select * from staff",
			dest: "staff_copy",
		},
		{
			dbName: "pgx",
			setupQueries: []setupQuery{
				{query: "DROP TABLE staff_copy"},
				{query: "CREATE TABLE staff_copy AS SELECT * FROM staff WHERE 0=1", check: true},
			},
			src:  "select * from staff",
			dest: "staff_copy",
		},
		{
			dbName: "mysql",
			setupQueries: []setupQuery{
				{query: "DROP TABLE staff_copy"},
				{query: "CREATE TABLE staff_copy AS SELECT * FROM staff WHERE 0=1", check: true},
			},
			src:  "select staff_id, first_name, last_name, address_id, picture, email, store_id, active, username, password, last_update from staff",
			dest: "staff_copy(staff_id, first_name, last_name, address_id, picture, email, store_id, active, username, password, last_update)",
		},
		{
			dbName: "sqlserver",
			setupQueries: []setupQuery{
				{query: "DROP TABLE staff_copy"},
				{query: "SELECT * INTO staff_copy FROM staff WHERE 0=1", check: true},
			},
			src:  "select first_name, last_name, address_id, picture, email, store_id, active, username, password, last_update from staff",
			dest: "staff_copy(first_name, last_name, address_id, picture, email, store_id, active, username, password, last_update)",
		},
	}
	for _, test := range testCases {
		db, ok := dbs[test.dbName]
		if !ok {
			continue
		}

		// TODO test copy from a different DB, maybe csvq?
		// TODO test copy from same DB

		for _, q := range test.setupQueries {
			_, err := db.DB.Exec(q.query)
			if q.check && err != nil {
				log.Fatalf("Failed to run setup query `%s`: %v", q.query, err)
			}
		}
		rows, err := pg.DB.Query(test.src)
		if err != nil {
			log.Fatalf("Could not get rows to copy: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var rlen int64 = 1
		n, err := drivers.Copy(ctx, db.URL, nil, nil, rows, test.dest)
		if err != nil {
			log.Fatalf("Could not copy: %v", err)
		}
		if n != rlen {
			log.Fatalf("Expected to copy %d rows but got %d", rlen, n)
		}
	}
}

// filesEqual compares the files at paths a and b and returns an error if
// the content is not equal. Ignore is a regex. All matches will be removed
// from the file contents before comparison.
func filesEqual(a, b, ignore string) error {
	// per comment, better to not read an entire file into memory
	// this is simply a trivial example.
	f1, err := os.ReadFile(a)
	if err != nil {
		return fmt.Errorf("Cannot read file %s: %w", a, err)
	}

	f2, err := os.ReadFile(b)
	if err != nil {
		return fmt.Errorf("Cannot read file %s: %w", b, err)
	}

	if ignore != "" {
		reg, err := regexp.Compile(ignore)
		if err != nil {
			return fmt.Errorf("Cannot compile regex (%s): %w", ignore, err)
		}
		f1 = reg.ReplaceAllLiteral(f1, []byte{})
		f2 = reg.ReplaceAllLiteral(f2, []byte{})
	}

	if !bytes.Equal(f1, f2) {
		return fmt.Errorf("Files %s and %s have different contents", a, b)
	}
	return nil
}
