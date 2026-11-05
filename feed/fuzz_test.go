package feed

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/xo/dbmeta"
	_ "github.com/xo/dbmeta/all"
	"github.com/xo/transit"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

// The fuzz tests feed scripts through a Source, as usql reads a file given
// with -f, in the syntax of every dialect that dbmeta has a model for. go test
// runs only the seeds below. To search for more inputs, run one test at a
// time:
//
//	go test -run '^$' -fuzz FuzzQuotedNames ./feed

// exotic are names and strings that have tripped a SQL splitter, or could:
// quotes and separators of every dialect, a backslash, and Unicode that is not
// a letter, such as joiners, marks that change the direction of the text,
// line separators, a byte order mark, NUL, and look-alikes of ' and ;.
var exotic = []string{
	"n° d'entrée",
	"l'état",
	`a"b`,
	"a`b",
	"a]b",
	"a]]b",
	"a[b",
	"a;b",
	"a--b",
	"a/*b",
	"a*/b",
	"a#b",
	"a//b",
	`a\b`,
	`a\`,
	`a\;b`,
	"a:b",
	":name",
	":'name'",
	"$$a$$",
	"$t$a",
	"a\nb",
	"a\r\nb",
	"a\rb",
	"a\tb",
	"日本語",
	"中文名字",
	"한국어",
	"Ωμέγα",
	"Привет",
	"עברית",
	"عربي",
	"😀",
	"👩‍👩‍👧",
	"🇫🇷",
	"e\u0301",
	"\u200b",
	"\u200d",
	"\u202e",
	"\u2028",
	"\u2029",
	"\u0085",
	"\ufeff",
	"\u00a0",
	"\u3000",
	"\x00",
	"；",
	"＇",
	"＂",
	"’",
	"“",
	"［",
	"］",
	"\xff\xfe",
	"",
}

// dialect is the syntax of one dialect.
type dialect struct {
	name    dbmeta.Dialect
	syntax  dbmeta.Syntax
	batches bool
}

// language returns the usql grammar for d.
func (d dialect) language() *transit.Language {
	return Language(d.syntax, d.batches)
}

// dialects returns every dialect that dbmeta has a syntax for, by name.
func dialects(tb testing.TB) []dialect {
	tb.Helper()
	var v []dialect
	for _, name := range dbmeta.Dialects() {
		s, ok := name.Syntax()
		if !ok {
			continue
		}
		info, _ := name.Info()
		v = append(v, dialect{name: name, syntax: s, batches: len(info.Batches) != 0})
	}
	if len(v) == 0 {
		tb.Fatal("expected dbmeta to have dialects")
	}
	slices.SortFunc(v, func(a, b dialect) int { return strings.Compare(string(a.name), string(b.name)) })
	return v
}

// feedScript returns every item that a Source feeds from src, read one line at
// a time, as from a file.
func feedScript(tb testing.TB, d dialect, src string) []Item {
	tb.Helper()
	s, err := NewSource(NewLines(strings.NewReader(src)), d.language(), nil)
	if err != nil {
		tb.Fatal(err)
	}
	var items []Item
	for {
		it, err := s.Next()
		if errors.Is(err, io.EOF) {
			return items
		}
		if err != nil {
			tb.Fatalf("%s: %v", d.name, err)
		}
		items = append(items, it)
	}
}

// asLines returns src as a Source holds it after reading it with Lines: each
// line ends with \n, and a \r before a \n is gone.
func asLines(src string) string {
	if src == "" {
		return ""
	}
	lines := strings.Split(strings.TrimSuffix(src, "\n"), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSuffix(line, "\r")
	}
	return strings.Join(lines, "\n") + "\n"
}

// FuzzSplit checks that any input, valid UTF-8 or not, feeds the same items
// line by line as it parses in one piece, in every dialect.
func FuzzSplit(f *testing.F) {
	for _, s := range exotic {
		f.Add("select '" + s + "' as x;\nselect 2;\n")
		f.Add("select 1 as \"" + s + "\";\nselect 2;\n")
		f.Add("select 1 as [" + s + "];\nselect 2;\n")
		f.Add("select 1 as `" + s + "`;\nselect 2;\n")
		f.Add("-- " + s + "\nselect 1; /* " + s + " */ select 2;\n")
		f.Add("\\echo " + s + "\nselect 1 \\g\n")
	}
	// the dump of GitHub discussion xo/usql#600, with the line ends of Windows
	f.Add("-- dbo.export_order definition\r\n\r\n-- Drop table\r\n\r\n-- DROP TABLE dbo.export_order;\r\n\r\n" +
		"CREATE TABLE export_order (\r\n\t\"n° d'entrée\" nvarchar(255) COLLATE French_CI_AS NULL,\r\n" +
		"\t[l'état] nvarchar(50) NULL\r\n);\r\n\r\nINSERT INTO export_order VALUES (N'a''b', N'é');\r\n")
	f.Add("\ufeffselect 1;\r\nselect 2")
	f.Add("create procedure p() begin\n  select ';';\nend;\nselect 1;\n")
	f.Add("begin batch insert into t (a) values ('x;y'); apply batch;\n")
	f.Add("select $$a;b$$, $t$c;$t$;\nselect 'open")
	dias := dialects(f)
	f.Fuzz(func(t *testing.T, src string) {
		for _, d := range dias {
			p, err := NewParser(d.language())
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := p.Parse(asLines(src))
			if err != nil {
				t.Fatalf("%s: %v", d.name, err)
			}
			if got, exp := summary(feedScript(t, d, src)), summary(parsed); got != exp {
				t.Errorf("%s: %q\n got: %q\nwant: %q", d.name, src, got, exp)
			}
		}
	})
}

// quoteForms returns each way that d writes name as a quoted identifier.
func quoteForms(d dialect, name string) []string {
	v := []string{`"` + strings.ReplaceAll(name, `"`, `""`) + `"`}
	if d.syntax.Backticks {
		v = append(v, "`"+strings.ReplaceAll(name, "`", "``")+"`")
	}
	if d.syntax.Brackets {
		v = append(v, "["+strings.ReplaceAll(name, "]", "]]")+"]")
	}
	return v
}

