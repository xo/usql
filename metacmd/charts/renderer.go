package charts

import (
	"context"
	"image"
	"image/color"

	"github.com/xo/usql/text"
)

// Renderer turns a chart into something that can be displayed.
//
// A renderer is not part of every usql build. The implementation draws the
// chart with a JavaScript engine and rasterizes the result with an SVG
// renderer, and those two together are a large part of the size of the binary,
// so they are behind the charts build tag. Register is called from the init of
// the implementation package, which the generated file internal/charts.go
// imports when that tag is set.
type Renderer interface {
	// RenderSVG draws the ECharts option document opts as an SVG document.
	RenderSVG(ctx context.Context, opts string, width, height int) ([]byte, error)

	// Rasterize draws the SVG document svg as an image, on background bg.
	Rasterize(svg []byte, bg color.Color) (image.Image, error)
}

// renderer is the registered renderer. It is nil when usql was built without
// the charts build tag.
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

// RenderSVG draws the ECharts option document opts as an SVG document. It
// returns text.ErrChartsNotSupported when usql was built without a renderer.
func RenderSVG(ctx context.Context, opts string, width, height int) ([]byte, error) {
	if renderer == nil {
		return nil, text.ErrChartsNotSupported
	}
	return renderer.RenderSVG(ctx, opts, width, height)
}

// Rasterize draws the SVG document svg as an image, on background bg. It
// returns text.ErrChartsNotSupported when usql was built without a renderer.
func Rasterize(svg []byte, bg color.Color) (image.Image, error) {
	if renderer == nil {
		return nil, text.ErrChartsNotSupported
	}
	return renderer.Rasterize(svg, bg)
}
