package dameng

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

// TestMetadataReaderClassifiesMaterializedViews verifies DM8's VIEW compatibility classification.
func TestMetadataReaderClassifiesMaterializedViews(t *testing.T) {
	// db and mock provide deterministic database/sql rows without requiring a live DM8 server.
	db, mock, err := sqlmock.New()
	// A broken SQL mock setup makes the regression test invalid.
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	// Close releases the test-only database handle after all expectations are checked.
	defer db.Close()

	// Base metadata rows reproduce DM8 reporting a materialized view as VIEW in ALL_OBJECTS.
	mock.ExpectQuery(`(?s)FROM all_objects o.*o\.object_type IN`).
		WithArgs("APPDB", "VIEW", "MATERIALIZED VIEW").
		WillReturnRows(sqlmock.NewRows([]string{"table_schem", "table_name", "table_type"}).
			AddRow("APPDB", "ORDINARY_VIEW", "VIEW").
			AddRow("APPDB", "SALES_MV", "VIEW"))
	// The DM system catalog identifies which compatible VIEW row is materialized.
	mock.ExpectQuery(`(?s)FROM SYS\.SYSOBJECTS materialized_view.*schema_object\.NAME LIKE :1`).
		WithArgs("APPDB").
		WillReturnRows(sqlmock.NewRows([]string{"owner", "mview_name"}).AddRow("APPDB", "SALES_MV"))

	// reader is the same registered metadata reader used by interactive usql commands.
	reader := drivers.Available()["dm"].NewMetadataReader(db)
	// tables executes the public metadata contract with both view categories requested.
	tables, err := reader.(metadata.TableReader).Tables(metadata.Filter{
		Schema: "APPDB",
		Types:  []string{"VIEW", "MATERIALIZED VIEW"},
	})
	// Metadata classification must complete before result types can be inspected.
	if err != nil {
		t.Fatalf("Tables() error = %v", err)
	}
	// Close releases the in-memory result set after iteration.
	defer tables.Close()

	// types records the classified type returned for each test object.
	types := map[string]string{}
	// Iterate over every metadata object to verify ordinary and materialized views together.
	for tables.Next() {
		// table is the current classified metadata record.
		table := tables.Get()
		types[table.Name] = table.Type
	}
	// SALES_MV must become the exact type understood by DefaultWriter's \dm command.
	if types["SALES_MV"] != "MATERIALIZED VIEW" {
		t.Fatalf("SALES_MV type = %q, want %q", types["SALES_MV"], "MATERIALIZED VIEW")
	}
	// ORDINARY_VIEW must remain visible to \dv instead of being over-classified.
	if types["ORDINARY_VIEW"] != "VIEW" {
		t.Fatalf("ORDINARY_VIEW type = %q, want %q", types["ORDINARY_VIEW"], "VIEW")
	}
	// All expected catalog calls must occur, guarding the classification query itself.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("metadata SQL expectations were not met: %v", err)
	}
}

// TestMetadataReaderFallsBackForRestrictedCatalog verifies ordinary accounts do not require SYS access.
func TestMetadataReaderFallsBackForRestrictedCatalog(t *testing.T) {
	// db and mock provide deterministic primary and fallback catalog responses.
	db, mock, err := sqlmock.New()
	// A broken SQL mock setup makes the fallback regression test invalid.
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	// Close releases the test-only database handle after all expectations are checked.
	defer db.Close()

	// ALL_OBJECTS reproduces DM8 exposing the requested materialized view as VIEW.
	mock.ExpectQuery(`(?s)FROM all_objects o.*o\.object_type IN`).
		WithArgs("APPDB", "MATERIALIZED VIEW", "VIEW").
		WillReturnRows(sqlmock.NewRows([]string{"table_schem", "table_name", "table_type"}).
			AddRow("APPDB", "SALES_MV", "VIEW"))
	// SYSOBJECTS can be denied to an otherwise valid ordinary database account.
	mock.ExpectQuery(`(?s)FROM SYS\.SYSOBJECTS materialized_view`).
		WithArgs("APPDB").
		WillReturnError(errors.New("SYSOBJECTS 权限不足"))
	// ALL_DEPENDENCIES is DBX's accessible fallback for materialized-view classification.
	mock.ExpectQuery(`(?s)FROM ALL_DEPENDENCIES.*OWNER LIKE :1`).
		WithArgs("APPDB").
		WillReturnRows(sqlmock.NewRows([]string{"owner", "mview_name"}).AddRow("APPDB", "SALES_MV"))

	// reader is the registered production reader used by usql metadata commands.
	reader := drivers.Available()["dm"].NewMetadataReader(db)
	// tables requests only materialized views to prove the fallback still filters correctly.
	tables, err := reader.(metadata.TableReader).Tables(metadata.Filter{
		Schema: "APPDB",
		Types:  []string{"MATERIALIZED VIEW"},
	})
	// Restricted catalog access must not break metadata discovery.
	if err != nil {
		t.Fatalf("Tables() error = %v", err)
	}
	// Close releases the in-memory result set after the assertion.
	defer tables.Close()
	// The fallback must return the requested object before its corrected type can be checked.
	if !tables.Next() {
		t.Fatal("fallback returned no materialized view")
	}
	// table is the one object classified by the accessible catalog fallback.
	table := tables.Get()
	// The fallback must preserve the exact type understood by DefaultWriter's \dm command.
	if table.Type != "MATERIALIZED VIEW" {
		t.Fatalf("fallback table = %#v, want MATERIALIZED VIEW", table)
	}
	// All primary and fallback catalog calls must occur in the intended order.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("fallback SQL expectations were not met: %v", err)
	}
}

