package informationschema_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	dt "github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
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
}

var (
	dbs = map[string]*Database{
		"pg": {
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
		},
		"my": {
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
			},
		},
	}
)

func TestMain(m *testing.M) {
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
	}
	//exitCode, err := resource.Exec([]string{"psql", "-f", "/postgres-sakila-schema.sql"}, dt.ExecOptions{
	//	StdIn:  os.Stdin,
	//	StdOut: os.Stdout,
	//	StdErr: os.Stderr,
	//	TTY:    true,
	//	Env:    []string{"PGUSER=postgres", "PGPASSWORD=pw"},
	//})
	//if err != nil || exitCode != 0 {
	//	log.Fatal("Could not load schema: ", err)
	//}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	for _, db := range dbs {
		if err := pool.Purge(db.Resource); err != nil {
			log.Fatal("Could not purge resource: ", err)
		}
	}

	os.Exit(code)
}

func TestSchemas(t *testing.T) {
	expected := map[string]string{
		"pg": "information_schema, pg_catalog, pg_toast, public",
		"my": "information_schema, mysql, performance_schema, sakila, sys",
	}
	for dbName, db := range dbs {
		r := informationschema.New(db.Opts...)(db.DB)

		result, err := r.Schemas()
		if err != nil {
			log.Fatalf("Could not read %s schemas: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Schema)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s schema names, expected:\n  %v, got:\n  %v", dbName, expected[dbName], names)
		}
	}
}

func TestTables(t *testing.T) {
	schemas := map[string]string{
		"pg": "public",
		"my": "sakila",
	}
	expected := map[string]string{
		"pg": "actor, address, category, city, country, customer, film, film_actor, film_category, inventory, language, payment, payment_p2007_01, payment_p2007_02, payment_p2007_03, payment_p2007_04, payment_p2007_05, payment_p2007_06, rental, staff, store, actor_info, customer_list, film_list, nicer_but_slower_film_list, sales_by_film_category, sales_by_store, staff_list",
		"my": "actor, address, category, city, country, customer, film, film_actor, film_category, film_text, inventory, language, payment, rental, staff, store, actor_info, customer_list, film_list, nicer_but_slower_film_list, sales_by_film_category, sales_by_store, staff_list",
	}
	for dbName, db := range dbs {
		r := informationschema.New(db.Opts...)(db.DB)

		result, err := r.Tables("", schemas[dbName], "", []string{"BASE TABLE", "TABLE", "VIEW"})
		if err != nil {
			log.Fatalf("Could not read %s tables: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Name)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s table names, expected:\n  %v, got:\n  %v", dbName, expected[dbName], names)
		}
	}
}

func TestColumns(t *testing.T) {
	schemas := map[string]string{
		"pg": "public",
		"my": "sakila",
	}
	expected := map[string]string{
		"pg": "film_id, title, description, release_year, language_id, original_language_id, rental_duration, rental_rate, length, replacement_cost, rating, last_update, special_features, fulltext, actor_id, film_id, last_update, film_id, category_id, last_update, fid, title, description, category, price, length, rating, actors",
		"my": "film_id, title, description, release_year, language_id, original_language_id, rental_duration, rental_rate, length, replacement_cost, rating, special_features, last_update, actor_id, film_id, last_update, film_id, category_id, last_update, FID, title, description, category, price, length, rating, actors, film_id, title, description",
	}
	for dbName, db := range dbs {
		r := informationschema.New(db.Opts...)(db.DB)

		result, err := r.Columns("", schemas[dbName], "film%")
		if err != nil {
			log.Fatalf("Could not read %s columns: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Name)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s column names, expected:\n  %v, got:\n  %v", expected[dbName], expected, names)
		}
	}
}
