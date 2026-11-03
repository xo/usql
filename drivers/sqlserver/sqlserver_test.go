package sqlserver

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/xo/tblfmt"
)

// TestUniqueIdentifier checks that a uniqueidentifier prints in lower case,
// as a uuid of PostgreSQL does, with its bytes in the order that SQL Server's
// own text shows.
func TestUniqueIdentifier(t *testing.T) {
	t.Parallel()
	// the bytes that SQL Server sends for 47EB423E-F36B-1410-8944-00F8E7EB68FD
	raw := []byte{0x3e, 0x42, 0xeb, 0x47, 0x6b, 0xf3, 0x10, 0x14, 0x89, 0x44, 0x00, 0xf8, 0xe7, 0xeb, 0x68, 0xfd}
	const exp = "47eb423e-f36b-1410-8944-00f8e7eb68fd"
	var u uniqueIdentifier
	if err := u.Scan(raw); err != nil {
		t.Fatal(err)
	}
	if got := u.String(); got != exp {
		t.Errorf("expected %s, got: %s", exp, got)
	}
	// tblfmt prints the value, and a null, through sql.Null
	var n sql.Null[uniqueIdentifier]
	if err := n.Scan(raw); err != nil {
		t.Fatal(err)
	}
	vals, err := tblfmt.NewEscapeFormatter().Format([]any{&u, &n, &sql.Null[uniqueIdentifier]{}})
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range vals[:2] {
		if got := string(bytes.TrimSpace(v.Buf)); got != exp {
			t.Errorf("value %d: expected %s, got: %s", i, exp, got)
		}
	}
	if vals[2] != nil {
		t.Errorf("expected a null, got: %q", vals[2].Buf)
	}
}
