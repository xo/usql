// Package drivers_test runs integration tests for drivers package
// on real databases running in containers. During development, add the
// `-cleanup=false` flag when calling `go test github.com/xo/usql/drivers` to
// leave the containers running for inspection afterwards; they are labeled
// `usql-test`, and can be removed with:
//
//	docker container rm -f $(docker container ls -aq --filter label=usql-test)
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

	dt "github.com/ory/dockertest/v4"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	_ "github.com/xo/usql/internal"
)

type Database struct {
	BuildArgs map[string]string
	Image     string
	RunOpts   []dt.RunOption
	DSN       string
	ReadyDSN  string
	Exec      []string

	DockerPort string
	Resource   dt.ClosableResource
	URL        *dburl.URL
	DB         *sql.DB
}

// maxWait is how long to wait for a container to become ready. Loading the
// sakila schema into mysql or starting trino takes well over a minute.
const maxWait = 5 * time.Minute

// buildArgs converts name/value pairs to the form dockertest expects.
func buildArgs(args map[string]string) map[string]*string {
	m := make(map[string]*string, len(args))
	for k, v := range args {
		m[k] = &v
	}
	return m
}

// startResource starts the container for db. A database that declares no
// SCHEMA_URL has nothing to add to its base image, and the shared build file
// always runs ADD, so that case runs the base image directly instead.
func startResource(ctx context.Context, pool dt.Pool, contextDir string, db *Database) (dt.ClosableResource, error) {
	if db.BuildArgs["SCHEMA_URL"] == "" {
		// Run takes a repository and a separate tag option, so a tagged
		// reference has to be split or it ends up as "repo:tag:latest".
		repository, tag := db.BuildArgs["BASE_IMAGE"], ""
		if i := strings.LastIndex(repository, ":"); i != -1 && !strings.Contains(repository[i+1:], "/") {
			repository, tag = repository[:i], repository[i+1:]
		}
		opts := db.RunOpts
		if tag != "" {
			opts = append(append([]dt.RunOption{}, opts...), dt.WithTag(tag))
		}
		return pool.Run(ctx, repository, opts...)
	}
	return pool.BuildAndRun(ctx, db.Image, &dt.BuildOptions{
		ContextDir: contextDir,
		Dockerfile: "Containerfile",
		BuildArgs:  buildArgs(db.BuildArgs),
	}, db.RunOpts...)
}

const (
	pw = "yourStrong123_Password"
)

