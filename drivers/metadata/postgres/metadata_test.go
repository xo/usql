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
