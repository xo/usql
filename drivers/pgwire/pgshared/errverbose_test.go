package pgshared

import "testing"

// TestVerbose checks the error of \errverbose against what psql 18 writes for
// the same error.
func TestVerbose(t *testing.T) {
	tests := []struct {
		name  string
		e     ErrorFields
		query string
		exp   string
	}{
		{
			"position",
			ErrorFields{Severity: "ERROR", Code: "42P01", Message: `relation "nope" does not exist`, Position: 15, File: "parse_relation.c", Line: 1449, Routine: "parserOpenTable"},
			"select * from nope;",
			"ERROR:  42P01: relation \"nope\" does not exist\nLINE 1: select * from nope;\n                      ^\nLOCATION:  parserOpenTable, parse_relation.c:1449",
		},
		{
			"second line",
			ErrorFields{Severity: "ERROR", Code: "42703", Message: `column "b" does not exist`, Position: 16},
			"select a,\n  é, b from t",
			"ERROR:  42703: column \"b\" does not exist\nLINE 2:   é, b from t\n             ^",
		},
		{
			"fields",
			ErrorFields{Severity: "ERROR", Code: "23505", Message: `duplicate key value violates unique constraint "t_pkey"`, Detail: "Key (id)=(1) already exists.", SchemaName: "public", TableName: "t", ConstraintName: "t_pkey", File: "nbtinsert.c", Line: 666, Routine: "_bt_check_unique"},
			"insert into t values (1)",
			"ERROR:  23505: duplicate key value violates unique constraint \"t_pkey\"\nDETAIL:  Key (id)=(1) already exists.\nSCHEMA NAME:  public\nTABLE NAME:  t\nCONSTRAINT NAME:  t_pkey\nLOCATION:  _bt_check_unique, nbtinsert.c:666",
		},
	}
	for _, test := range tests {
		if got := Verbose(test.e, test.query); got != test.exp {
			t.Errorf("%s:\n got: %q\nwant: %q", test.name, got, test.exp)
		}
	}
}