var (
	dbs = map[string]*Database{
		"pgsql": {
			BuildArgs: map[string]string{
				"BASE_IMAGE": "postgres:13",
				"SCHEMA_URL": "https://raw.githubusercontent.com/jOOQ/sakila/main/postgres-sakila-db/postgres-sakila-schema.sql",
				"TARGET":     "/docker-entrypoint-initdb.d",
				"USER":       "root",
			},
			Image: "usql-pgsql",
			RunOpts: []dt.RunOption{
				dt.WithCmd([]string{"-c", "log_statement=all", "-c", "log_min_duration_statement=0"}),
				dt.WithEnv([]string{"POSTGRES_PASSWORD=pw"}),
				dt.WithLabels(map[string]string{"usql-test": "pgsql"}),
			},
			DSN:        "postgres://postgres:pw@localhost:%s/postgres?sslmode=disable",
			DockerPort: "5432/tcp",
		},
		"pgx": {
			BuildArgs: map[string]string{
				"BASE_IMAGE": "postgres:13",
				"SCHEMA_URL": "https://raw.githubusercontent.com/jOOQ/sakila/main/postgres-sakila-db/postgres-sakila-schema.sql",
				"TARGET":     "/docker-entrypoint-initdb.d",
				"USER":       "root",
			},
			Image: "usql-pgsql",
			RunOpts: []dt.RunOption{
				dt.WithCmd([]string{"-c", "log_statement=all", "-c", "log_min_duration_statement=0"}),
				dt.WithEnv([]string{"POSTGRES_PASSWORD=pw"}),
				dt.WithLabels(map[string]string{"usql-test": "pgsql"}),
			},
			DSN:        "pgx://postgres:pw@localhost:%s/postgres?sslmode=disable",
			DockerPort: "5432/tcp",
		},
		"mysql": {
			BuildArgs: map[string]string{
				"BASE_IMAGE": "mysql:8",
				"SCHEMA_URL": "https://raw.githubusercontent.com/jOOQ/sakila/main/mysql-sakila-db/mysql-sakila-schema.sql",
				"TARGET":     "/docker-entrypoint-initdb.d",
				"USER":       "root",
			},
			Image: "usql-mysql",
			RunOpts: []dt.RunOption{
				// InnoDB cannot allocate AIO contexts when the host is near its
				// fs.aio-max-nr limit, which is common under rootless podman, and
				// mysqld then aborts at startup instead of starting slowly.
				dt.WithCmd([]string{"--general-log=1", "--general-log-file=/var/lib/mysql/mysql.log", "--innodb-use-native-aio=0"}),
				dt.WithEnv([]string{"MYSQL_ROOT_PASSWORD=pw"}),
				dt.WithLabels(map[string]string{"usql-test": "mysql"}),
			},
			DSN:        "mysql://root:pw@localhost:%s/sakila?parseTime=true",
			DockerPort: "3306/tcp",
		},
		"sqlserver": {
			BuildArgs: map[string]string{
				"BASE_IMAGE": "mcr.microsoft.com/mssql/server:2019-latest",
				"SCHEMA_URL": "https://raw.githubusercontent.com/jOOQ/sakila/main/sql-server-sakila-db/sql-server-sakila-schema.sql",
				"TARGET":     "/schema",
				"USER":       "mssql:0",
			},
			Image: "usql-sqlserver",
			RunOpts: []dt.RunOption{
				dt.WithEnv([]string{"ACCEPT_EULA=Y", "SA_PASSWORD=" + pw}),
				dt.WithLabels(map[string]string{"usql-test": "sqlserver"}),
			},
			DSN:        "sqlserver://sa:" + url.QueryEscape(pw) + "@127.0.0.1:%s?database=sakila",
			ReadyDSN:   "sqlserver://sa:" + url.QueryEscape(pw) + "@127.0.0.1:%s?database=master",
			Exec:       []string{"/opt/mssql-tools18/bin/sqlcmd", "-C", "-S", "localhost", "-U", "sa", "-P", pw, "-d", "master", "-i", "/schema/sql-server-sakila-schema.sql"},
			DockerPort: "1433/tcp",
		},
		"trino": {
			BuildArgs: map[string]string{
				"BASE_IMAGE": "trinodb/trino:359",
			},
			Image: "usql-trino",
			RunOpts: []dt.RunOption{
				dt.WithLabels(map[string]string{"usql-test": "trino"}),
			},
			DSN:        "trino://test@localhost:%s/tpch/sf1",
			DockerPort: "8080/tcp",
		},
		"csvq": {
			// go test sets working directory to current package regardless of initial working directory
			DSN: "csvq://./testdata/csvq",
		},
	}
	cleanup bool
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

// run sets up, runs the tests and tears down, returning the code to exit with.
//
// TestMain cannot do this itself. os.Exit skips deferred functions, so any
// fatal error during setup used to leave behind every container started before
// it, and those then starve the next run of memory. Everything here returns an
// error instead, so the deferred purge always happens.
func run(m *testing.M) int {
	// Keep the standard logger on stderr, so that a failure here cannot exit
	// silently. A driver can take it away: calling slog.SetDefault from an
	// init also redirects the standard log package into that handler, and one
	// pointed anywhere but stderr swallows every log.Print and log.Fatalf in
	// the process. The driver that did it has been removed, but this is one
	// line and nothing stops the next one.
	log.SetOutput(os.Stderr)

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

	ctx := context.Background()
	pool, err := dt.NewPool(ctx, "")
	if err != nil {
		log.Printf("Could not connect to docker: %s", err)
		return 1
	}
	// From here on a container may exist, so nothing may exit without this.
	defer func() {
		if !cleanup {
			return
		}
		if err := pool.Close(ctx); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	if err := setup(ctx, pool); err != nil {
		log.Print(err)
		return 1
	}
	return m.Run()
}

// setup starts a container for each database and waits for it to answer.
func setup(ctx context.Context, pool dt.Pool) error {
	for dbName, db := range dbs {
		dsn, hostPort, err := connInfo(ctx, dbName, db, pool)
		if err != nil {
			return err
		}
		if db.URL, err = dburl.Parse(dsn); err != nil {
			return fmt.Errorf("parsing the %s URL %s: %w", dbName, db.DSN, err)
		}

		if len(db.Exec) != 0 {
			readyDSN := db.ReadyDSN
			if db.ReadyDSN == "" {
				readyDSN = db.DSN
			}
			if hostPort != "" {
				readyDSN = fmt.Sprintf(db.ReadyDSN, hostPort)
			}
			readyURL, err := dburl.Parse(readyDSN)
			if err != nil {
				return fmt.Errorf("parsing the %s ready URL %s: %w", dbName, db.ReadyDSN, err)
			}
			if err := pool.Retry(ctx, maxWait, func() error {
				readyDB, err := drivers.Open(ctx, readyURL, nil, nil)
				if err != nil {
					return err
				}
				return readyDB.Ping()
			}); err != nil {
				return fmt.Errorf("waiting for %s to be ready: %w", dbName, err)
			}
			res, err := db.Resource.Exec(ctx, db.Exec)
			if err != nil || res.ExitCode != 0 {
				return fmt.Errorf("loading the schema for %s: %w\n%s\n%s", dbName, err, res.StdOut, res.StdErr)
			}
		}

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		var openErr error
		if retryErr := pool.Retry(ctx, maxWait, func() error {
			db.DB, openErr = drivers.Open(ctx, db.URL, nil, nil)
			if openErr != nil {
				return openErr
			}
			if err := db.DB.Ping(); err != nil {
				return err
			}
			// Ping alone is not a readiness signal. Trino answers it while it
			// is still starting and then resets the first real query.
			var ok int
			return db.DB.QueryRowContext(ctx, "SELECT 1").Scan(&ok)
		}); retryErr != nil {
			return fmt.Errorf("waiting for %s: %w (open: %v)", dbName, retryErr, openErr)
		}
	}
	return nil
}

// connInfo starts the container for db, if it needs one, and returns its DSN
// and the host port it was published on.
func connInfo(ctx context.Context, dbName string, db *Database, pool dt.Pool) (string, string, error) {
	if db.Image == "" {
		return db.DSN, "", nil
	}

	// containers are reused within a run, keyed on the built image -- dbs
	// sharing an Image (pgsql and pgx) share a single container
	var err error
	db.Resource, err = startResource(ctx, pool, "./testdata/docker", db)
	if err != nil {
		return "", "", fmt.Errorf("starting %s: %w", dbName, err)
	}
	hostPort := db.Resource.GetPort(db.DockerPort)
	return fmt.Sprintf(db.DSN, hostPort), hostPort, nil
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
				t.Fatalf("Cannot create writer %s %s: %v", test.dbName, testFunc.label, err)
			}

			err = testFunc.f(w, db.URL)
			if err != nil {
				t.Fatalf("Cannot write %s %s: %v", test.dbName, testFunc.label, err)
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
			dbName: "pgsql",
			setupQueries: []setupQuery{
				{query: "DROP TABLE staff_copy"},
				{query: "CREATE TABLE staff_copy AS SELECT * FROM staff WHERE 0=1", check: true},
			},
			src:  "select * from staff",
			dest: "public.staff_copy",
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
			dbName: "pgx",
			setupQueries: []setupQuery{
				{query: "DROP TABLE staff_copy"},
				{query: "CREATE TABLE staff_copy AS SELECT * FROM staff WHERE 0=1", check: true},
			},
			src:  "select * from staff",
			dest: "public.staff_copy",
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
		{
			dbName: "csvq",
			setupQueries: []setupQuery{
				{query: "CREATE TABLE IF NOT EXISTS staff_copy AS SELECT * FROM `staff.csv` WHERE 0=1", check: true},
			},
			src:  "select first_name, last_name, address_id, email, store_id, active, username, password, last_update from staff",
			dest: "staff_copy",
		},
	}
	for _, test := range testCases {
		db, ok := dbs[test.dbName]
		if !ok {
			continue
		}

		t.Run(test.dbName, func(t *testing.T) {

			// TODO test copy from a different DB, maybe csvq?
			// TODO test copy from same DB

			for _, q := range test.setupQueries {
				_, err := db.DB.Exec(q.query)
				if q.check && err != nil {
					t.Fatalf("Failed to run setup query `%s`: %v", q.query, err)
				}
			}
			rows, err := pg.DB.Query(test.src)
			if err != nil {
				t.Fatalf("Could not get rows to copy: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var rlen int64 = 1
			n, err := drivers.Copy(ctx, db.URL, nil, nil, rows, test.dest)
			if err != nil {
				t.Fatalf("Could not copy: %v", err)
			}
			if n != rlen {
				t.Fatalf("Expected to copy %d rows but got %d", rlen, n)
			}
		})
	}
}

// filesEqual compares the files at paths a and b and returns an error if
// the content is not equal. Ignore is a regex. All matches will be removed
// from the file contents before comparison.
func filesEqual(a, b, ignore string) error {
	// Reading both files whole is not how this should work on a large file.
	// The fixtures here are small enough that it does not matter.
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
