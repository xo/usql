package charts

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/kenshaw/colors"
	"github.com/xo/usql/text"
)

type ChartConfig struct {
	Title      string
	Subtitle   string
	W, H       int
	Background color.Color
	Type       string
	Prec       int

	File string
}

func ParseArgs(opts map[string]string) (ChartConfig, error) {
	cfg := ChartConfig{
		Title:      opts["title"],
		Subtitle:   opts["subtitle"],
		W:          800,
		H:          600,
		Background: color.White,
		Type:       opts["type"],
	}
	if size, ok := opts["size"]; ok {
		b, a, ok := strings.Cut(size, "x")
		if !ok {
			return ChartConfig{}, fmt.Errorf(text.ChartParseFailed, "size", "provide size as NxN")
		}
		var err error
		cfg.W, err = strconv.Atoi(b)
		if err != nil {
			return ChartConfig{}, fmt.Errorf(text.ChartParseFailed, "size", err)
		}
		cfg.H, err = strconv.Atoi(a)
		if err != nil {
			return ChartConfig{}, fmt.Errorf(text.ChartParseFailed, "size", err)
		}
	}
	if c, ok := opts["bg"]; ok {
		var err error
		cfg.Background, err = colors.Parse(c)
		if err != nil {
			return ChartConfig{}, fmt.Errorf(text.ChartParseFailed, "bg", err)
		}
	}
	if prec, ok := opts["prec"]; ok {
		p, err := strconv.Atoi(prec)
		if err != nil {
			return ChartConfig{}, fmt.Errorf(text.ChartParseFailed, "prec", err)
		}
		cfg.Prec = p
	}
	if file, ok := opts["file"]; ok {
		cfg.File = file
	}
	return cfg, nil
}

type Chart struct {
	Title    string
	Subtitle string
	Legend   []string
	XAxis    Series
	YAxis    Series
	Series   []Series
}

type Series struct {
	Name string
	Type string
	Data any
}

func MakeChart(cfg ChartConfig, cols []string, transposed [][]string) (*Chart, error) {
	numCols := make([][]float64, len(cols))
	for i, col := range transposed {
		for _, v := range col {
			f, err := parseFloat(v, cfg.Prec)
			if err != nil {
				numCols[i] = nil
				break
			}
			if numCols[i] == nil {
				// don't allocate slice unless we have at least some valid data
				numCols[i] = make([]float64, 0, len(col))
			}
			numCols[i] = append(numCols[i], f)
		}
	}
	firstReg, firstNumeric := -1, -1
	for i, c := range numCols {
		if firstReg == -1 && c == nil {
			firstReg = i
		}
		if firstNumeric == -1 && c != nil {
			firstNumeric = i
		}
	}
	c := &Chart{
		Title:    cfg.Title,
		Subtitle: cfg.Subtitle,
	}
	var x int
	var chartType string
	switch {
	case firstNumeric == -1:
		return nil, text.ErrNoNumericColumns
	case firstReg >= 0:
		x = firstReg
		chartType = "bar"
	default:
		x = firstNumeric
		chartType = "line"
	}
	if cfg.Type != "" {
		chartType = cfg.Type
	}
	c.XAxis = Series{
		Name: cols[x],
		Type: "category",
		Data: transposed[x],
	}
	c.YAxis = Series{
		Type: "value",
	}
	for i, col := range cols {
		if i == x {
			continue
		}
		c.Legend = append(c.Legend, col)
		c.Series = append(c.Series, Series{
			Name: col,
			Type: chartType,
			Data: numCols[i],
		})
	}
	return c, nil
}

