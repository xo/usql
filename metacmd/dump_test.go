package metacmd

import (
	"strings"
	"testing"

	"github.com/xo/usql/dump"
)

// TestDumpHelpNamesEveryOption checks that the help of \dump names each
// option of a dump, with its value, as -NAME. The help is a doc comment, and
// the options are a table in the package dump, so a new option must be added
// to both.
func TestDumpHelpNamesEveryOption(t *testing.T) {
	var help string
	for i := range sections {
		for _, desc := range descs[i] {
			if desc.Name == "dump" {
				help = desc.Desc
			}
		}
	}
	if help == "" {
		t.Fatal(`\dump has no help`)
	}
	for _, f := range dump.Flags {
		s := "-" + f.Name
		if f.Value != "" {
			s += " " + f.Value
		}
		if !strings.Contains(help, s+",") && !strings.Contains(help, s+")") {
			t.Errorf(`the help of \dump does not name %s: %s`, s, help)
		}
	}
}

// TestDumpOption reads the options of \dump.
func TestDumpOption(t *testing.T) {
	for s, want := range map[string]string{
		"-table":        "table",
		"--schema-only": "schema-only",
		"-":             "",
		"file.sql":      "",
		"|gzip":         "",
	} {
		name, ok := dumpOption(s)
		if name != want || ok != (want != "") {
			t.Errorf("dumpOption(%q) = %q, %t, want %q", s, name, ok, want)
		}
	}
}
