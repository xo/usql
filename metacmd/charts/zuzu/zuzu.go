// Package zuzu defines and registers usql's chart renderer.
//
// The renderer draws a Vega-Lite spec with zuzu, which is pure Go. It is
// reached only through the generated file internal/charts.go, and the build
// tag no_charts leaves it out.
//
// See: https://github.com/xo/zuzu
package zuzu

import (
	"bytes"
	"context"
	"fmt"
	"image"

	"github.com/xo/usql/metacmd/charts"
	"github.com/xo/zuzu"
)

func init() {
	charts.Register(renderer{})
}

// renderer draws charts with zuzu.
type renderer struct{}

// SVG satisfies the charts.Renderer interface.
func (renderer) SVG(ctx context.Context, spec []byte) ([]byte, error) {
	c, err := parse(spec)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if _, err := c.SVG(ctx, &b); err != nil {
		return nil, fmt.Errorf("drawing the chart: %w", err)
	}
	return b.Bytes(), nil
}

// Image satisfies the charts.Renderer interface.
func (renderer) Image(ctx context.Context, spec []byte) (image.Image, error) {
	c, err := parse(spec)
	if err != nil {
		return nil, err
	}
	img, _, err := c.Image(ctx)
	if err != nil {
		return nil, fmt.Errorf("drawing the chart: %w", err)
	}
	return img, nil
}

// parse reads the Vega-Lite spec.
func parse(spec []byte) (*zuzu.Chart, error) {
	c, _, err := zuzu.Parse(spec, zuzu.WithDialect(zuzu.VegaLite))
	if err != nil {
		return nil, fmt.Errorf("reading the chart: %w", err)
	}
	return c, nil
}
