// Package informationschema_test runs integration tests for informationschema package
// on real databases running in containers. During development, to avoid rebuilding
// containers every run, add the `-cleanup=false` flags when calling `go test`.
package informationschema_test

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	Reader     metadata.BasicReader
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
				informationschema.WithTypeDetails(false),
				informationschema.WithPlaceholder(func(int) string { return "?" }),
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
		db.Reader = informationschema.New(db.Opts...)(db.DB).(metadata.BasicReader)
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
	if cleanup {
		for _, db := range dbs {
			if err := pool.Purge(db.Resource); err != nil {
				log.Fatal("Could not purge resource: ", err)
			}
		}
	}

	os.Exit(code)
}

func TestSchemas(t *testing.T) {
	expected := map[string]string{
		"pgsql": "information_schema, pg_catalog, pg_toast, public",
		"mysql": "information_schema, mysql, performance_schema, sakila, sys",
		"trino": "information_schema, sf1, sf100, sf1000, sf10000, sf100000, sf300, sf3000, sf30000, tiny",
	}
	for dbName, db := range dbs {
		r := db.Reader

		result, err := r.Schemas("", "")
		if err != nil {
			log.Fatalf("Could not read %s schemas: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Schema)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s schema names, expected:\n  %v\ngot:\n  %v", dbName, expected[dbName], names)
		}
	}
}

func TestTables(t *testing.T) {
	schemas := map[string]string{
		"pgsql": "public",
		"mysql": "sakila",
		"trino": "sf1",
	}
	expected := map[string]string{
		"pgsql": "actor, address, category, city, country, customer, film, film_actor, film_category, inventory, language, payment, payment_p2007_01, payment_p2007_02, payment_p2007_03, payment_p2007_04, payment_p2007_05, payment_p2007_06, rental, staff, store, actor_info, customer_list, film_list, nicer_but_slower_film_list, sales_by_film_category, sales_by_store, staff_list",
		"mysql": "actor, address, category, city, country, customer, film, film_actor, film_category, film_text, inventory, language, payment, rental, staff, store, actor_info, customer_list, film_list, nicer_but_slower_film_list, sales_by_film_category, sales_by_store, staff_list",
		"trino": "customer, lineitem, nation, orders, part, partsupp, region, supplier",
	}
	for dbName, db := range dbs {
		r := db.Reader

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
			t.Errorf("Wrong %s table names, expected:\n  %v\ngot:\n  %v", dbName, expected[dbName], names)
		}
	}
}

func TestColumns(t *testing.T) {
	schemas := map[string]string{
		"pgsql": "public",
		"mysql": "sakila",
		"trino": "sf1",
	}
	tables := map[string]string{
		"pgsql": "film%",
		"mysql": "film%",
		"trino": "orders",
	}
	expected := map[string]string{
		"pgsql": "film_id, title, description, release_year, language_id, original_language_id, rental_duration, rental_rate, length, replacement_cost, rating, last_update, special_features, fulltext, actor_id, film_id, last_update, film_id, category_id, last_update, fid, title, description, category, price, length, rating, actors",
		"mysql": "film_id, title, description, release_year, language_id, original_language_id, rental_duration, rental_rate, length, replacement_cost, rating, special_features, last_update, actor_id, film_id, last_update, film_id, category_id, last_update, FID, title, description, category, price, length, rating, actors, film_id, title, description",
		"trino": "orderkey, custkey, orderstatus, totalprice, orderdate, orderpriority, clerk, shippriority, comment",
	}
	for dbName, db := range dbs {
		r := db.Reader

		result, err := r.Columns("", schemas[dbName], tables[dbName])
		if err != nil {
			log.Fatalf("Could not read %s columns: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Name)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s column names, expected:\n  %v, got:\n  %v", dbName, expected[dbName], names)
		}
	}
}

func TestFunctions(t *testing.T) {
	schemas := map[string]string{
		"pgsql": "public",
		"mysql": "sakila",
	}
	expected := map[string]string{
		"pgsql": "_group_concat, film_in_stock, film_not_in_stock, get_customer_balance, inventory_held_by_customer, inventory_in_stock, last_day, last_updated, rewards_report, group_concat",
		"mysql": "get_customer_balance, inventory_held_by_customer, inventory_in_stock, film_in_stock, film_not_in_stock, rewards_report",
	}
	for dbName, db := range dbs {
		if schemas[dbName] == "" {
			continue
		}
		r := informationschema.New(db.Opts...)(db.DB).(metadata.FunctionReader)

		result, err := r.Functions("", schemas[dbName], "", []string{})
		if err != nil {
			log.Fatalf("Could not read %s functions: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Name)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s function names, expected:\n  %v\ngot:\n  %v", dbName, expected[dbName], names)
		}
	}
}

