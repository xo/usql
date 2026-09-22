package echarts

import (
	"context"
	"image/color"
	"strings"
	"testing"

	"github.com/xo/usql/metacmd/charts"
)

// TestRegistered checks that importing this package registers the renderer.
// Nothing else does: usql reaches the renderer only through the generated file
// internal/charts.go, so a registration that stopped happening would show up
// as \chart reporting that the build has no renderer.
func TestRegistered(t *testing.T) {
	if !charts.Available() {
		t.Fatal("importing the package did not register a renderer")
	}
}

// TestRoundTrip draws a chart and rasterizes it, because the two steps are
// what the handler calls and neither is covered anywhere else.
func TestRoundTrip(t *testing.T) {
	const opts = `{"xAxis":{"type":"category","data":["a","b"]},` +
		`"yAxis":{"type":"value"},` +
		`"series":[{"type":"bar","data":[1,2]}]}`
	svg, err := charts.RenderSVG(context.Background(), opts, 320, 240)
	if err != nil {
		t.Fatalf("rendering: %v", err)
	}
	if !strings.Contains(string(svg), "<svg") {
		t.Fatalf("the result is not an SVG document: %.80q", svg)
	}
	img, err := charts.Rasterize(svg, color.White)
	if err != nil {
		t.Fatalf("rasterizing: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 320 || b.Dy() != 240 {
		t.Errorf("the image is %dx%d, want 320x240", b.Dx(), b.Dy())
	}
}
