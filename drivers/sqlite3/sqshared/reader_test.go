package sqshared

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3" // DRIVER: sqlite3
	"github.com/xo/usql/drivers/metadata"
)

var (
	db     *sql.DB
	reader *MetadataReader
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
	reader = &MetadataReader{LoggingReader: metadata.NewLoggingReader(db)}

	code := m.Run()
	os.Exit(code)
}

// schemaURL is the sakila schema used to build the test database.
const schemaURL = "https://raw.githubusercontent.com/jOOQ/sakila/main/sqlite-sakila-db/sqlite-sakila-schema.sql"

// createDb builds the sakila test database directly with the sqlite3 driver.
// It needs no container, so the test runs anywhere usql builds. The schema is
// cached next to the database so repeat runs need no network.
func createDb(location, name string) error {
	if err := os.MkdirAll(location, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", location, err)
	}
	schema, err := loadSchema(filepath.Join(location, "sqlite-sakila-schema.sql"))
	if err != nil {
		return err
	}
	dbPath := filepath.Join(location, name)
	if err := os.Remove(dbPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("removing %s: %w", dbPath, err)
	}
	sakila, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", dbPath, err)
	}
	defer sakila.Close()
	if _, err := sakila.Exec(string(schema)); err != nil {
		return fmt.Errorf("loading schema into %s: %w", dbPath, err)
	}
	return nil
}

// loadSchema returns the cached schema, retrieving it once when absent.
func loadSchema(cache string) ([]byte, error) {
	switch buf, err := os.ReadFile(cache); {
	case err == nil:
		return buf, nil
	case !errors.Is(err, fs.ErrNotExist):
		return nil, fmt.Errorf("reading %s: %w", cache, err)
	}
	req, err := http.NewRequest(http.MethodGet, schemaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for %s: %w", schemaURL, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %w", schemaURL, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("retrieving %s: status %d", schemaURL, res.StatusCode)
	}
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", schemaURL, err)
	}
	if err := os.WriteFile(cache, buf, 0o644); err != nil {
		return nil, fmt.Errorf("writing %s: %w", cache, err)
	}
	return buf, nil
}

func TestSchemas(t *testing.T) {
	result, err := reader.Schemas(metadata.Filter{})
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
	result, err := reader.Tables(metadata.Filter{Types: []string{"BASE TABLE", "TABLE", "VIEW"}})
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
	result, err := reader.Columns(metadata.Filter{Parent: "film%"})
	if err != nil {
		log.Fatalf("Could not read columns: %v", err)
	}

	names := []string{}
	for result.Next() {
		names = append(names, result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	expected := "description, film_id, language_id, last_update, length, original_language_id, rating, release_year, rental_duration, rental_rate, replacement_cost, special_features, title, actor_id, film_id, last_update, category_id, film_id, last_update, description, film_id, title, FID, actors, category, description, length, price, rating, title"
	if actual != expected {
		t.Errorf("Wrong column names, expected:\n  %v, got:\n  %v", expected, names)
	}
}

func TestFunctions(t *testing.T) {
	result, err := reader.Functions(metadata.Filter{})
	if err != nil {
		t.Fatalf("Could not read functions: %v", err)
	}

	names := map[string]bool{}
	for result.Next() {
		names[result.Get().Name] = true
	}
	// The full set of built-in functions belongs to the SQLite library that is
	// linked in, so it changes with the SQLite version and with the sqlite_fts5,
	// sqlite_json1 and sqlite_math_functions build tags. Assert that the reader
	// returns the functions every build has instead of matching an exact list.
	for _, want := range []string{
		"abs", "changes", "coalesce", "count", "glob", "group_concat", "hex",
		"ifnull", "instr", "last_insert_rowid", "length", "like", "lower",
		"ltrim", "max", "min", "nullif", "printf", "quote", "random", "replace",
		"round", "rtrim", "sqlite_source_id", "sqlite_version", "substr", "sum",
		"trim", "typeof", "upper", "zeroblob",
	} {
		if !names[want] {
			t.Errorf("Missing function %q", want)
		}
	}
	if len(names) < 50 {
		t.Errorf("Expected at least 50 functions, got %d", len(names))
	}
}

func TestIndexes(t *testing.T) {
	result, err := reader.Indexes(metadata.Filter{})
	if err != nil {
		t.Fatalf("Could not read indexes: %v", err)
	}

	names := []string{}
	for result.Next() {
		// Skip the indexes SQLite creates for primary keys. Whether one exists
		// depends on the declared column type, because an INTEGER PRIMARY KEY is
		// an alias for the rowid and gets no index, so the set changes whenever
		// the sakila schema changes its column types upstream.
		if strings.HasPrefix(result.Get().Name, "sqlite_autoindex_") {
			continue
		}
		names = append(names, result.Get().Table+"."+result.Get().Name)
	}
	actual := strings.Join(names, ", ")
	expected := "actor.idx_actor_last_name, address.idx_fk_city_id, city.idx_fk_country_id, customer.idx_customer_last_name, customer.idx_customer_fk_address_id, customer.idx_customer_fk_store_id, film.idx_fk_original_language_id, film.idx_fk_language_id, film_actor.idx_fk_film_actor_actor, film_actor.idx_fk_film_actor_film, film_category.idx_fk_film_category_category, film_category.idx_fk_film_category_film, inventory.idx_fk_film_id_store_id, inventory.idx_fk_film_id, payment.idx_fk_customer_id, payment.idx_fk_staff_id, rental.idx_rental_uq, rental.idx_rental_fk_staff_id, rental.idx_rental_fk_customer_id, rental.idx_rental_fk_inventory_id, staff.idx_fk_staff_address_id, staff.idx_fk_staff_store_id, store.idx_fk_store_address, store.idx_store_fk_manager_staff_id"
	if actual != expected {
		t.Errorf("Wrong index names, expected:\n  %v\ngot:\n  %v", expected, names)
	}
}

func TestIndexColumns(t *testing.T) {
	result, err := reader.IndexColumns(metadata.Filter{Name: "idx%"})
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
