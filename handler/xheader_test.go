package handler

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xo/usql/env"
)

// TestXheaderWidth writes the header line of an expanded record as psql 18.6
// writes it for each value of xheader_width. The outputs were recorded with
// psql 18.6 on 2026-11-09.
func TestXheaderWidth(t *testing.T) {
	h, stdout, stderr := newConfigHandler(t)
	if err := h.Open(context.Background(), "moderncsqlite:"+filepath.Join(t.TempDir(), "x.db")); err != nil {
		t.Fatal(err)
	}
	// Windows cannot remove the file of the database while it is open
	defer h.Close()
	t.Cleanup(func() {
		for _, opt := range []string{"print.expanded=off", "print.border=1", "print.columns=0", "print.xheader_width=full"} {
			_ = env.Vars().SetOption(opt)
		}
	})
	const query = "select 1 as a, 'a much longer value that is wide' as b;"
	const rows = "a | 1\nb | a much longer value that is wide\n"
	const rows0 = "a 1\nb a much longer value that is wide\n"
	const rows2 = "| a | 1                                |\n| b | a much longer value that is wide |\n"
	for _, test := range []struct {
		border, width, columns string
		want                   string
	}{
		{"1", "full", "0", "-[ RECORD 1 ]-----------------------\n" + rows},
		{"0", "full", "0", "* Record 1                       \n" + rows0},
		{"2", "full", "0", "+-[ RECORD 1 ]-------------------------+\n" + rows2 + "+---+----------------------------------+\n"},
		{"1", "column", "0", "-[ RECORD 1 ]\n" + rows},
		{"0", "column", "0", "* Record 1\n" + rows0},
		{"2", "column", "0", "+-[ RECORD 1 ]\n" + rows2 + "+---+\n"},
		{"1", "page", "30", "-[ RECORD 1 ]-----------------\n" + rows},
		{"1", "20", "0", "-[ RECORD 1 ]-------\n" + rows},
		{"1", "5", "0", "-[ RECORD 1 ]\n" + rows},
		{"1", "40", "0", "-[ RECORD 1 ]-----------------------\n" + rows},
		{"1", "-3", "0", "-[ RECORD 1 ]-----------------------\n" + rows},
		{"0", "20", "0", "* Record 1          \n" + rows0},
		{"2", "20", "0", "+-[ RECORD 1 ]-----+\n" + rows2 + "+---+--------------+\n"},
	} {
		for _, opt := range []string{"print.expanded=on", "print.border=" + test.border, "print.columns=" + test.columns, "print.xheader_width=" + test.width} {
			if err := env.Vars().SetOption(opt); err != nil {
				t.Fatalf("%s: %v", opt, err)
			}
		}
		stdout.Reset()
		if err := h.RunCommand(query); err != nil {
			t.Fatalf("%v\n%s", err, stderr)
		}
		// tblfmt ends each line with "\r\n" on Windows
		if got := strings.ReplaceAll(stdout.String(), "\r\n", "\n"); got != test.want+"\n" {
			t.Errorf("border %s, xheader_width %s, columns %s:\n%s\nwant:\n%s", test.border, test.width, test.columns, got, test.want)
		}
	}
}

// TestXheaderWidthValue reads the values of xheader_width as psql reads them.
func TestXheaderWidthValue(t *testing.T) {
	for value, want := range map[string]string{
		"full":   "full",
		"COLUMN": "column",
		"Page":   "page",
		"20":     "20",
		"20abc":  "20",
		" +7":    "7",
		"-3":     "-3",
		"0":      "",
		"":       "",
		"wide":   "",
	} {
		got, err := env.XheaderWidth(value)
		if got != want || (err != nil) != (want == "") {
			t.Errorf("XheaderWidth(%q) = %q, %v, want %q", value, got, err, want)
		}
	}
}
