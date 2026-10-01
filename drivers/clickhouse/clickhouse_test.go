package clickhouse_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"

	dt "github.com/ory/dockertest/v4"
	"github.com/yookoala/realpath"

	_ "github.com/xo/usql/drivers/csvq"
	_ "github.com/xo/usql/drivers/moderncsqlite"
)

// db is the database connection.
var db struct {
	db  *sql.DB
	res dt.ClosableResource
}

// maxWait is how long to wait for a container to become ready.
const maxWait = time.Minute

func TestMain(m *testing.M) {
	cleanup := flag.Bool("cleanup", true, "cleanup when finished")
	flag.Parse()
	code, err := doMain(m, *cleanup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

func doMain(m *testing.M, cleanup bool) (int, error) {
	dir, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	dir, err = realpath.Realpath(dir)
	if err != nil {
		return 0, err
	}
	ctx := context.Background()
	pool, err := dt.NewPool(ctx, "")
	if err != nil {
		return 0, fmt.Errorf("could not connect to docker: %w", err)
	}
	db.res, err = pool.Run(
		ctx,
		"clickhouse/clickhouse-server",
		dt.WithTag("22.7"),
		dt.WithMounts([]string{filepath.Join(dir, "testdata") + ":/docker-entrypoint-initdb.d"}),
		dt.WithLabels(map[string]string{"usql-test": "clickhouse"}),
	)
	if err != nil {
		return 0, fmt.Errorf("unable to run: %w", err)
	}
	if cleanup {
		defer func() {
			if err := pool.Close(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "error: could not purge resource: %v\n", err)
			}
		}()
	}
	// exponential backoff-retry, because the application in the container
	// might not be ready to accept connections yet
	if err := pool.Retry(ctx, maxWait, func() error {
		port := db.res.GetPort("9000/tcp")
		var err error
		if db.db, err = sql.Open("clickhouse", fmt.Sprintf("clickhouse://127.0.0.1:%s", port)); err != nil {
			return err
		}
		return db.db.Ping()
	}); err != nil {
		return 0, fmt.Errorf("unable to open database: %w", err)
	}
	code := m.Run()
	return code, nil
}

func TestCopy(t *testing.T) {
	// Tests with csvq source DB. That driver doesn't support ScanType()
	for _, test := range []struct {
		destTableSpec string
		skip          string
	}{
		{destTableSpec: "copy_test.dest"},
		{destTableSpec: "copy_test.dest(StringCol, NumCol)"},
		{destTableSpec: "insert into copy_test.dest VALUES (?, ?)"},
		{
			destTableSpec: "insert into copy_test.dest values(?, ?)",
			// clickhouse-go strips the parameter list with the regexp
			// `\sVALUES\s.*$`, which is case sensitive and needs whitespace
			// after VALUES. A lowercase "values(" therefore reaches the server
			// with its placeholders intact and is rejected. Remove this skip
			// once clickhouse-go accepts the lowercase form.
			skip: "clickhouse-go only truncates an uppercase VALUES followed by whitespace",
		},
	} {
		t.Run("csvq_"+test.destTableSpec, func(t *testing.T) {
			if test.skip != "" {
				t.Skip(test.skip)
			}
			testCopy(t, test.destTableSpec, "csvq:.")
		})
	}
	// Test with a driver that supports ScanType()
	t.Run("sqlite", func(t *testing.T) {
		testCopy(t, "copy_test.dest", "moderncsqlite://:memory:")
	})
}

func testCopy(t *testing.T, destTableSpec string, sourceDbUrlStr string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := db.db.ExecContext(ctx, "truncate table copy_test.dest")
	if err != nil {
		t.Fatalf("could not truncate copy_test table: %v", err)
	}
	// Prepare copy destination URL
	port := db.res.GetPort("9000/tcp")
	dbUrlStr := fmt.Sprintf("clickhouse://127.0.0.1:%s", port)
	dbUrl, err := dburl.Parse(dbUrlStr)
	if err != nil {
		t.Fatalf("could not parse clickhouse url %s: %v", dbUrlStr, err)
	}
	// Prepare source data
	sourceDbUrl, err := dburl.Parse(sourceDbUrlStr)
	if err != nil {
		t.Fatalf("could not parse source DB url %s: %v", sourceDbUrlStr, err)
	}
	sourceDb, err := drivers.Open(ctx, sourceDbUrl, nil, nil)
	if err != nil {
		t.Fatalf("could not open sourceDb: %v", err)
	}
	defer sourceDb.Close()
	rows, err := sourceDb.QueryContext(ctx, "select 'string', 1")
	if err != nil {
		t.Fatalf("could not retrieve source rows: %v", err)
	}
	// Do Copy, ignoring copied rows count because clickhouse driver doesn't report RowsAffected
	_, err = drivers.Copy(ctx, dbUrl, nil, nil, rows, destTableSpec)
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}
	rows, err = db.db.QueryContext(ctx, "select StringCol, NumCol from copy_test.dest")
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	defer rows.Close()
	var copiedString string
	var copiedNum int
	if !rows.Next() {
		t.Fatalf("nothing copied")
	}
	err = rows.Scan(&copiedString, &copiedNum)
	if err != nil {
		t.Fatalf("could not read copied data: %v", err)
	}
	if copiedString != "string" || copiedNum != 1 {
		t.Fatalf("copied data differs: %s != string, %d != 1", copiedString, copiedNum)
	}
}
