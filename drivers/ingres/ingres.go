// Package ingres defines and registers usql's Ingres (Actian X, Vector, VectorH) driver.
// Requires CGO. Uses platform's Ingres libraries.
//
// See: https://github.com/ildus/ingres
// Group: all
package ingres

import (
	_ "github.com/ildus/ingres" // DRIVER
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("ingres", drivers.Driver{
		NewMetadataReader: NewMetadataReader,
	})
}