// quoteString returns value as a string between single quotes. A backslash
// escapes the next character in a string of every family (transit's D105),
// so it is doubled as well as the quote.
func quoteString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

// checkScript checks that a script that writes name and value in the quotes of
// d, and in its comments, feeds three statements, and that the quoted name and
// the string are each one part of their statement.
func checkScript(t *testing.T, d dialect, name, value string, crlf, bom bool) {
	t.Helper()
	str := quoteString(value)
	for _, q := range quoteForms(d, name) {
		stmts := []string{
			"create table " + q + " (" + q + " int);",
			"insert into " + q + " values (" + str + ");",
			"select " + str + " as " + q + ";",
		}
		// how many names and strings each statement holds, which counting q
		// in the text gets wrong when the string holds q
		parts := [][2]int{{2, 0}, {1, 1}, {1, 1}}
		var lines []string
		if d.syntax.DashComments {
			lines = append(lines, "-- "+strings.NewReplacer("\r", "", "\n", "").Replace(value))
		}
		lines = append(lines, stmts[0], stmts[1])
		if d.syntax.BlockComments && !strings.Contains(value, "/*") && !strings.Contains(value, "*/") {
			lines = append(lines, "/* "+value+" */")
		}
		lines = append(lines, stmts[2])
		eol := "\n"
		if crlf {
			eol = "\r\n"
		}
		src := strings.Join(lines, eol) + eol
		if bom {
			src = "\ufeff" + src
		}
		items := feedScript(t, d, src)
		if len(items) != len(stmts) {
			t.Errorf("%s %s: expected %d statements, got %d: %q\n%s", d.name, q, len(stmts), len(items), src, summary(items))
			continue
		}
		for i, it := range items {
			exp := strings.ReplaceAll(stmts[i], "\r\n", "\n")
			if it.Kind != Statement || !it.Terminated || !strings.HasSuffix(strings.TrimSpace(it.Text), exp) {
				t.Errorf("%s %s: statement %d:\n got: %q\nwant: %q", d.name, q, i, it.Text, exp)
				continue
			}
			names, strs := countParts(it, strings.ReplaceAll(q, "\r\n", "\n"), strings.ReplaceAll(str, "\r\n", "\n"))
			if expNames, expStrs := parts[i][0], parts[i][1]; names != expNames || strs != expStrs {
				t.Errorf("%s %s: statement %d: expected %d names and %d strings, got %d and %d: %+v",
					d.name, q, i, expNames, expStrs, names, strs, it.Parts)
			}
		}
	}
}

