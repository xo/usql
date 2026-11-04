package highlight

import (
	"context"
	"slices"
	"strings"

	"github.com/xo/transit"
)

// Keywords returns the keywords that the SQL grammar of dialect accepts at
// the end of src, the text of one statement up to the word that is being
// typed, such as TABLE, INDEX and VIEW after CREATE on SQLite. The completer
// offers them, so that each dialect gets its own words from its grammar
// (transit's docs/API.md).
//
// The parse tables of a grammar are minimized, so a list can hold a few words
// more than the dialect allows at that place.
func Keywords(ctx context.Context, dialect string, src []byte) ([]string, error) {
	g := grammarFor(dialect)
	lang := g.language()
	p := transit.NewParser()
	if err := p.SetLanguage(lang); err != nil {
		return nil, err
	}
	states, err := p.StatesAt(ctx, src, len(src), nil)
	if err != nil {
		return nil, err
	}
	known := make(map[string]bool)
	for _, k := range g.keywords() {
		known[k] = true
	}
	var words []string
	for _, state := range states {
		it, ok := lang.LookaheadIterator(state)
		if !ok {
			continue
		}
		for sym := range it.Symbols() {
			name := lang.SymbolName(sym)
			if !known[name] && (lang.SymbolType(sym) != transit.SymbolAnonymous || !isWord(name)) {
				continue
			}
			if w := keywordText(name); !slices.Contains(words, w) {
				words = append(words, w)
			}
		}
	}
	slices.Sort(words)
	return words, nil
}

// keywordText returns the text of the keyword of the symbol name. The grammars
// of the dialects name a keyword by its text, such as SELECT, or as
// kw_select or keyword_select.
func keywordText(name string) string {
	for _, prefix := range []string{"keyword_", "kw_"} {
		if s, ok := strings.CutPrefix(name, prefix); ok {
			return strings.ToUpper(s)
		}
	}
	return strings.ToUpper(name)
}

// isWord reports whether s is a word of letters and underscores, as a
// keyword is, and not punctuation such as ( or ;.
func isWord(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r == '_' || 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z') {
			return false
		}
	}
	return true
}
