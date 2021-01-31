// Package metadata_test runs integration tests for metadata package
// on real databases running in containers. During development, to avoid rebuilding
// containers every run, add the `-cleanup=false` flags when calling `go test`.
package metadata_test

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	dt "github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	_ "github.com/trinodb/trino-go-client/trino"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/drivers/metadata/informationschema"
	_ "github.com/xo/usql/drivers/postgres"
)

type Database struct {
	BuildArgs  []dc.BuildArg
	RunOptions *dt.RunOptions
	Driver     string
	URL        string
	DockerPort string
	Resource   *dt.Resource
	DB         *sql.DB
	Opts       []informationschema.Option
	Reader     metadata.Reader
	WriterOpts []metadata.Option
}

var (
	dbs = map[string]*Database{
		"pgsql": {
			BuildArgs: []dc.BuildArg{
				{Name: "BASE_IMAGE", Value: "postgres:13"},
				{Name: "SCHEMA_URL", Value: "https://raw.githubusercontent.com/jOOQ/jOOQ/main/jOOQ-examples/Sakila/postgres-sakila-db/postgres-sakila-schema.sql"},
				{Name: "TARGET", Value: "/docker-entrypoint-initdb.d"},
			},
			RunOptions: &dt.RunOptions{
				Name: "usql-pgsql",
				Cmd:  []string{"-c", "log_statement=all", "-c", "log_min_duration_statement=0"},
				Env:  []string{"POSTGRES_PASSWORD=pw"},
			},
			Driver:     "postgres",
			URL:        "postgres://postgres:pw@localhost:%s/postgres?sslmode=disable",
			DockerPort: "5432/tcp",
			Opts: []informationschema.Option{
				informationschema.WithIndexes(false),
			},
			WriterOpts: []metadata.Option{
				metadata.WithSystemSchemas([]string{"pg_catalog", "pg_toast", "information_schema"}),
			},
		},
		"mysql": {
			BuildArgs: []dc.BuildArg{
				{Name: "BASE_IMAGE", Value: "mysql:8"},
				{Name: "SCHEMA_URL", Value: "https://raw.githubusercontent.com/jOOQ/jOOQ/main/jOOQ-examples/Sakila/mysql-sakila-db/mysql-sakila-schema.sql"},
				{Name: "TARGET", Value: "/docker-entrypoint-initdb.d"},
			},
			RunOptions: &dt.RunOptions{
				Name: "usql-mysql",
				Cmd:  []string{"--general-log=1", "--general-log-file=/var/lib/mysql/mysql.log"},
				Env:  []string{"MYSQL_ROOT_PASSWORD=pw"},
			},
			Driver:     "mysql",
			URL:        "root:pw@(localhost:%s)/mysql?parseTime=true",
			DockerPort: "3306/tcp",
			Opts: []informationschema.Option{
				informationschema.WithPlaceholder(func(int) string { return "?" }),
				informationschema.WithSequences(false),
			},
			WriterOpts: []metadata.Option{
				metadata.WithSystemSchemas([]string{"mysql", "performance_schema", "information_schema"}),
			},
		},
		"trino": {
			BuildArgs: []dc.BuildArg{
				{Name: "BASE_IMAGE", Value: "trinodb/trino:351"},
			},
			RunOptions: &dt.RunOptions{
				Name: "usql-trino",
			},
			Driver:     "trino",
			URL:        "http://test@localhost:%s?catalog=tpch&schema=sf1",
			DockerPort: "8080/tcp",
			Opts: []informationschema.Option{
				informationschema.WithPlaceholder(func(int) string { return "?" }),
				informationschema.WithTypeDetails(false),
				informationschema.WithFunctions(false),
				informationschema.WithSequences(false),
				informationschema.WithIndexes(false),
			},
		},
	}
	cleanup bool
)

