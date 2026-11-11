package charts

import (
	"context"
	"image"

	"github.com/xo/usql/text"
)

// Renderer draws a chart.
//
// A renderer is not part of every usql build. The build tag no_charts leaves
// it out. Register is called from the init of the implementation package,
// which the generated file internal/charts.go imports unless the build leaves
// it out.
type Renderer interface {
	// SVG draws the Vega-Lite spec as an SVG document.
	SVG(ctx context.Context, spec []byte) ([]byte, error)

	// Image draws the Vega-Lite spec as an image.
	Image(ctx context.Context, spec []byte) (image.Image, error)
}

// renderer is the registered renderer. It is nil when the build left the
// charts out.
var renderer Renderer

// Register registers the chart renderer. It panics when a renderer is already
// registered, because two implementations in one binary means the build gated
// them wrongly.
func Register(r Renderer) {
	if r == nil {
		panic("charts: Register called with a nil renderer")
	}
	if renderer != nil {
		panic("charts: a renderer is already registered")
	}
	renderer = r
}

// Available reports whether usql was built with a chart renderer.
func Available() bool {
	return renderer != nil
}

// SVG draws the Vega-Lite spec as an SVG document. It returns
// text.ErrChartsNotSupported when usql was built without a renderer.
func SVG(ctx context.Context, spec []byte) ([]byte, error) {
	if renderer == nil {
		return nil, text.ErrChartsNotSupported
	}
	return renderer.SVG(ctx, spec)
}

// Image draws the Vega-Lite spec as an image. It returns
// text.ErrChartsNotSupported when usql was built without a renderer.
func Image(ctx context.Context, spec []byte) (image.Image, error) {
	if renderer == nil {
		return nil, text.ErrChartsNotSupported
	}
	return renderer.Image(ctx, spec)
}
