package sqlserver_test

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"testing"

	dt "github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/drivers/sqlserver"
)

type Database struct {
	BuildArgs    []dc.BuildArg
	RunOptions   *dt.RunOptions
	Exec         []string
	Driver       string
	URL          string
	ReadinessURL string
	DockerPort   string
	Resource     *dt.Resource
	DB           *sql.DB
	Opts         []metadata.ReaderOption
	Reader       metadata.BasicReader
}

var dbName string = "sakila"

const pw = "yourStrong123_Password"

var db = Database{
	BuildArgs: []dc.BuildArg{
		{Name: "BASE_IMAGE", Value: "mcr.microsoft.com/mssql/server:2019-latest"},
		{Name: "SCHEMA_URL", Value: "https://raw.githubusercontent.com/jOOQ/sakila/main/sql-server-sakila-db/sql-server-sakila-schema.sql"},
		{Name: "TARGET", Value: "/schema"},
		{Name: "USER", Value: "mssql:0"},
	},
	RunOptions: &dt.RunOptions{
		Name: "usql-sqlserver",
		Env:  []string{"ACCEPT_EULA=Y", "SA_PASSWORD=" + pw},
	},
	Exec:         []string{"/opt/mssql-tools/bin/sqlcmd", "-S", "localhost", "-U", "sa", "-P", pw, "-d", "master", "-i", "/schema/sql-server-sakila-schema.sql"},
	Driver:       "sqlserver",
	URL:          "sqlserver://sa:" + url.QueryEscape(pw) + "@127.0.0.1:%s?database=" + dbName,
	ReadinessURL: "sqlserver://sa:" + url.QueryEscape(pw) + "@127.0.0.1:%s",
	DockerPort:   "1433/tcp",
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
			ContextDir: "../testdata/docker",
			BuildArgs:  db.BuildArgs,
		}
		db.Resource, err = pool.BuildAndRunWithBuildOptions(buildOpts, db.RunOptions)
		if err != nil {
			log.Fatal("Could not start resource: ", err)
		}
	}

	url := db.URL
	if db.ReadinessURL != "" {
		url = db.ReadinessURL
	}
	port := db.Resource.GetPort(db.DockerPort)
	if db.DB, err = waitForDbConnection(db.Driver, pool, url, port); err != nil {
		log.Fatal("Timed out waiting for db: ", err)
	}

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

	// Reconnect with actual URL if a separate URL for readiness checking was used
	if db.ReadinessURL != "" {
		if db.DB, err = waitForDbConnection(db.Driver, pool, db.URL, port); err != nil {
			log.Fatal("Timed out waiting for db: ", err)
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

func waitForDbConnection(driver string, pool *dt.Pool, url string, port string) (*sql.DB, error) {
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	var db *sql.DB
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open(driver, fmt.Sprintf(url, port))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		return nil, err
	}
	return db, nil
}

func TestColumns(t *testing.T) {
	// Only testing sqlserver specific datatype formatting.
	// The rest of the functionality is covered by informationschema/metadata_test.go:TestColumns
	type test struct {
		typeDef string
		want    string
	}
	schema := "dbo"
	table := "test_dtypes"
	tests := []test{
		{typeDef: "bigint", want: "bigint"},
		{typeDef: "numeric", want: "numeric"},
		{typeDef: "numeric(4,2)", want: "numeric(4,2)"},
		{typeDef: "numeric(18,0)", want: "numeric"},
		{typeDef: "decimal", want: "decimal"},
		{typeDef: "decimal(4,2)", want: "decimal(4,2)"},
		{typeDef: "decimal(18,0)", want: "decimal"},
		{typeDef: "bit", want: "bit"},
		{typeDef: "smallint", want: "smallint"},
		{typeDef: "smallmoney", want: "smallmoney"},
		{typeDef: "int", want: "int"},
		{typeDef: "tinyint", want: "tinyint"},
		{typeDef: "money", want: "money"},
		{typeDef: "float", want: "float"},
		{typeDef: "float(11)", want: "real"},
		{typeDef: "float(30)", want: "float"},
		{typeDef: "real", want: "real"},
		{typeDef: "date", want: "date"},
		{typeDef: "datetimeoffset", want: "datetimeoffset"},
		{typeDef: "datetimeoffset(5)", want: "datetimeoffset(5)"},
		{typeDef: "datetimeoffset(7)", want: "datetimeoffset"},
		{typeDef: "datetime2", want: "datetime2"},
		{typeDef: "datetime2(5)", want: "datetime2(5)"},
		{typeDef: "datetime2(7)", want: "datetime2"},
		{typeDef: "smalldatetime", want: "smalldatetime"},
		{typeDef: "datetime", want: "datetime"},
		{typeDef: "time", want: "time"},
		{typeDef: "time(5)", want: "time(5)"},
		{typeDef: "time(7)", want: "time"},
		{typeDef: "char", want: "char"},
		{typeDef: "char(3)", want: "char(3)"},
		{typeDef: "char(1)", want: "char"},
		{typeDef: "varchar", want: "varchar"},
		{typeDef: "varchar(12)", want: "varchar(12)"},
		{typeDef: "varchar(1)", want: "varchar"},
		{typeDef: "varchar(max)", want: "varchar(max)"},
		{typeDef: "text", want: "text"},
		{typeDef: "nchar", want: "nchar"},
		{typeDef: "nchar(2)", want: "nchar(2)"},
		{typeDef: "nchar(1)", want: "nchar"},
		{typeDef: "nvarchar", want: "nvarchar"},
		{typeDef: "nvarchar(12)", want: "nvarchar(12)"},
		{typeDef: "nvarchar(1)", want: "nvarchar"},
		{typeDef: "nvarchar(max)", want: "nvarchar(max)"},
		{typeDef: "ntext", want: "ntext"},
		{typeDef: "binary", want: "binary"},
		{typeDef: "binary(12)", want: "binary(12)"},
		{typeDef: "binary(1)", want: "binary"},
		{typeDef: "varbinary", want: "varbinary"},
		{typeDef: "varbinary(12)", want: "varbinary(12)"},
		{typeDef: "varbinary(1)", want: "varbinary"},
		{typeDef: "varbinary(max)", want: "varbinary(max)"},
		{typeDef: "image", want: "image"},
		{typeDef: "rowversion", want: "timestamp"},
		{typeDef: "hierarchyid", want: "hierarchyid"},
		{typeDef: "uniqueidentifier", want: "uniqueidentifier"},
		{typeDef: "sql_variant", want: "sql_variant"},
		{typeDef: "xml", want: "xml"},
		{typeDef: "geometry", want: "geometry"},
		{typeDef: "geography", want: "geography"},
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
	r := sqlserver.NewReader(db.DB).(metadata.ColumnReader)
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
