// Package styles provides chroma styles based on the chroma styles but removing
// the backgrounds.
package styles

import (
	"sync"

	"github.com/alecthomas/chroma"
	cstyles "github.com/alecthomas/chroma/styles"
)

// styles is the set of styles with their background colors removed.
var styles = struct {
	styles map[string]*chroma.Style
	sync.Mutex
}{
	styles: make(map[string]*chroma.Style),
}

// Get retrieves the equivalent chroma style.
func Get(name string) *chroma.Style {
	styles.Lock()
	defer styles.Unlock()

	if _, ok := styles.styles[name]; !ok {
		// get original style
		s := cstyles.Get(name)

		// create new entry map
		m := make(chroma.StyleEntries)
		for _, typ := range s.Types() {
			// skip background
			if typ == chroma.Background {
				continue
			}
			z := s.Get(typ)

			// unset background
			z.Background = chroma.Colour(0)
			m[typ] = z.String()
		}

		styles.styles[name] = chroma.MustNewStyle(s.Name, m)
	}

	return styles.styles[name]
}