// countParts returns how many parts of it are the quoted identifier q and the
// string str.
func countParts(it Item, q, str string) (int, int) {
	var names, strs int
	for _, p := range it.Parts {
		switch text := it.Text[p.Start:p.End]; {
		case p.Kind == QuotedIdentifier && text == q:
			names++
		case p.Kind == String && text == str:
			strs++
		}
	}
	return names, strs
}

// FuzzQuotedNames checks that a quote, a separator or exotic Unicode inside a
// quoted name, a string or a comment does not change where a statement ends,
// in every dialect, with either line end, and after a byte order mark.
func FuzzQuotedNames(f *testing.F) {
	for i, s := range exotic {
		f.Add(s, exotic[(i+1)%len(exotic)], i%2 == 0, i%3 == 0)
	}
	dias := dialects(f)
	f.Fuzz(func(t *testing.T, name, value string, crlf, bom bool) {
		for _, d := range dias {
			checkScript(t, d, name, value, crlf, bom)
		}
	})
}

// legacy are encodings that a script can be saved in. In each one a byte under
// 0x80 is always the ASCII character, so usql, which reads UTF-8, finds the
// same quotes and separators as in the UTF-8 text.
var legacy = []struct {
	name string
	enc  encoding.Encoding
}{
	{"Windows-1252", charmap.Windows1252},
	{"ISO-8859-1", charmap.ISO8859_1},
	{"ISO-8859-15", charmap.ISO8859_15},
	{"Windows-1251", charmap.Windows1251},
	{"KOI8-R", charmap.KOI8R},
	{"Windows-1253", charmap.Windows1253},
	{"EUC-JP", japanese.EUCJP},
	{"EUC-KR", korean.EUCKR},
}

// multibyte are encodings where the second byte of a character can be an
// ASCII byte, such as 0x5C, the backslash, in the Shift JIS 表 (0x95 0x5C). A
// backslash escapes the next character in a string (transit's D105), so these
// are checked only in a name between double quotes, where a backslash is
// plain text.
var multibyte = []struct {
	name string
	enc  encoding.Encoding
}{
	{"Shift JIS", japanese.ShiftJIS},
	{"Big5", traditionalchinese.Big5},
	{"GBK", simplifiedchinese.GBK},
	{"GB18030", simplifiedchinese.GB18030},
}

// FuzzEncodings checks scripts saved in other encodings than UTF-8. text is
// encoded first, and a character that the encoding lacks becomes a
// replacement.
func FuzzEncodings(f *testing.F) {
	f.Add("n° d'entrée")
	f.Add("Привет, мир")
	f.Add("Ωμέγα")
	f.Add("表ソ能予申十貼構")
	f.Add("許功蓋中文")
	f.Add("中文名字誠")
	f.Add("한국어")
	f.Add("日本語表")
	dias := dialects(f)
	f.Fuzz(func(t *testing.T, text string) {
		for _, e := range legacy {
			b, err := encoding.ReplaceUnsupported(e.enc.NewEncoder()).String(text)
			if err != nil {
				t.Skip()
			}
			for _, d := range dias {
				checkScript(t, d, b, b, false, false)
			}
		}
		for _, e := range multibyte {
			b, err := encoding.ReplaceUnsupported(e.enc.NewEncoder()).String(text)
			if err != nil {
				t.Skip()
			}
			for _, d := range dias {
				q := `"` + strings.ReplaceAll(b, `"`, `""`) + `"`
				src := "select 1 as " + q + ";\nselect 2;\n"
				items := feedScript(t, d, asLines(src))
				if got, exp := summary(items), summary([]Item{
					{Kind: Statement, Text: strings.TrimSuffix(asLines("select 1 as "+q+";"), "\n"), Terminated: true},
					{Kind: Statement, Text: "select 2;", Terminated: true},
				}); got != exp {
					t.Errorf("%s %s:\n got: %q\nwant: %q", e.name, d.name, got, exp)
				}
			}
		}
	})
}