// TestMetadataReaderSummarizesPrivileges verifies \dp object and column privilege aggregation.
func TestMetadataReaderSummarizesPrivileges(t *testing.T) {
	// db and mock provide controlled table and privilege catalog rows.
	db, mock, err := sqlmock.New()
	// A broken SQL mock setup makes the privilege regression test invalid.
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	// Close releases the test-only database handle after the assertion completes.
	defer db.Close()

	// reader is the production DM metadata reader selected by usql.
	reader := drivers.Available()["dm"].NewMetadataReader(db)
	// privilegeReader proves the registered reader exposes the same \dp capability as existing databases.
	privilegeReader, ok := reader.(metadata.PrivilegeSummaryReader)
	// The original Oracle-only registration reproduces this review finding as a failed assertion.
	if !ok {
		t.Fatal("DM metadata reader does not implement metadata.PrivilegeSummaryReader")
	}

	// ALL_OBJECTS supplies both privileged and unprivileged tables so \dp lists a complete object set.
	mock.ExpectQuery(`(?s)FROM all_objects o.*o\.object_type IN`).
		WithArgs("APPDB", "TABLE").
		WillReturnRows(sqlmock.NewRows([]string{"table_schem", "table_name", "table_type"}).
			AddRow("APPDB", "CUSTOMER", "TABLE").
			AddRow("APPDB", "EMPTY_TABLE", "TABLE"))
	// ALL_TAB_PRIVS supplies object-level grants using DM8's OWNER column.
	mock.ExpectQuery(`(?s)FROM ALL_TAB_PRIVS.*OWNER LIKE :1`).
		WithArgs("APPDB").
		WillReturnRows(sqlmock.NewRows([]string{"owner", "table_name", "grantee", "grantor", "privilege", "grantable"}).
			AddRow("APPDB", "CUSTOMER", "REPORTER", "SYSDBA", "SELECT", "YES"))
	// ALL_COL_PRIVS supplies column-level grants using DM8's TABLE_SCHEMA column.
	mock.ExpectQuery(`(?s)FROM ALL_COL_PRIVS.*TABLE_SCHEMA LIKE :1`).
		WithArgs("APPDB").
		WillReturnRows(sqlmock.NewRows([]string{"table_schema", "table_name", "column_name", "grantee", "grantor", "privilege", "grantable"}).
			AddRow("APPDB", "CUSTOMER", "EMAIL", "AUDITOR", "SYSDBA", "SELECT", "NO"))

	// summaries executes the public privilege metadata contract used by \dp.
	summaries, err := privilegeReader.PrivilegeSummaries(metadata.Filter{Schema: "APPDB", Types: []string{"TABLE"}})
	// Privilege discovery must succeed before aggregation can be inspected.
	if err != nil {
		t.Fatalf("PrivilegeSummaries() error = %v", err)
	}
	// Close releases the in-memory privilege result set after iteration.
	defer summaries.Close()

	// byName records each returned summary so both populated and empty objects can be checked.
	byName := map[string]*metadata.PrivilegeSummary{}
	// Iterate over every summary to validate aggregation by database object.
	for summaries.Next() {
		// summary is the current table's combined object and column privilege record.
		summary := summaries.Get()
		byName[summary.Name] = summary
	}
	// customer is the populated privilege summary expected from the DM8 catalog rows.
	customer := byName["CUSTOMER"]
	// CUSTOMER must exist before its aggregated grant collections are inspected.
	if customer == nil {
		t.Fatal("CUSTOMER privilege summary is missing")
	}
	// CUSTOMER must contain the one object-level and one column-level grant returned by DM8.
	if len(customer.ObjectPrivileges) != 1 || len(customer.ColumnPrivileges) != 1 {
		t.Fatalf("CUSTOMER privileges = %#v", byName["CUSTOMER"])
	}
	// Grantability must preserve DM8's YES/NO values at both privilege levels.
	if !customer.ObjectPrivileges[0].IsGrantable || customer.ColumnPrivileges[0].IsGrantable {
		t.Fatalf("CUSTOMER grantability = %#v", byName["CUSTOMER"])
	}
	// EMPTY_TABLE must remain present even when no explicit grants exist.
	if byName["EMPTY_TABLE"] == nil {
		t.Fatal("EMPTY_TABLE privilege summary is missing")
	}
	// All expected catalog calls must occur, guarding both DM8 privilege views.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("privilege SQL expectations were not met: %v", err)
	}
}
