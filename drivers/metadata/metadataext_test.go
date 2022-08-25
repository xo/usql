package metadata_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/xo/usql/drivers/metadata"
	"github.com/xo/usql/drivers/metadata/postgres"
	_ "github.com/xo/usql/drivers/postgres"
	"github.com/xo/usql/drivers/sqlite3/sqshared"
)

type error struct {
	strError  string
	msg       string
	sql       string
	codeError int32
	trace     string
}

func (err *error) GetTrace() string {
	return err.trace
}
func (err *error) Error() string {
	return fmt.Sprintf("H2 SQL Exception: [%s] %s", err.strError, err.msg)
}

// Test that postgres conversion of internal to external datatype works
// To setup, import the northwind sql in testdata/northwind/nortwind.sql into a database called northwind
// also don't forget to add user and password.
func TestGetPgExternalDataType(t *testing.T) {

	db, err := sql.Open("postgres", "postgres://<user>:<password>@localhost:5432/northwind")
	assert.NoError(t, err)
	assert.NotEmpty(t, db)

	reader := postgres.NewReader()(db).(metadata.BasicReader)
	assert.NotEmpty(t, reader)

	columnSet, err := reader.Columns(metadata.Filter{Catalog: "northwind", Schema: "public", Types: []string{"TABLE"}})
	assert.NoError(t, err)
	assert.NotEmpty(t, columnSet)

	for columnSet.Next() {
		assert.NotEmpty(t, columnSet.Get().ExternalDataType, fmt.Sprintf("Should not be empty: %s", columnSet.Get().DataType))
		assert.Equal(t, postgres.Mapping[columnSet.Get().InternalDataType], columnSet.Get().ExternalDataType)
	}

	err = db.Close()
	assert.NoError(t, err)
}

// Test that sqlite conversion of internal to external datatype works
// Does not use the information schema functionality since it seems to be bypassed in general for sqlite
func TestGetSqliteExternalDataType(t *testing.T) {
	path, err := os.Getwd()
	assert.NoError(t, err)
	url := fmt.Sprintf("file:%s/../../testdata/northwind/northwind.sqlite", path)

	db, err := sql.Open("sqlite3", url)
	assert.NoError(t, err)
	assert.NotEmpty(t, db)

	reader := &sqshared.MetadataReader{LoggingReader: metadata.NewLoggingReader(db)}
	assert.NotEmpty(t, reader)

	columnSet, err := reader.Columns(metadata.Filter{Parent: "Customer", Types: []string{"TABLE"}, WithSystem: false, OnlyVisible: true})
	assert.NoError(t, err)
	assert.NotEmpty(t, columnSet)

	for columnSet.Next() {
		assert.NotEmpty(t, columnSet.Get().ExternalDataType, fmt.Sprintf("Should not be empty: %s", columnSet.Get().DataType))
		assert.Equal(t, sqshared.Mapping[columnSet.Get().InternalDataType], columnSet.Get().ExternalDataType)
	}

	err = db.Close()
	assert.NoError(t, err)
}
