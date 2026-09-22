// Package echarts defines and registers usql's ECharts chart renderer.
//
// The renderer draws a chart by running Apache ECharts in a JavaScript engine
// to produce an SVG document, then rasterizing that document. Both of those
// are large: together they are roughly 12 MB of every binary they are linked
// into, which is why this package is behind the charts build tag and is
// reached only through the generated file internal/charts.go.
//
// See: https://github.com/xo/echartsgoja
// Group: all
package echarts

import (
	"context"
	"fmt"
	"image"
	"image/color"

	"github.com/xo/echartsgoja"
	"github.com/xo/resvg"
	"github.com/xo/usql/metacmd/charts"
)

func init() {
	charts.Register(renderer{})
}

// renderer draws charts with ECharts and resvg.
type renderer struct{}

// RenderSVG satisfies the charts.Renderer interface.
func (renderer) RenderSVG(ctx context.Context, opts string, width, height int) ([]byte, error) {
	svg, err := echartsgoja.New(echartsgoja.WithWidthHeight(width, height)).RenderOptions(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("rendering the chart: %w", err)
	}
	return []byte(svg), nil
}

// Rasterize satisfies the charts.Renderer interface.
func (renderer) Rasterize(svg []byte, bg color.Color) (image.Image, error) {
	img, err := resvg.Render(svg, resvg.WithBackground(bg))
	if err != nil {
		return nil, fmt.Errorf("rasterizing the chart: %w", err)
	}
	return img, nil
}
