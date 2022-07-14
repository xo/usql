package postgres_test

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	dt "github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/drivers/metadata/postgres"
	_ "github.com/xo/usql/drivers/postgres"
)

type Database struct {
	BuildArgs  []dc.BuildArg
	RunOptions *dt.RunOptions
	Exec       []string
	Driver     string
	URL        string
	DockerPort string
	Resource   *dt.Resource
	DB         *sql.DB
	Opts       []metadata.ReaderOption
	Reader     metadata.BasicReader
}

var dbName string = "postgres"

var db = Database{
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
	Driver:     "postgres",
	URL:        "postgres://postgres:pw@localhost:%s/postgres?sslmode=disable",
	DockerPort: "5432/tcp",
}

func TestMain(m *testing.M) {
	cleanup := true
	flag.BoolVar(&cleanup, "cleanup", true, "delete containers when finished")
	flag.Parse()
	pool, err := dt.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	var ok bool
	db.Resource, ok = pool.ContainerByName(db.RunOptions.Name)
	if !ok {
		buildOpts := &dt.BuildOptions{
			ContextDir: "../../testdata/docker",
			BuildArgs:  db.BuildArgs,
		}
		db.Resource, err = pool.BuildAndRunWithBuildOptions(buildOpts, db.RunOptions)
		if err != nil {
			log.Fatal("Could not start resource: ", err)
		}
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		hostPort := db.Resource.GetPort(db.DockerPort)
		var err error
		db.DB, err = sql.Open(db.Driver, fmt.Sprintf(db.URL, hostPort))
		if err != nil {
			return err
		}
		return db.DB.Ping()
	}); err != nil {
		log.Fatal("Timed out waiting for db: ", err)
	}
	db.Reader = postgres.NewReader()(db.DB).(metadata.BasicReader)

	if len(db.Exec) != 0 {
		exitCode, err := db.Resource.Exec(db.Exec, dt.ExecOptions{
			StdIn:  os.Stdin,
			StdOut: os.Stdout,
			StdErr: os.Stderr,
			TTY:    true,
		})
		if err != nil || exitCode != 0 {
			log.Fatal("Could not load schema: ", err)
		}
	}
	code := m.Run()
	// You can't defer this because os.Exit doesn't care for defer
	if cleanup {
		if err := pool.Purge(db.Resource); err != nil {
			log.Fatal("Could not purge resource: ", err)
		}
	}
	os.Exit(code)
}

func TestTriggers(t *testing.T) {
	schema := "public"
	expected := "film_fulltext_trigger, last_updated"
	parent := "film"
	r := postgres.NewReader()(db.DB).(metadata.TriggerReader)

	result, err := r.Triggers(metadata.Filter{Schema: schema, Parent: parent})
	if err != nil {
		log.Fatalf("Could not read %s triggers: %v", dbName, err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	if actual != expected {
		t.Errorf("Wrong %s trigger names, expected:\n  %v\ngot:\n  %v", dbName, expected, names)
	}
}

func TestColumns(t *testing.T) {
	// Only testing postgres specific datatype formatting.
	// The rest of the functionality is covered by informationschema/metadata_test.go:TestColumns
	type test struct {
		typeDef string
		want    string
	}
	schema := "public"
	table := "test_dtypes"
	tests := []test{
		{typeDef: "bit", want: "bit(1)"},
		{typeDef: "bit(1)", want: "bit(1)"},
		{typeDef: "bit varying", want: "bit varying"},
		{typeDef: "bit varying(2)", want: "bit varying(2)"},
		{typeDef: "character", want: "character(1)"},
		{typeDef: "character(3)", want: "character(3)"},
		{typeDef: "character varying", want: "character varying"},
		{typeDef: "character varying(4)", want: "character varying(4)"},
		{typeDef: "numeric", want: "numeric"},
		{typeDef: "numeric(1,0)", want: "numeric(1,0)"},
		{typeDef: "time", want: "time(6) without time zone"},
		{typeDef: "time(4)", want: "time(4) without time zone"},
		{typeDef: "time(6)", want: "time(6) without time zone"},
		{typeDef: "time with time zone", want: "time(6) with time zone"},
		{typeDef: "time(3) with time zone", want: "time(3) with time zone"},
		{typeDef: "timestamp", want: "timestamp(6) without time zone"},
		{typeDef: "timestamp(2)", want: "timestamp(2) without time zone"},
		{typeDef: "timestamp with time zone", want: "timestamp(6) with time zone"},
		{typeDef: "timestamp(1) with time zone", want: "timestamp(1) with time zone"},
		{typeDef: "bigint", want: "bigint"},
		{typeDef: "bigserial", want: "bigint"},
		{typeDef: "boolean", want: "boolean"},
		{typeDef: "box", want: "box"},
		{typeDef: "bytea", want: "bytea"},
		{typeDef: "cidr", want: "cidr"},
		{typeDef: "circle", want: "circle"},
		{typeDef: "date", want: "date"},
		{typeDef: "double precision", want: "double precision"},
		{typeDef: "inet", want: "inet"},
		{typeDef: "integer", want: "integer"},
		{typeDef: "json", want: "json"},
		{typeDef: "jsonb", want: "jsonb"},
		{typeDef: "line", want: "line"},
		{typeDef: "lseg", want: "lseg"},
		{typeDef: "macaddr", want: "macaddr"},
		{typeDef: "macaddr8", want: "macaddr8"},
		{typeDef: "money", want: "money"},
		{typeDef: "path", want: "path"},
		{typeDef: "pg_lsn", want: "pg_lsn"},
		{typeDef: "pg_snapshot", want: "pg_snapshot"},
		{typeDef: "point", want: "point"},
		{typeDef: "polygon", want: "polygon"},
		{typeDef: "real", want: "real"},
		{typeDef: "smallint", want: "smallint"},
		{typeDef: "smallserial", want: "smallint"},
		{typeDef: "serial", want: "integer"},
		{typeDef: "text", want: "text"},
		{typeDef: "tsvector", want: "tsvector"},
		{typeDef: "txid_snapshot", want: "txid_snapshot"},
		{typeDef: "uuid", want: "uuid"},
		{typeDef: "xml", want: "xml"},
	}

	// Create table
	colExpressions := []string{}
	for i, test := range tests {
		colExpressions = append(colExpressions, fmt.Sprintf("column_%d %s", i, test.typeDef))
	}
	query := fmt.Sprintf("CREATE TABLE %s.%s (%s)", schema, table, strings.Join(colExpressions, ", "))
	db.DB.Exec(query)
	defer db.DB.Exec(fmt.Sprintf("DROP TABLE %s.%s", schema, table))

	// Read data types
	r := postgres.NewReader()(db.DB).(metadata.ColumnReader)
	result, err := r.Columns(metadata.Filter{Schema: schema, Parent: table})
	if err != nil {
		log.Fatalf("Could not read %s columns: %v", dbName, err)
	}
	actualTypes := []string{}
	for result.Next() {
		actualTypes = append(actualTypes, result.Get().DataType)
	}

	// Compare
	for i, test := range tests {
		if actualTypes[i] != test.want {
			t.Errorf("Wrong %s column data type, expected:\n  %s, got:\n  %s", dbName, test.want, actualTypes[i])
		}
	}
}
