package sqlite3

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/xo/usql/drivers/metadata"
)

var (
	db     *sql.DB
	reader *metaReader
)

func TestMain(m *testing.M) {
	err := createDb("testdata", "sakila.db")
	if err != nil {
		log.Fatalf("Could not prepare the database: %s", err)
	}
	db, err = sql.Open("sqlite3", "testdata/sakila.db")
	if err != nil {
		log.Fatalf("Could not open the database: %s", err)
	}
	reader = &metaReader{LoggingReader: metadata.NewLoggingReader(db)}

	code := m.Run()
	os.Exit(code)
}

func createDb(location, name string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	tar, err := archive.TarWithOptions("../metadata/testdata/docker", &archive.TarOptions{})
	if err != nil {
		return err
	}
	baseImage := "centos:7"
	schemaURL := "https://raw.githubusercontent.com/jOOQ/jOOQ/main/jOOQ-examples/Sakila/sqlite-sakila-db/sqlite-sakila-schema.sql"
	target := "/schema"
	buildOptions := types.ImageBuildOptions{
		Tags: []string{"usql-sqlite"},
		BuildArgs: map[string]*string{
			"BASE_IMAGE": &baseImage,
			"SCHEMA_URL": &schemaURL,
			"TARGET":     &target,
		},
	}

	res, err := cli.ImageBuild(ctx, tar, buildOptions)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	u, err := user.Current()
	if err != nil {
		return err
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:           "usql-sqlite",
		Cmd:             []string{"bash", "-xc", "sqlite3 -batch -echo -init /schema/sqlite-sakila-schema.sql /data/" + name},
		User:            u.Uid + ":" + u.Gid,
		NetworkDisabled: true,
	}, &container.HostConfig{
		Binds: []string{
			path.Join(cwd, location) + ":/data",
		},
	}, nil, nil, "")
	if err != nil {
		return err
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case status := <-statusCh:
		fmt.Println(status.StatusCode, status.Error)
	}

	//out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	//if err != nil {
	//	return err
	//}

	//_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	//if err != nil {
	//	return err
	//}

	return cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
}

func TestSchemas(t *testing.T) {
	result, err := reader.Schemas("", "")
	if err != nil {
		log.Fatalf("Could not read schemas: %v", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Schema)
	}
	actual := strings.Join(names, ", ")
	expected := "main"
	if actual != expected {
		t.Errorf("Wrong schema names, expected:\n  %v\ngot:\n  %v", expected, names)
	}
}

func TestTables(t *testing.T) {
	result, err := reader.Tables("", "", "", []string{"BASE TABLE", "TABLE", "VIEW"})
	if err != nil {
		log.Fatalf("Could not read tables: %v", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	expected := "actor, address, category, city, country, customer, film, film_actor, film_category, film_text, inventory, language, payment, rental, staff, store, customer_list, film_list, sales_by_film_category, sales_by_store, staff_list"
	if actual != expected {
		t.Errorf("Wrong table names, expected:\n  %v\ngot:\n  %v", expected, names)
	}
}

func TestColumns(t *testing.T) {
	result, err := reader.Columns("", "", "film%")
	if err != nil {
		log.Fatalf("Could not read columns: %v", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	expected := "film_id, title, description, release_year, language_id, original_language_id, rental_duration, rental_rate, length, replacement_cost, rating, special_features, last_update, actor_id, film_id, last_update, film_id, category_id, last_update, film_id, title, description, FID, title, description, category, price, length, rating, actors"
	if actual != expected {
		t.Errorf("Wrong column names, expected:\n  %v, got:\n  %v", expected, names)
	}
}

func TestFunctions(t *testing.T) {
	result, err := reader.Functions("", "", "", []string{})
	if err != nil {
		log.Fatalf("Could not read functions: %v", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	expected := "abs, auth_enabled, auth_user_add, auth_user_change, auth_user_delete, authenticate, avg, changes, char, coalesce, count, count, cume_dist, current_date, current_time, current_timestamp, date, datetime, dense_rank, first_value, fts3_tokenizer, fts3_tokenizer, glob, group_concat, group_concat, hex, ifnull, instr, julianday, lag, lag, lag, last_insert_rowid, last_value, lead, lead, lead, length, like, like, likelihood, likely, load_extension, load_extension, lower, ltrim, ltrim, match, matchinfo, matchinfo, max, max, min, min, nth_value, ntile, nullif, offsets, optimize, percent_rank, printf, quote, random, randomblob, rank, replace, round, round, row_number, rtreecheck, rtreedepth, rtreenode, rtrim, rtrim, snippet, sqlite_compileoption_get, sqlite_compileoption_used, sqlite_log, sqlite_source_id, sqlite_version, strftime, substr, substr, sum, time, total, total_changes, trim, trim, typeof, unicode, unlikely, upper, zeroblob"
	if actual != expected {
		t.Errorf("Wrong function names, expected:\n  %v\ngot:\n  %v", expected, names)
	}
}

func TestIndexes(t *testing.T) {
	result, err := reader.Indexes("", "", "", "")
	if err != nil {
		log.Fatalf("Could not read indexes: %v", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Table+"."+result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	expected := "actor.idx_actor_last_name, actor.sqlite_autoindex_actor_1, address.idx_fk_city_id, address.sqlite_autoindex_address_1, category.sqlite_autoindex_category_1, city.idx_fk_country_id, city.sqlite_autoindex_city_1, country.sqlite_autoindex_country_1, customer.idx_customer_last_name, customer.idx_customer_fk_address_id, customer.idx_customer_fk_store_id, customer.sqlite_autoindex_customer_1, film.idx_fk_original_language_id, film.idx_fk_language_id, film.sqlite_autoindex_film_1, film_actor.idx_fk_film_actor_actor, film_actor.idx_fk_film_actor_film, film_actor.sqlite_autoindex_film_actor_1, film_category.idx_fk_film_category_category, film_category.idx_fk_film_category_film, film_category.sqlite_autoindex_film_category_1, film_text.sqlite_autoindex_film_text_1, inventory.idx_fk_film_id_store_id, inventory.idx_fk_film_id, inventory.sqlite_autoindex_inventory_1, language.sqlite_autoindex_language_1, payment.idx_fk_customer_id, payment.idx_fk_staff_id, payment.sqlite_autoindex_payment_1, rental.idx_rental_uq, rental.idx_rental_fk_staff_id, rental.idx_rental_fk_customer_id, rental.idx_rental_fk_inventory_id, rental.sqlite_autoindex_rental_1, staff.idx_fk_staff_address_id, staff.idx_fk_staff_store_id, staff.sqlite_autoindex_staff_1, store.idx_fk_store_address, store.idx_store_fk_manager_staff_id, store.sqlite_autoindex_store_1"
	if actual != expected {
		t.Errorf("Wrong index names, expected:\n  %v\ngot:\n  %v", expected, names)
	}
}

func TestIndexColumns(t *testing.T) {
	result, err := reader.IndexColumns("", "", "", "idx%")
	if err != nil {
		log.Fatalf("Could not read index columns: %v", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	expected := "last_name, city_id, country_id, last_name, address_id, store_id, original_language_id, language_id, actor_id, film_id, category_id, film_id, store_id, film_id, film_id, customer_id, staff_id, rental_date, inventory_id, customer_id, staff_id, customer_id, inventory_id, address_id, store_id, address_id, manager_staff_id"
	if actual != expected {
		t.Errorf("Wrong index column names, expected:\n  %v, got:\n  %v", expected, names)
	}
}
