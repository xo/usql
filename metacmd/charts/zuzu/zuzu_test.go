package zuzu

import (
	"context"
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

// TestDraw draws a chart of two series as SVG and as an image, because those
// are the two steps that the handler calls.
func TestDraw(t *testing.T) {
	cfg, err := charts.ParseArgs(map[string]string{"title": "Rain", "size": "320x240"})
	if err != nil {
		t.Fatal(err)
	}
	cols := []string{"month", "rain", "sun"}
	rows := [][]string{{"Jan", "Feb", "Mar"}, {"2", "4.9", "7"}, {"9", "8", "6.5"}}
	c, err := charts.MakeChart(cfg, cols, rows)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := c.ToVegaLite(cfg.W, cfg.H)
	if err != nil {
		t.Fatal(err)
	}
	svg, err := charts.SVG(context.Background(), spec)
	if err != nil {
		t.Fatalf("drawing the SVG: %v", err)
	}
	for _, s := range []string{"<svg", "Rain", "Feb", "rain", "sun"} {
		if !strings.Contains(string(svg), s) {
			t.Errorf("the SVG does not hold %q", s)
		}
	}
	img, err := charts.Image(context.Background(), spec)
	if err != nil {
		t.Fatalf("drawing the image: %v", err)
	}
	if b := img.Bounds(); b.Dx() != cfg.W || b.Dy() != cfg.H {
		t.Errorf("the image is %dx%d, want %dx%d", b.Dx(), b.Dy(), cfg.W, cfg.H)
	}
}
