package pgshared

import (
	"testing"
	"time"
)

// TestFormatClock checks the text of a time of day against psql 18.6, which
// prints these values for '06:35:11'::time, '20:47:40.0612'::time,
// '24:00:00'::time and the same times with the offsets of timetz.
func TestFormatClock(t *testing.T) {
	tests := []struct {
		t    time.Time
		zone bool
		exp  string
	}{
		{time.Date(0, 1, 1, 6, 35, 11, 0, time.UTC), false, "06:35:11"},
		{time.Date(0, 1, 1, 20, 47, 40, 61200000, time.UTC), false, "20:47:40.0612"},
		{time.Date(0, 1, 2, 0, 0, 0, 0, time.UTC), false, "24:00:00"},
		{time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC), false, "00:00:00"},
		{time.Date(0, 1, 1, 23, 48, 40, 0, time.FixedZone("", 5*3600+30*60)), true, "23:48:40+05:30"},
		{time.Date(0, 1, 1, 12, 42, 39, 644483000, time.FixedZone("", -8*3600)), true, "12:42:39.644483-08"},
		{time.Date(0, 1, 1, 5, 16, 48, 0, time.FixedZone("", -(15*3600+59*60))), true, "05:16:48-15:59"},
		{time.Date(0, 1, 1, 1, 0, 0, 0, time.FixedZone("", 3600+2*60+3)), true, "01:00:00+01:02:03"},
		{time.Date(0, 1, 1, 1, 0, 0, 0, time.UTC), true, "01:00:00+00"},
	}
	for _, test := range tests {
		if s := formatClock(test.t, test.zone); s != test.exp {
			t.Errorf("formatClock(%v, %t) = %q, want %q", test.t, test.zone, s, test.exp)
		}
	}
}

// TestScanClock checks that the text that pgx gives is kept, and that a type
// that is not a time of day is an error.
func TestScanClock(t *testing.T) {
	var c clock
	if err := c.Scan("10:00:00"); err != nil || c.String() != "10:00:00" {
		t.Errorf("Scan(string) = %q, %v", c, err)
	}
	var z clockZone
	if err := z.Scan([]byte("10:00:00+05")); err != nil || z.String() != "10:00:00+05" {
		t.Errorf("Scan([]byte) = %q, %v", z, err)
	}
	if err := c.Scan(42); err == nil {
		t.Error("Scan(int) returned no error")
	}
}
