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

/* echarts */

type echarts struct {
	Title  *echartsTitle  `json:"title,omitempty"`
	Legend *echartsLegend `json:"legend,omitempty"`
	XAxis  *echartsAxis   `json:"xAxis,omitempty"`
	YAxis  *echartsAxis   `json:"yAxis,omitempty"`
	Series []echartsAxis  `json:"series,omitempty"`
}

type echartsTitle struct {
	Title   string `json:"text,omitempty"`
	Subtext string `json:"subtext,omitempty"`
}

type echartsLegend struct {
	Data []string `json:"data,omitempty"`
}

type echartsAxis struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Data any    `json:"data,omitempty"`
}

func (c Chart) ToEcharts() (string, error) {
	ec := echarts{}
	if c.Title != "" || c.Subtitle != "" {
		ec.Title = &echartsTitle{c.Title, c.Subtitle}
	}
	if len(c.Legend) > 0 {
		ec.Legend = &echartsLegend{c.Legend}
	}
	if c.XAxis.Data != nil || c.YAxis.Type != "" {
		ec.XAxis = &echartsAxis{
			Name: c.XAxis.Name,
			Type: c.XAxis.Type,
			Data: c.XAxis.Data,
		}
	}
	if c.YAxis.Data != nil || c.YAxis.Type != "" {
		ec.YAxis = &echartsAxis{
			Name: c.YAxis.Name,
			Type: c.YAxis.Type,
			Data: c.YAxis.Data,
		}
	}
	if len(c.Series) > 0 {
		ec.Series = make([]echartsAxis, 0, len(c.Series))
		for _, s := range c.Series {
			ec.Series = append(ec.Series, echartsAxis{
				Name: s.Name,
				Type: s.Type,
				Data: s.Data,
			})
		}
	}
	buf, err := json.Marshal(ec)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func parseFloat(v string, prec int) (f float64, err error) {
	f, err = strconv.ParseFloat(v, 64)
	if err != nil || prec == 0 {
		return
	}
	r := math.Pow(10, float64(prec))
	return math.Round(f*r) / r, nil
}

const basicBarTemplate = `
{
  "title": {
    "text": {{ printf "%q" .Title }},
    "subtext": {{ printf "%q" .Subtitle }}
  },
  {{- if .Legend }}
  "legend": {
    "data": [
      {{ range .Legend }}{{ printf "%q" . }}{{ end }}
    ]
  },
  {{- end }}
  "xAxis": [
    {
      "type": "category",
      "data": [
        "Jan",
        "Feb",
        "Mar",
        "Apr",
        "May",
        "Jun",
        "Jul",
        "Aug",
        "Sep",
        "Oct",
        "Nov",
        "Dec"
      ]
    }
  ],
  "yAxis": [
    {
      "type": "value"
    }
  ],
  "series": [
    {
      "name": "Rainfall",
      "type": "bar",
      "data": [
        2,
        4.9,
        7,
        23.2,
        25.6,
        76.7,
        135.6,
        162.2,
        32.6,
        20,
        6.4,
        3.3
      ],
    },
    {
      "name": "Evaporation",
      "type": "bar",
      "data": [
        2.6,
        5.9,
        9,
        26.4,
        28.7,
        70.7,
        175.6,
        182.2,
        48.7,
        18.8,
        6,
        2.3
      ],
    }
  ]
}
`
