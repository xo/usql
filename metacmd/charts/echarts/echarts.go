// Package echarts defines and registers usql's ECharts chart renderer.
//
// The renderer draws a chart by running Apache ECharts in a JavaScript engine
// to produce an SVG document, then rasterizing that document. Both of those
// are large: together they are roughly 12 MiB of every binary they are linked
// into. So this package is reached only through the generated file
// internal/charts.go, and the build tag no_charts leaves it out.
//
// See: https://github.com/xo/echartsgoja
package echarts

import (
	"context"
	"fmt"
	"image"
	"image/color"

	"github.com/xo/echartsgoja"
	"github.com/xo/okage/svg"
	"github.com/xo/usql/metacmd/charts"
)

func init() {
	charts.Register(renderer{})
}

// renderer draws charts with ECharts and rasterizes them with okage's SVG
// renderer, which is pure Go.
type renderer struct{}

// RenderSVG satisfies the charts.Renderer interface.
func (renderer) RenderSVG(ctx context.Context, opts string, width, height int) ([]byte, error) {
	s, err := echartsgoja.New(echartsgoja.WithWidthHeight(width, height)).RenderOptions(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("rendering the chart: %w", err)
	}
	return []byte(s), nil
}

// Rasterize satisfies the charts.Renderer interface. Each call makes a
// renderer, which loads the fonts of the system and of the home directory.
func (renderer) Rasterize(data []byte, bg color.Color) (image.Image, error) {
	img, err := svg.New(svg.WithBackground(bg)).Render(data)
	if err != nil {
		return nil, fmt.Errorf("rasterizing the chart: %w", err)
	}
	return img, nil
}
