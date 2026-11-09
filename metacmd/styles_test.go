package metacmd

import (
	"bytes"
	"slices"
	"strings"
	"testing"
)

// TestStyles lists the names in columns that read down, each name once, and
// names the style in use.
func TestStyles(t *testing.T) {
	names := []string{"a", "bb", "c", "dddd", "e"}
	var b bytes.Buffer
	if err := Styles(&b, names, "bb"); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(b.String(), "\n")
	if lines[0] != "Syntax highlighting styles:" {
		t.Errorf("title = %q", lines[0])
	}
	// five short names fit in one row of columns 6 wide
	if want := "  a     bb    c     dddd  e"; lines[1] != want {
		t.Errorf("row = %q, want %q", lines[1], want)
	}
	if !strings.Contains(b.String(), "The style in use is bb.\n") {
		t.Errorf("no line names the style in use:\n%s", b.String())
	}
	// 40 names of 27 letters fill three columns of 14 rows, read down
	names = nil
	for i := range 40 {
		names = append(names, strings.Repeat(string(rune('a'+i%26)), 26)+string(rune('0'+i/26)))
	}
	slices.Sort(names)
	b.Reset()
	if err := Styles(&b, names, "x"); err != nil {
		t.Fatal(err)
	}
	lines = strings.Split(b.String(), "\n")
	if got := strings.Fields(lines[1]); !slices.Equal(got, []string{names[0], names[14], names[28]}) {
		t.Errorf("first row = %q, want the names 0, 14 and 28", got)
	}
	var all []string
	for _, line := range lines[1:15] {
		all = append(all, strings.Fields(line)...)
	}
	slices.Sort(all)
	if !slices.Equal(all, names) {
		t.Errorf("the rows hold %d names, want each of the %d once", len(all), len(names))
	}
}
