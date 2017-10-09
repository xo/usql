package handler

import (
	"strings"
)

func explode(r []rune, s, n string) []string {
	z := string(r)
	if z != "" {
		z += n
	}
	return strings.Split(z+s, n)
}
