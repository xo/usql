package stmt

import (
	"unicode"
)

// max returns the maximum of a, b.
func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// min returns the maximum of a, b.
func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// IsSpace is a special test for either a space or a control (ie, \b)
// characters.
func IsSpace(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r)
}