// ToVegaLite returns the chart as a Vega-Lite spec of width by height pixels,
// with the axes, the title and the legend inside that size. The rows are in
// the spec, one for each value of each series, in the order of the result. A
// bar chart of more than one series puts the bars of each category side by
// side.
func (c Chart) ToVegaLite(width, height int) ([]byte, error) {
	categories, _ := c.XAxis.Data.([]string)
	var values []vegaLiteRow
	var series []string
	for _, s := range c.Series {
		data, ok := s.Data.([]float64)
		if !ok {
			continue
		}
		series = append(series, s.Name)
		for i, v := range data {
			if i < len(categories) {
				values = append(values, vegaLiteRow{X: categories[i], Series: s.Name, Value: v})
			}
		}
	}
	mark := "line"
	if len(c.Series) > 0 {
		mark = c.Series[0].Type
	}
	spec := vegaLite{
		Schema:   "https://vega.github.io/schema/vega-lite/v5.json",
		Width:    width,
		Height:   height,
		Padding:  padding,
		Autosize: vegaLiteAutosize{Type: "fit", Contains: "padding"},
		Data:     vegaLiteData{Values: values},
		Mark:     mark,
		Encoding: vegaLiteEncoding{
			// a sort of null keeps the order of the rows
			X:     &vegaLiteChannel{Field: "x", Type: "nominal", Title: c.XAxis.Name, Sort: json.RawMessage("null")},
			Y:     &vegaLiteChannel{Field: "value", Type: "quantitative", Title: c.YAxis.Name},
			Color: &vegaLiteChannel{Field: "series", Type: "nominal", Sort: series},
		},
	}
	if mark == "bar" && len(series) > 1 {
		spec.Encoding.XOffset = &vegaLiteChannel{Field: "series", Sort: series}
	}
	if c.Title != "" || c.Subtitle != "" {
		spec.Title = &vegaLiteTitle{Text: c.Title, Subtitle: c.Subtitle}
	}
	return json.Marshal(spec)
}

// padding is the space around the chart, in pixels, inside its width and
// height.
const padding = 5

// vegaLite is the part of a Vega-Lite spec that ToVegaLite writes.
type vegaLite struct {
	Schema   string           `json:"$schema"`
	Width    int              `json:"width"`
	Height   int              `json:"height"`
	Padding  int              `json:"padding"`
	Autosize vegaLiteAutosize `json:"autosize"`
	Title    *vegaLiteTitle   `json:"title,omitempty"`
	Data     vegaLiteData     `json:"data"`
	Mark     string           `json:"mark"`
	Encoding vegaLiteEncoding `json:"encoding"`
}

// vegaLiteAutosize says how the chart fits its width and height.
type vegaLiteAutosize struct {
	Type     string `json:"type"`
	Contains string `json:"contains"`
}

// vegaLiteTitle is the title of a Vega-Lite spec.
type vegaLiteTitle struct {
	Text     string `json:"text"`
	Subtitle string `json:"subtitle,omitempty"`
}

// vegaLiteData holds the rows of a Vega-Lite spec.
type vegaLiteData struct {
	Values []vegaLiteRow `json:"values"`
}

// vegaLiteRow is one value of one series.
type vegaLiteRow struct {
	X      string  `json:"x"`
	Series string  `json:"series"`
	Value  float64 `json:"value"`
}

// vegaLiteEncoding maps the fields of a row to the channels of the mark.
type vegaLiteEncoding struct {
	X       *vegaLiteChannel `json:"x,omitempty"`
	Y       *vegaLiteChannel `json:"y,omitempty"`
	Color   *vegaLiteChannel `json:"color,omitempty"`
	XOffset *vegaLiteChannel `json:"xOffset,omitempty"`
}

// vegaLiteChannel is one channel of an encoding. Sort is the order of the
// values, or null for the order of the rows.
type vegaLiteChannel struct {
	Field string `json:"field"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
	Sort  any    `json:"sort,omitempty"`
}

func parseFloat(v string, prec int) (f float64, err error) {
	f, err = strconv.ParseFloat(v, 64)
	if err != nil || prec == 0 {
		return
	}
	r := math.Pow(10, float64(prec))
	return math.Round(f*r) / r, nil
}
