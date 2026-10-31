package feed

import (
	"strings"
	"testing"

	"github.com/xo/dbmeta"
)

// TestBacktickIdentifierWithQuoteKeepsStringEscapes is the regression for
// xo/usql#587. Before the transit grammar honored Syntax.Backticks, a MySQL
// identifier such as `we'ird` opened a phantom string and the following
// literal's \\ escapes were collapsed before the server saw them.
func TestBacktickIdentifierWithQuoteKeepsStringEscapes(t *testing.T) {
	t.Parallel()
	const q = "select 1 as `we'ird`, hex('a\\\\b');"
	mysql := dbmeta.Syntax{BlockComments: true, HashComments: true, Backticks: true}
	p, err := NewParser(Language(mysql, false))
	if err != nil {
		t.Fatal(err)
	}
	items, err := p.Parse(q)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items=%d summary=%s", len(items), summary(items))
	}
	got := items[0].Expand(func(string) (string, bool) { return "", false })
	if !strings.Contains(got, `a\\b`) {
		t.Fatalf("expected a\\\\b preserved in %q", got)
	}
	var sawIdent bool
	for _, part := range items[0].Parts {
		if part.Kind == QuotedIdentifier {
			sawIdent = true
			text := items[0].Text[part.Start:part.End]
			if !strings.Contains(text, "'") {
				t.Fatalf("quoted ident missing embedded quote: %q", text)
			}
		}
	}
	if !sawIdent {
		t.Fatalf("expected QuotedIdentifier part, parts=%+v", items[0].Parts)
	}
}