func TestFunctionColumns(t *testing.T) {
	schemas := map[string]string{
		"pgsql": "public",
		"mysql": "sakila",
	}
	tables := map[string]string{
		"pgsql": "film%",
		"mysql": "film%",
	}
	expected := map[string]string{
		"pgsql": "p_film_id, p_store_id, p_film_count, p_film_id, p_store_id, p_film_count",
		"mysql": "p_film_id, p_store_id, p_film_count, p_film_id, p_store_id, p_film_count",
	}
	for dbName, db := range dbs {
		if schemas[dbName] == "" {
			continue
		}
		r := informationschema.New(db.Opts...)(db.DB).(metadata.FunctionColumnReader)

		result, err := r.FunctionColumns("", schemas[dbName], tables[dbName])
		if err != nil {
			log.Fatalf("Could not read %s function columns: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Name)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s function column names, expected:\n  %v, got:\n  %v", dbName, expected[dbName], names)
		}
	}
}

func TestIndexes(t *testing.T) {
	schemas := map[string]string{
		"mysql": "sakila",
	}
	expected := map[string]string{
		"mysql": "actor.idx_actor_last_name, actor.PRIMARY, address.idx_fk_city_id, address.PRIMARY, category.PRIMARY, city.idx_fk_country_id, city.PRIMARY, country.PRIMARY, customer.idx_fk_address_id, customer.idx_fk_store_id, customer.idx_last_name, customer.PRIMARY, film.idx_fk_language_id, film.idx_fk_original_language_id, film.idx_title, film.PRIMARY, film_actor.idx_fk_film_id, film_actor.PRIMARY, film_category.fk_film_category_category, film_category.PRIMARY, film_text.idx_title_description, film_text.PRIMARY, inventory.idx_fk_film_id, inventory.idx_store_id_film_id, inventory.PRIMARY, language.PRIMARY, payment.fk_payment_rental, payment.idx_fk_customer_id, payment.idx_fk_staff_id, payment.PRIMARY, rental.idx_fk_customer_id, rental.idx_fk_inventory_id, rental.idx_fk_staff_id, rental.PRIMARY, rental.rental_date, staff.idx_fk_address_id, staff.idx_fk_store_id, staff.PRIMARY, store.idx_fk_address_id, store.idx_unique_manager, store.PRIMARY",
	}
	for dbName, db := range dbs {
		if schemas[dbName] == "" {
			continue
		}
		r := informationschema.New(db.Opts...)(db.DB).(metadata.IndexReader)

		result, err := r.Indexes("", schemas[dbName], "")
		if err != nil {
			log.Fatalf("Could not read %s indexes: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Table+"."+result.Get().Name)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s index names, expected:\n  %v\ngot:\n  %v", dbName, expected[dbName], names)
		}
	}
}

func TestIndexColumns(t *testing.T) {
	schemas := map[string]string{
		"mysql": "sakila",
	}
	tables := map[string]string{
		"mysql": "idx%",
	}
	expected := map[string]string{
		"mysql": "last_name, address_id, address_id, address_id, city_id, country_id, customer_id, customer_id, film_id, film_id, inventory_id, language_id, original_language_id, staff_id, staff_id, store_id, store_id, last_name, store_id, film_id, title, title, description, manager_staff_id",
	}
	for dbName, db := range dbs {
		if schemas[dbName] == "" {
			continue
		}
		r := informationschema.New(db.Opts...)(db.DB).(metadata.IndexColumnReader)

		result, err := r.IndexColumns("", schemas[dbName], tables[dbName])
		if err != nil {
			log.Fatalf("Could not read %s index columns: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Name)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s index column names, expected:\n  %v, got:\n  %v", dbName, expected[dbName], names)
		}
	}
}

func TestSequences(t *testing.T) {
	schemas := map[string]string{
		"pgsql": "public",
	}
	expected := map[string]string{
		"pgsql": "actor_actor_id_seq, address_address_id_seq, category_category_id_seq, city_city_id_seq, country_country_id_seq, customer_customer_id_seq, film_film_id_seq, inventory_inventory_id_seq, language_language_id_seq, payment_payment_id_seq, rental_rental_id_seq, staff_staff_id_seq, store_store_id_seq",
	}
	for dbName, db := range dbs {
		if schemas[dbName] == "" {
			continue
		}
		r := informationschema.New(db.Opts...)(db.DB).(metadata.SequenceReader)

		result, err := r.Sequences("", schemas[dbName], "")
		if err != nil {
			log.Fatalf("Could not read %s sequences: %v", dbName, err)
		}

		names := []string{}
		for result.Next() {
			names = append(names, result.Get().Name)
		}
		actual := strings.Join(names, ", ")
		if actual != expected[dbName] {
			t.Errorf("Wrong %s sequence names, expected:\n  %v\ngot:\n  %v", dbName, expected[dbName], names)
		}
	}
}