func TestMain(m *testing.M) {
	flag.BoolVar(&cleanup, "cleanup", true, "delete containers when finished")
	flag.Parse()

	pool, err := dt.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	for _, db := range dbs {
		var ok bool
		db.Resource, ok = pool.ContainerByName(db.RunOptions.Name)
		if !ok {
			buildOpts := &dt.BuildOptions{
				ContextDir: ".",
				BuildArgs:  db.BuildArgs,
			}
			db.Resource, err = pool.BuildAndRunWithBuildOptions(buildOpts, db.RunOptions)
			if err != nil {
				log.Fatal("Could not start resource: ", err)
			}
		}

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		if err := pool.Retry(func() error {
			var err error
			db.DB, err = sql.Open(db.Driver, fmt.Sprintf(db.URL, db.Resource.GetPort(db.DockerPort)))
			if err != nil {
				return err
			}
			return db.DB.Ping()
		}); err != nil {
			log.Fatal("Could not connect to docker: ", err)
		}
		db.Reader = informationschema.New(db.Opts...)(db.DB)
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
		label string
		f     func(w metadata.Writer) error
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
					f: func(w metadata.Writer) error {
						return w.DescribeTableDetails("film*", true, false)
					},
				},
				{
					label: "listTables",
					f: func(w metadata.Writer) error {
						return w.ListTables("tvmsE", "film*", true, false)
					},
				},
				{
					label: "listFuncs",
					f: func(w metadata.Writer) error {
						return w.DescribeFunctions("", "", false, false)
					},
				},
				{
					label: "listSchemas",
					f: func(w metadata.Writer) error {
						return w.ListSchemas("", true, false)
					},
				},
			},
		},
		{
			dbName: "mysql",
			funcs: []testFunc{
				{
					label: "descTable",
					f: func(w metadata.Writer) error {
						return w.DescribeTableDetails("film*", true, false)
					},
				},
				{
					label: "listTables",
					f: func(w metadata.Writer) error {
						return w.ListTables("tvmsE", "film*", true, false)
					},
				},
				{
					label: "listFuncs",
					f: func(w metadata.Writer) error {
						return w.DescribeFunctions("", "", false, false)
					},
				},
				{
					label: "listIndexes",
					f: func(w metadata.Writer) error {
						return w.ListIndexes("", true, false)
					},
				},
				{
					label: "listSchemas",
					f: func(w metadata.Writer) error {
						return w.ListSchemas("", true, false)
					},
				},
			},
		},
		{
			dbName: "trino",
			funcs: []testFunc{
				{
					label: "descTable",
					f: func(w metadata.Writer) error {
						return w.DescribeTableDetails("order*", true, false)
					},
				},
				{
					label: "listTables",
					f: func(w metadata.Writer) error {
						return w.ListTables("tvmsE", "order*", true, false)
					},
				},
				{
					label: "listSchemas",
					f: func(w metadata.Writer) error {
						return w.ListSchemas("", true, false)
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

			db := dbs[test.dbName]
			w := metadata.NewDefaultWriter(db.Reader, db.WriterOpts...)(db.DB, fo)

			err = testFunc.f(w)
			if err != nil {
				log.Fatalf("Could not write %s %s: %v", test.dbName, testFunc.label, err)
			}
			err = fo.Close()
			if err != nil {
				t.Fatalf("Cannot close results file %s: %v", actual, err)
			}

			expected := fmt.Sprintf("testdata/%s.%s.expected.txt", test.dbName, testFunc.label)
			err = filesEqual(expected, actual)
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func filesEqual(a, b string) error {
	// per comment, better to not read an entire file into memory
	// this is simply a trivial example.
	f1, err := ioutil.ReadFile(a)
	if err != nil {
		return fmt.Errorf("Cannot read file %s: %w", a, err)
	}

	f2, err := ioutil.ReadFile(b)
	if err != nil {
		return fmt.Errorf("Cannot read file %s: %w", b, err)
	}

	if !bytes.Equal(f1, f2) {
		return fmt.Errorf("Files %s and %s have different contents", a, b)
	}
	return nil
}
