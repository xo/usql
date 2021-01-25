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

var db *sql.DB

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dt.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	buildOpts := &dt.BuildOptions{
		ContextDir: ".",
		BuildArgs: []dc.BuildArg{
			{Name: "BASE_IMAGE", Value: "postgres:13"},
			{Name: "SCHEMA_URL", Value: "https://raw.githubusercontent.com/jOOQ/jOOQ/main/jOOQ-examples/Sakila/postgres-sakila-db/postgres-sakila-schema.sql"},
			{Name: "TARGET", Value: "/docker-entrypoint-initdb.d"},
		},
	}
	resource, err := pool.BuildAndRunWithBuildOptions(buildOpts, &dt.RunOptions{
		Name: "usql-test",
		Cmd:  []string{"-c", "log_statement=all", "-c", "log_min_duration_statement=0"},
		Env:  []string{"POSTGRES_PASSWORD=pw"},
	})
	if err != nil {
		log.Fatal("Could not start resource: ", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", fmt.Sprintf("postgres://postgres:pw@localhost:%s/%s?sslmode=disable", resource.GetPort("5432/tcp"), "postgres"))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatal("Could not connect to docker: ", err)
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
	if err := pool.Purge(resource); err != nil {
		log.Fatal("Could not purge resource: ", err)
	}

	os.Exit(code)
}

func TestSchemas(t *testing.T) {
	r := informationschema.New(db)

	result, err := r.Schemas()
	if err != nil {
		log.Fatal("Could not read schemas", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Schema)
	}
	actual := strings.Join(names, ", ")
	expected := "information_schema, pg_catalog, pg_toast, public"
	if actual != expected {
		t.Errorf("Wrong schema names, expected %v, got %v", expected, names)
	}
}

func TestTables(t *testing.T) {
	r := informationschema.New(db)

	result, err := r.Tables("", "public", "", []string{"BASE TABLE", "TABLE", "VIEW"})
	if err != nil {
		log.Fatal("Could not read tables", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	expected := "actor, address, category, city, country, customer, film, film_actor, film_category, inventory, language, payment, payment_p2007_01, payment_p2007_02, payment_p2007_03, payment_p2007_04, payment_p2007_05, payment_p2007_06, rental, staff, store, actor_info, customer_list, film_list, nicer_but_slower_film_list, sales_by_film_category, sales_by_store, staff_list"
	if actual != expected {
		t.Errorf("Wrong table names, expected %v, got %v", expected, names)
	}
}

func TestColumns(t *testing.T) {
	r := informationschema.New(db)

	result, err := r.Columns("", "public", "film%")
	if err != nil {
		log.Fatal("Could not read columns: ", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	expected := "film_id, title, description, release_year, language_id, original_language_id, rental_duration, rental_rate, length, replacement_cost, rating, last_update, special_features, fulltext, actor_id, film_id, last_update, film_id, category_id, last_update, fid, title, description, category, price, length, rating, actors"
	if actual != expected {
		t.Errorf("Wrong column names, expected %v, got %v", expected, names)
	}
}
